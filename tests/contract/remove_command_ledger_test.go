package contract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/history"
)

func TestRemoveCommandLedgerMetadata(t *testing.T) {
	tmp := t.TempDir()

	createRemoveFile(t, filepath.Join(tmp, "report copy draft.txt"))
	createRemoveFile(t, filepath.Join(tmp, "notes draft.txt"))

	root := renamercmd.NewRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"remove", " copy", " draft", "--path", tmp, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("remove command error: %v\noutput: %s", err, out.String())
	}

	ledgerPath := filepath.Join(tmp, ".renamer")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(lines) == 0 {
		t.Fatalf("expected ledger entry written")
	}

	var entry history.Entry
	if err := json.Unmarshal(lines[len(lines)-1], &entry); err != nil {
		t.Fatalf("decode ledger entry: %v", err)
	}

	if entry.Command != "remove" {
		t.Fatalf("expected remove command recorded, got %q", entry.Command)
	}

	tokensVal, ok := entry.Metadata["tokens"].([]any)
	if !ok {
		t.Fatalf("expected tokens metadata, got %#v", entry.Metadata)
	}
	tokens := make([]string, len(tokensVal))
	for i, v := range tokensVal {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("token entry not string: %#v", v)
		}
		tokens[i] = s
	}
	if len(tokens) != 2 || tokens[0] != " copy" || tokens[1] != " draft" {
		t.Fatalf("unexpected tokens metadata: %#v", tokens)
	}

	matchesVal, ok := entry.Metadata["matches"].(map[string]any)
	if !ok {
		t.Fatalf("expected matches metadata, got %#v", entry.Metadata)
	}
	if len(matchesVal) != 2 {
		t.Fatalf("unexpected matches metadata: %#v", matchesVal)
	}
	if toFloat(matchesVal[" copy"]) != 1 || toFloat(matchesVal[" draft"]) != 2 {
		t.Fatalf("unexpected match counts: %#v", matchesVal)
	}

	// Ensure undo restores originals for automation workflows.
	undo := renamercmd.NewRootCommand()
	undo.SetOut(&bytes.Buffer{})
	undo.SetErr(&bytes.Buffer{})
	undo.SetArgs([]string{"undo", "--path", tmp})
	if err := undo.Execute(); err != nil {
		t.Fatalf("undo command error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "report copy draft.txt")); err != nil {
		t.Fatalf("expected original restored after undo: %v", err)
	}
}

func toFloat(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return -1
	}
}
