package remove

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

// Candidate represents a file or directory eligible for token removal.
type Candidate struct {
	RelativePath string
	OriginalPath string
	BaseName     string
	IsDir        bool
	Depth        int
}

// Traverse walks the working directory according to the request scope and invokes fn for each
// candidate (files by default, directories when IncludeDirectories is true).
func Traverse(ctx context.Context, req *Request, fn func(Candidate) error) error {
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
