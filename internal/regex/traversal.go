package regex

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/traversal"
)

// Candidate represents a file or directory subject to regex evaluation.
type Candidate struct {
	RelativePath string
	OriginalPath string
	BaseName     string
	Stem         string
	Extension    string
	IsDir        bool
	Depth        int
}

// TraverseCandidates walks the working directory and invokes fn for each eligible candidate.
func TraverseCandidates(ctx context.Context, req *Request, fn func(Candidate) error) error {
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
			name := entry.Name()
			stem := name
			ext := ""
			if !isDir {
				if dot := strings.IndexRune(name, '.'); dot > 0 {
					ext = name[dot:]
					stem = name[:dot]
				}

				if len(extensions) > 0 {
					lower := strings.ToLower(ext)
					if _, ok := extensions[lower]; !ok {
						return nil
					}
				}
			}

			rel := filepath.ToSlash(relPath)
			if rel == "." {
				return nil
			}

			candidate := Candidate{
				RelativePath: rel,
				OriginalPath: filepath.Join(req.WorkingDir, relPath),
				BaseName:     name,
				Stem:         stem,
				Extension:    ext,
				IsDir:        isDir,
				Depth:        depth,
			}

			return fn(candidate)
		},
	)
}
