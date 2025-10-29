package contract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogeecn/renamer/internal/replace"
)

func TestPreviewSummaryCounts(t *testing.T) {
	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "draft.txt"))
	createFile(t, filepath.Join(tmp, "Draft.md"))
	createFile(t, filepath.Join(tmp, "notes", "DRAFT.log"))

	args := []string{"draft", "Draft", "DRAFT", "final"}
	parsed, err := replace.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	req := &replace.ReplaceRequest{
		WorkingDir:  tmp,
		Patterns:    parsed.Patterns,
		Replacement: parsed.Replacement,
		Recursive:   true,
	}

	var buf bytes.Buffer
	summary, planned, err := replace.Preview(context.Background(), req, parsed, &buf)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	if summary.TotalCandidates == 0 {
		t.Fatalf("expected candidates to be processed")
	}

	if summary.ChangedCount != len(planned) {
		t.Fatalf("changed count mismatch: %d vs %d", summary.ChangedCount, len(planned))
	}

	for _, pattern := range []string{"draft", "Draft", "DRAFT"} {
		if summary.PatternMatches[pattern] == 0 {
			t.Fatalf("expected matches recorded for %s", pattern)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "draft.txt -> final.txt") {
		t.Fatalf("expected preview output to list replacements, got: %s", output)
	}

	if summary.ReplacementWasEmpty(parsed.Replacement) {
		t.Fatalf("replacement should not be empty warning for this test")
	}
}

func TestPreviewWarnsOnEmptyReplacement(t *testing.T) {
	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "foo.txt"))

	args := []string{"foo", ""}
	parsed, err := replace.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs error: %v", err)
	}

	req := &replace.ReplaceRequest{
		WorkingDir:  tmp,
		Patterns:    parsed.Patterns,
		Replacement: parsed.Replacement,
	}

	var buf bytes.Buffer
	summary, _, err := replace.Preview(context.Background(), req, parsed, &buf)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	if !summary.EmptyReplacement {
		t.Fatalf("expected empty replacement flag to be set")
	}

	if !strings.Contains(buf.String(), "Warning: replacement string is empty") {
		t.Fatalf("expected empty replacement warning in preview output")
	}
}

func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
