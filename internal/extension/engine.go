package extension

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

// PlannedRename describes a filesystem rename operation produced during planning.
type PlannedRename struct {
	OriginalRelative string
	OriginalAbsolute string
	ProposedRelative string
	ProposedAbsolute string
	SourceExtension  string
	IsDir            bool
	Depth            int
}

// PlanResult captures the preview summary and concrete operations required for apply.
type PlanResult struct {
	Summary    *ExtensionSummary
	Operations []PlannedRename
}

// BuildPlan walks the scoped filesystem, collecting preview entries and rename operations.
func BuildPlan(ctx context.Context, req *ExtensionRequest) (*PlanResult, error) {
	if req == nil {
		return nil, errors.New("extension request cannot be nil")
	}
	if err := req.Normalize(); err != nil {
		return nil, err
	}

	summary := NewSummary()
	operations := make([]PlannedRename, 0)
	detector := newConflictDetector()

	targetExt := NormalizeTargetExtension(req.TargetExtension)
	targetCanonical := CanonicalExtension(targetExt)

	sourceSet := make(map[string]struct{}, len(req.SourceExtensions))
	for _, source := range req.SourceExtensions {
		sourceSet[CanonicalExtension(source)] = struct{}{}
	}

	filterSet := make(map[string]struct{}, len(req.ExtensionFilter))
	for _, filter := range req.ExtensionFilter {
		filterSet[CanonicalExtension(filter)] = struct{}{}
	}

	walker := traversal.NewWalker()

	err := walker.Walk(
		req.WorkingDir,
		req.Recursive,
		req.IncludeDirs,
		req.IncludeHidden,
		0,
		func(relPath string, entry fs.DirEntry, depth int) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if relPath == "." {
				return nil
			}

			isDir := entry.IsDir()
			if isDir && !req.IncludeDirs {
				return nil
			}

			name := entry.Name()
			rawExt := strings.TrimSpace(filepath.Ext(name))
			canonicalExt := CanonicalExtension(rawExt)

			if !isDir && len(filterSet) > 0 {
				if _, ok := filterSet[canonicalExt]; !ok {
					return nil
				}
			}

			_, sourceMatch := sourceSet[canonicalExt]
			targetMatch := canonicalExt == targetCanonical && rawExt == targetExt

			if !sourceMatch && !targetMatch {
				return nil
			}

			relative := filepath.ToSlash(relPath)
			originalAbsolute := filepath.Join(req.WorkingDir, filepath.FromSlash(relative))

			status := PreviewStatusChanged
			targetRelative := relative
			targetAbsolute := originalAbsolute

			if targetMatch {
				status = PreviewStatusNoChange
			}

			if status == PreviewStatusChanged {
				base := strings.TrimSuffix(name, rawExt)
				targetName := base + targetExt
				dir := filepath.Dir(relative)
				if dir == "." {
					dir = ""
				}
				if dir == "" {
					targetRelative = filepath.ToSlash(targetName)
				} else {
					targetRelative = filepath.ToSlash(filepath.Join(dir, targetName))
				}
				targetAbsolute = filepath.Join(req.WorkingDir, filepath.FromSlash(targetRelative))

				allowed, err := detector.evaluateTarget(summary, relative, targetRelative, originalAbsolute, targetAbsolute)
				if err != nil {
					return err
				}
				if allowed {
					operations = append(operations, PlannedRename{
						OriginalRelative: relative,
						OriginalAbsolute: originalAbsolute,
						ProposedRelative: targetRelative,
						ProposedAbsolute: targetAbsolute,
						SourceExtension:  rawExt,
						IsDir:            isDir,
						Depth:            depth,
					})
				} else {
					status = PreviewStatusSkipped
				}
			}

			entrySummary := PreviewEntry{
				OriginalPath:    relative,
				ProposedPath:    targetRelative,
				Status:          status,
				SourceExtension: rawExt,
			}
			summary.RecordEntry(entrySummary)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &PlanResult{
		Summary:    summary,
		Operations: operations,
	}, nil
}
