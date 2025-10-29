package replace

import "sort"

// ConflictDetail describes a rename that could not be applied.
type ConflictDetail struct {
	OriginalPath string
	ProposedPath string
	Reason       string
}

// Summary aggregates metrics for previews, applies, and ledger entries.
type Summary struct {
	TotalCandidates  int
	ChangedCount     int
	PatternMatches   map[string]int
	Conflicts        []ConflictDetail
	Duplicates       []string
	EmptyReplacement bool
}

// NewSummary constructs an initialized summary.
func NewSummary() Summary {
	return Summary{
		PatternMatches: make(map[string]int),
	}
}

// AddDuplicate records a duplicate pattern supplied by the user.
func (s *Summary) AddDuplicate(pattern string) {
	if pattern == "" {
		return
	}
	s.Duplicates = append(s.Duplicates, pattern)
}

// AddResult incorporates an individual candidate replacement result.
// RecordCandidate updates aggregate counts for a processed candidate and any matches.
func (s *Summary) RecordCandidate(res Result) {
	s.TotalCandidates++
	if !res.Changed {
		return
	}
	s.ChangedCount++
	for pattern, count := range res.Matches {
		s.PatternMatches[pattern] += count
	}
}

// AddConflict appends a conflict detail to the summary.
func (s *Summary) AddConflict(conflict ConflictDetail) {
	s.Conflicts = append(s.Conflicts, conflict)
}

// SortedDuplicates returns de-duplicated duplicates list for reporting.
func (s *Summary) SortedDuplicates() []string {
	if len(s.Duplicates) == 0 {
		return nil
	}
	copyList := make([]string, 0, len(s.Duplicates))
	seen := make(map[string]struct{}, len(s.Duplicates))
	for _, dup := range s.Duplicates {
		if _, ok := seen[dup]; ok {
			continue
		}
		seen[dup] = struct{}{}
		copyList = append(copyList, dup)
	}
	sort.Strings(copyList)
	return copyList
}

// ReplacementWasEmpty records whether the replacement string is empty and returns true.
func (s *Summary) ReplacementWasEmpty(replacement string) bool {
	if replacement == "" {
		s.EmptyReplacement = true
		return true
	}
	return false
}
