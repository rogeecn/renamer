package contract

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogeecn/renamer/internal/extension"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestExtensionPreviewAndApply(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeTestFile(t, filepath.Join(tmp, "photo.jpeg"))
	writeTestFile(t, filepath.Join(tmp, "banner.JPG"))
	writeTestFile(t, filepath.Join(tmp, "logo.jpg"))
	writeTestFile(t, filepath.Join(tmp, "notes.txt"))

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

	req := extension.NewRequest(scope)
	req.SetExecutionMode(true, false)

	sources := []string{".jpeg", ".JPG", ".jpg"}
	canonical, display, duplicates := extension.NormalizeSourceExtensions(sources)
	target := extension.NormalizeTargetExtension(".jpg")
	targetCanonical := extension.CanonicalExtension(target)

	filteredCanonical := make([]string, 0, len(canonical))
	filteredDisplay := make([]string, 0, len(display))
	noOps := make([]string, 0)
	for i, canon := range canonical {
		if canon == targetCanonical {
			noOps = append(noOps, display[i])
			continue
		}
		filteredCanonical = append(filteredCanonical, canon)
		filteredDisplay = append(filteredDisplay, display[i])
	}

	if len(filteredCanonical) == 0 {
		t.Fatalf("expected canonical sources after filtering")
	}

	req.SetExtensions(filteredCanonical, filteredDisplay, target)
	req.SetWarnings(duplicates, noOps)

	var buf bytes.Buffer
	summary, planned, err := extension.Preview(context.Background(), req, &buf)
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}

	if summary.TotalCandidates != 3 {
		t.Fatalf("expected 3 candidates, got %d", summary.TotalCandidates)
	}
	if summary.TotalChanged != 2 {
		t.Fatalf("expected 2 changed entries, got %d", summary.TotalChanged)
	}
	if summary.NoChange != 1 {
		t.Fatalf("expected 1 no-change entry, got %d", summary.NoChange)
	}
	if len(planned) != 2 {
		t.Fatalf("expected 2 planned renames, got %d", len(planned))
	}

	output := buf.String()
	if !strings.Contains(output, "photo.jpeg -> photo.jpg") {
		t.Fatalf("expected preview to include photo rename, output: %s", output)
	}
	if !strings.Contains(output, "banner.JPG -> banner.jpg") {
		t.Fatalf("expected preview to include banner rename, output: %s", output)
	}
	if !strings.Contains(output, "logo.jpg (no change)") {
		t.Fatalf("expected preview to mark logo as no change, output: %s", output)
	}
	if !strings.Contains(output, "Summary: 3 candidates, 2 will change, 1 already target extension") {
		t.Fatalf("expected summary line, output: %s", output)
	}

	if len(summary.Warnings) == 0 {
		t.Fatalf("expected warnings for duplicates/no-ops")
	}

	req.SetExecutionMode(false, true)
	entry, err := extension.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if len(entry.Operations) != len(planned) {
		t.Fatalf("expected %d ledger operations, got %d", len(planned), len(entry.Operations))
	}

	if _, err := os.Stat(filepath.Join(tmp, "photo.jpg")); err != nil {
		t.Fatalf("expected photo.jpg after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "banner.jpg")); err != nil {
		t.Fatalf("expected banner.jpg after apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "logo.jpg")); err != nil {
		t.Fatalf("expected logo.jpg to remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "photo.jpeg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected photo.jpeg to be renamed, err=%v", err)
	}

	ledger := filepath.Join(tmp, ".renamer")
	if _, err := os.Stat(ledger); err != nil {
		t.Fatalf("expected ledger file to be created: %v", err)
	}
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
