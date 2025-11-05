package contract

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/ai"
	"github.com/rogeecn/renamer/internal/ai/flow"
)

func TestAIMetadataPersistedInLedgerEntry(t *testing.T) {
	t.Setenv("RENAMER_AI_KEY", "test-key")

	tmp := t.TempDir()
	createFile(t, filepath.Join(tmp, "clip.mov"))

	suggestions := []flow.Suggestion{
		{Original: "clip.mov", Suggested: "highlight-01.mov"},
	}

	validation := ai.ValidateSuggestions([]string{"clip.mov"}, suggestions)
	if len(validation.Conflicts) != 0 {
		t.Fatalf("expected no conflicts, got %#v", validation)
	}

	entry, err := ai.Apply(context.Background(), tmp, suggestions, validation, ai.ApplyMetadata{
		Prompt:            "Highlight Reel",
		PromptHistory:     []string{"Highlight Reel", "Celebration Cut"},
		Notes:             []string{"accepted preview"},
		Model:             "googleai/gemini-1.5-flash",
		SequenceSeparator: "_",
	}, io.Discard)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if entry.Command != "ai" {
		t.Fatalf("expected command 'ai', got %q", entry.Command)
	}

	if entry.Metadata == nil {
		t.Fatalf("expected metadata to be recorded")
	}

	if got := entry.Metadata["prompt"]; got != "Highlight Reel" {
		t.Fatalf("unexpected prompt metadata: %#v", got)
	}

	history, ok := entry.Metadata["promptHistory"].([]string)
	if !ok || len(history) != 2 {
		t.Fatalf("unexpected prompt history: %#v", entry.Metadata["promptHistory"])
	}

	model, _ := entry.Metadata["model"].(string)
	if model == "" {
		t.Fatalf("expected model metadata to be present")
	}

	if sep, ok := entry.Metadata["sequenceSeparator"].(string); !ok || sep != "_" {
		t.Fatalf("expected sequence separator metadata, got %#v", entry.Metadata["sequenceSeparator"])
	}

	if _, err := ai.Apply(context.Background(), tmp, suggestions, validation, ai.ApplyMetadata{Prompt: "irrelevant"}, io.Discard); err == nil {
		t.Fatalf("expected error when renaming non-existent file")
	}
}
