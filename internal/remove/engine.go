package remove

import "strings"

// Result captures the outcome of applying sequential removals to a candidate.
type Result struct {
	Candidate    Candidate
	ProposedName string
	Matches      map[string]int
	Changed      bool
}

// ApplyTokens removes each token sequentially from the candidate's basename.
func ApplyTokens(candidate Candidate, tokens []string) Result {
	current := candidate.BaseName
	matches := make(map[string]int, len(tokens))

	for _, token := range tokens {
		if token == "" {
			continue
		}
		count := strings.Count(current, token)
		if count == 0 {
			continue
		}
		current = strings.ReplaceAll(current, token, "")
		matches[token] += count
	}

	return Result{
		Candidate:    candidate,
		ProposedName: current,
		Matches:      matches,
		Changed:      current != candidate.BaseName,
	}
}
