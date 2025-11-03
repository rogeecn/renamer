package plan

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/traversal"
)

// CollectCandidates walks the scope described by req and returns eligible file candidates.
func CollectCandidates(ctx context.Context, req *listing.ListingRequest) ([]Candidate, error) {
	if req == nil {
		return nil, errors.New("collect candidates: request cannot be nil")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	w := traversal.NewWalker()
	extensions := make(map[string]struct{}, len(req.Extensions))
	for _, ext := range req.Extensions {
		extensions[ext] = struct{}{}
	}

	candidates := make([]Candidate, 0)

	err := w.Walk(
		req.WorkingDir,
		req.Recursive,
		false, // directories are not considered candidates
		req.IncludeHidden,
		req.MaxDepth,
		func(relPath string, entry fs.DirEntry, depth int) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if entry.IsDir() {
				return nil
			}

			relSlash := filepath.ToSlash(relPath)
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if len(extensions) > 0 {
				if _, match := extensions[ext]; !match {
					return nil
				}
			}

			info, err := entry.Info()
			if err != nil {
				return err
			}

			candidates = append(candidates, Candidate{
				OriginalPath: relSlash,
				SizeBytes:    info.Size(),
				Depth:        depth,
				Extension:    filepath.Ext(entry.Name()),
			})

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return candidates, nil
}
