package remove

import "sort"

// ConflictDetail describes a rename that cannot proceed.
type ConflictDetail struct {
	OriginalPath string
	ProposedPath string
	Reason       string
}

// Summary aggregates preview/apply metrics for reporting and ledger metadata.
type Summary struct {
	TotalCandidates int
	ChangedCount    int
	TokenMatches    map[string]int
	Conflicts       []ConflictDetail
	Empties         []string
	Duplicates      []string
}

// NewSummary constructs an initialized summary instance.
func NewSummary() Summary {
	return Summary{
		TokenMatches: make(map[string]int),
	}
}

// RecordCandidate updates aggregate counts based on a candidate result.
func (s *Summary) RecordCandidate(res Result) {
	s.TotalCandidates++
	if !res.Changed {
		return
	}
	s.ChangedCount++
	for token, count := range res.Matches {
		s.TokenMatches[token] += count
	}
}

// AddConflict registers a conflict for reporting.
func (s *Summary) AddConflict(conflict ConflictDetail) {
	s.Conflicts = append(s.Conflicts, conflict)
}

// AddEmpty records a path whose resulting name would be empty.
func (s *Summary) AddEmpty(path string) {
	s.Empties = append(s.Empties, path)
}

// AddDuplicate stores duplicate tokens captured during parsing.
func (s *Summary) AddDuplicate(token string) {
	if token == "" {
		return
	}
	s.Duplicates = append(s.Duplicates, token)
}

// SortedDuplicates returns unique duplicate tokens sorted for deterministic output.
func (s *Summary) SortedDuplicates() []string {
	if len(s.Duplicates) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(s.Duplicates))
	result := make([]string, 0, len(s.Duplicates))
	for _, dup := range s.Duplicates {
		if _, ok := seen[dup]; ok {
			continue
		}
		seen[dup] = struct{}{}
		result = append(result, dup)
	}
	sort.Strings(result)
	return result
}

// SortedTokenMatches returns token match counts sorted alphabetically by token.
func (s *Summary) SortedTokenMatches() []struct {
	Token string
	Count int
} {
	if len(s.TokenMatches) == 0 {
		return nil
	}
	result := make([]struct {
		Token string
		Count int
	}, 0, len(s.TokenMatches))
	for token, count := range s.TokenMatches {
		result = append(result, struct {
			Token string
			Count int
		}{Token: token, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Token < result[j].Token
	})
	return result
}
