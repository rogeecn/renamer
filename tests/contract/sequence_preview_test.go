package contract

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequencePreviewDefaultNumbering(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "draft.txt"))
	createFile(t, filepath.Join(tmp, "notes.txt"))
	createFile(t, filepath.Join(tmp, "plan.txt"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	if plan.Summary.TotalCandidates != 3 {
		t.Fatalf("expected 3 candidates, got %d", plan.Summary.TotalCandidates)
	}
	if plan.Summary.RenamedCount != 3 {
		t.Fatalf("expected 3 renamed entries, got %d", plan.Summary.RenamedCount)
	}

	expected := []string{"001_draft.txt", "002_notes.txt", "003_plan.txt"}
	if len(plan.Candidates) != 3 {
		t.Fatalf("expected 3 planned candidates, got %d", len(plan.Candidates))
	}
	for i, candidate := range plan.Candidates {
		if candidate.ProposedPath != expected[i] {
			t.Fatalf("candidate %d proposed %s, expected %s", i, candidate.ProposedPath, expected[i])
		}
	}

	if plan.Summary.AppliedWidth != 3 {
		t.Fatalf("expected applied width 3, got %d", plan.Summary.AppliedWidth)
	}

	if plan.Config.Start != 1 {
		t.Fatalf("expected start 1, got %d", plan.Config.Start)
	}
	if plan.Config.Placement != sequence.PlacementPrefix {
		t.Fatalf("expected prefix placement, got %s", plan.Config.Placement)
	}
}
