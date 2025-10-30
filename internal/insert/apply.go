package insert

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/rogeecn/renamer/internal/history"
)

// Apply performs planned insert operations and records them in the ledger.
func Apply(ctx context.Context, req *Request, planned []PlannedOperation, summary *Summary) (history.Entry, error) {
	entry := history.Entry{Command: "insert"}

	if len(planned) == 0 {
		return entry, nil
	}

	sort.SliceStable(planned, func(i, j int) bool {
		return planned[i].Depth > planned[j].Depth
	})

	done := make([]history.Operation, 0, len(planned))

	revert := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			op := done[i]
			source := filepath.Join(req.WorkingDir, filepath.FromSlash(op.To))
			destination := filepath.Join(req.WorkingDir, filepath.FromSlash(op.From))
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

		if op.OriginalAbsolute == op.ProposedAbsolute {
			continue
		}

		if err := os.Rename(op.OriginalAbsolute, op.ProposedAbsolute); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		done = append(done, history.Operation{
			From: op.OriginalRelative,
			To:   op.ProposedRelative,
		})
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done
	if summary != nil {
		meta := make(map[string]any, len(summary.LedgerMetadata))
		for k, v := range summary.LedgerMetadata {
			meta[k] = v
		}
		meta["totalCandidates"] = summary.TotalCandidates
		meta["totalChanged"] = summary.TotalChanged
		meta["noChange"] = summary.NoChange
		if len(summary.Warnings) > 0 {
			meta["warnings"] = append([]string(nil), summary.Warnings...)
		}
		entry.Metadata = meta
	}

	if err := history.Append(req.WorkingDir, entry); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}
