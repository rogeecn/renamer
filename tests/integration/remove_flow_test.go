package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/remove"
)

func TestRemoveApplyAndUndo(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "report copy draft.txt"))
	createFile(t, filepath.Join(tmp, "nested", "notes draft.txt"))

	parsed, err := remove.ParseArgs([]string{" copy", " draft"})
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	req := &remove.Request{
		WorkingDir: tmp,
		Tokens:     parsed.Tokens,
		Recursive:  true,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("request validation error: %v", err)
	}

	summary, planned, err := remove.Preview(context.Background(), req, parsed, nil)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	if summary.ChangedCount != 2 {
		t.Fatalf("expected 2 changes, got %d", summary.ChangedCount)
	}

	entry, err := remove.Apply(context.Background(), req, planned, summary, parsed.Tokens)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations recorded, got %d", len(entry.Operations))
	}
	if entry.Metadata == nil {
		t.Fatalf("expected metadata to be recorded")
	}
	tokens, ok := entry.Metadata["tokens"].([]string)
	if !ok {
		t.Fatalf("expected ordered tokens metadata, got %#v", entry.Metadata)
	}
	if len(tokens) != 2 || tokens[0] != " copy" || tokens[1] != " draft" {
		t.Fatalf("unexpected tokens metadata: %#v", tokens)
	}
	matches, ok := entry.Metadata["matches"].(map[string]int)
	if !ok {
		t.Fatalf("expected matches metadata, got %#v", entry.Metadata)
	}
	if matches[" copy"] != 1 || matches[" draft"] != 2 {
		t.Fatalf("unexpected match counts: %#v", matches)
	}

	if _, err := os.Stat(filepath.Join(tmp, "report.txt")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "nested", "notes.txt")); err != nil {
		t.Fatalf("expected nested rename exists: %v", err)
	}

	if _, err := history.Undo(tmp); err != nil {
		t.Fatalf("undo error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "report copy draft.txt")); err != nil {
		t.Fatalf("expected original restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "nested", "notes draft.txt")); err != nil {
		t.Fatalf("expected nested original restored: %v", err)
	}
}
