package contract

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogeecn/renamer/cmd"
)

func TestAICommandPreviewTable(t *testing.T) {
	t.Setenv("RENAMER_AI_KEY", "test-key")

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "IMG_0001.jpg"))
	createFile(t, filepath.Join(tmp, "trip-notes.txt"))

	root := cmd.NewRootCommand()
	root.SetArgs([]string{"ai", "--path", tmp, "--prompt", "Travel Memories", "--dry-run"})

	var buf bytes.Buffer
	root.SetIn(strings.NewReader("\n"))
	root.SetOut(&buf)
	root.SetErr(&buf)

	if err := root.Execute(); err != nil {
		t.Fatalf("ai command returned error: %v\noutput: %s", err, buf.String())
	}

	output := buf.String()

	if !strings.Contains(output, "IMG_0001.jpg") {
		t.Fatalf("expected original filename in preview, got: %s", output)
	}

	if !strings.Contains(output, "trip-notes.txt") {
		t.Fatalf("expected secondary filename in preview, got: %s", output)
	}

	if !strings.Contains(output, "01.travel-memories-img-0001.jpg") {
		t.Fatalf("expected deterministic suggestion in preview, got: %s", output)
	}

	if !strings.Contains(output, "02.travel-memories-trip-notes.txt") {
		t.Fatalf("expected sequential suggestion for second file, got: %s", output)
	}
}
