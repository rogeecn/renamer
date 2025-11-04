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
	"github.com/rogeecn/renamer/internal/ai/plan"
	"github.com/rogeecn/renamer/internal/ai/prompt"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestAIApplyAndUndoFlow(t *testing.T) {
	initialWorkflow := stubWorkflow{
		response: prompt.RenameResponse{
			Items: []prompt.RenameItem{
				{
					Original: "draft_one.txt",
					Proposed: "001_initial.txt",
					Sequence: 1,
				},
				{
					Original: "draft_two.txt",
					Proposed: "002_initial.txt",
					Sequence: 2,
				},
			},
			Model: "test-model",
		},
	}

	genkit.OverrideWorkflowFactory(func(ctx context.Context, opts genkit.Options) (genkit.WorkflowRunner, error) {
		return initialWorkflow, nil
	})
	t.Cleanup(genkit.ResetWorkflowFactory)

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "draft_one.txt"))
	writeFile(t, filepath.Join(root, "draft_two.txt"))

	planPath := filepath.Join(root, "renamer.plan.json")

	preview := renamercmd.NewRootCommand()
	var previewOut, previewErr bytes.Buffer
	preview.SetOut(&previewOut)
	preview.SetErr(&previewErr)
	preview.SetArgs([]string{
		"ai",
		"--path", root,
		"--dry-run",
	})

	if err := preview.Execute(); err != nil {
		if previewOut.Len() > 0 {
			t.Logf("preview stdout: %s", previewOut.String())
		}
		if previewErr.Len() > 0 {
			t.Logf("preview stderr: %s", previewErr.String())
		}
		t.Fatalf("initial preview: %v", err)
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read plan: %v", err)
	}
	var exported prompt.RenameResponse
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}

	if len(exported.Items) != 2 {
		t.Fatalf("expected two plan items, got %d", len(exported.Items))
	}
	// Simulate operator edit.
	exported.Items[0].Proposed = "001_final-one.txt"
	exported.Items[1].Proposed = "002_final-two.txt"
	exported.Items[0].Notes = "custom edit"

	modified, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		t.Fatalf("marshal modified plan: %v", err)
	}
	if err := os.WriteFile(planPath, append(modified, '\n'), 0o644); err != nil {
		t.Fatalf("write modified plan: %v", err)
	}

	req := &listing.ListingRequest{WorkingDir: root}
	if err := req.Validate(); err != nil {
		t.Fatalf("validate listing request: %v", err)
	}
	currentCandidates, err := plan.CollectCandidates(context.Background(), req)
	if err != nil {
		t.Fatalf("collect candidates: %v", err)
	}
	filtered := make([]plan.Candidate, 0, len(currentCandidates))
	for _, cand := range currentCandidates {
		if strings.EqualFold(cand.OriginalPath, filepath.Base(planPath)) {
			continue
		}
		filtered = append(filtered, cand)
	}
	originals := make([]string, 0, len(filtered))
	for _, cand := range filtered {
		originals = append(originals, cand.OriginalPath)
	}
	validator := plan.NewValidator(originals, prompt.NamingPolicyConfig{Casing: "kebab"}, nil)
	if _, err := validator.Validate(exported); err != nil {
		t.Fatalf("pre-validation of edited plan: %v", err)
	}

	previewEdited := renamercmd.NewRootCommand()
	var editedOut, editedErr bytes.Buffer
	previewEdited.SetOut(&editedOut)
	previewEdited.SetErr(&editedErr)
	previewEdited.SetArgs([]string{
		"ai",
		"--path", root,
		"--dry-run",
	})

	if err := previewEdited.Execute(); err != nil {
		if editedOut.Len() > 0 {
			t.Logf("edited stdout: %s", editedOut.String())
		}
		if editedErr.Len() > 0 {
			t.Logf("edited stderr: %s", editedErr.String())
		}
		t.Fatalf("preview edited plan: %v", err)
	}

	if !strings.Contains(editedOut.String(), "001_final-one.txt") {
		t.Fatalf("expected edited preview to show final name, got: %s", editedOut.String())
	}

	applyCmd := renamercmd.NewRootCommand()
	var applyOut, applyErr bytes.Buffer
	applyCmd.SetOut(&applyOut)
	applyCmd.SetErr(&applyErr)
	applyCmd.SetArgs([]string{
		"ai",
		"--path", root,
		"--yes",
	})

	if err := applyCmd.Execute(); err != nil {
		if applyOut.Len() > 0 {
			t.Logf("apply stdout: %s", applyOut.String())
		}
		if applyErr.Len() > 0 {
			t.Logf("apply stderr: %s", applyErr.String())
		}
		t.Fatalf("apply plan: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "001_final-one.txt")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "002_final-two.txt")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}

	undo := renamercmd.NewRootCommand()
	var undoOut bytes.Buffer
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", root})

	if err := undo.Execute(); err != nil {
		t.Fatalf("undo command: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "draft_one.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "draft_two.txt")); err != nil {
		t.Fatalf("expected original file after undo: %v", err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
