package output

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// tableFormatter renders aligned columns for human-friendly review.
type tableFormatter struct {
	writer *tabwriter.Writer
}

// NewTableFormatter constructs a table formatter.
func NewTableFormatter() Formatter {
	return &tableFormatter{}
}

func (f *tableFormatter) Begin(w io.Writer) error {
	f.writer = tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	_, err := fmt.Fprintln(f.writer, "PATH\tTYPE\tSIZE")
	return err
}

func (f *tableFormatter) WriteEntry(w io.Writer, entry Entry) error {
	if f.writer == nil {
		return fmt.Errorf("table formatter not initialized")
	}

	size := "-"
	if entry.Type == "file" && entry.SizeBytes >= 0 {
		size = fmt.Sprintf("%d", entry.SizeBytes)
	}

	_, err := fmt.Fprintf(f.writer, "%s\t%s\t%s\n", entry.Path, entry.Type, size)
	return err
}

func (f *tableFormatter) WriteSummary(w io.Writer, summary Summary) error {
	if f.writer == nil {
		return fmt.Errorf("table formatter not initialized")
	}
	if err := f.writer.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, DefaultSummaryLine(summary))
	return err
}
