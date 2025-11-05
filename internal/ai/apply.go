package ai

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/rogeecn/renamer/internal/ai/flow"
	"github.com/rogeecn/renamer/internal/history"
	"github.com/rogeecn/renamer/internal/output"
)

// ApplyMetadata captures contextual information persisted alongside ledger entries.
type ApplyMetadata struct {
	Prompt            string
	PromptHistory     []string
	Notes             []string
	Model             string
	SequenceSeparator string
}

// toMap converts metadata into a ledger-friendly map.
func (m ApplyMetadata) toMap(warnings []string) map[string]any {
	data := history.BuildAIMetadata(m.Prompt, m.PromptHistory, m.Notes, m.Model, warnings)
	if m.SequenceSeparator != "" {
		data["sequenceSeparator"] = m.SequenceSeparator
	}
	return data
}

// Apply executes the rename suggestions, records a ledger entry, and emits progress updates.
func Apply(ctx context.Context, workingDir string, suggestions []flow.Suggestion, validation ValidationResult, meta ApplyMetadata, writer io.Writer) (history.Entry, error) {
	entry := history.Entry{Command: "ai"}

	if len(suggestions) == 0 {
		return entry, nil
	}

	reporter := output.NewProgressReporter(writer, len(suggestions))

	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Original > suggestions[j].Original
	})

	operations := make([]history.Operation, 0, len(suggestions))

	revert := func() error {
		for i := len(operations) - 1; i >= 0; i-- {
			op := operations[i]
			source := filepath.Join(workingDir, filepath.FromSlash(op.To))
			destination := filepath.Join(workingDir, filepath.FromSlash(op.From))
			if err := os.Rename(source, destination); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}

	for _, suggestion := range suggestions {
		if err := ctx.Err(); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		fromRel := flowToKey(suggestion.Original)
		toRel := flowToKey(suggestion.Suggested)

		fromAbs := filepath.Join(workingDir, filepath.FromSlash(fromRel))
		toAbs := filepath.Join(workingDir, filepath.FromSlash(toRel))

		if fromAbs == toAbs {
			continue
		}

		if err := ensureParentDir(toAbs); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		if err := os.Rename(fromAbs, toAbs); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		operations = append(operations, history.Operation{From: fromRel, To: toRel})
		if err := reporter.Step(fromRel, toRel); err != nil {
			_ = revert()
			return history.Entry{}, err
		}
	}

	if len(operations) == 0 {
		return entry, reporter.Complete()
	}

	if err := reporter.Complete(); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	entry.Operations = operations
	entry.Metadata = meta.toMap(validation.Warnings)

	if err := history.Append(workingDir, entry); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
