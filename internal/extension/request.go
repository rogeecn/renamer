package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rogeecn/renamer/internal/listing"
)

// ExtensionRequest captures all inputs required to evaluate an extension normalization run.
type ExtensionRequest struct {
	WorkingDir              string
	SourceExtensions        []string
	DisplaySourceExtensions []string
	TargetExtension         string
	DuplicateSources        []string
	NoOpSources             []string

	IncludeDirs   bool
	Recursive     bool
	IncludeHidden bool

	ExtensionFilter []string

	DryRun      bool
	AutoConfirm bool

	Timestamp time.Time
}

// NewRequest seeds an ExtensionRequest using the shared listing scope settings.
func NewRequest(scope *listing.ListingRequest) *ExtensionRequest {
	if scope == nil {
		return &ExtensionRequest{}
	}

	filterCopy := append([]string(nil), scope.Extensions...)

	return &ExtensionRequest{
		WorkingDir:      scope.WorkingDir,
		IncludeDirs:     scope.IncludeDirectories,
		Recursive:       scope.Recursive,
		IncludeHidden:   scope.IncludeHidden,
		ExtensionFilter: filterCopy,
	}
}

// SetExecutionMode records dry-run/auto-apply preferences inherited from CLI flags.
func (r *ExtensionRequest) SetExecutionMode(dryRun, autoConfirm bool) {
	r.DryRun = dryRun
	r.AutoConfirm = autoConfirm
}

// SetExtensions stores source and target extensions before normalization.
func (r *ExtensionRequest) SetExtensions(canonical []string, display []string, target string) {
	r.SourceExtensions = append(r.SourceExtensions[:0], canonical...)
	r.DisplaySourceExtensions = append(r.DisplaySourceExtensions[:0], display...)
	r.TargetExtension = target
}

// SetWarnings captures duplicate or no-op tokens for later surfacing in preview output.
func (r *ExtensionRequest) SetWarnings(dupes, noOps []string) {
	r.DuplicateSources = append(r.DuplicateSources[:0], dupes...)
	r.NoOpSources = append(r.NoOpSources[:0], noOps...)
}

// Normalize ensures working directory and timestamp fields are ready for execution.
func (r *ExtensionRequest) Normalize() error {
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
