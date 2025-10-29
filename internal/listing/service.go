package listing

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/output"
	"github.com/rogeecn/renamer/internal/traversal"
)

// walker abstracts traversal implementation for easier testing.
type walker interface {
	Walk(root string, recursive bool, includeDirs bool, includeHidden bool, maxDepth int, fn func(relPath string, entry fs.DirEntry, depth int) error) error
}

// Service orchestrates filesystem traversal, filtering, and output formatting
// for the read-only `renamer list` command.
type Service struct {
	walker walker
}

// Option configures optional dependencies for the Service.
type Option func(*Service)

// WithWalker provides a custom traversal walker (useful for tests).
func WithWalker(w walker) Option {
	return func(s *Service) {
		s.walker = w
	}
}

// NewService initializes a listing Service with default dependencies.
func NewService(opts ...Option) *Service {
	service := &Service{
		walker: traversal.NewWalker(),
	}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

// List executes a listing request, writing formatted output to sink while
// returning the computed summary for downstream consumers.
func (s *Service) List(ctx context.Context, req *ListingRequest, formatter output.Formatter, sink io.Writer) (output.Summary, error) {
	var summary output.Summary

	if formatter == nil {
		return summary, errors.New("formatter cannot be nil")
	}

	if sink == nil {
		sink = io.Discard
	}

	if err := req.Validate(); err != nil {
		return summary, err
	}

	if err := formatter.Begin(sink); err != nil {
		return summary, err
	}

	extensions := make(map[string]struct{}, len(req.Extensions))
	for _, ext := range req.Extensions {
		extensions[ext] = struct{}{}
	}

	err := s.walker.Walk(
		req.WorkingDir,
		req.Recursive,
		req.IncludeDirectories,
		req.IncludeHidden,
		req.MaxDepth,
		func(relPath string, entry fs.DirEntry, depth int) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			listingEntry, err := s.toListingEntry(req.WorkingDir, relPath, entry, depth)
			if err != nil {
				return err
			}

			if req.IncludeDirectories && listingEntry.Type != EntryTypeDir {
				return nil
			}

			// Apply extension filtering to files only.
			if listingEntry.Type == EntryTypeFile && len(extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if _, match := extensions[ext]; !match {
					return nil
				}
				listingEntry.MatchedExtension = ext
			}

			outEntry := toOutputEntry(listingEntry)

			if err := formatter.WriteEntry(sink, outEntry); err != nil {
				return err
			}
			summary.Add(outEntry)
			return nil
		},
	)
	if err != nil {
		return summary, err
	}

	if err := formatter.WriteSummary(sink, summary); err != nil {
		return summary, err
	}

	return summary, nil
}

func (s *Service) toListingEntry(root, rel string, entry fs.DirEntry, depth int) (ListingEntry, error) {
	fullPath := filepath.Join(root, rel)
	entryType := classifyEntry(entry)

	var size int64
	if entryType == EntryTypeFile {
		info, err := entry.Info()
		if err != nil {
			return ListingEntry{}, err
		}
		size = info.Size()
	}
	if entryType == EntryTypeSymlink {
		info, err := os.Lstat(fullPath)
		if err == nil {
			size = info.Size()
		}
	}

	return ListingEntry{
		Path:      filepath.ToSlash(rel),
		Type:      entryType,
		SizeBytes: size,
		Depth:     depth,
	}, nil
}

func classifyEntry(entry fs.DirEntry) EntryType {
	if entry.Type()&os.ModeSymlink != 0 {
		return EntryTypeSymlink
	}
	if entry.IsDir() {
		return EntryTypeDir
	}
	return EntryTypeFile
}

func toOutputEntry(entry ListingEntry) output.Entry {
	return output.Entry{
		Path:             entry.Path,
		Type:             string(entry.Type),
		SizeBytes:        entry.SizeBytes,
		Depth:            entry.Depth,
		MatchedExtension: entry.MatchedExtension,
	}
}
