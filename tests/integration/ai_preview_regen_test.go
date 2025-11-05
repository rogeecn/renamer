package integration

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogeecn/renamer/cmd"
)

func TestAICommandSupportsPromptRefinement(t *testing.T) {
	t.Setenv("RENAMER_AI_KEY", "test-key")

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "IMG_1024.jpg"))
	createFile(t, filepath.Join(tmp, "notes/day1.txt"))

	root := cmd.NewRootCommand()
	root.SetArgs([]string{"ai", "--path", tmp})

	// Simulate editing the prompt then quitting.
	var output bytes.Buffer
	input := strings.NewReader("e\nVacation Highlights\nq\n")
	root.SetIn(input)
	root.SetOut(&output)
	root.SetErr(&output)

	if err := root.Execute(); err != nil {
		t.Fatalf("ai command returned error: %v\noutput: %s", err, output.String())
	}

	got := output.String()

	if !strings.Contains(got, "Current prompt: \"Vacation Highlights\"") {
		t.Fatalf("expected updated prompt in output, got: %s", got)
	}

	if !strings.Contains(got, "01.vacation-highlights-img-1024.jpg") {
		t.Fatalf("expected regenerated suggestion with new prefix, got: %s", got)
	}

	if !strings.Contains(got, "Session ended without applying changes.") {
		t.Fatalf("expected session completion message, got: %s", got)
	}
}
