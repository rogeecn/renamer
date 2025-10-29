package contract

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/output"
)

type captureFormatter struct {
	entries []output.Entry
	summary output.Summary
}

func (f *captureFormatter) Begin(io.Writer) error {
	return nil
}

func (f *captureFormatter) WriteEntry(_ io.Writer, entry output.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

func (f *captureFormatter) WriteSummary(_ io.Writer, summary output.Summary) error {
	f.summary = summary
	return nil
}

func TestListServiceFiltersByExtension(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "keep.jpg"))
	mustWriteFile(t, filepath.Join(tmp, "skip.txt"))

	formatter := &captureFormatter{}

	svc := listing.NewService()
	req := &listing.ListingRequest{
		WorkingDir: tmp,
		Extensions: []string{".jpg"},
		Format:     listing.FormatPlain,
	}

	summary, err := svc.List(context.Background(), req, formatter, io.Discard)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(formatter.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(formatter.entries))
	}

	entry := formatter.entries[0]
	if entry.Path != "keep.jpg" {
		t.Fatalf("expected path keep.jpg, got %q", entry.Path)
	}
	if entry.MatchedExtension != ".jpg" {
		t.Fatalf("expected matched extension .jpg, got %q", entry.MatchedExtension)
	}

	if summary.Files != 1 || summary.Total() != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestListServiceFormatParity(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "a.txt"))

	svc := listing.NewService()

	plainReq := &listing.ListingRequest{
		WorkingDir: tmp,
		Format:     listing.FormatPlain,
	}

	plainSummary, err := svc.List(context.Background(), plainReq, output.NewPlainFormatter(), io.Discard)
	if err != nil {
		t.Fatalf("plain list error: %v", err)
	}

	tableReq := &listing.ListingRequest{
		WorkingDir: tmp,
		Format:     listing.FormatTable,
	}

	var buf bytes.Buffer
	tableSummary, err := svc.List(context.Background(), tableReq, output.NewTableFormatter(), &buf)
	if err != nil {
		t.Fatalf("table list error: %v", err)
	}

	if plainSummary.Total() != tableSummary.Total() {
		t.Fatalf("summary total mismatch: plain %d vs table %d", plainSummary.Total(), tableSummary.Total())
	}

	header := buf.String()
	if !strings.Contains(header, "PATH") || !strings.Contains(header, "TYPE") {
		t.Fatalf("expected table header in output, got: %s", header)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := ensureParent(path); err != nil {
		t.Fatalf("ensure parent: %v", err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func ensureParent(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}
