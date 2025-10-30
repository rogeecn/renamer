package insert

// Status represents the preview outcome for a candidate entry.
type Status string

const (
	StatusChanged  Status = "changed"
	StatusNoChange Status = "no_change"
	StatusSkipped  Status = "skipped"
)

// PreviewEntry describes a single original → proposed mapping.
type PreviewEntry struct {
	OriginalPath string
	ProposedPath string
	Status       Status
	InsertedText string
}

// Conflict captures a conflicting rename outcome.
type Conflict struct {
	OriginalPath string
	ProposedPath string
	Reason       string
}

// Summary aggregates counts, warnings, conflicts, and ledger metadata for insert operations.
type Summary struct {
	TotalCandidates int
	TotalChanged    int
	NoChange        int

	Entries   []PreviewEntry
	Conflicts []Conflict
	Warnings  []string

	LedgerMetadata map[string]any
}

// NewSummary constructs an empty summary with initialized maps.
func NewSummary() *Summary {
	return &Summary{
		Entries:        make([]PreviewEntry, 0),
		Conflicts:      make([]Conflict, 0),
		Warnings:       make([]string, 0),
		LedgerMetadata: make(map[string]any),
	}
}

// RecordEntry appends a preview entry and updates aggregate counts.
func (s *Summary) RecordEntry(entry PreviewEntry) {
	s.Entries = append(s.Entries, entry)
	s.TotalCandidates++

	switch entry.Status {
	case StatusChanged:
		s.TotalChanged++
	case StatusNoChange:
		s.NoChange++
	}
}

// AddConflict records a blocking conflict.
func (s *Summary) AddConflict(conflict Conflict) {
	s.Conflicts = append(s.Conflicts, conflict)
}

// AddWarning adds a warning if not already present.
func (s *Summary) AddWarning(msg string) {
	if msg == "" {
		return
	}
	for _, existing := range s.Warnings {
		if existing == msg {
			return
		}
	}
	s.Warnings = append(s.Warnings, msg)
}

// HasConflicts indicates whether apply should be blocked.
func (s *Summary) HasConflicts() bool {
	return len(s.Conflicts) > 0
}
