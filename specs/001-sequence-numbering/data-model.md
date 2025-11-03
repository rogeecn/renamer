# Data Model: Sequence Numbering Command

## SequenceRequest
- **Fields**
  - `Path` (string): Root path for traversal (defaults to current directory); must exist and be accessible.
  - `Recursive` (bool): Includes subdirectories when true.
  - `IncludeDirs` (bool): Includes directories in traversal results without numbering.
  - `Hidden` (bool): Includes hidden files when true.
  - `Extensions` ([]string): Optional `.`-prefixed extension filter; deduplicated and validated.
  - `DryRun` (bool): Indicates preview-only execution.
  - `Yes` (bool): Confirmation flag for apply mode.
  - `Start` (int): First sequence value; must be ≥1.
  - `Width` (int): Minimum digits for zero padding; must be ≥1 when provided.
  - `Placement` (enum: `prefix` | `suffix`): Determines where the sequence number is inserted; default `suffix`.
  - `Separator` (string): Separator between number and filename; defaults to `_`; must comply with filesystem rules (no path separators, non-empty).
- **Relationships**: Consumed by traversal service to produce candidates and by sequence rule to generate `SequencePlan`.
- **Validations**: Numeric fields validated before preview; separator sanitized; conflicting flags (dry-run vs yes) rejected.

## SequencePlan
- **Fields**
  - `Candidates` ([]SequenceCandidate): Ordered list of files considered for numbering.
  - `SkippedConflicts` ([]SequenceConflict): Files skipped due to target path collisions.
  - `Summary` (SequenceSummary): Counts for total candidates, renamed files, and skipped items.
  - `Config` (SequenceConfig): Snapshot of sequence settings (start, width, placement, separator).
- **Relationships**: Passed to output package for preview rendering and to history package for ledger persistence.
- **Validations**: Candidate ordering must match traversal ordering; conflicts identified before apply.

### SequenceCandidate
- **Fields**
  - `OriginalPath` (string)
  - `ProposedPath` (string)
  - `Index` (int): Zero-based position used to derive padded number.
- **Constraints**: Proposed path must differ from original to be considered a rename; duplicates flagged as conflicts.

### SequenceConflict
- **Fields**
  - `OriginalPath` (string)
  - `ConflictingPath` (string)
  - `Reason` (string enum: `existing_target`, `invalid_separator`, `width_overflow`)

### SequenceSummary
- **Fields**
  - `TotalCandidates` (int)
  - `RenamedCount` (int)
  - `SkippedCount` (int)
  - `Warnings` ([]string)

## SequenceLedgerEntry
- **Fields**
  - `Timestamp` (time.Time)
  - `Rule` (string): Fixed value `sequence`.
  - `Config` (SequenceConfig): Stored to support undo.
  - `Operations` ([]SequenceOperation): Each captures `From` and `To` paths actually renamed.
- **Relationships**: Append-only entry written by history package; consumed by undo command.
- **Validations**: Only include successful renames (skipped conflicts omitted). Undo must verify files still exist before attempting reversal.

### SequenceOperation
- **Fields**
  - `From` (string)
  - `To` (string)

## SequenceConfig
- **Fields**
  - `Start` (int)
  - `Width` (int)
  - `Placement` (string)
  - `Separator` (string)
- **Usage**: Embedded in plan summaries, ledger entries, and undo operations to ensure consistent behavior across preview and apply.
