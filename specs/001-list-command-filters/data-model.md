# Data Model: Cobra List Command with Global Filters

## Entities

### ListingRequest
- **Description**: Captures the user’s desired listing scope and presentation options.
- **Fields**:
  - `workingDir` (string): Absolute or relative path where traversal begins. Default: current dir.
  - `includeDirectories` (bool): Mirrors `-d` flag; when true, directories appear in results.
  - `recursive` (bool): Mirrors `-r` flag; enables depth-first traversal of subdirectories.
  - `extensions` ([]string): Parsed, normalized list of `.`-prefixed extensions from `-e`.
  - `format` (enum): Output format requested (`table`, `plain`).
  - `maxDepth` (int, optional): Future-safe guard to prevent runaway recursion; defaults to unlimited.
- **Validations**:
  - `extensions` MUST NOT contain empty strings or duplicates after normalization.
  - `format` MUST be one of the supported enum values; invalid values trigger validation errors.
  - `workingDir` MUST resolve to an accessible directory before traversal begins.

### ListingEntry
- **Description**: Represents a single filesystem node returned by the list command.
- **Fields**:
  - `path` (string): Relative path from `workingDir`.
  - `type` (enum): `file`, `directory`, or `symlink`.
  - `sizeBytes` (int64): File size in bytes; directories report aggregated size only if available.
  - `depth` (int): Depth level from the root directory (root = 0).
  - `matchedExtension` (string, optional): Extension that satisfied the filter when applicable.
- **Validations**:
  - `path` MUST be unique within a single command invocation.
  - `type` MUST align with actual filesystem metadata; symlinks MUST be flagged to avoid confusion.

## Relationships
- `ListingRequest` produces a stream of `ListingEntry` items based on traversal rules.
- `ListingEntry` items may reference parent directories implicitly via `path` hierarchy; no explicit
  parent pointer is stored to keep payload lightweight.

## State Transitions
1. **Initialization**: CLI parses flags into `ListingRequest` and validates inputs.
2. **Traversal**: Request feeds traversal engine, emitting raw filesystem metadata.
3. **Filtering**: Raw entries filtered against `includeDirectories`, `recursive`, and `extensions`.
4. **Formatting**: Filtered entries passed to output renderer selecting table or plain layout.
