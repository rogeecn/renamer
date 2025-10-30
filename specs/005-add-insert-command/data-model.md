# Data Model – Insert Command

## Entity: InsertRequest
- **Fields**
  - `WorkingDir string` — Absolute path derived from CLI `--path` or current directory.
  - `PositionToken string` — Raw user input (`^`, `$`, positive int, negative int) describing insertion point.
  - `InsertText string` — Unicode string to insert.
  - `IncludeDirs bool` — Mirrors `--include-dirs` scope flag.
  - `Recursive bool` — Mirrors `--recursive`.
  - `IncludeHidden bool` — True only when `--hidden` supplied.
  - `ExtensionFilter []string` — Normalized tokens from `--extensions`.
  - `DryRun bool` — Preview-only execution state.
  - `AutoConfirm bool` — Captures `--yes` for non-interactive runs.
  - `Timestamp time.Time` — Invocation timestamp for ledger correlation.
- **Validation Rules**
  - Reject empty `InsertText`, path separators, or control characters.
  - Ensure `PositionToken` parses to a valid rune index relative to the stem (`^` = 0, `$` = len, positive 1-based, negative = offset from end).
  - Confirm resulting filename is non-empty and does not change extension semantics.
- **Relationships**
  - Consumed by insert engine to plan operations.
  - Serialized into ledger metadata with `InsertSummary`.

## Entity: InsertSummary
- **Fields**
  - `TotalCandidates int` — Items inspected after scope filtering.
  - `TotalChanged int` — Entries that will change after insertion.
  - `NoChange int` — Entries already containing target string at position (if applicable).
  - `Conflicts []Conflict` — Target collisions or invalid positions.
  - `Warnings []string` — Validation notices (duplicates, trimmed tokens, skipped hidden items).
  - `Entries []PreviewEntry` — Ordered original/proposed mappings with status.
  - `LedgerMetadata map[string]any` — Snapshot persisted with ledger entry (position, text, scope flags).
- **Validation Rules**
  - Conflicts must be empty before apply.
  - `TotalChanged + NoChange` equals count of entries with status `changed` or `no_change`.
  - Entries sorted deterministically by original path.
- **Relationships**
  - Emitted to preview renderer and ledger writer.
  - Input for undo verification.

## Entity: Conflict
- **Fields**
  - `OriginalPath string`
  - `ProposedPath string`
  - `Reason string` — (`duplicate_target`, `invalid_position`, `existing_file`, etc.)
- **Validation Rules**
  - `ProposedPath` unique among planned operations.
  - Reason restricted to known enum values for messaging.
- **Relationships**
  - Reported in preview output and used to block apply.

## Entity: PreviewEntry
- **Fields**
  - `OriginalPath string`
  - `ProposedPath string`
  - `Status string` — `changed`, `no_change`, `skipped`.
  - `InsertedText string` — Text segment inserted (for auditing).
- **Validation Rules**
  - `ProposedPath` equals `OriginalPath` when `Status == "no_change"`.
  - `InsertedText` empty only for `no_change` or `skipped`.
- **Relationships**
  - Displayed in preview output.
  - Persisted with ledger metadata for undo playback.
