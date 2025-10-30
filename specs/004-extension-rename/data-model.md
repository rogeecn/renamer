# Data Model – Extension Command

## Entity: ExtensionRequest
- **Fields**
  - `WorkingDir string` — Absolute path resolved from CLI `--path` or current directory.
  - `SourceExtensions []string` — Ordered list of unique, dot-prefixed source extensions (case-insensitive comparisons).
  - `TargetExtension string` — Dot-prefixed extension applied verbatim during rename.
  - `IncludeDirs bool` — Mirrors `--include-dirs` scope flag.
  - `Recursive bool` — Mirrors `--recursive` traversal flag.
  - `IncludeHidden bool` — True only when `--hidden` supplied.
  - `ExtensionFilter []string` — Normalized extension filter from `--extensions`, applied before source matching.
  - `DryRun bool` — Indicates preview-only execution (`--dry-run`).
  - `AutoConfirm bool` — Captures `--yes` for non-interactive apply.
  - `Timestamp time.Time` — Invocation timestamp for ledger correlation.
- **Validation Rules**
  - Require at least one source extension and one target extension (total args ≥ 2).
  - All extensions MUST start with `.` and contain ≥1 alphanumeric character after the dot.
  - Deduplicate source extensions case-insensitively; warn on duplicates/no-ops.
  - Target extension MUST NOT be empty and MUST include leading dot.
  - Reject empty string tokens after trimming whitespace.
- **Relationships**
  - Consumed by the extension rule engine to enumerate candidate files.
  - Serialized into ledger metadata alongside `ExtensionSummary`.

## Entity: ExtensionSummary
- **Fields**
  - `TotalCandidates int` — Number of filesystem entries examined post-scope filtering.
  - `TotalChanged int` — Count of files scheduled for rename (target extension applied).
  - `NoChange int` — Count of files already matching `TargetExtension`.
  - `PerExtensionCounts map[string]int` — Matches per source extension (case-insensitive key).
  - `Conflicts []Conflict` — Entries describing colliding target paths.
  - `Warnings []string` — Validation and scope warnings (duplicates, no matches, hidden skipped).
  - `PreviewEntries []PreviewEntry` — Ordered list of original/new paths with status `changed|no_change|skipped`.
  - `LedgerMetadata map[string]any` — Snapshot persisted with ledger entry (sources, target, flags).
- **Validation Rules**
  - Conflicts list MUST be empty before allowing apply.
  - `NoChange + TotalChanged` MUST equal count of entries included in preview.
  - Preview entries MUST be deterministic (stable sort by original path).
- **Relationships**
  - Emitted to preview renderer for CLI output formatting.
  - Persisted with ledger entry for undo operations and audits.

## Entity: Conflict
- **Fields**
  - `OriginalPath string` — Existing file path causing the collision.
  - `ProposedPath string` — Target path that clashes with another candidate.
  - `Reason string` — Short code (e.g., `duplicate_target`, `existing_file`) describing conflict type.
- **Validation Rules**
  - `ProposedPath` MUST be unique across non-conflicting entries.
  - Reasons limited to known enum for consistent messaging.
- **Relationships**
  - Referenced within `ExtensionSummary.Conflicts`.
  - Propagated to CLI preview warnings.

## Entity: PreviewEntry
- **Fields**
  - `OriginalPath string`
  - `ProposedPath string`
  - `Status string` — `changed`, `no_change`, or `skipped`.
  - `SourceExtension string` — Detected source extension (normalized lowercase).
- **Validation Rules**
  - `Status` MUST be one of the defined constants.
  - `ProposedPath` MUST equal `OriginalPath` when `Status == "no_change"`.
- **Relationships**
  - Aggregated into `ExtensionSummary.PreviewEntries`.
  - Used by preview renderer and ledger writer.
