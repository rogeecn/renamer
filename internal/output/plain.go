package output

import (
	"fmt"
	"io"
)

// plainFormatter emits one entry per line suitable for piping into other tools.
type plainFormatter struct{}

// NewPlainFormatter constructs a formatter for plain output.
func NewPlainFormatter() Formatter {
	return &plainFormatter{}
}

func (plainFormatter) Begin(io.Writer) error {
	return nil
}

func (plainFormatter) WriteEntry(w io.Writer, entry Entry) error {
	_, err := fmt.Fprintln(w, entry.Path)
	return err
}

func (plainFormatter) WriteSummary(w io.Writer, summary Summary) error {
	_, err := fmt.Fprintln(w, DefaultSummaryLine(summary))
	return err
}
