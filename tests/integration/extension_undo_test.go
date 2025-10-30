package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestExtensionAutomationUndo(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "config.yaml"))
	createFile(t, filepath.Join(tmp, "notes.yml"))

	var applyOut bytes.Buffer
	apply := renamercmd.NewRootCommand()
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"extension", ".yaml", ".yml", ".yml", "--yes", "--path", tmp})

	if err := apply.Execute(); err != nil {
		t.Fatalf("automation apply failed: %v\noutput: %s", err, applyOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "config.yml")); err != nil {
		t.Fatalf("expected config.yml after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "config.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected config.yaml renamed, err=%v", err)
	}

	var undoOut bytes.Buffer
	undo := renamercmd.NewRootCommand()
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", tmp})

	if err := undo.Execute(); err != nil {
		t.Fatalf("undo failed: %v\noutput: %s", err, undoOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "config.yaml")); err != nil {
		t.Fatalf("expected config.yaml after undo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "config.yml")); !os.IsNotExist(err) {
		t.Fatalf("expected config.yml removed after undo, err=%v", err)
	}
}
