# Data Model: Remove Command with Sequential Multi-Pattern Support

## Entities

### RemoveRequest
- **Description**: Captures inputs driving a remove operation.
- **Fields**:
  - `workingDir` (string): Absolute path where traversal begins.
  - `tokens` ([]string): Ordered list of literal substrings to remove sequentially.
  - `includeDirectories` (bool): Whether directory names participate.
  - `recursive` (bool): Whether to traverse subdirectories.
  - `includeHidden` (bool): Include hidden files/directories when true.
  - `extensions` ([]string): Optional extension filters inherited from global scope flags.
  - `dryRun` (bool): Preview flag; true during preview, false for apply.
- **Validations**:
  - `tokens` MUST contain at least one non-empty string after trimming.
  - `workingDir` MUST exist and be readable prior to traversal.
  - `tokens` are deduplicated case-sensitively but order of first occurrence preserved.

### RemoveSummary
- **Description**: Aggregates preview/apply outcomes for reporting and ledger.
- **Fields**:
  - `totalCandidates` (int): Count of files/directories evaluated.
  - `changedCount` (int): Count of items whose names change after removals.
  - `tokenMatches` (map[string]int): Number of occurrences removed per token (ordered in ledger metadata).
  - `conflicts` ([]ConflictDetail): Detected rename conflicts preventing apply.
  - `empties` ([]string): Relative paths where removal would lead to empty basename (skipped).

### ConflictDetail
- **Description**: Captures rename conflicts detected during preview.
- **Fields**:
  - `originalPath` (string)
  - `proposedPath` (string)
  - `reason` (string)

## Relationships
- `RemoveRequest` feeds traversal utilities to produce candidate names.
- `RemoveSummary` aggregates results from sequential removal engine and is persisted to ledger entries.
- `ConflictDetail` entries inform preview output and determine which renames are skipped.

## State Transitions
1. **Parse**: CLI arguments parsed into `RemoveRequest`; validations ensure tokens and scope are valid.
2. **Preview**: Sequential removal engine produces proposed names, conflicts, and warnings recorded in
   `RemoveSummary`.
3. **Apply**: Upon confirmation/`--yes`, renames execute (in dependency order), ledger entry written
   with ordered token metadata.
4. **Undo**: Ledger reverse operation uses stored operations to restore original names.
