package regex

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PlannedRename represents a proposed rename resulting from preview.
type PlannedRename struct {
	SourceRelative string
	SourceAbsolute string
	TargetRelative string
	TargetAbsolute string
	MatchGroups    []string
	Depth          int
}

// Preview evaluates the regex rename request and returns a summary plus the planned operations.
func Preview(ctx context.Context, req Request, out io.Writer) (Summary, []PlannedRename, error) {
	reqCopy := req
	if err := reqCopy.Validate(); err != nil {
		return Summary{}, nil, err
	}

	engine, err := NewEngine(reqCopy.Pattern, reqCopy.Template)
	if err != nil {
		return Summary{}, nil, err
	}

	summary := Summary{
		LedgerMetadata: map[string]any{
			"pattern":  reqCopy.Pattern,
			"template": reqCopy.Template,
		},
		Entries: make([]PreviewEntry, 0),
	}

	planned := make([]PlannedRename, 0)
	plannedTargets := make(map[string]string)
	plannedTargetsFold := make(map[string]string)

	err = TraverseCandidates(ctx, &reqCopy, func(candidate Candidate) error {
		summary.TotalCandidates++

		rendered, groups, matched, err := engine.Apply(candidate.Stem)
		if err != nil {
			summary.Warnings = append(summary.Warnings, err.Error())
			summary.Skipped++
			summary.Entries = append(summary.Entries, PreviewEntry{
				OriginalPath: candidate.RelativePath,
				ProposedPath: candidate.RelativePath,
				Status:       EntrySkipped,
			})
			return nil
		}

		if !matched {
			summary.Skipped++
			summary.Entries = append(summary.Entries, PreviewEntry{
				OriginalPath: candidate.RelativePath,
				ProposedPath: candidate.RelativePath,
				Status:       EntrySkipped,
			})
			return nil
		}

		summary.Matched++

		proposedName := rendered
		if !candidate.IsDir && candidate.Extension != "" {
			proposedName += candidate.Extension
		}

		dir := filepath.Dir(candidate.RelativePath)
		if dir == "." {
			dir = ""
		}

		var proposedRelative string
		if dir != "" {
			proposedRelative = filepath.ToSlash(filepath.Join(dir, proposedName))
		} else {
			proposedRelative = filepath.ToSlash(proposedName)
		}

		matchEntry := PreviewEntry{
			OriginalPath: candidate.RelativePath,
			ProposedPath: proposedRelative,
			MatchGroups:  groups,
		}

		if proposedRelative == candidate.RelativePath {
			summary.Entries = append(summary.Entries, PreviewEntry{
				OriginalPath: candidate.RelativePath,
				ProposedPath: candidate.RelativePath,
				Status:       EntryNoChange,
				MatchGroups:  groups,
			})
			return nil
		}

		if proposedName == "" || proposedRelative == "" {
			summary.Conflicts = append(summary.Conflicts, Conflict{
				OriginalPath: candidate.RelativePath,
				ProposedPath: proposedRelative,
				Reason:       ConflictInvalidTemplate,
			})
			summary.Skipped++
			matchEntry.Status = EntrySkipped
			summary.Entries = append(summary.Entries, matchEntry)
			return nil
		}

		if existing, ok := plannedTargets[proposedRelative]; ok && existing != candidate.RelativePath {
			summary.Conflicts = append(summary.Conflicts, Conflict{
				OriginalPath: candidate.RelativePath,
				ProposedPath: proposedRelative,
				Reason:       ConflictDuplicateTarget,
			})
			summary.Skipped++
			matchEntry.Status = EntrySkipped
			summary.Entries = append(summary.Entries, matchEntry)
			return nil
		}

		casefoldKey := strings.ToLower(proposedRelative)
		if existing, ok := plannedTargetsFold[casefoldKey]; ok && existing != candidate.RelativePath {
			summary.Conflicts = append(summary.Conflicts, Conflict{
				OriginalPath: candidate.RelativePath,
				ProposedPath: proposedRelative,
				Reason:       ConflictDuplicateTarget,
			})
			summary.Skipped++
			matchEntry.Status = EntrySkipped
			summary.Entries = append(summary.Entries, matchEntry)
			return nil
		}

		targetAbsolute := filepath.Join(reqCopy.WorkingDir, filepath.FromSlash(proposedRelative))
		if info, statErr := os.Stat(targetAbsolute); statErr == nil {
			origInfo, origErr := os.Stat(candidate.OriginalPath)
			if origErr != nil || !os.SameFile(info, origInfo) {
				reason := ConflictExistingFile
				if info.IsDir() {
					reason = ConflictExistingDir
				}
				summary.Conflicts = append(summary.Conflicts, Conflict{
					OriginalPath: candidate.RelativePath,
					ProposedPath: proposedRelative,
					Reason:       reason,
				})
				summary.Skipped++
				matchEntry.Status = EntrySkipped
				summary.Entries = append(summary.Entries, matchEntry)
				return nil
			}
		}

		plannedTargets[proposedRelative] = candidate.RelativePath
		plannedTargetsFold[casefoldKey] = candidate.RelativePath

		matchEntry.Status = EntryChanged
		summary.Entries = append(summary.Entries, matchEntry)
		summary.Changed++

		planned = append(planned, PlannedRename{
			SourceRelative: candidate.RelativePath,
			SourceAbsolute: candidate.OriginalPath,
			TargetRelative: proposedRelative,
			TargetAbsolute: targetAbsolute,
			MatchGroups:    groups,
			Depth:          candidate.Depth,
		})

		if out != nil {
			fmt.Fprintf(out, "%s -> %s\n", candidate.RelativePath, proposedRelative)
		}

		return nil
	})
	if err != nil {
		return Summary{}, nil, err
	}

	return summary, planned, nil
}
