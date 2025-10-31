package regex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/rogeecn/renamer/internal/history"
)

// Apply executes the planned regex renames and writes a ledger entry.
func Apply(ctx context.Context, req Request, planned []PlannedRename, summary Summary) (history.Entry, error) {
	reqCopy := req
	if err := reqCopy.Validate(); err != nil {
		return history.Entry{}, err
	}

	entry := history.Entry{Command: "regex"}

	if len(planned) == 0 {
		return entry, nil
	}

	sort.SliceStable(planned, func(i, j int) bool {
		return planned[i].Depth > planned[j].Depth
	})

	done := make([]history.Operation, 0, len(planned))
	groupsMeta := make(map[string][]string, len(planned))

	revert := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			op := done[i]
			source := filepath.Join(reqCopy.WorkingDir, filepath.FromSlash(op.To))
			destination := filepath.Join(reqCopy.WorkingDir, filepath.FromSlash(op.From))
			if err := os.Rename(source, destination); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}

	for _, plan := range planned {
		if err := ctx.Err(); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		if err := os.MkdirAll(filepath.Dir(plan.TargetAbsolute), 0o755); err != nil {
			_ = revert()
			return history.Entry{}, fmt.Errorf("prepare target directory: %w", err)
		}

		if err := os.Rename(plan.SourceAbsolute, plan.TargetAbsolute); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		relFrom := filepath.ToSlash(plan.SourceRelative)
		relTo := filepath.ToSlash(plan.TargetRelative)
		done = append(done, history.Operation{From: relFrom, To: relTo})
		if len(plan.MatchGroups) > 0 {
			groupsMeta[relFrom] = append([]string(nil), plan.MatchGroups...)
		}
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done
	metadata := map[string]any{
		"pattern":  reqCopy.Pattern,
		"template": reqCopy.Template,
		"matched":  summary.Matched,
		"changed":  summary.Changed,
	}
	if len(groupsMeta) > 0 {
		metadata["matchGroups"] = groupsMeta
	}
	entry.Metadata = metadata

	if err := history.Append(reqCopy.WorkingDir, entry); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}
