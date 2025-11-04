package integration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/ai/genkit"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

type violatingWorkflow struct{}

func (violatingWorkflow) Run(ctx context.Context, req genkit.Request) (genkit.Result, error) {
	return genkit.Result{
		Response: prompt.RenameResponse{
			Items: []prompt.RenameItem{
				{
					Original: "video.mov",
					Proposed: "001_clickbait-offer.mov",
					Sequence: 1,
				},
			},
			Warnings: []string{"model returned promotional phrasing"},
		},
	}, nil
}

func TestAIPolicyValidationFailsWithActionableMessage(t *testing.T) {
	genkit.OverrideWorkflowFactory(func(ctx context.Context, opts genkit.Options) (genkit.WorkflowRunner, error) {
		return violatingWorkflow{}, nil
	})
	t.Cleanup(genkit.ResetWorkflowFactory)

	rootDir := t.TempDir()
	createAIPolicyFixture(t, filepath.Join(rootDir, "video.mov"))

	rootCmd := renamercmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"ai",
		"--path", rootDir,
		"--dry-run",
	})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatalf("expected policy violation error")
	}

	lines := stderr.String()
	if !strings.Contains(lines, "Policy violation (banned)") {
		t.Fatalf("expected banned token message in stderr, got: %s", lines)
	}
	if !strings.Contains(err.Error(), "policy violations") {
		t.Fatalf("expected error to mention policy violations, got: %v", err)
	}

	if stdout.Len() != 0 {
		t.Logf("stdout: %s", stdout.String())
	}
}

func createAIPolicyFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
