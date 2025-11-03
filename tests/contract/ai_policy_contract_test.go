package contract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/ai/genkit"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

type captureWorkflow struct {
	request genkit.Request
}

func (c *captureWorkflow) Run(ctx context.Context, req genkit.Request) (genkit.Result, error) {
	c.request = req
	return genkit.Result{
		Response: prompt.RenameResponse{
			Items: []prompt.RenameItem{
				{
					Original: "alpha.txt",
					Proposed: "proj_001_sample_file.txt",
					Sequence: 1,
				},
			},
			Model: "test-model",
		},
	}, nil
}

func TestAICommandAppliesNamingPoliciesToPrompt(t *testing.T) {
	genkit.ResetWorkflowFactory()
	stub := &captureWorkflow{}
	genkit.OverrideWorkflowFactory(func(ctx context.Context, opts genkit.Options) (genkit.WorkflowRunner, error) {
		return stub, nil
	})
	t.Cleanup(genkit.ResetWorkflowFactory)

	rootDir := t.TempDir()
	createPolicyTestFile(t, filepath.Join(rootDir, "alpha.txt"))

	rootCmd := renamercmd.NewRootCommand()
	var stdout, stderr bytes.Buffer
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{
		"ai",
		"--path", rootDir,
		"--dry-run",
		"--naming-casing", "snake",
		"--naming-prefix", "proj",
		"--naming-allow-spaces",
		"--naming-keep-order",
		"--banned", "alpha",
	})

	if err := rootCmd.Execute(); err != nil {
		if stdout.Len() > 0 {
			t.Logf("stdout: %s", stdout.String())
		}
		if stderr.Len() > 0 {
			t.Logf("stderr: %s", stderr.String())
		}
		t.Fatalf("command execute: %v", err)
	}

	req := stub.request
	policies := req.Payload.Policies
	if policies.Prefix != "proj" {
		t.Fatalf("expected prefix proj, got %q", policies.Prefix)
	}
	if policies.Casing != "snake" {
		t.Fatalf("expected casing snake, got %q", policies.Casing)
	}
	if !policies.AllowSpaces {
		t.Fatalf("expected allow spaces flag to propagate")
	}
	if !policies.KeepOriginalOrder {
		t.Fatalf("expected keep original order flag to propagate")
	}
	if len(policies.ForbiddenTokens) != 1 || policies.ForbiddenTokens[0] != "alpha" {
		t.Fatalf("expected forbidden tokens to capture user list, got %#v", policies.ForbiddenTokens)
	}

	banned := req.Payload.BannedTerms
	containsDefault := false
	containsUser := false
	for _, term := range banned {
		switch term {
		case "alpha":
			containsUser = true
		case "clickbait":
			containsDefault = true
		}
	}
	if !containsUser {
		t.Fatalf("expected banned terms to include user-provided token")
	}
	if !containsDefault {
		t.Fatalf("expected banned terms to retain default tokens")
	}
}

func createPolicyTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("demo"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
