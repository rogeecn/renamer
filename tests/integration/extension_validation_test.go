package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestExtensionCommandBlocksConflicts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "doc.jpeg"))
	createFile(t, filepath.Join(tmp, "doc.jpg"))

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"extension", ".jpeg", ".jpg", ".jpg", "--yes", "--path", tmp})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected conflict to produce an error")
	}

	if !strings.Contains(out.String(), "existing") && !strings.Contains(out.String(), "conflict") {
		t.Fatalf("expected conflict messaging in output, got: %s", out.String())
	}

	// Ensure files unchanged after failed apply.
	if _, err := os.Stat(filepath.Join(tmp, "doc.jpeg")); err != nil {
		t.Fatalf("expected doc.jpeg to remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "doc.jpg")); err != nil {
		t.Fatalf("expected doc.jpg to remain: %v", err)
	}
}
