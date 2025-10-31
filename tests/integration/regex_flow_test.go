package integration

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRegexPreviewCommand(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	copyRegexFixtureIntegration(t, "baseline", tmp)

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"regex", "^(\\w+)-(\\d+)", "@2_@1", "--dry-run", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("regex preview command failed: %v\noutput: %s", err, out.String())
	}

	expected := []string{
		"alpha-123.log -> 123_alpha.log",
		"beta-456.log -> 456_beta.log",
		"gamma-789.log -> 789_gamma.log",
		"Preview complete: 3 matched, 3 changed, 0 skipped.",
		"Preview complete. Re-run with --yes to apply.",
	}

	for _, token := range expected {
		if !bytes.Contains(out.Bytes(), []byte(token)) {
			t.Fatalf("expected output to contain %q, got: %s", token, out.String())
		}
	}
}

func copyRegexFixtureIntegration(t *testing.T, name, dest string) {
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

		return os.WriteFile(targetPath, content, 0o644)
	}); err != nil {
		t.Fatalf("copy regex fixture: %v", err)
	}
}
