package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestInsertAutomationUndo(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createInsertAutomationFile(t, filepath.Join(tmp, "config.yaml"))

	var applyOut bytes.Buffer
	apply := renamercmd.NewRootCommand()
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"insert", "$", "_ARCHIVE", "--yes", "--path", tmp})

	if err := apply.Execute(); err != nil {
		t.Fatalf("apply command failed: %v\noutput: %s", err, applyOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "config_ARCHIVE.yaml")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}

	var undoOut bytes.Buffer
	undo := renamercmd.NewRootCommand()
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", tmp})

	if err := undo.Execute(); err != nil {
		t.Fatalf("undo command failed: %v\noutput: %s", err, undoOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "config.yaml")); err != nil {
		t.Fatalf("expected original file restored: %v", err)
	}

	if !strings.Contains(undoOut.String(), "Inserted text") {
		t.Fatalf("expected undo output to describe inserted text, got: %s", undoOut.String())
	}
}

func createInsertAutomationFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("automation"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
