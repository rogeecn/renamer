package contract

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rogeecn/renamer/internal/insert"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestInsertPreviewAndApply(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeInsertFile(t, filepath.Join(tmp, "项目A报告.docx"))
	writeInsertFile(t, filepath.Join(tmp, "项目B报告.docx"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Extensions:         nil,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := insert.NewRequest(scope)
	req.SetExecutionMode(true, false)
	req.SetPositionAndText("^", "2025-")

	var buf bytes.Buffer
	summary, planned, err := insert.Preview(context.Background(), req, &buf)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	if summary.TotalCandidates != 2 {
		t.Fatalf("expected 2 candidates, got %d", summary.TotalCandidates)
	}
	if summary.TotalChanged != 2 {
		t.Fatalf("expected 2 changes, got %d", summary.TotalChanged)
	}

	output := buf.String()
	if !containsAll(output, "2025-项目A报告.docx", "2025-项目B报告.docx") {
		t.Fatalf("preview output missing expected names: %s", output)
	}

	req.SetExecutionMode(false, true)
	entry, err := insert.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}

	if len(entry.Operations) != 2 {
		t.Fatalf("expected 2 ledger operations, got %d", len(entry.Operations))
	}

	if _, err := os.Stat(filepath.Join(tmp, "项目A报告.docx")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected original name to be renamed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "2025-项目A报告.docx")); err != nil {
		t.Fatalf("expected renamed file: %v", err)
	}
}

func writeInsertFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, n := range needles {
		if !bytes.Contains([]byte(haystack), []byte(n)) {
			return false
		}
	}
	return true
}
