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

func TestRegexCommandLedgerMetadata(t *testing.T) {
	tmp := t.TempDir()
	copyRegexFixture(t, "mixed", tmp)

	cmd := renamercmd.NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"regex", "^build_(\\d+)_(.*)$", "release-@1-@2", "--yes", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("regex apply command failed: %v\noutput: %s", err, out.String())
	}

	ledgerPath := filepath.Join(tmp, ".renamer")
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(lines) == 0 {
		t.Fatalf("expected ledger entries written")
	}

	var entry history.Entry
	if err := json.Unmarshal(lines[len(lines)-1], &entry); err != nil {
		t.Fatalf("decode ledger entry: %v", err)
	}

	if entry.Command != "regex" {
		t.Fatalf("expected regex command recorded, got %q", entry.Command)
	}

	if entry.Metadata["pattern"] != "^build_(\\d+)_(.*)$" {
		t.Fatalf("unexpected pattern metadata: %#v", entry.Metadata["pattern"])
	}
	if entry.Metadata["template"] != "release-@1-@2" {
		t.Fatalf("unexpected template metadata: %#v", entry.Metadata["template"])
	}

	if toFloat(entry.Metadata["matched"]) != 2 || toFloat(entry.Metadata["changed"]) != 2 {
		t.Fatalf("unexpected match/change counts: %#v", entry.Metadata)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(entry.Operations))
	}
}
