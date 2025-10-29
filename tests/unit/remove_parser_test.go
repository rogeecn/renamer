package replace_test

import (
	"testing"

	"github.com/rogeecn/renamer/internal/remove"
)

func TestParseArgsDeduplicatesPreservingOrder(t *testing.T) {
	args := []string{" draft", " draft", " copy", " draft"}

	result, err := remove.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	expected := []string{" draft", " copy"}
	if len(result.Tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(result.Tokens))
	}
	for i, token := range expected {
		if result.Tokens[i] != token {
			t.Fatalf("token[%d] mismatch: expected %q, got %q", i, token, result.Tokens[i])
		}
	}

	if len(result.Duplicates) != 2 {
		t.Fatalf("expected 2 duplicates recorded, got %d", len(result.Duplicates))
	}
	if result.Duplicates[0] != " draft" || result.Duplicates[1] != " draft" {
		t.Fatalf("unexpected duplicates: %#v", result.Duplicates)
	}
}

func TestParseArgsSkipsWhitespaceOnlyTokens(t *testing.T) {
	args := []string{"  ", "\t", "foo"}

	result, err := remove.ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if len(result.Tokens) != 1 || result.Tokens[0] != "foo" {
		t.Fatalf("expected single token 'foo', got %#v", result.Tokens)
	}
}
