package sequence

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Placement controls where a sequence number is inserted.
type Placement string

const (
	// PlacementSuffix appends the sequence number after the filename stem.
	PlacementSuffix Placement = "suffix"
	// PlacementPrefix inserts the sequence number before the filename stem.
	PlacementPrefix Placement = "prefix"
)

// Options captures configuration for numbering operations.
type Options struct {
	WorkingDir         string
	Start              int
	Width              int
	WidthSet           bool
	NumberPrefix       string
	NumberSuffix       string
	Placement          Placement
	Separator          string
	IncludeHidden      bool
	IncludeDirectories bool
	Recursive          bool
	Extensions         []string
	DryRun             bool
	AutoApply          bool
}

// DefaultOptions returns a copy of the default configuration.
func DefaultOptions() Options {
	return Options{
		Start:     1,
		Width:     3,
		Placement: PlacementPrefix,
		Separator: "_",
	}
}

func validateOptions(opts *Options) error {
	if opts == nil {
		return errors.New("options cannot be nil")
	}

	if opts.WorkingDir == "" {
		return errors.New("working directory must be provided")
	}

	abs, err := filepath.Abs(opts.WorkingDir)
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	opts.WorkingDir = abs

	if opts.Start < 1 {
		return errors.New("start must be >= 1")
	}

	if opts.Width < 0 {
		return errors.New("width cannot be negative")
	}
	if opts.WidthSet && opts.Width < 1 {
		return errors.New("width must be >= 1 when specified")
	}

	switch opts.Placement {
	case PlacementPrefix, PlacementSuffix:
		// ok
	case "":
		opts.Placement = PlacementSuffix
	default:
		return fmt.Errorf("unsupported placement %q", opts.Placement)
	}

	if strings.ContainsAny(opts.Separator, "/\\") {
		return errors.New("separator cannot contain path separators")
	}

	if strings.ContainsAny(opts.NumberPrefix, "/\\") {
		return errors.New("number prefix cannot contain path separators")
	}
	if strings.ContainsAny(opts.NumberSuffix, "/\\") {
		return errors.New("number suffix cannot contain path separators")
	}

	return nil
}
