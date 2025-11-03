package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/ai/genkit"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

type stubWorkflow struct {
	response prompt.RenameResponse
}

func (s stubWorkflow) Run(ctx context.Context, req genkit.Request) (genkit.Result, error) {
	return genkit.Result{Response: s.response}, nil
}

func TestAIPreviewFlowRendersSequenceTable(t *testing.T) {
	workflow := stubWorkflow{
		response: prompt.RenameResponse{
			Items: []prompt.RenameItem{
				{
					Original: "promo SALE 01.JPG",
					Proposed: "001_summer-session.jpg",
					Sequence: 1,
					Notes:    "Removed promotional flair",
				},
				{
					Original: "family_photo.png",
					Proposed: "002_family-photo.png",
					Sequence: 2,
					Notes:    "Normalized casing",
				},
			},
			Warnings:   []string{"AI warning: trimmed banned tokens"},
			PromptHash: "",
		},
	}

	genkit.OverrideWorkflowFactory(func(ctx context.Context, opts genkit.Options) (genkit.WorkflowRunner, error) {
		return workflow, nil
	})
	defer genkit.ResetWorkflowFactory()

	root := t.TempDir()
	createAIPreviewFile(t, filepath.Join(root, "promo SALE 01.JPG"))
	createAIPreviewFile(t, filepath.Join(root, "family_photo.png"))

	t.Setenv("default_MODEL_AUTH_TOKEN", "test-token")

	rootCmd := renamercmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	exportPath := filepath.Join(root, "plan.json")
	rootCmd.SetArgs([]string{
		"ai",
		"--path", root,
		"--dry-run",
		"--debug-genkit",
		"--export-plan", exportPath,
	})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command execute: %v", err)
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported plan: %v", err)
	}

	var exported prompt.RenameResponse
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("unmarshal exported plan: %v", err)
	}
	if len(exported.Items) != len(workflow.response.Items) {
		t.Fatalf("expected exported items %d, got %d", len(workflow.response.Items), len(exported.Items))
	}

	out := stdout.String()
	if !strings.Contains(out, "SEQ") || !strings.Contains(out, "ORIGINAL") || !strings.Contains(out, "SANITIZED") {
		t.Fatalf("expected table headers in output, got:\n%s", out)
	}
	if !strings.Contains(out, "001") || !strings.Contains(out, "promo SALE 01.JPG") || !strings.Contains(out, "001_summer-session.jpg") {
		t.Fatalf("expected first entry in output, got:\n%s", out)
	}
	if !strings.Contains(out, "removed: promo sale") {
		t.Fatalf("expected sanitization notes in output, got:\n%s", out)
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "Prompt hash:") {
		t.Fatalf("expected prompt hash in debug output, got:\n%s", errOut)
	}
	if !strings.Contains(errOut, "AI warning: trimmed banned tokens") {
		t.Fatalf("expected warning surfaced in debug output, got:\n%s", errOut)
	}
}

func createAIPreviewFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
