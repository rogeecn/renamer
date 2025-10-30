package extension

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/rogeecn/renamer/internal/history"
)

// Apply executes planned renames and records the operations in the ledger.
func Apply(ctx context.Context, req *ExtensionRequest, planned []PlannedRename, summary *ExtensionSummary) (history.Entry, error) {
	entry := history.Entry{Command: "extension"}

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
			From: filepath.ToSlash(op.OriginalRelative),
			To:   filepath.ToSlash(op.ProposedRelative),
		})
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done

	if summary != nil {
		meta := make(map[string]any)
		if len(req.DisplaySourceExtensions) > 0 {
			meta["sourceExtensions"] = append([]string(nil), req.DisplaySourceExtensions...)
		}
		if req.TargetExtension != "" {
			meta["targetExtension"] = req.TargetExtension
		}
		meta["totalCandidates"] = summary.TotalCandidates
		meta["totalChanged"] = summary.TotalChanged
		meta["noChange"] = summary.NoChange

		if len(summary.PerExtensionCounts) > 0 {
			counts := make(map[string]int, len(summary.PerExtensionCounts))
			for ext, count := range summary.PerExtensionCounts {
				counts[ext] = count
			}
			meta["perExtensionCounts"] = counts
		}

		scope := map[string]any{
			"includeDirs":   req.IncludeDirs,
			"recursive":     req.Recursive,
			"includeHidden": req.IncludeHidden,
		}
		if len(req.ExtensionFilter) > 0 {
			scope["extensionFilter"] = append([]string(nil), req.ExtensionFilter...)
		}
		meta["scope"] = scope

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
