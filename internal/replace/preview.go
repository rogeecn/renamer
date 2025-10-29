package replace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// PlannedOperation represents a rename that will be executed during apply.
type PlannedOperation struct {
	Result         Result
	TargetRelative string
	TargetAbsolute string
}

// Preview computes replacements and writes a human-readable summary to out.
func Preview(ctx context.Context, req *ReplaceRequest, parseResult ParseArgsResult, out io.Writer) (Summary, []PlannedOperation, error) {
	summary := NewSummary()
	for _, dup := range parseResult.Duplicates {
		summary.AddDuplicate(dup)
	}

	planned := make([]PlannedOperation, 0)
	plannedTargets := make(map[string]string) // target rel -> source rel to detect duplicates

	err := TraverseCandidates(ctx, req, func(candidate Candidate) error {
		res := ApplyPatterns(candidate, parseResult.Patterns, parseResult.Replacement)
		summary.RecordCandidate(res)

		if !res.Changed {
			return nil
		}

		dir := filepath.Dir(candidate.RelativePath)
		if dir == "." {
			dir = ""
		}

		targetRelative := res.ProposedName
		if dir != "" {
			targetRelative = filepath.ToSlash(filepath.Join(dir, res.ProposedName))
		} else {
			targetRelative = filepath.ToSlash(res.ProposedName)
		}

		if targetRelative == candidate.RelativePath {
			return nil
		}

		if existing, ok := plannedTargets[targetRelative]; ok && existing != candidate.RelativePath {
			summary.AddConflict(ConflictDetail{
				OriginalPath: candidate.RelativePath,
				ProposedPath: targetRelative,
				Reason:       "duplicate target generated in preview",
			})
			return nil
		}

		targetAbsolute := filepath.Join(req.WorkingDir, filepath.FromSlash(targetRelative))
		if info, err := os.Stat(targetAbsolute); err == nil {
			if candidate.OriginalPath != targetAbsolute {
				// Case-only renames are allowed on case-insensitive filesystems. Compare file identity.
				if origInfo, origErr := os.Stat(candidate.OriginalPath); origErr == nil {
					if os.SameFile(info, origInfo) {
						// Same file—case-only update permitted.
						goto recordOperation
					}
				}

				reason := "target already exists"
				if info.IsDir() {
					reason = "target directory already exists"
				}
				summary.AddConflict(ConflictDetail{
					OriginalPath: candidate.RelativePath,
					ProposedPath: targetRelative,
					Reason:       reason,
				})
				return nil
			}
		}

		plannedTargets[targetRelative] = candidate.RelativePath

		if out != nil {
			fmt.Fprintf(out, "%s -> %s\n", candidate.RelativePath, targetRelative)
		}

	recordOperation:
		planned = append(planned, PlannedOperation{
			Result:         res,
			TargetRelative: targetRelative,
			TargetAbsolute: targetAbsolute,
		})

		return nil
	})
	if err != nil {
		return Summary{}, nil, err
	}

	if summary.ReplacementWasEmpty(parseResult.Replacement) {
		if out != nil {
			fmt.Fprintln(out, "Warning: replacement string is empty; matched patterns will be removed.")
		}
	}

	return summary, planned, nil
}
