package insert

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
)

// Preview computes planned insert operations, renders preview output, and returns the summary.
func Preview(ctx context.Context, req *Request, out io.Writer) (*Summary, []PlannedOperation, error) {
	if req == nil {
		return nil, nil, errors.New("insert request cannot be nil")
	}

	summary, operations, err := BuildPlan(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	summary.LedgerMetadata["positionToken"] = req.PositionToken
	summary.LedgerMetadata["insertText"] = req.InsertText
	scope := map[string]any{
		"includeDirs":   req.IncludeDirs,
		"recursive":     req.Recursive,
		"includeHidden": req.IncludeHidden,
	}
	if len(req.ExtensionFilter) > 0 {
		scope["extensionFilter"] = append([]string(nil), req.ExtensionFilter...)
	}
	summary.LedgerMetadata["scope"] = scope

	if out != nil {
		conflictReasons := make(map[string]string, len(summary.Conflicts))
		for _, conflict := range summary.Conflicts {
			key := conflict.OriginalPath + "->" + conflict.ProposedPath
			conflictReasons[key] = conflict.Reason
		}

		entries := append([]PreviewEntry(nil), summary.Entries...)
		sort.SliceStable(entries, func(i, j int) bool {
			return entries[i].OriginalPath < entries[j].OriginalPath
		})

		for _, entry := range entries {
			switch entry.Status {
			case StatusChanged:
				fmt.Fprintf(out, "%s -> %s\n", entry.OriginalPath, entry.ProposedPath)
			case StatusNoChange:
				fmt.Fprintf(out, "%s (no change)\n", entry.OriginalPath)
			case StatusSkipped:
				reason := conflictReasons[entry.OriginalPath+"->"+entry.ProposedPath]
				if reason == "" {
					reason = "skipped"
				}
				fmt.Fprintf(out, "%s -> %s (skipped: %s)\n", entry.OriginalPath, entry.ProposedPath, reason)
			}
		}

		if summary.TotalCandidates > 0 {
			fmt.Fprintf(out, "\nSummary: %d candidates, %d will change, %d already target position\n",
				summary.TotalCandidates, summary.TotalChanged, summary.NoChange)
		} else {
			fmt.Fprintln(out, "No candidates found.")
		}

		if len(summary.Warnings) > 0 {
			fmt.Fprintln(out)
			for _, warning := range summary.Warnings {
				fmt.Fprintf(out, "Warning: %s\n", warning)
			}
		}
	}

	return summary, operations, nil
}
