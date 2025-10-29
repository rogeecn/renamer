package replace_test

import (
	"testing"

	"github.com/rogeecn/renamer/internal/remove"
)

func TestApplyTokensSequentialRemoval(t *testing.T) {
	candidate := remove.Candidate{
		BaseName:     "report copy draft.txt",
		RelativePath: "report copy draft.txt",
	}

	result := remove.ApplyTokens(candidate, []string{" copy", " draft"})

	if !result.Changed {
		t.Fatalf("expected result to be marked as changed")
	}

	if result.ProposedName != "report.txt" {
		t.Fatalf("expected proposed name to be report.txt, got %q", result.ProposedName)
	}

	if result.Matches[" copy"] != 1 {
		t.Fatalf("expected match count for ' copy' to be 1, got %d", result.Matches[" copy"])
	}

	if result.Matches[" draft"] != 1 {
		t.Fatalf("expected match count for ' draft' to be 1, got %d", result.Matches[" draft"])
	}
}

func TestApplyTokensNoChange(t *testing.T) {
	candidate := remove.Candidate{
		BaseName:     "notes.txt",
		RelativePath: "notes.txt",
	}

	result := remove.ApplyTokens(candidate, []string{" copy"})

	if result.Changed {
		t.Fatalf("expected no change for candidate without matches")
	}

	if len(result.Matches) != 0 {
		t.Fatalf("expected no matches to be recorded, got %#v", result.Matches)
	}

	if result.ProposedName != candidate.BaseName {
		t.Fatalf("expected proposed name to remain %q, got %q", candidate.BaseName, result.ProposedName)
	}
}

func TestApplyTokensEmptyName(t *testing.T) {
	candidate := remove.Candidate{
		BaseName:     "draft",
		RelativePath: "draft",
	}

	result := remove.ApplyTokens(candidate, []string{"draft"})

	if !result.Changed {
		t.Fatalf("expected change when removing the full name")
	}

	if result.ProposedName != "" {
		t.Fatalf("expected proposed name to be empty, got %q", result.ProposedName)
	}

	if result.Matches["draft"] != 1 {
		t.Fatalf("expected matches to record removal, got %#v", result.Matches)
	}
}
