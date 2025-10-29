package replace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/rogeecn/renamer/internal/history"
)

// Apply executes the planned operations and records them in the ledger.
func Apply(ctx context.Context, req *ReplaceRequest, planned []PlannedOperation, summary Summary) (history.Entry, error) {
	entry := history.Entry{Command: "replace"}

	if len(planned) == 0 {
		return entry, nil
	}

	sort.SliceStable(planned, func(i, j int) bool {
		return planned[i].Result.Candidate.Depth > planned[j].Result.Candidate.Depth
	})

	done := make([]history.Operation, 0, len(planned))

	revert := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			op := done[i]
			source := filepath.Join(req.WorkingDir, op.To)
			destination := filepath.Join(req.WorkingDir, op.From)
			if err := os.Rename(source, destination); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}

	for _, op := range planned {
		if err := ctx.Err(); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		from := op.Result.Candidate.OriginalPath
		to := op.TargetAbsolute

		if from == to {
			continue
		}

		if err := os.Rename(from, to); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		done = append(done, history.Operation{
			From: filepath.ToSlash(op.Result.Candidate.RelativePath),
			To:   filepath.ToSlash(op.TargetRelative),
		})
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done

	metadataPatterns := make(map[string]int, len(summary.PatternMatches))
	for pattern, count := range summary.PatternMatches {
		metadataPatterns[pattern] = count
	}
	entry.Metadata = map[string]any{
		"patterns":        metadataPatterns,
		"changed":         summary.ChangedCount,
		"totalCandidates": summary.TotalCandidates,
	}

	if err := history.Append(req.WorkingDir, entry); err != nil {
		// Attempt to undo renames if ledger append fails.
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}
