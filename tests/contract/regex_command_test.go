package contract

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/regex"
)

func TestRegexPreviewUsesCaptureGroups(t *testing.T) {
	tmp := t.TempDir()
	copyRegexFixture(t, "baseline", tmp)

	req := regex.NewRequest(tmp)
	req.Pattern = "^(\\w+)-(\\d+)"
	req.Template = "@2_@1"
	req.IncludeDirectories = false
	req.Recursive = false

	var buf bytes.Buffer
	summary, planned, err := regex.Preview(context.Background(), req, &buf)
	if err != nil {
		t.Fatalf("regex preview returned error: %v", err)
	}

	if summary.TotalCandidates != len(summary.Entries) {
		t.Fatalf("expected summary entries to equal candidates: %d vs %d", summary.TotalCandidates, len(summary.Entries))
	}

	expected := map[string]string{
		"alpha-123.log": "123_alpha.log",
		"beta-456.log":  "456_beta.log",
		"gamma-789.log": "789_gamma.log",
	}

	if summary.Changed != len(planned) {
		t.Fatalf("expected changed count %d to equal plan length %d", summary.Changed, len(planned))
	}

	for _, entry := range summary.Entries {
		target, ok := expected[filepath.Base(entry.OriginalPath)]
		if !ok {
			t.Fatalf("unexpected candidate in preview: %s", entry.OriginalPath)
		}
		if entry.ProposedPath != filepath.Join(filepath.Dir(entry.OriginalPath), target) {
			t.Fatalf("expected proposed path %s, got %s", target, entry.ProposedPath)
		}
		if entry.Status != regex.EntryChanged {
			t.Fatalf("expected entry status 'changed', got %s", entry.Status)
		}
	}

	if len(planned) != len(expected) {
		t.Fatalf("expected plan length %d, got %d", len(expected), len(planned))
	}

	for _, plan := range planned {
		base := filepath.Base(plan.SourceRelative)
		target, ok := expected[base]
		if !ok {
			t.Fatalf("unexpected plan entry: %s", base)
		}
		if plan.TargetRelative != filepath.Join(filepath.Dir(plan.SourceRelative), target) {
			t.Fatalf("expected planned target %s, got %s", target, plan.TargetRelative)
		}
		if len(plan.MatchGroups) != 2 {
			t.Fatalf("expected 2 match groups in plan, got %d", len(plan.MatchGroups))
		}
	}

	output := buf.String()
	for _, target := range expected {
		if !bytes.Contains([]byte(output), []byte(target)) {
			t.Fatalf("expected preview output to contain %s, got %s", target, output)
		}
	}
}

func copyRegexFixture(t *testing.T, name, dest string) {
	t.Helper()
	src := filepath.Join("..", "fixtures", "regex", name)
	if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, rel)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
}
