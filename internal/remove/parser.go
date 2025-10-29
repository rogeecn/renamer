package remove

import "strings"

// ParseArgsResult captures parser output for sequential removals.
type ParseArgsResult struct {
	Tokens     []string
	Duplicates []string
}

// ParseArgs splits, trims, and deduplicates tokens while preserving order.
func ParseArgs(args []string) (ParseArgsResult, error) {
	result := ParseArgsResult{Tokens: make([]string, 0, len(args))}

	seen := make(map[string]int)
	for _, raw := range args {
		token := strings.TrimSpace(raw)
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			result.Duplicates = append(result.Duplicates, token)
			continue
		}
		seen[token] = len(result.Tokens)
		result.Tokens = append(result.Tokens, token)
	}

	if len(result.Tokens) == 0 {
		return ParseArgsResult{}, ErrNoTokens
	}

	return result, nil
}
