package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/history"
)

func TestAIRenameApplyAndUndo(t *testing.T) {
	t.Setenv("RENAMER_AI_KEY", "test-key")

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "IMG_2001.jpg"))
	createFile(t, filepath.Join(tmp, "session-notes.txt"))

	root := renamercmd.NewRootCommand()
	root.SetArgs([]string{"ai", "--path", tmp, "--prompt", "Album Shots", "--yes"})
	root.SetIn(strings.NewReader(""))
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)

	if err := root.Execute(); err != nil {
		t.Fatalf("ai command returned error: %v\noutput: %s", err, output.String())
	}

	ledgerPath := filepath.Join(tmp, ".renamer")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 {
		t.Fatalf("expected ledger entries")
	}
	var entry history.Entry
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("decode entry: %v", err)
	}
	if entry.Command != "ai" {
		t.Fatalf("expected command 'ai', got %q", entry.Command)
	}
	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations recorded, got %d", len(entry.Operations))
	}
	if entry.Metadata == nil || entry.Metadata["prompt"] != "Album Shots" {
		t.Fatalf("expected prompt metadata recorded, got %#v", entry.Metadata)
	}

	if sep, ok := entry.Metadata["sequenceSeparator"].(string); !ok || sep != "." {
		t.Fatalf("expected sequence separator metadata, got %#v", entry.Metadata["sequenceSeparator"])
	}

	for _, op := range entry.Operations {
		dest := filepath.Join(tmp, filepath.FromSlash(op.To))
		if _, err := os.Stat(dest); err != nil {
			t.Fatalf("expected destination %q to exist: %v", dest, err)
		}
	}

	undoCmd := renamercmd.NewRootCommand()
	undoCmd.SetArgs([]string{"undo", "--path", tmp})
	undoCmd.SetIn(strings.NewReader(""))
	undoCmd.SetOut(&output)
	undoCmd.SetErr(&output)
	if err := undoCmd.Execute(); err != nil {
		t.Fatalf("undo command error: %v\noutput: %s", err, output.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "IMG_2001.jpg")); err != nil {
		t.Fatalf("expected original root file restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "session-notes.txt")); err != nil {
		t.Fatalf("expected original secondary file restored: %v", err)
	}
}
