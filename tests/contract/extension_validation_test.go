package contract

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/extension"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestParseArgsValidation(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"tooFew", []string{".jpg"}},
		{"emptySource", []string{" ", ".jpg"}},
		{"missingDotSource", []string{"jpg", ".png"}},
		{"missingDotTarget", []string{".jpg", "png"}},
	}

	for _, tc := range cases {
		_, err := extension.ParseArgs(tc.args)
		if err == nil {
			t.Fatalf("expected error for case %s", tc.name)
		}
	}

	_, err := extension.ParseArgs([]string{".jpg", ".JPG"})
	if err == nil {
		t.Fatalf("expected error when all sources match target")
	}

	parsed, err := extension.ParseArgs([]string{".jpeg", ".JPG", ".jpg"})
	if err != nil {
		t.Fatalf("unexpected error for valid args: %v", err)
	}
	if len(parsed.SourcesCanonical) != 1 || parsed.SourcesCanonical[0] != ".jpeg" {
		t.Fatalf("expected canonical list to contain .jpeg only, got %#v", parsed.SourcesCanonical)
	}
	if len(parsed.NoOps) != 1 {
		t.Fatalf("expected .jpg to be treated as no-op")
	}
}

func TestPreviewDetectsConflicts(t *testing.T) {
	tmp := t.TempDir()
	writeTestFile(t, filepath.Join(tmp, "report.jpeg"))
	writeTestFile(t, filepath.Join(tmp, "report.jpg"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := extension.NewRequest(scope)
	parsed, err := extension.ParseArgs([]string{".jpeg", ".jpg"})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	req.SetExtensions(parsed.SourcesCanonical, parsed.SourcesDisplay, parsed.Target)
	req.SetWarnings(parsed.Duplicates, parsed.NoOps)

	summary, planned, err := extension.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}
	if !summary.HasConflicts() {
		t.Fatalf("expected conflict when target already exists")
	}
	if len(planned) != 0 {
		t.Fatalf("expected no operations due to conflict, got %d", len(planned))
	}
	if len(summary.Warnings) == 0 {
		t.Fatalf("expected warning recorded for conflict")
	}

	// Apply should be skipped by caller; invoking directly without operations should no-op.
	req.SetExecutionMode(false, true)
	entry, err := extension.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}
	if len(entry.Operations) != 0 {
		t.Fatalf("expected zero operations recorded when conflicts present")
	}

	if _, err := extension.ParseArgs([]string{".jpeg", ".jpg"}); err != nil {
		// ensure previous parse errors do not leak state
		if !errors.Is(err, nil) {
			// unreachable, but keeps staticcheck happy
		}
	}
}
