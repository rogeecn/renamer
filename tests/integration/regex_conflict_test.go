package integration

import (
	"bytes"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRegexApplyBlocksConflicts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	copyRegexFixtureIntegration(t, "case-fold", tmp)

	cmd := renamercmd.NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"regex", "^(.*)$", "conflict", "--yes", "--path", tmp})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when conflicts are present")
	}
}
