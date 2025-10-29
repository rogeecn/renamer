package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRemoveCommandEmptyBasenameWarning(t *testing.T) {
	tmp := t.TempDir()

	createValidationFile(t, filepath.Join(tmp, "draft"))
	createValidationFile(t, filepath.Join(tmp, "draft copy.txt"))

	root := renamercmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"remove", "draft", "--path", tmp, "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("remove dry-run failed: %v\noutput: %s", err, out.String())
	}

	if !strings.Contains(out.String(), "Warning: draft would become empty; skipping") {
		t.Fatalf("expected empty basename warning, got: %s", out.String())
	}
}

func TestRemoveCommandDuplicateWarning(t *testing.T) {
	tmp := t.TempDir()

	createValidationFile(t, filepath.Join(tmp, "foo draft draft.txt"))

	root := renamercmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"remove", " draft", " draft", "--path", tmp, "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("remove dry-run failed: %v\noutput: %s", err, out.String())
	}

	if !strings.Contains(out.String(), "Warning: token \" draft\" provided multiple times") {
		t.Fatalf("expected duplicate warning, got: %s", out.String())
	}
}

func createValidationFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
