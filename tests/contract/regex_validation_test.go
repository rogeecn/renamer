package contract

import (
	"context"
	"testing"

	"github.com/rogeecn/renamer/internal/regex"
)

func TestRegexTemplateRejectsUndefinedGroup(t *testing.T) {
	req := regex.NewRequest(t.TempDir())
	req.Pattern = "^(\\w+)-(\\d+)"
	req.Template = "@3"

	_, _, err := regex.Preview(context.Background(), req, nil)
	if err == nil {
		t.Fatalf("expected error for undefined capture group")
	}
}

func TestRegexPreviewHandlesInvalidPattern(t *testing.T) {
	req := regex.NewRequest(t.TempDir())
	req.Pattern = "(([" // invalid pattern
	req.Template = "@1"

	_, _, err := regex.Preview(context.Background(), req, nil)
	if err == nil {
		t.Fatalf("expected error for invalid pattern")
	}
}

func TestRegexPreviewSkipsUnmatchedOptionalGroup(t *testing.T) {
	req := regex.NewRequest(t.TempDir())
	req.Pattern = "^(\\w+)(?:-(\\d+))?"
	req.Template = "@1_@2"

	summary, _, err := regex.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalCandidates != 0 {
		t.Fatalf("expected no candidates without files, got %d", summary.TotalCandidates)
	}
}
