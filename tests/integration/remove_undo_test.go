package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRemoveCommandAutomationUndo(t *testing.T) {
	tmp := t.TempDir()

	createFile(t, filepath.Join(tmp, "alpha copy.txt"))
	createFile(t, filepath.Join(tmp, "nested", "beta draft.txt"))

	preview := renamercmd.NewRootCommand()
	var previewOut bytes.Buffer
	preview.SetOut(&previewOut)
	preview.SetErr(&previewOut)
	preview.SetArgs([]string{"remove", " copy", " draft", "--path", tmp, "--recursive", "--dry-run"})
	if err := preview.Execute(); err != nil {
		t.Fatalf("preview failed: %v\noutput: %s", err, previewOut.String())
	}

	apply := renamercmd.NewRootCommand()
	var applyOut bytes.Buffer
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"remove", " copy", " draft", "--path", tmp, "--recursive", "--yes"})
	if err := apply.Execute(); err != nil {
		t.Fatalf("apply failed: %v\noutput: %s", err, applyOut.String())
	}

	if !fileExists(filepath.Join(tmp, "alpha.txt")) || !fileExists(filepath.Join(tmp, "nested", "beta.txt")) {
		t.Fatalf("expected files renamed after apply")
	}

	undo := renamercmd.NewRootCommand()
	var undoOut bytes.Buffer
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", tmp})
	if err := undo.Execute(); err != nil {
		t.Fatalf("undo failed: %v\noutput: %s", err, undoOut.String())
	}

	if !fileExists(filepath.Join(tmp, "alpha copy.txt")) || !fileExists(filepath.Join(tmp, "nested", "beta draft.txt")) {
		t.Fatalf("expected originals restored after undo")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
