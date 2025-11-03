package contract

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequencePreviewWithExplicitWidth(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "cutA.mov"))
	createFile(t, filepath.Join(tmp, "cutB.mov"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp
	opts.Width = 4

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	expected := []string{"0001_cutA.mov", "0002_cutB.mov"}
	if len(plan.Candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d", len(expected), len(plan.Candidates))
	}
	for i, candidate := range plan.Candidates {
		if candidate.ProposedPath != expected[i] {
			t.Fatalf("candidate %d proposed %s, expected %s", i, candidate.ProposedPath, expected[i])
		}
	}

	if plan.Summary.AppliedWidth != 4 {
		t.Fatalf("expected applied width 4, got %d", plan.Summary.AppliedWidth)
	}
}
