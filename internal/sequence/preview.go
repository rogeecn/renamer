package sequence

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Preview computes the numbering plan for the provided options, returning the
// plan without mutating the filesystem. The writer is reserved for future
// preview output integration and may be nil.
func Preview(ctx context.Context, opts Options, w io.Writer) (Plan, error) {
	merged := mergeOptions(opts)

	if err := validateOptions(&merged); err != nil {
		return Plan{}, err
	}

	traversalCandidates, err := collectTraversalCandidates(ctx, merged)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Candidates: make([]Candidate, 0, len(traversalCandidates)),
		Config: Config{
			Start:        merged.Start,
			Width:        merged.Width,
			Placement:    merged.Placement,
			Separator:    merged.Separator,
			NumberPrefix: merged.NumberPrefix,
			NumberSuffix: merged.NumberSuffix,
		},
	}

	plannedTargets := make(map[string]string)
	plannedTargetsFold := make(map[string]string)

	widthUsed := merged.Width
	widthWarned := false
	nextValue := merged.Start
	sequenceIndex := 0

	for _, entry := range traversalCandidates {
		if entry.IsDir {
			// Directories remain untouched for numbering purposes.
			continue
		}

		plan.Summary.TotalCandidates++

		number, appliedWidth := formatNumber(nextValue, merged.Width)
		if merged.WidthSet && appliedWidth > merged.Width && !widthWarned {
			plan.Summary.Warnings = append(plan.Summary.Warnings, fmt.Sprintf("requested width %d expanded to %d for %s", merged.Width, appliedWidth, entry.RelativePath))
			widthWarned = true
		}
		if appliedWidth > widthUsed {
			widthUsed = appliedWidth
		}

		formattedNumber := merged.NumberPrefix + number + merged.NumberSuffix

		proposed := buildProposedPath(entry, merged, formattedNumber)

		candidate := Candidate{
			OriginalPath: entry.RelativePath,
			ProposedPath: proposed,
			Index:        sequenceIndex,
			Status:       CandidatePending,
			IsDir:        entry.IsDir,
		}

		if proposed == entry.RelativePath {
			candidate.Status = CandidateUnchanged
			plan.Candidates = append(plan.Candidates, candidate)
			sequenceIndex++
			nextValue++
			continue
		}

		if existing, ok := plannedTargets[proposed]; ok && existing != entry.RelativePath {
			plan.appendConflict(entry.RelativePath, proposed, ConflictExistingTarget)
			plan.Summary.SkippedCount++
			candidate.Status = CandidateSkipped
			plan.Candidates = append(plan.Candidates, candidate)
			sequenceIndex++
			nextValue++
			continue
		}

		lowerKey := strings.ToLower(proposed)
		if existing, ok := plannedTargetsFold[lowerKey]; ok && existing != entry.RelativePath {
			plan.appendConflict(entry.RelativePath, proposed, ConflictExistingTarget)
			plan.Summary.SkippedCount++
			candidate.Status = CandidateSkipped
			plan.Candidates = append(plan.Candidates, candidate)
			sequenceIndex++
			nextValue++
			continue
		}

		targetAbs := filepath.Join(merged.WorkingDir, filepath.FromSlash(proposed))
		if info, statErr := os.Stat(targetAbs); statErr == nil {
			origInfo, origErr := os.Stat(entry.AbsolutePath)
			if origErr != nil || !os.SameFile(info, origInfo) {
				plan.appendConflict(entry.RelativePath, proposed, ConflictExistingTarget)
				plan.Summary.SkippedCount++
				candidate.Status = CandidateSkipped
				plan.Candidates = append(plan.Candidates, candidate)
				sequenceIndex++
				nextValue++
				continue
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return Plan{}, statErr
		}

		plan.Candidates = append(plan.Candidates, candidate)
		plan.Summary.RenamedCount++
		plannedTargets[proposed] = entry.RelativePath
		plannedTargetsFold[lowerKey] = entry.RelativePath
		sequenceIndex++
		nextValue++
	}

	plan.Summary.AppliedWidth = widthUsed
	return plan, nil
}

func mergeOptions(opts Options) Options {
	merged := DefaultOptions()
	if opts.Start != 0 {
		merged.Start = opts.Start
	}
	merged.Width = opts.Width
	merged.WidthSet = opts.WidthSet
	if opts.Placement != "" {
		merged.Placement = opts.Placement
	}
	if opts.Separator != "" {
		merged.Separator = opts.Separator
	}
	merged.NumberPrefix = opts.NumberPrefix
	merged.NumberSuffix = opts.NumberSuffix
	merged.WorkingDir = opts.WorkingDir
	merged.IncludeDirectories = opts.IncludeDirectories
	merged.IncludeHidden = opts.IncludeHidden
	merged.Recursive = opts.Recursive
	merged.Extensions = append([]string(nil), opts.Extensions...)
	merged.DryRun = opts.DryRun
	merged.AutoApply = opts.AutoApply
	return merged
}

func buildProposedPath(entry traversalCandidate, opts Options, formattedNumber string) string {
	dir := filepath.Dir(entry.RelativePath)
	if dir == "." {
		dir = ""
	}

	stem := entry.Stem
	if opts.Placement == PlacementPrefix {
		stem = formattedNumber + joinIfNeeded(opts.Separator, stem)
	} else {
		stem = stem + joinIfNeeded(opts.Separator, formattedNumber)
	}

	if !entry.IsDir && entry.Extension != "" {
		stem += entry.Extension
	}

	if dir == "" {
		return filepath.ToSlash(stem)
	}
	return filepath.ToSlash(filepath.Join(dir, stem))
}

func joinIfNeeded(separator, value string) string {
	if separator == "" {
		return value
	}
	return separator + value
}

func (p *Plan) appendConflict(original, proposed string, reason ConflictReason) {
	p.SkippedConflicts = append(p.SkippedConflicts, Conflict{
		OriginalPath:    original,
		ConflictingPath: proposed,
		Reason:          reason,
	})
}
