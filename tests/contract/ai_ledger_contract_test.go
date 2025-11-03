package contract

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/ai/plan"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

func TestAIApplyLedgerMetadataContract(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "sample.txt"))

	candidate := plan.Candidate{
		OriginalPath: "sample.txt",
		SizeBytes:    4,
		Depth:        0,
	}

	response := prompt.RenameResponse{
		Items: []prompt.RenameItem{
			{
				Original: "sample.txt",
				Proposed: "001_sample-final.txt",
				Sequence: 1,
			},
		},
		Model:      "test-model",
		PromptHash: "prompt-hash-123",
	}

	policy := prompt.NamingPolicyConfig{Prefix: "proj", Casing: "kebab"}

	entry, err := plan.Apply(context.Background(), plan.ApplyOptions{
		WorkingDir: root,
		Candidates: []plan.Candidate{candidate},
		Response:   response,
		Policies:   policy,
		PromptHash: response.PromptHash,
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}

	if len(entry.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(entry.Operations))
	}

	planFile := filepath.Join(root, "001_sample-final.txt")
	if _, err := os.Stat(planFile); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}

	ledgerPath := filepath.Join(root, ".renamer")
	file, err := os.Open(ledgerPath)
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan ledger: %v", err)
	}

	var recorded map[string]any
	if err := json.Unmarshal([]byte(lastLine), &recorded); err != nil {
		t.Fatalf("unmarshal ledger entry: %v", err)
	}

	metaRaw, ok := recorded["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata in ledger entry")
	}
	aiRaw, ok := metaRaw["ai"].(map[string]any)
	if !ok {
		t.Fatalf("expected ai metadata in ledger entry")
	}

	if aiRaw["model"] != "test-model" {
		t.Fatalf("expected model test-model, got %v", aiRaw["model"])
	}
	if aiRaw["promptHash"] != "prompt-hash-123" {
		t.Fatalf("expected prompt hash recorded, got %v", aiRaw["promptHash"])
	}
	if aiRaw["batchSize"].(float64) != 1 {
		t.Fatalf("expected batch size 1, got %v", aiRaw["batchSize"])
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
