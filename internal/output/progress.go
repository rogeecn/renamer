package output

import (
	"fmt"
	"io"
)

// ProgressReporter prints textual progress for rename operations.
type ProgressReporter struct {
	writer io.Writer
	total  int
	count  int
}

// NewProgressReporter constructs a reporter bound to the supplied writer.
func NewProgressReporter(w io.Writer, total int) *ProgressReporter {
	if w == nil {
		w = io.Discard
	}
	return &ProgressReporter{writer: w, total: total}
}

// Step registers a completed operation and prints the progress.
func (r *ProgressReporter) Step(from, to string) error {
	if r == nil {
		return nil
	}
	r.count++
	_, err := fmt.Fprintf(r.writer, "[%d/%d] %s -> %s\n", r.count, r.total, from, to)
	return err
}

// Complete emits a summary line after all operations finish.
func (r *ProgressReporter) Complete() error {
	if r == nil {
		return nil
	}
	_, err := fmt.Fprintf(r.writer, "Completed %d rename(s).\n", r.count)
	return err
}
