package contract

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/sequence"
)

func TestSequencePreviewWithPrefixPlacement(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "cover.png"))
	createFile(t, filepath.Join(tmp, "index.png"))

	opts := sequence.DefaultOptions()
	opts.WorkingDir = tmp
	opts.Start = 10
	opts.Placement = sequence.PlacementPrefix
	opts.Separator = "-"

	plan, err := sequence.Preview(context.Background(), opts, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	expected := []string{"010-cover.png", "011-index.png"}
	if len(plan.Candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d", len(expected), len(plan.Candidates))
	}
	for i, candidate := range plan.Candidates {
		if candidate.ProposedPath != expected[i] {
			t.Fatalf("candidate %d proposed %s, expected %s", i, candidate.ProposedPath, expected[i])
		}
	}

	if plan.Config.Start != 10 {
		t.Fatalf("expected config start 10, got %d", plan.Config.Start)
	}
	if plan.Config.Placement != sequence.PlacementPrefix {
		t.Fatalf("expected prefix placement in config")
	}
}

func TestSequencePreviewWithNumberPrefixLabel(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "cover.png"))
	createFile(t, filepath.Join(tmp, "index.png"))

	ops := sequence.DefaultOptions()
	ops.WorkingDir = tmp
	ops.Placement = sequence.PlacementPrefix
	ops.Separator = "-"
	ops.NumberPrefix = "seq"

	plan, err := sequence.Preview(context.Background(), ops, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}

	expected := []string{"seq001-cover.png", "seq002-index.png"}
	if len(plan.Candidates) != len(expected) {
		t.Fatalf("expected %d candidates, got %d", len(expected), len(plan.Candidates))
	}
	for i, candidate := range plan.Candidates {
		if candidate.ProposedPath != expected[i] {
			t.Fatalf("candidate %d proposed %s, expected %s", i, candidate.ProposedPath, expected[i])
		}
	}

	if warnings := plan.Summary.Warnings; len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", warnings)
	}
}
