package filters

import (
	"fmt"
	"strings"
)

// ParseExtensions converts a raw `|`-delimited extension string into a
// normalized, deduplicated slice. Tokens must be prefixed with a dot.
func ParseExtensions(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	tokens := strings.Split(raw, "|")
	seen := make(map[string]struct{}, len(tokens))
	result := make([]string, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			return nil, fmt.Errorf("extensions string contains empty token")
		}
		if !strings.HasPrefix(trimmed, ".") {
			return nil, fmt.Errorf("extension %q must start with '.'", trimmed)
		}
		normalized := strings.ToLower(trimmed)
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result, nil
}

// MergeExtensions merges two extension slices, deduplicating case-insensitively.
func MergeExtensions(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}

	seen := make(map[string]struct{}, len(base)+len(extra))
	merged := make([]string, 0, len(base)+len(extra))

	for _, ext := range base {
		lower := strings.ToLower(ext)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		merged = append(merged, lower)
	}

	for _, ext := range extra {
		lower := strings.ToLower(ext)
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		merged = append(merged, lower)
	}

	return merged
}
