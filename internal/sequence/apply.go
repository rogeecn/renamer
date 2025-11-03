package sequence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogeecn/renamer/internal/history"
)

// Apply executes the planned numbering operations and records them in the ledger.
func Apply(ctx context.Context, opts Options, plan Plan) (history.Entry, error) {
	merged := mergeOptions(opts)
	if err := validateOptions(&merged); err != nil {
		return history.Entry{}, err
	}

	entry := history.Entry{Command: "sequence"}

	type renameOp struct {
		fromAbs string
		toAbs   string
		fromRel string
		toRel   string
		depth   int
	}

	ops := make([]renameOp, 0, len(plan.Candidates))
	for _, candidate := range plan.Candidates {
		if candidate.Status != CandidatePending {
			continue
		}
		ops = append(ops, renameOp{
			fromAbs: filepath.Join(merged.WorkingDir, filepath.FromSlash(candidate.OriginalPath)),
			toAbs:   filepath.Join(merged.WorkingDir, filepath.FromSlash(candidate.ProposedPath)),
			fromRel: candidate.OriginalPath,
			toRel:   candidate.ProposedPath,
			depth:   strings.Count(candidate.OriginalPath, "/"),
		})
	}

	if len(ops) == 0 {
		return entry, nil
	}

	sort.SliceStable(ops, func(i, j int) bool {
		return ops[i].depth > ops[j].depth
	})

	done := make([]history.Operation, 0, len(ops))

	revert := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			op := done[i]
			source := filepath.Join(merged.WorkingDir, filepath.FromSlash(op.To))
			destination := filepath.Join(merged.WorkingDir, filepath.FromSlash(op.From))
			if err := os.Rename(source, destination); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}

	for _, op := range ops {
		if err := ctx.Err(); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		if op.fromAbs == op.toAbs {
			continue
		}

		if err := os.Rename(op.fromAbs, op.toAbs); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		done = append(done, history.Operation{
			From: filepath.ToSlash(op.fromRel),
			To:   filepath.ToSlash(op.toRel),
		})
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done
	entry.Metadata = map[string]any{
		"sequence": map[string]any{
			"start":     plan.Config.Start,
			"width":     plan.Summary.AppliedWidth,
			"placement": string(plan.Config.Placement),
			"separator": plan.Config.Separator,
			"prefix":    plan.Config.NumberPrefix,
			"suffix":    plan.Config.NumberSuffix,
		},
		"totalCandidates": plan.Summary.TotalCandidates,
		"renamed":         plan.Summary.RenamedCount,
		"skipped":         plan.Summary.SkippedCount,
	}

	if err := history.Append(merged.WorkingDir, entry); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}
