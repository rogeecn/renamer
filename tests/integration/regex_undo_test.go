package integration

import (
	"bytes"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestRegexUndoRestoresAutomationRun(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	copyRegexFixtureIntegration(t, "mixed", tmp)

	apply := renamercmd.NewRootCommand()
	var applyOut bytes.Buffer
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"regex", "^build_(\\d+)_(.*)$", "release-@1-@2", "--yes", "--path", tmp})
	if err := apply.Execute(); err != nil {
		t.Fatalf("regex apply failed: %v\noutput: %s", err, applyOut.String())
	}

	if !fileExistsTestHelper(filepath.Join(tmp, "release-101-release.tar.gz")) || !fileExistsTestHelper(filepath.Join(tmp, "release-102-hotfix.tar.gz")) {
		t.Fatalf("expected renamed files after apply")
	}

	undo := renamercmd.NewRootCommand()
	var undoOut bytes.Buffer
	undo.SetOut(&undoOut)
	undo.SetErr(&undoOut)
	undo.SetArgs([]string{"undo", "--path", tmp})
	if err := undo.Execute(); err != nil {
		t.Fatalf("undo failed: %v\noutput: %s", err, undoOut.String())
	}

	if !fileExistsTestHelper(filepath.Join(tmp, "build_101_release.tar.gz")) || !fileExistsTestHelper(filepath.Join(tmp, "build_102_hotfix.tar.gz")) {
		t.Fatalf("expected originals restored after undo")
	}
}
