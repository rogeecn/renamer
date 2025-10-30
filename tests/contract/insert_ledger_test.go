package contract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/insert"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestInsertLedgerMetadataAndUndo(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeInsertLedgerFile(t, filepath.Join(tmp, "doc.txt"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := insert.NewRequest(scope)
	req.SetExecutionMode(false, true)
	req.SetPositionAndText("$", "_ARCHIVE")

	summary, planned, err := insert.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	entry, err := insert.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	meta := entry.Metadata
	if meta == nil {
		t.Fatalf("expected metadata to be recorded")
	}

	if got := meta["insertText"]; got != "_ARCHIVE" {
		t.Fatalf("expected insertText metadata, got %v", got)
	}
	if got := meta["positionToken"]; got != "$" {
		t.Fatalf("expected position token metadata, got %v", got)
	}
	scopeMeta, ok := meta["scope"].(map[string]any)
	if !ok {
		t.Fatalf("expected scope metadata, got %T", meta["scope"])
	}
	if includeHidden, _ := scopeMeta["includeHidden"].(bool); includeHidden {
		t.Fatalf("expected includeHidden to be false")
	}

	undoEntry, err := history.Undo(tmp)
	if err != nil {
		t.Fatalf("undo error: %v", err)
	}
	if undoEntry.Command != "insert" {
		t.Fatalf("expected undo command to be insert, got %s", undoEntry.Command)
	}
	if _, err := os.Stat(filepath.Join(tmp, "doc.txt")); err != nil {
		t.Fatalf("expected original file restored: %v", err)
	}
}

func TestInsertZeroMatchExitsSuccessfully(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	var out bytes.Buffer

	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"insert", "^", "TEST", "--yes", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected zero-match insert to succeed, err=%v output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "No candidates found.") {
		t.Fatalf("expected zero-match notice, output=%s", out.String())
	}
}

func writeInsertLedgerFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("ledger"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
