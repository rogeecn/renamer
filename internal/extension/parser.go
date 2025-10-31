package extension

import (
	"errors"
	"fmt"
	"strings"
)

// ParseResult captures normalized CLI arguments for the extension command.
type ParseResult struct {
	SourcesCanonical []string
	SourcesDisplay   []string
	Duplicates       []string
	NoOps            []string
	Target           string
}

// ParseArgs validates extension command arguments and returns normalized tokens.
func ParseArgs(args []string) (ParseResult, error) {
	if len(args) < 2 {
		return ParseResult{}, errors.New("at least one source extension and a target extension are required")
	}

	target := NormalizeTargetExtension(args[len(args)-1])
	if target == "" {
		return ParseResult{}, errors.New("target extension cannot be empty")
	}
	if !strings.HasPrefix(target, ".") {
		return ParseResult{}, fmt.Errorf("target extension %q must start with '.'", target)
	}
	if len(target) == 1 {
		return ParseResult{}, fmt.Errorf("target extension %q must include characters after '.'", target)
	}

	rawSources := args[:len(args)-1]
	trimmedSources := make([]string, len(rawSources))
	for i, src := range rawSources {
		trimmed := strings.TrimSpace(src)
		if trimmed == "" {
			return ParseResult{}, errors.New("source extensions cannot be empty")
		}
		if !strings.HasPrefix(trimmed, ".") {
			return ParseResult{}, fmt.Errorf("source extension %q must start with '.'", trimmed)
		}
		if len(trimmed) == 1 {
			return ParseResult{}, fmt.Errorf("source extension %q must include characters after '.'", trimmed)
		}
		trimmedSources[i] = trimmed
	}

	canonical, display, duplicates := NormalizeSourceExtensions(trimmedSources)
	targetCanonical := CanonicalExtension(target)

	filteredCanonical := make([]string, 0, len(canonical))
	filteredDisplay := make([]string, 0, len(display))
	noOps := make([]string, 0)
	for i, canon := range canonical {
		if canon == targetCanonical && display[i] == target {
			noOps = append(noOps, display[i])
			continue
		}
		filteredCanonical = append(filteredCanonical, canon)
		filteredDisplay = append(filteredDisplay, display[i])
	}

	if len(filteredCanonical) == 0 {
		return ParseResult{}, errors.New("all source extensions match the target extension; provide at least one distinct source extension")
	}

	return ParseResult{
		SourcesCanonical: filteredCanonical,
		SourcesDisplay:   filteredDisplay,
		Duplicates:       duplicates,
		NoOps:            noOps,
		Target:           target,
	}, nil
}
