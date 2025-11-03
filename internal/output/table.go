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

// AIPlanRow represents a single AI plan preview row.
type AIPlanRow struct {
	Sequence  string
	Original  string
	Proposed  string
	Sanitized string
}

// AIPlanTable renders AI plan previews in a tabular format.
type AIPlanTable struct {
	writer *tabwriter.Writer
}

// NewAIPlanTable constructs a table for AI plan previews.
func NewAIPlanTable() *AIPlanTable {
	return &AIPlanTable{}
}

// Begin writes the header for the AI plan table.
func (t *AIPlanTable) Begin(w io.Writer) error {
	if t.writer != nil {
		return fmt.Errorf("ai plan table already initialized")
	}
	t.writer = tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	_, err := fmt.Fprintln(t.writer, "SEQ\tORIGINAL\tPROPOSED\tSANITIZED")
	return err
}

// WriteRow appends a plan row to the table.
func (t *AIPlanTable) WriteRow(row AIPlanRow) error {
	if t.writer == nil {
		return fmt.Errorf("ai plan table not initialized")
	}
	_, err := fmt.Fprintf(t.writer, "%s\t%s\t%s\t%s\n", row.Sequence, row.Original, row.Proposed, row.Sanitized)
	return err
}

// End flushes the table to the underlying writer.
func (t *AIPlanTable) End(w io.Writer) error {
	if t.writer == nil {
		return fmt.Errorf("ai plan table not initialized")
	}
	if err := t.writer.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w)
	t.writer = nil
	return err
}
