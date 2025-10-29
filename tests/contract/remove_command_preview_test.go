package contract

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRemoveCommandDryRunPreview(t *testing.T) {
	tmp := t.TempDir()
	createRemoveFile(t, filepath.Join(tmp, "report copy draft.txt"))
	createRemoveFile(t, filepath.Join(tmp, "notes draft.txt"))

	root := renamercmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"remove", " copy", " draft", "--path", tmp, "--dry-run"})

	if err := root.Execute(); err != nil {
		t.Fatalf("remove command returned error: %v (output: %s)", err, out.String())
	}

	output := out.String()
	if !strings.Contains(output, "report copy draft.txt -> report.txt") {
		t.Fatalf("expected preview mapping in output, got: %s", output)
	}

	if !strings.Contains(output, "Preview complete") {
		t.Fatalf("expected dry-run completion message, got: %s", output)
	}
}

func createRemoveFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
