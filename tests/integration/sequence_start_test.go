package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequenceApplyWithStartOffset(t *testing.T) {
	tmp := t.TempDir()

	createIntegrationFile(t, filepath.Join(tmp, "shotA.exr"))
	createIntegrationFile(t, filepath.Join(tmp, "shotB.exr"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp
	opts.Start = 10

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	expected := []string{"010_shotA.exr", "011_shotB.exr"}
	if len(plan.Candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d", len(expected), len(plan.Candidates))
	}
	for i, candidate := range plan.Candidates {
		if candidate.ProposedPath != expected[i] {
			t.Fatalf("candidate %d proposed %s, expected %s", i, candidate.ProposedPath, expected[i])
		}
	}

	entry, err := sequence.Apply(context.Background(), opts, plan)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(entry.Operations))
	}

	if _, err := os.Stat(filepath.Join(tmp, "010_shotA.exr")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "011_shotB.exr")); err != nil {
		t.Fatalf("expected renamed file exists: %v", err)
	}

	meta, ok := entry.Metadata["sequence"].(map[string]any)
	if !ok {
		t.Fatalf("sequence metadata missing")
	}
	if start, ok := meta["start"].(int); !ok || start != 10 {
		t.Fatalf("expected metadata start 10, got %#v", meta["start"])
	}

	if _, err := history.Undo(tmp); err != nil {
		t.Fatalf("undo error: %v", err)
	}
}
