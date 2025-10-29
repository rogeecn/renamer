package output

import (
	"fmt"
	"io"
)

// Formatter renders listing entries in a chosen representation.
type Formatter interface {
	Begin(w io.Writer) error
	WriteEntry(w io.Writer, entry Entry) error
	WriteSummary(w io.Writer, summary Summary) error
}

// Entry represents a single listing output record in a formatter-agnostic form.
type Entry struct {
	Path             string
	Type             string
	SizeBytes        int64
	Depth            int
	MatchedExtension string
}

// Summary aggregates counts for final reporting.
type Summary struct {
	Files       int
	Directories int
	Symlinks    int
}

// Add records a new entry in the summary counters.
func (s *Summary) Add(entry Entry) {
	switch entry.Type {
	case "file":
		s.Files++
	case "directory":
		s.Directories++
	case "symlink":
		s.Symlinks++
	}
}

// Total returns the sum of all entry classifications.
func (s Summary) Total() int {
	return s.Files + s.Directories + s.Symlinks
}

// DefaultSummaryLine produces a human-readable summary string for any format.
func DefaultSummaryLine(summary Summary) string {
	return fmt.Sprintf("Total: %d entries (files: %d, directories: %d, symlinks: %d)",
		summary.Total(), summary.Files, summary.Directories, summary.Symlinks)
}
