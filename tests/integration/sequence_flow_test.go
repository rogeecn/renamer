package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequenceApplyAndUndo(t *testing.T) {
	tmp := t.TempDir()

	createIntegrationFile(t, filepath.Join(tmp, "draft.txt"))
	createIntegrationFile(t, filepath.Join(tmp, "notes.txt"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	if plan.Summary.RenamedCount != 2 {
		t.Fatalf("expected 2 planned renames, got %d", plan.Summary.RenamedCount)
	}

	entry, err := sequence.Apply(context.Background(), opts, plan)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 recorded operations, got %d", len(entry.Operations))
	}
	if entry.Command != "sequence" {
		t.Fatalf("expected ledger command 'sequence', got %s", entry.Command)
	}

	if _, err := os.Stat(filepath.Join(tmp, "001_draft.txt")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "002_notes.txt")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	if _, err := history.Undo(tmp); err != nil {
		t.Fatalf("undo error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "draft.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "notes.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
}

func createIntegrationFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
