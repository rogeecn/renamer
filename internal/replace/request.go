package replace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ReplaceRequest captures all inputs needed to evaluate a replace operation.
type ReplaceRequest struct {
	WorkingDir         string
	Patterns           []string
	Replacement        string
	IncludeDirectories bool
	Recursive          bool
	IncludeHidden      bool
	Extensions         []string
}

// Validate ensures the request is well-formed before preview/apply.
func (r *ReplaceRequest) Validate() error {
	if r == nil {
		return errors.New("replace request cannot be nil")
	}

	if len(r.Patterns) == 0 {
		return errors.New("at least one pattern is required")
	}

	if r.Replacement == "" {
		// Allow empty replacement but make sure caller has surfaced warnings elsewhere.
		// No error returned; preview will message accordingly.
	}

	if r.WorkingDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		r.WorkingDir = cwd
	}

	if !filepath.IsAbs(r.WorkingDir) {
		abs, err := filepath.Abs(r.WorkingDir)
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
		r.WorkingDir = abs
	}

	info, err := os.Stat(r.WorkingDir)
	if err != nil {
		return fmt.Errorf("stat working directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("working directory %q is not a directory", r.WorkingDir)
	}

	return nil
}
