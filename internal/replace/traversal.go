package replace

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

// Candidate represents a file or directory that may be renamed.
type Candidate struct {
	RelativePath string
	OriginalPath string
	BaseName     string
	IsDir        bool
	Depth        int
}

// TraverseCandidates walks the working directory according to the request scope and invokes fn for
// every eligible candidate (files by default, directories when IncludeDirectories is true).
func TraverseCandidates(ctx context.Context, req *ReplaceRequest, fn func(Candidate) error) error {
	if err := req.Validate(); err != nil {
		return err
	}

	extensions := make(map[string]struct{}, len(req.Extensions))
	for _, ext := range req.Extensions {
		lower := strings.ToLower(ext)
		extensions[lower] = struct{}{}
	}

	walker := traversal.NewWalker()

	return walker.Walk(
		req.WorkingDir,
		req.Recursive,
		req.IncludeDirectories,
		req.IncludeHidden,
		0,
		func(relPath string, entry fs.DirEntry, depth int) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			isDir := entry.IsDir()
			ext := strings.ToLower(filepath.Ext(entry.Name()))

			if !isDir && len(extensions) > 0 {
				if _, ok := extensions[ext]; !ok {
					return nil
				}
			}

			candidate := Candidate{
				RelativePath: filepath.ToSlash(relPath),
				OriginalPath: filepath.Join(req.WorkingDir, relPath),
				BaseName:     entry.Name(),
				IsDir:        isDir,
				Depth:        depth,
			}

			if candidate.RelativePath == "." {
				return nil
			}

			return fn(candidate)
		},
	)
}
