package regex

// Summary describes the outcome of previewing or applying a regex rename request.
type Summary struct {
	TotalCandidates int
	Matched         int
	Changed         int
	Skipped         int
	Conflicts       []Conflict
	Warnings        []string
	Entries         []PreviewEntry
	LedgerMetadata  map[string]any
}

// ConflictReason enumerates reasons a proposed rename cannot proceed.
type ConflictReason string

const (
	ConflictDuplicateTarget ConflictReason = "duplicate_target"
	ConflictExistingFile    ConflictReason = "existing_file"
	ConflictExistingDir     ConflictReason = "existing_directory"
	ConflictInvalidTemplate ConflictReason = "invalid_template"
)

// Conflict reports a blocked rename candidate to the CLI and callers.
type Conflict struct {
	OriginalPath string
	ProposedPath string
	Reason       ConflictReason
}

// EntryStatus captures the preview disposition for a candidate path.
type EntryStatus string

const (
	EntryChanged  EntryStatus = "changed"
	EntryNoChange EntryStatus = "no_change"
	EntrySkipped  EntryStatus = "skipped"
)

// PreviewEntry documents a single rename candidate and the proposed output.
type PreviewEntry struct {
	OriginalPath string
	ProposedPath string
	Status       EntryStatus
	MatchGroups  []string
}
