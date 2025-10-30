package insert

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

// PlannedOperation captures a filesystem rename to be applied.
type PlannedOperation struct {
	OriginalRelative string
	OriginalAbsolute string
	ProposedRelative string
	ProposedAbsolute string
	InsertedText     string
	IsDir            bool
	Depth            int
}

// BuildPlan enumerates candidates, computes preview entries, and prepares filesystem operations.
func BuildPlan(ctx context.Context, req *Request) (*Summary, []PlannedOperation, error) {
	if req == nil {
		return nil, nil, errors.New("insert request cannot be nil")
	}
	if err := req.Normalize(); err != nil {
		return nil, nil, err
	}

	summary := NewSummary()
	operations := make([]PlannedOperation, 0)
	detector := NewConflictDetector()

	filterSet := make(map[string]struct{}, len(req.ExtensionFilter))
	for _, ext := range req.ExtensionFilter {
		filterSet[strings.ToLower(ext)] = struct{}{}
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

			relative := filepath.ToSlash(relPath)
			name := entry.Name()

			ext := ""
			stem := name
			if !isDir {
				ext = filepath.Ext(name)
				stem = strings.TrimSuffix(name, ext)
			}

			if !isDir && len(filterSet) > 0 {
				lowerExt := strings.ToLower(ext)
				if _, ok := filterSet[lowerExt]; !ok {
					return nil
				}
			}

			stemRunes := []rune(stem)
			if err := ParseInputs(req.PositionToken, req.InsertText, len(stemRunes)); err != nil {
				return err
			}

			position, err := ResolvePosition(req.PositionToken, len(stemRunes))
			if err != nil {
				return err
			}

			status := StatusChanged
			proposedRelative := relative
			proposedAbsolute := filepath.Join(req.WorkingDir, filepath.FromSlash(relative))

			var builder strings.Builder
			builder.WriteString(string(stemRunes[:position.Index]))
			builder.WriteString(req.InsertText)
			builder.WriteString(string(stemRunes[position.Index:]))
			proposedName := builder.String() + ext

			dir := filepath.Dir(relative)
			if dir == "." {
				dir = ""
			}
			if dir == "" {
				proposedRelative = filepath.ToSlash(proposedName)
			} else {
				proposedRelative = filepath.ToSlash(filepath.Join(dir, proposedName))
			}
			proposedAbsolute = filepath.Join(req.WorkingDir, filepath.FromSlash(proposedRelative))

			if proposedRelative == relative {
				status = StatusNoChange
			}

			if status == StatusChanged {
				if reason, ok := detector.Register(relative, proposedRelative); !ok {
					summary.AddConflict(Conflict{
						OriginalPath: relative,
						ProposedPath: proposedRelative,
						Reason:       "duplicate_target",
					})
					summary.AddWarning(reason)
					status = StatusSkipped
				} else if info, err := os.Stat(proposedAbsolute); err == nil {
					origInfo, origErr := os.Stat(filepath.Join(req.WorkingDir, filepath.FromSlash(relative)))
					if origErr != nil {
						return origErr
					}
					if !os.SameFile(info, origInfo) {
						reason := "existing_file"
						if info.IsDir() {
							reason = "existing_directory"
						}
						summary.AddConflict(Conflict{
							OriginalPath: relative,
							ProposedPath: proposedRelative,
							Reason:       reason,
						})
						summary.AddWarning("target already exists")
						detector.Forget(proposedRelative)
						status = StatusSkipped
					}
				} else if !errors.Is(err, os.ErrNotExist) {
					return err
				}

				if status == StatusChanged {
					operations = append(operations, PlannedOperation{
						OriginalRelative: relative,
						OriginalAbsolute: filepath.Join(req.WorkingDir, filepath.FromSlash(relative)),
						ProposedRelative: proposedRelative,
						ProposedAbsolute: proposedAbsolute,
						InsertedText:     req.InsertText,
						IsDir:            isDir,
						Depth:            depth,
					})
				}
			}

			entrySummary := PreviewEntry{
				OriginalPath: relative,
				ProposedPath: proposedRelative,
				Status:       status,
			}
			if status == StatusChanged {
				entrySummary.InsertedText = req.InsertText
			}

			summary.RecordEntry(entrySummary)
			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	sort.SliceStable(summary.Entries, func(i, j int) bool {
		return summary.Entries[i].OriginalPath < summary.Entries[j].OriginalPath
	})

	return summary, operations, nil
}
