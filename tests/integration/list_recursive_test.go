package integration

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/output"
)

type captureFormatter struct {
	paths []string
}

func (f *captureFormatter) Begin(io.Writer) error { return nil }

func (f *captureFormatter) WriteEntry(_ io.Writer, entry output.Entry) error {
	f.paths = append(f.paths, entry.Path)
	return nil
}

func (f *captureFormatter) WriteSummary(io.Writer, output.Summary) error { return nil }

func TestListServiceRecursiveTraversal(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "root.txt"))
	mustWriteFile(t, filepath.Join(tmp, "nested", "child.txt"))
	mustWriteDir(t, filepath.Join(tmp, "nested", "inner"))

	svc := listing.NewService()
	req := &listing.ListingRequest{
		WorkingDir: tmp,
		Recursive:  true,
		Format:     listing.FormatPlain,
	}

	formatter := &captureFormatter{}
	summary, err := svc.List(context.Background(), req, formatter, io.Discard)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	sort.Strings(formatter.paths)
	expected := []string{"nested/child.txt", "root.txt"}
	if len(formatter.paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d (%v)", len(expected), len(formatter.paths), formatter.paths)
	}
	for i, path := range expected {
		if formatter.paths[i] != path {
			t.Fatalf("expected path %q at index %d, got %q", path, i, formatter.paths[i])
		}
	}

	if summary.Total() != len(expected) {
		t.Fatalf("unexpected summary total: %d", summary.Total())
	}
}

func TestListServiceDirectoryOnlyMode(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "file.txt"))
	mustWriteDir(t, filepath.Join(tmp, "folder"))

	svc := listing.NewService()
	req := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: true,
		Format:             listing.FormatPlain,
	}

	formatter := &captureFormatter{}
	_, err := svc.List(context.Background(), req, formatter, io.Discard)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if len(formatter.paths) != 1 || formatter.paths[0] != "folder" {
		t.Fatalf("expected only directory entry, got %v", formatter.paths)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func mustWriteDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir dir %s: %v", path, err)
	}
}
