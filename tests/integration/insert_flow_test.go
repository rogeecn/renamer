package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
)

func TestInsertCommandFlow(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createInsertFile(t, filepath.Join(tmp, "holiday.jpg"))
	createInsertFile(t, filepath.Join(tmp, "trip.jpg"))

	var previewOut bytes.Buffer
	preview := renamercmd.NewRootCommand()
	preview.SetOut(&previewOut)
	preview.SetErr(&previewOut)
	preview.SetArgs([]string{"insert", "3", "_tag", "--dry-run", "--path", tmp})

	if err := preview.Execute(); err != nil {
		t.Fatalf("preview command failed: %v\noutput: %s", err, previewOut.String())
	}

	if !contains(t, previewOut.String(), "hol_tagiday.jpg", "tri_tagp.jpg") {
		t.Fatalf("preview output missing expected inserts: %s", previewOut.String())
	}

	var applyOut bytes.Buffer
	apply := renamercmd.NewRootCommand()
	apply.SetOut(&applyOut)
	apply.SetErr(&applyOut)
	apply.SetArgs([]string{"insert", "3", "_tag", "--yes", "--path", tmp})

	if err := apply.Execute(); err != nil {
		t.Fatalf("apply command failed: %v\noutput: %s", err, applyOut.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "hol_tagiday.jpg")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "tri_tagp.jpg")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}
}

func TestInsertCommandTailOffsetToken(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	createInsertFile(t, filepath.Join(tmp, "demo.txt"))

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"insert", "1$", "_TAIL", "--yes", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("insert command failed: %v\noutput: %s", err, out.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "dem_TAILo.txt")); err != nil {
		t.Fatalf("expected dem_TAILo.txt after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "demo.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected original demo.txt to be renamed, err=%v", err)
	}
}

func contains(t *testing.T, haystack string, expected ...string) bool {
	t.Helper()
	for _, s := range expected {
		if !bytes.Contains([]byte(haystack), []byte(s)) {
			return false
		}
	}
	return true
}

func createInsertFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
