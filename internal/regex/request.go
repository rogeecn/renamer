package regex

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Request captures the inputs required to evaluate regex-based rename operations.
type Request struct {
	WorkingDir         string
	Pattern            string
	Template           string
	IncludeDirectories bool
	Recursive          bool
	IncludeHidden      bool
	Extensions         []string
	DryRun             bool
	AutoConfirm        bool
	Timestamp          time.Time
}

// NewRequest constructs a Request with the supplied working directory and defaults the
// timestamp to the current UTC time. Additional fields should be set by the caller.
func NewRequest(workingDir string) Request {
	return Request{
		WorkingDir: workingDir,
		Timestamp:  time.Now().UTC(),
	}
}

// Validate ensures the request has usable defaults and a resolvable working directory.
func (r *Request) Validate() error {
	if r == nil {
		return errors.New("regex request cannot be nil")
	}

	if r.Pattern == "" {
		return errors.New("regex pattern is required")
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
