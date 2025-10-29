package integration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/replace"
)

func TestReplaceApplyAndUndo(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "foo_draft.txt"))
	createFile(t, filepath.Join(tmp, "bar_draft.txt"))

	parsed, err := replace.ParseArgs([]string{"draft", "final"})
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	req := &replace.ReplaceRequest{
		WorkingDir:  tmp,
		Patterns:    parsed.Patterns,
		Replacement: parsed.Replacement,
	}

	summary, planned, err := replace.Preview(context.Background(), req, parsed, nil)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}
	if summary.ChangedCount != 2 {
		t.Fatalf("expected 2 changes, got %d", summary.ChangedCount)
	}

	entry, err := replace.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations recorded, got %d", len(entry.Operations))
	}
	if entry.Metadata == nil {
		t.Fatalf("expected metadata to be recorded")
	}
	counts, ok := entry.Metadata["patterns"].(map[string]int)
	if !ok {
		t.Fatalf("patterns metadata missing or wrong type: %#v", entry.Metadata)
	}
	if counts["draft"] != 2 {
		t.Fatalf("expected pattern count for 'draft' to be 2, got %d", counts["draft"])
	}

	if _, err := os.Stat(filepath.Join(tmp, "foo_final.txt")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "bar_final.txt")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	_, err = history.Undo(tmp)
	if err != nil {
		t.Fatalf("undo error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "foo_draft.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "bar_draft.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
}

func TestReplaceCommandInvalidArgs(t *testing.T) {
	root := renamercmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"replace", "onlyone"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error for insufficient arguments")
	}
}

func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
