package replace_test

import (
	"testing"

	"github.com/rogeecn/renamer/internal/replace"
)

func TestParseArgsHandlesWhitespaceAndDuplicates(t *testing.T) {
	args := []string{"  draft  ", "Draft", "draft", "final"}
	result, err := replace.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if len(result.Patterns) != 2 {
		t.Fatalf("expected 2 unique patterns, got %d", len(result.Patterns))
	}
	if len(result.Duplicates) != 1 {
		t.Fatalf("expected duplicate reported, got %d", len(result.Duplicates))
	}
	if result.Replacement != "final" {
		t.Fatalf("replacement mismatch: %s", result.Replacement)
	}
}

func TestParseArgsRequiresSufficientTokens(t *testing.T) {
	if _, err := replace.ParseArgs([]string{"onlyone"}); err == nil {
		t.Fatalf("expected error when replacement missing")
	}
}
