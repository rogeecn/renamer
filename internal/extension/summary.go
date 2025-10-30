package extension

import (
	"strings"
)

// PreviewStatus represents the outcome for a single preview entry.
type PreviewStatus string

const (
	PreviewStatusChanged  PreviewStatus = "changed"
	PreviewStatusNoChange PreviewStatus = "no_change"
	PreviewStatusSkipped  PreviewStatus = "skipped"
)

// PreviewEntry captures a single original → proposed path mapping for preview output.
type PreviewEntry struct {
	OriginalPath    string
	ProposedPath    string
	Status          PreviewStatus
	SourceExtension string
}

// Conflict describes a proposed rename that cannot be applied safely.
type Conflict struct {
	OriginalPath string
	ProposedPath string
	Reason       string
}

// ExtensionSummary aggregates counts, conflicts, warnings, and ledger metadata.
type ExtensionSummary struct {
	TotalCandidates int
	TotalChanged    int
	NoChange        int

	PerExtensionCounts map[string]int
	Conflicts          []Conflict
	Warnings           []string
	Entries            []PreviewEntry

	LedgerMetadata map[string]any
}

// NewSummary constructs an empty ExtensionSummary with initialized maps.
func NewSummary() *ExtensionSummary {
	return &ExtensionSummary{
		PerExtensionCounts: make(map[string]int),
		LedgerMetadata:     make(map[string]any),
	}
}

// RecordEntry appends a preview entry and updates aggregate counters.
func (s *ExtensionSummary) RecordEntry(entry PreviewEntry) {
	s.Entries = append(s.Entries, entry)
	s.TotalCandidates++

	switch entry.Status {
	case PreviewStatusChanged:
		s.TotalChanged++
	case PreviewStatusNoChange:
		s.NoChange++
	}

	if entry.SourceExtension != "" {
		key := strings.ToLower(entry.SourceExtension)
		s.PerExtensionCounts[key]++
	}
}

// AddConflict registers a new conflict encountered during planning.
func (s *ExtensionSummary) AddConflict(conflict Conflict) {
	s.Conflicts = append(s.Conflicts, conflict)
}

// AddWarning ensures warning messages are collected without duplication.
func (s *ExtensionSummary) AddWarning(message string) {
	if message == "" {
		return
	}
	for _, existing := range s.Warnings {
		if existing == message {
			return
		}
	}
	s.Warnings = append(s.Warnings, message)
}

// HasConflicts reports whether any blocking conflicts were recorded.
func (s *ExtensionSummary) HasConflicts() bool {
	return len(s.Conflicts) > 0
}
