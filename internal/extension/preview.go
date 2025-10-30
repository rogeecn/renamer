package extension

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Preview generates a summary and planned operations for an extension normalization run.
func Preview(ctx context.Context, req *ExtensionRequest, out io.Writer) (*ExtensionSummary, []PlannedRename, error) {
	if req == nil {
		return nil, nil, errors.New("extension request cannot be nil")
	}

	plan, err := BuildPlan(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	summary := plan.Summary

	for _, dup := range req.DuplicateSources {
		summary.AddWarning(fmt.Sprintf("duplicate source extension ignored: %s", dup))
	}
	for _, noop := range req.NoOpSources {
		summary.AddWarning(fmt.Sprintf("source extension matches target and is skipped: %s", noop))
	}

	if summary.TotalCandidates == 0 {
		summary.AddWarning("no candidates found for provided extensions")
	}

	if out != nil {
		// Ensure deterministic ordering for preview output.
		sort.SliceStable(summary.Entries, func(i, j int) bool {
			if summary.Entries[i].OriginalPath == summary.Entries[j].OriginalPath {
				return summary.Entries[i].ProposedPath < summary.Entries[j].ProposedPath
			}
			return summary.Entries[i].OriginalPath < summary.Entries[j].OriginalPath
		})

		conflictReasons := make(map[string]string, len(summary.Conflicts))
		for _, conflict := range summary.Conflicts {
			key := fmt.Sprintf("%s->%s", conflict.OriginalPath, conflict.ProposedPath)
			conflictReasons[key] = conflict.Reason
		}

		for _, entry := range summary.Entries {
			switch entry.Status {
			case PreviewStatusChanged:
				if entry.ProposedPath == entry.OriginalPath {
					fmt.Fprintf(out, "%s (pending extension update)\n", entry.OriginalPath)
				} else {
					fmt.Fprintf(out, "%s -> %s\n", entry.OriginalPath, entry.ProposedPath)
				}
			case PreviewStatusNoChange:
				fmt.Fprintf(out, "%s (no change)\n", entry.OriginalPath)
			case PreviewStatusSkipped:
				reason := conflictReasons[fmt.Sprintf("%s->%s", entry.OriginalPath, entry.ProposedPath)]
				if reason == "" {
					reason = "skipped"
				}
				fmt.Fprintf(out, "%s -> %s (skipped: %s)\n", entry.OriginalPath, entry.ProposedPath, reason)
			default:
				fmt.Fprintf(out, "%s (status: %s)\n", entry.OriginalPath, entry.Status)
			}
		}

		if summary.TotalCandidates > 0 {
			fmt.Fprintf(out, "\nSummary: %d candidates, %d will change, %d already target extension\n",
				summary.TotalCandidates, summary.TotalChanged, summary.NoChange)
		} else {
			fmt.Fprintln(out, "No candidates found.")
		}

		if len(summary.Warnings) > 0 {
			fmt.Fprintln(out)
			for _, warning := range summary.Warnings {
				if !strings.HasPrefix(strings.ToLower(warning), "warning:") {
					fmt.Fprintf(out, "Warning: %s\n", warning)
				} else {
					fmt.Fprintln(out, warning)
				}
			}
		}
	}

	return summary, plan.Operations, nil
}
