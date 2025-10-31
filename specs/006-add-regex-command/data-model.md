# Data Model – Regex Command

## Entity: RegexRequest
- **Fields**
  - `WorkingDir string` — Absolute path derived from CLI `--path` or current directory.
  - `Pattern string` — User-supplied regular expression.
  - `Template string` — Replacement string with `@n` placeholders.
  - `IncludeDirs bool` — Mirrors `--include-dirs` flag.
  - `Recursive bool` — Mirrors `--recursive` flag.
  - `IncludeHidden bool` — True only when `--hidden` is supplied.
  - `ExtensionFilter []string` — Filter tokens from `--extensions`.
  - `DryRun bool` — Preview-only execution state.
  - `AutoConfirm bool` — Captures `--yes` for non-interactive runs.
  - `Timestamp time.Time` — Invocation timestamp for ledger correlation.
- **Validation Rules**
  - Regex must compile; invalid patterns produce errors.
  - Template may reference `@0` (full match) and numbered groups; referencing undefined groups is invalid.
  - Prohibit control characters and path separators in resulting names.
- **Relationships**
  - Consumed by regex engine to build rename plan.
  - Serialized into ledger metadata alongside summary output.

## Entity: RegexSummary
- **Fields**
  - `TotalCandidates int` — Items inspected after scope filtering.
  - `Matched int` — Files whose names matched the regex.
  - `Changed int` — Entries that will change after template substitution.
  - `Skipped int` — Non-matching or invalid-template entries.
  - `Conflicts []Conflict` — Rename collisions or generated duplicates.
  - `Warnings []string` — Validation notices (unused groups, truncated templates).
  - `Entries []PreviewEntry` — Original/proposed mappings with status.
  - `LedgerMetadata map[string]any` — Snapshot persisted with ledger entry (pattern, template, scope flags).
- **Validation Rules**
  - Conflicts must be empty before apply.
  - `Matched = Changed + (matched entries with no change)` for consistency.
- **Relationships**
  - Drives preview rendering.
  - Input for ledger writer and undo verification.

## Entity: Conflict
- **Fields**
  - `OriginalPath string`
  - `ProposedPath string`
  - `Reason string` — (`duplicate_target`, `existing_file`, `invalid_template`).
- **Validation Rules**
  - `ProposedPath` unique among planned operations.
  - Reason drawn from known enum for consistent messaging.
- **Relationships**
  - Reported in preview output and blocks apply.

## Entity: PreviewEntry
- **Fields**
  - `OriginalPath string`
  - `ProposedPath string`
  - `Status string` — `changed`, `no_change`, `skipped`.
  - `MatchGroups []string` — Captured groups applied to template.
- **Validation Rules**
  - `ProposedPath` equals `OriginalPath` when `Status == "no_change"`.
  - `MatchGroups` length must equal number of captured groups.
- **Relationships**
  - Displayed in preview output.
  - Persisted alongside ledger metadata for undo.
