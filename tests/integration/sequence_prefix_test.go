package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequenceApplyWithNumberPrefix(t *testing.T) {
	tmp := t.TempDir()

	createIntegrationFile(t, filepath.Join(tmp, "cover.png"))
	createIntegrationFile(t, filepath.Join(tmp, "index.png"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp
	opts.Placement = sequence.PlacementPrefix
	opts.Separator = "-"
	opts.NumberPrefix = "seq"

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	entry, err := sequence.Apply(context.Background(), opts, plan)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "seq001-cover.png")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "seq002-index.png")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	meta, ok := entry.Metadata["sequence"].(map[string]any)
	if !ok {
		t.Fatalf("sequence metadata missing")
	}
	if prefix, ok := meta["prefix"].(string); !ok || prefix != "seq" {
		t.Fatalf("expected metadata prefix 'seq', got %#v", meta["prefix"])
	}

	if _, err := history.Undo(tmp); err != nil {
		t.Fatalf("undo error: %v", err)
	}
}
