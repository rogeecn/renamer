package output

import (
	"fmt"
	"io"
	"strings"
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

// WriteAIPlanDebug emits prompt hashes and warnings to the provided writer.
func WriteAIPlanDebug(w io.Writer, promptHash string, warnings []string) {
	if w == nil {
		return
	}
	if promptHash != "" {
		fmt.Fprintf(w, "Prompt hash: %s\n", promptHash)
	}
	for _, warning := range warnings {
		if strings.TrimSpace(warning) == "" {
			continue
		}
		fmt.Fprintf(w, "%s\n", warning)
	}
}

// PolicyViolationMessage describes a single policy failure for display purposes.
type PolicyViolationMessage struct {
	Original string
	Proposed string
	Rule     string
	Message  string
}

// WritePolicyViolations prints detailed policy failure information to the writer.
func WritePolicyViolations(w io.Writer, violations []PolicyViolationMessage) {
	if w == nil {
		return
	}
	for _, violation := range violations {
		rule := violation.Rule
		if rule == "" {
			rule = "policy"
		}
		fmt.Fprintf(w, "Policy violation (%s): %s -> %s (%s)\n", rule, violation.Original, violation.Proposed, violation.Message)
	}
}
