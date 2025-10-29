package remove

import "sort"

// Summary aggregates results across preview/apply phases.
type Summary struct {
	totalCandidates int
	changedCount    int
	conflicts       []Conflict
	empties         []string
	tokenMatches    map[string]int
	duplicates      []string
}

// Conflict describes a rename conflict detected during planning.
type Conflict struct {
	Original string
	Proposed string
	Reason   string
}

// NewSummary constructs a ready-to-use Summary.
func NewSummary() Summary {
	return Summary{
		tokenMatches: make(map[string]int),
		conflicts:    make([]Conflict, 0),
		empties:      make([]string, 0),
		duplicates:   make([]string, 0),
	}
}

// RecordCandidate increments the total candidate count.
func (s *Summary) RecordCandidate() {
	s.totalCandidates++
}

// RecordChange increments changed items.
func (s *Summary) RecordChange() {
	s.changedCount++
}

// AddTokenMatch records the number of matches for a token.
func (s *Summary) AddTokenMatch(token string, count int) {
	s.tokenMatches[token] += count
}

// AddConflict registers a detected conflict.
func (s *Summary) AddConflict(c Conflict) {
	s.conflicts = append(s.conflicts, c)
}

// AddEmpty registers a path skipped due to empty result names.
func (s *Summary) AddEmpty(path string) {
	s.empties = append(s.empties, path)
}

// AddDuplicate tracks duplicate tokens encountered during parsing.
func (s *Summary) AddDuplicate(token string) {
	s.duplicates = append(s.duplicates, token)
}

// TotalCandidates returns how many items were considered.
func (s Summary) TotalCandidates() int {
	return s.totalCandidates
}

// ChangedCount returns the number of items whose names changed.
func (s Summary) ChangedCount() int {
	return s.changedCount
}

// Conflicts returns a copy of conflict info.
func (s Summary) Conflicts() []Conflict {
	out := make([]Conflict, len(s.conflicts))
	copy(out, s.conflicts)
	return out
}

// Empties returns paths skipped for empty basename results.
func (s Summary) Empties() []string {
	out := make([]string, len(s.empties))
	copy(out, s.empties)
	return out
}

// TokenMatches returns a sorted slice of tokens and counts.
func (s Summary) TokenMatches() []struct {
	Token string
	Count int
} {
	pairs := make([]struct {
		Token string
		Count int
	}, 0, len(s.tokenMatches))
	for token, count := range s.tokenMatches {
		pairs = append(pairs, struct {
			Token string
			Count int
		}{Token: token, Count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Token < pairs[j].Token
	})
	return pairs
}

// Duplicates returns duplicates flagged by the parser.
func (s Summary) Duplicates() []string {
	out := make([]string, len(s.duplicates))
	copy(out, s.duplicates)
	sort.Strings(out)
	return out
}
