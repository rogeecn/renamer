package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestExtensionCommandFlow(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "image.jpeg"))
	createFile(t, filepath.Join(tmp, "poster.JPG"))
	createFile(t, filepath.Join(tmp, "logo.jpg"))

	var previewOut bytes.Buffer
	preview := renamercmd.NewRootCommand()
	preview.SetOut(&previewOut)
	preview.SetErr(&previewOut)
	preview.SetArgs([]string{"extension", ".jpeg", ".JPG", ".jpg", "--dry-run", "--path", tmp})

	if err := preview.Execute(); err != nil {
		t.Fatalf("preview command failed: %v\noutput: %s", err, previewOut.String())
	}

	output := previewOut.String()
	if !strings.Contains(output, "image.jpeg -> image.jpg") {
		t.Fatalf("expected preview output to include image rename, got:\n%s", output)
	}
	if !strings.Contains(output, "poster.JPG -> poster.jpg") {
		t.Fatalf("expected preview output to include poster rename, got:\n%s", output)
	}
	if !strings.Contains(output, "logo.jpg (no change)") {
		t.Fatalf("expected preview output to include no-change row for logo, got:\n%s", output)
	}

	var applyOut bytes.Buffer
	apply := renamercmd.NewRootCommand()
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"extension", ".jpeg", ".JPG", ".jpg", "--yes", "--path", tmp})

	if err := apply.Execute(); err != nil {
		t.Fatalf("apply command failed: %v\noutput: %s", err, applyOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "image.jpg")); err != nil {
		t.Fatalf("expected image.jpg after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "poster.jpg")); err != nil {
		t.Fatalf("expected poster.jpg after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "logo.jpg")); err != nil {
		t.Fatalf("expected logo.jpg to remain: %v", err)
	}

	var undoOut bytes.Buffer
	undo := renamercmd.NewRootCommand()
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", tmp})

	if err := undo.Execute(); err != nil {
		t.Fatalf("undo command failed: %v\noutput: %s", err, undoOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "image.jpeg")); err != nil {
		t.Fatalf("expected image.jpeg after undo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "poster.JPG")); err != nil {
		t.Fatalf("expected poster.JPG after undo: %v", err)
	}
}

func TestExtensionCommandSkipsCaseVariantsWithoutSource(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "image.jpeg"))
	createFile(t, filepath.Join(tmp, "poster.JPG"))
	createFile(t, filepath.Join(tmp, "logo.jpg"))

	var previewOut bytes.Buffer
	preview := renamercmd.NewRootCommand()
	preview.SetOut(&previewOut)
	preview.SetErr(&previewOut)
	preview.SetArgs([]string{"extension", ".jpeg", ".jpg", "--dry-run", "--path", tmp})

	if err := preview.Execute(); err != nil {
		t.Fatalf("preview command failed: %v\noutput: %s", err, previewOut.String())
	}

	output := previewOut.String()
	if !strings.Contains(output, "image.jpeg -> image.jpg") {
		t.Fatalf("expected preview output to include image rename, got:\n%s", output)
	}
	if strings.Contains(output, "poster.JPG -> poster.jpg") {
		t.Fatalf("expected poster.JPG to be excluded from preview, got:\n%s", output)
	}

	var applyOut bytes.Buffer
	apply := renamercmd.NewRootCommand()
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"extension", ".jpeg", ".jpg", "--yes", "--path", tmp})

	if err := apply.Execute(); err != nil {
		t.Fatalf("apply command failed: %v\noutput: %s", err, applyOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "image.jpg")); err != nil {
		t.Fatalf("expected image.jpg after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "poster.JPG")); err != nil {
		t.Fatalf("expected poster.JPG to remain uppercase: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "poster.jpg")); !os.IsNotExist(err) {
		t.Fatalf("expected poster.jpg not to exist, err=%v", err)
	}
}
