package sequence

// Plan represents the ordered numbering proposal produced during preview.
type Plan struct {
	Candidates       []Candidate
	SkippedConflicts []Conflict
	Summary          Summary
	Config           Config
}

// Candidate describes a single file considered for numbering.
type Candidate struct {
	OriginalPath string
	ProposedPath string
	Index        int
	IsDir        bool
	Status       CandidateStatus
}

// CandidateStatus indicates how a candidate was handled during preview.
type CandidateStatus string

const (
	// CandidatePending means the candidate will be renamed when applied.
	CandidatePending CandidateStatus = "pending"
	// CandidateSkipped indicates the candidate was skipped (e.g., conflict).
	CandidateSkipped CandidateStatus = "skipped"
	// CandidateUnchanged indicates the candidate already matches the target name.
	CandidateUnchanged CandidateStatus = "unchanged"
)

// Conflict captures a skipped item and the reason it could not be renamed.
type Conflict struct {
	OriginalPath    string
	ConflictingPath string
	Reason          ConflictReason
}

// ConflictReason enumerates known conflict types.
type ConflictReason string

const (
	// ConflictExistingTarget indicates the proposed name collides with an existing file.
	ConflictExistingTarget ConflictReason = "existing_target"
	// ConflictInvalidSeparator indicates the proposed separator produced an invalid path.
	ConflictInvalidSeparator ConflictReason = "invalid_separator"
	// ConflictWidthOverflow indicates numbering exceeded a fixed width.
	ConflictWidthOverflow ConflictReason = "width_overflow"
)

// Summary aggregates totals surfaced during preview.
type Summary struct {
	TotalCandidates int
	RenamedCount    int
	SkippedCount    int
	Warnings        []string
	AppliedWidth    int
}

// Config snapshots the numbering configuration for preview/apply/ledger.
type Config struct {
	Start        int
	Width        int
	Placement    Placement
	Separator    string
	NumberPrefix string
	NumberSuffix string
}
