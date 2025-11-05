package ai

import (
	"fmt"
	"io"

	"github.com/rogeecn/renamer/internal/ai/flow"
	"github.com/rogeecn/renamer/internal/output"
)

// PrintPreview renders suggestions in a tabular format alongside validation results.
func PrintPreview(w io.Writer, suggestions []flow.Suggestion, validation ValidationResult) error {
	table := output.NewAIPlanTable()
	if err := table.Begin(w); err != nil {
		return err
	}

	for idx, suggestion := range suggestions {
		if err := table.WriteRow(output.AIPlanRow{
			Sequence:  fmt.Sprintf("%02d", idx+1),
			Original:  suggestion.Original,
			Proposed:  suggestion.Suggested,
			Sanitized: flowToKey(suggestion.Suggested),
		}); err != nil {
			return err
		}
	}

	if err := table.End(w); err != nil {
		return err
	}

	for _, warn := range validation.Warnings {
		if _, err := fmt.Fprintf(w, "Warning: %s\n", warn); err != nil {
			return err
		}
	}

	if len(validation.Conflicts) > 0 {
		if _, err := fmt.Fprintln(w, "Conflicts detected:"); err != nil {
			return err
		}
		for _, conflict := range validation.Conflicts {
			if _, err := fmt.Fprintf(w, " - %s -> %s (%s)\n", conflict.Original, conflict.Suggested, conflict.Reason); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(w, "Previewed %d suggestion(s)\n", len(suggestions)); err != nil {
		return err
	}

	return nil
}
