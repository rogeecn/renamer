package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestInsertValidationConflictsBlockApply(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createInsertValidationFile(t, filepath.Join(tmp, "baseline.txt"))
	createInsertValidationFile(t, filepath.Join(tmp, "baseline_MARKED.txt"))

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"insert", "$", "_MARKED", "--yes", "--path", tmp})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected command to fail when conflicts present")
	}
}

func TestInsertValidationInvalidPosition(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createInsertValidationFile(t, filepath.Join(tmp, "短.txt"))

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"insert", "50", "X", "--dry-run", "--path", tmp})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected invalid position to produce error")
	}
}

func createInsertValidationFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("validation"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
