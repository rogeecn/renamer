package replace

import "strings"

// Result captures the outcome of applying patterns to a candidate name.
type Result struct {
	Candidate    Candidate
	ProposedName string
	Matches      map[string]int
	Changed      bool
}

// ApplyPatterns replaces every occurrence of the provided patterns within the candidate's base name.
func ApplyPatterns(candidate Candidate, patterns []string, replacement string) Result {
	current := candidate.BaseName
	matches := make(map[string]int, len(patterns))

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		count := strings.Count(current, pattern)
		if count == 0 {
			continue
		}
		current = strings.ReplaceAll(current, pattern, replacement)
		matches[pattern] += count
	}

	changed := current != candidate.BaseName

	return Result{
		Candidate:    candidate,
		ProposedName: current,
		Matches:      matches,
		Changed:      changed,
	}
}
