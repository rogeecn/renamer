package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequenceApplyWithExplicitWidth(t *testing.T) {
	tmp := t.TempDir()

	createIntegrationFile(t, filepath.Join(tmp, "cutA.mov"))
	createIntegrationFile(t, filepath.Join(tmp, "cutB.mov"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp
	opts.Width = 4

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	if plan.Summary.AppliedWidth != 4 {
		t.Fatalf("expected applied width 4, got %d", plan.Summary.AppliedWidth)
	}

	entry, err := sequence.Apply(context.Background(), opts, plan)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(entry.Operations))
	}

	if _, err := os.Stat(filepath.Join(tmp, "0001_cutA.mov")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "0002_cutB.mov")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	meta, ok := entry.Metadata["sequence"].(map[string]any)
	if !ok {
		t.Fatalf("sequence metadata missing")
	}
	if width, ok := meta["width"].(int); !ok || width != 4 {
		t.Fatalf("expected metadata width 4, got %#v", meta["width"])
	}

	if _, err := history.Undo(tmp); err != nil {
		t.Fatalf("undo error: %v", err)
	}
}
