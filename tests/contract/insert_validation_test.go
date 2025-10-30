package contract

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/insert"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestInsertRejectsOutOfRangePositions(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeInsertValidationFile(t, filepath.Join(tmp, "短.txt"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := insert.NewRequest(scope)
	req.SetExecutionMode(true, false)
	req.SetPositionAndText("50", "X")

	if _, _, err := insert.Preview(context.Background(), req, nil); err == nil {
		t.Fatalf("expected error for out-of-range position")
	}
}

func TestInsertBlocksExistingTargetConflicts(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeInsertValidationFile(t, filepath.Join(tmp, "report.txt"))
	writeInsertValidationFile(t, filepath.Join(tmp, "report_ARCHIVE.txt"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := insert.NewRequest(scope)
	req.SetExecutionMode(true, false)
	req.SetPositionAndText("$", "_ARCHIVE")

	summary, _, err := insert.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}
	if !summary.HasConflicts() {
		t.Fatalf("expected conflicts to be detected")
	}
}

func writeInsertValidationFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("validation"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
