package replace

import (
	"errors"
	"strings"
)

// ParseArgs splits CLI arguments into patterns and replacement while deduplicating patterns.
type ParseArgsResult struct {
	Patterns    []string
	Replacement string
	Duplicates  []string
}

// ParseArgs interprets positional arguments for the replace command.
// The final token is treated as the replacement; all preceding tokens are literal patterns.
func ParseArgs(args []string) (ParseArgsResult, error) {
	if len(args) < 2 {
		return ParseArgsResult{}, errors.New("provide at least one pattern and a replacement value")
	}

	replacement := args[len(args)-1]
	patternTokens := args[:len(args)-1]

	seen := make(map[string]struct{}, len(patternTokens))
	patterns := make([]string, 0, len(patternTokens))
	duplicates := make([]string, 0)

	for _, token := range patternTokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		key := trimmed
		if _, ok := seen[key]; ok {
			duplicates = append(duplicates, trimmed)
			continue
		}
		seen[key] = struct{}{}
		patterns = append(patterns, trimmed)
	}

	if len(patterns) == 0 {
		return ParseArgsResult{}, errors.New("at least one non-empty pattern is required before the replacement")
	}

	return ParseArgsResult{
		Patterns:    patterns,
		Replacement: replacement,
		Duplicates:  duplicates,
	}, nil
}
