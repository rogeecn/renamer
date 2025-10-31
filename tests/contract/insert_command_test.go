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

func TestInsertTailOffsetToken(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeInsertFile(t, filepath.Join(tmp, "code.txt"))

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
	req.SetPositionAndText("1$", "_TAIL")

	summary, planned, err := insert.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}
	if summary.TotalCandidates != 1 {
		t.Fatalf("expected 1 candidate, got %d", summary.TotalCandidates)
	}
	if summary.TotalChanged != 1 {
		t.Fatalf("expected 1 change, got %d", summary.TotalChanged)
	}
	if len(planned) != 1 {
		t.Fatalf("expected 1 planned operation, got %d", len(planned))
	}
	expected := filepath.ToSlash("cod_TAILe.txt")
	if planned[0].ProposedRelative != expected {
		t.Fatalf("expected proposed path %s, got %s", expected, planned[0].ProposedRelative)
	}

	req.SetExecutionMode(false, true)
	entry, err := insert.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}
	if len(entry.Operations) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(entry.Operations))
	}
	if _, err := os.Stat(filepath.Join(tmp, "cod_TAILe.txt")); err != nil {
		t.Fatalf("expected renamed file cod_TAILe.txt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "code.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected original file to be renamed, err=%v", err)
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
