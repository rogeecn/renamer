package insert

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rogeecn/renamer/internal/listing"
)

// Request encapsulates the inputs required to run an insert operation.
type Request struct {
	WorkingDir      string
	PositionToken   string
	InsertText      string
	IncludeDirs     bool
	Recursive       bool
	IncludeHidden   bool
	ExtensionFilter []string
	DryRun          bool
	AutoConfirm     bool
	Timestamp       time.Time
}

// NewRequest constructs a Request from shared listing scope.
func NewRequest(scope *listing.ListingRequest) *Request {
	if scope == nil {
		return &Request{}
	}

	extensions := append([]string(nil), scope.Extensions...)

	return &Request{
		WorkingDir:      scope.WorkingDir,
		IncludeDirs:     scope.IncludeDirectories,
		Recursive:       scope.Recursive,
		IncludeHidden:   scope.IncludeHidden,
		ExtensionFilter: extensions,
	}
}

// SetExecutionMode updates dry-run and auto-apply preferences.
func (r *Request) SetExecutionMode(dryRun, autoConfirm bool) {
	r.DryRun = dryRun
	r.AutoConfirm = autoConfirm
}

// SetPositionAndText stores the user-supplied position token and insert text.
func (r *Request) SetPositionAndText(positionToken, insertText string) {
	r.PositionToken = positionToken
	r.InsertText = insertText
}

// Normalize ensures working directory and timestamp fields are ready for execution.
func (r *Request) Normalize() error {
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

	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now().UTC()
	}

	return nil
}
