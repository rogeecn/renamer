package listing

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// FormatTable renders output in a column-aligned table for human review.
	FormatTable = "table"
	// FormatPlain renders newline-delimited paths for scripting.
	FormatPlain = "plain"
)

// ListingRequest captures scope and formatting preferences for a listing run.
type ListingRequest struct {
	WorkingDir         string
	IncludeDirectories bool
	Recursive          bool
	IncludeHidden      bool
	Extensions         []string
	Format             string
	MaxDepth           int
}

// ListingEntry represents a single filesystem node discovered during traversal.
type ListingEntry struct {
	Path             string
	Type             EntryType
	SizeBytes        int64
	Depth            int
	MatchedExtension string
}

// EntryType captures the classification of a filesystem node.
type EntryType string

const (
	EntryTypeFile    EntryType = "file"
	EntryTypeDir     EntryType = "directory"
	EntryTypeSymlink EntryType = "symlink"
)

// Validate ensures the request is well-formed before execution.
func (r *ListingRequest) Validate() error {
	if r == nil {
		return errors.New("listing request cannot be nil")
	}

	if r.WorkingDir == "" {
		return errors.New("working directory must be provided")
	}

	if !filepath.IsAbs(r.WorkingDir) {
		abs, err := filepath.Abs(r.WorkingDir)
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		r.WorkingDir = abs
	}

	if r.MaxDepth < 0 {
		return errors.New("max depth cannot be negative")
	}

	switch r.Format {
	case "":
		r.Format = FormatTable
	case FormatTable, FormatPlain:
		// ok
	default:
		return fmt.Errorf("unsupported format %q", r.Format)
	}

	seen := make(map[string]struct{})
	filtered := r.Extensions[:0]
	for _, ext := range r.Extensions {
		trimmed := strings.TrimSpace(ext)
		if trimmed == "" {
			return errors.New("extensions cannot include empty values")
		}
		if !strings.HasPrefix(trimmed, ".") {
			return fmt.Errorf("extension %q must start with '.'", ext)
		}
		lower := strings.ToLower(trimmed)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		filtered = append(filtered, lower)
	}
	r.Extensions = filtered

	return nil
}
