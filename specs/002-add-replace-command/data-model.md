# Data Model: Replace Command with Multi-Pattern Support

## Entities

### ReplaceRequest
- **Description**: Captures all inputs driving a replace operation.
- **Fields**:
  - `workingDir` (string, required): Absolute path where traversal begins.
  - `patterns` ([]string, min length 2): Ordered list of literal substrings to replace.
  - `replacement` (string, required): Literal value substituting each pattern.
  - `includeDirectories` (bool): Whether directories participate in replacement.
  - `recursive` (bool): Traverse subdirectories depth-first.
  - `includeHidden` (bool): Include hidden files during traversal.
  - `extensions` ([]string): Optional extension filters inherited from scope flags.
  - `dryRun` (bool): Preview mode flag; true during preview, false when applying changes.
- **Validations**:
  - `patterns` MUST be deduplicated case-sensitively before execution.
  - `replacement` MAY be empty, but command must warn the user during preview.
  - `workingDir` MUST exist and be readable before traversal.

### ReplaceSummary
- **Description**: Aggregates preview/apply outcomes for reporting and ledger entries.
- **Fields**:
  - `totalFiles` (int): Count of files/directories affected.
  - `patternMatches` (map[string]int): Total substitutions per pattern.
  - `conflicts` ([]ConflictDetail): Detected filename collisions with rationale.
  - `ledgerEntryID` (string, optional): Identifier once committed to `.renamer` ledger.

### ConflictDetail
- **Description**: Describes a file that could not be renamed due to collision or validation failure.
- **Fields**:
  - `originalPath` (string)
  - `proposedPath` (string)
  - `reason` (string): Human-readable description (e.g., "target already exists").

## Relationships
- `ReplaceRequest` produces a stream of candidate rename operations via traversal utilities.
- `ReplaceSummary` aggregates results from executing the request and is persisted inside ledger entries.
- `ConflictDetail` records subset of `ReplaceSummary` when collisions block application.

## State Transitions
1. **Parsing**: CLI args parsed into `ReplaceRequest`; validations run immediately.
2. **Preview**: Traversal + replacement simulation produce `ReplaceSummary` with proposed paths.
3. **Confirm**: Upon user confirmation (or `--yes`), operations apply atomically; ledger entry written.
4. **Undo**: Ledger reverse operation reads `ReplaceSummary` data to restore originals.
