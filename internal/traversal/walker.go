package traversal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Walker streams filesystem entries relative to a working directory.
type Walker struct{}

// NewWalker constructs a new Walker with default behavior.
func NewWalker() *Walker {
	return &Walker{}
}

// Walk traverses starting at root and invokes fn for each matching entry.
//
// The callback receives the relative path, os.DirEntry metadata, and depth.
// Directories that are symbolic links are not descended into when recursive is true.
func (w *Walker) Walk(
	root string,
	recursive bool,
	includeDirs bool,
	includeHidden bool,
	maxDepth int,
	fn func(relPath string, entry fs.DirEntry, depth int) error,
) error {
	if root == "" {
		return errors.New("walk root cannot be empty")
	}

	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("walk root must be a directory")
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	walker := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// propagate traversal errors to caller for logging/handling
			return err
		}

		rel, relErr := filepath.Rel(rootAbs, path)
		if relErr != nil {
			return relErr
		}

		rel = filepath.Clean(rel)
		depth := depthFor(rel)

		if rel == "." {
			// Skip emitting the root directory unless explicitly requested.
			if includeDirs {
				return fn(rel, d, depth)
			}
			return nil
		}

		if maxDepth > 0 && depth > maxDepth {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if !includeHidden && isHidden(rel) {
			if d.IsDir() && recursive {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if !recursive && depth > 0 {
				return fs.SkipDir
			}
			if !includeDirs {
				// continue traversal but don't emit directory
				return nil
			}
		}

		if d.Type()&os.ModeSymlink != 0 && d.IsDir() {
			// emit symlink entry but do not traverse into it
			if err := fn(rel, d, depth); err != nil {
				return err
			}
			if recursive {
				return nil
			}
			return fs.SkipDir
		}

		return fn(rel, d, depth)
	}

	if recursive {
		return filepath.WalkDir(rootAbs, walker)
	}

	entries, err := os.ReadDir(rootAbs)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		path := filepath.Join(rootAbs, entry.Name())
		if err := walker(path, entry, nil); err != nil {
			if errors.Is(err, fs.SkipDir) {
				continue
			}
			return err
		}
	}
	return nil
}

func depthFor(rel string) int {
	if rel == "." || rel == "" {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator))
}

func isHidden(rel string) bool {
	parts := strings.Split(rel, string(filepath.Separator))
	for _, part := range parts {
		if len(part) > 0 && part[0] == '.' {
			return true
		}
	}
	return false
}
