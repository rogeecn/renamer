package sequence

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

type traversalCandidate struct {
	RelativePath string
	AbsolutePath string
	Stem         string
	Extension    string
	IsDir        bool
	Depth        int
}

func collectTraversalCandidates(ctx context.Context, opts Options) ([]traversalCandidate, error) {
	if opts.WorkingDir == "" {
		return nil, errors.New("working directory must be provided")
	}

	absRoot, err := filepath.Abs(opts.WorkingDir)
	if err != nil {
		return nil, err
	}

	walker := traversal.NewWalker()

	allowedExts := make(map[string]struct{}, len(opts.Extensions))
	for _, ext := range opts.Extensions {
		lower := strings.ToLower(ext)
		allowedExts[lower] = struct{}{}
	}

	candidates := make([]traversalCandidate, 0)

	emit := func(relPath string, entry fs.DirEntry, depth int) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip the root placeholder emitted by the walker when include dirs is true.
		if relPath == "." {
			return nil
		}

		relSlash := filepath.ToSlash(relPath)
		absolute := filepath.Join(absRoot, relPath)
		candidate := traversalCandidate{
			RelativePath: relSlash,
			AbsolutePath: absolute,
			IsDir:        entry.IsDir(),
			Depth:        depth,
		}

		if !entry.IsDir() {
			rawExt := filepath.Ext(entry.Name())
			lowerExt := strings.ToLower(rawExt)
			if len(allowedExts) > 0 {
				if _, ok := allowedExts[lowerExt]; !ok {
					return nil
				}
			}

			candidate.Extension = rawExt
			stem := entry.Name()
			if rawExt != "" {
				stem = strings.TrimSuffix(stem, rawExt)
			}
			candidate.Stem = stem
		} else {
			candidate.Stem = entry.Name()
		}

		candidates = append(candidates, candidate)
		return nil
	}

	err = walker.Walk(absRoot, opts.Recursive, opts.IncludeDirectories, opts.IncludeHidden, 0, emit)
	if err != nil {
		return nil, err
	}

	return candidates, nil
}
