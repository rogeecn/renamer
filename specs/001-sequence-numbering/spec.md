# Feature Specification: Sequence Numbering Command

**Feature Branch**: `001-sequence-numbering`  
**Created**: 2025-10-31  
**Status**: Draft  
**Input**: User description: "添加 sequence 功能，为重命名文件添加序号，可以定义序列号长度 使用0左填充，可以指定序列号开始序号"

## Clarifications

### Session 2025-11-03

- Q: How should the command behave when the generated filename already exists outside the current batch (e.g., `file_001.txt`)? → A: Skip conflicting files, continue, and warn.
- Q: How are sequence numbers applied when new files appear between preview and apply, potentially altering traversal order? → A: Ignore new files and rename only the previewed set.
- Q: What validation and messaging occur when the starting number, width, or separator arguments are invalid (negative numbers, zero width, multi-character separator conflicting with filesystem rules)? → A: Hard error with non-zero exit and no preview output.
- Q: How is numbering handled for directories when `--include-dirs` is used alongside files within the same traversal scope? → A: Do not rename directories; sequence applies to files only.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Add Sequential Indices to Batch (Priority: P1)

As a content manager preparing assets for delivery, I want to append an auto-incrementing number to each filename within my selected scope so that downstream systems receive files in a predictable order.

**Why this priority**: Sequencing multiple files is the primary value of the feature and removes the need for external renaming tools for common workflows such as media preparation or document packaging.

**Independent Test**: Place three files in a working directory, run `renamer sequence --path <dir> --dry-run`, and verify the preview shows `001_`, `002_`, `003_` prefixes in deterministic order. Re-run with `--yes` and confirm the ledger captures the batch.

**Acceptance Scenarios**:

1. **Given** a directory containing `draft.txt`, `notes.txt`, and `plan.txt`, **When** the user runs `renamer sequence --dry-run --path <dir>`, **Then** the preview lists each file renamed with a `001_` prefix in alphabetical order and reports the candidate totals.
2. **Given** the same directory and preview, **When** the user re-executes the command with `--yes`, **Then** the CLI reports three files updated and the `.renamer` ledger stores the sequence configuration.

---

### User Story 2 - Control Number Formatting (Priority: P2)

As an archivist following strict naming standards, I want to define the sequence width and zero padding so the filenames meet fixed-length requirements without additional scripts.

**Why this priority**: Formatting options broaden adoption by matching industry conventions (e.g., four-digit reels) and avoid manual corrections after renaming.

**Independent Test**: Run `renamer sequence --width 4 --path <dir> --dry-run` and confirm every previewed filename contains a four-digit, zero-padded sequence value.

**Acceptance Scenarios**:

1. **Given** files `cutA.mov` and `cutB.mov`, **When** the user runs `renamer sequence --width 4 --dry-run`, **Then** the preview shows `0001_` and `0002_` prefixes despite having only two files.
2. **Given** the same files, **When** the user omits an explicit width, **Then** the preview pads only as needed to accommodate the highest sequence number (e.g., `1_`, `2_`, `10_`).

---

### User Story 3 - Configure Starting Number and Placement (Priority: P3)

As a production coordinator resuming interrupted work, I want to choose the starting sequence value and whether the number appears as a prefix or suffix so I can continue existing numbering schemes without renaming older assets.

**Why this priority**: Starting offsets and placement control reduce rework when numbering must align with partner systems or previously delivered batches.

**Independent Test**: Run `renamer sequence --start 10 --dry-run` and confirm the preview begins at `010_` and inserts the number before the filename stem with the configured separator.

**Acceptance Scenarios**:

1. **Given** files `shotA.exr` and `shotB.exr`, **When** the user runs `renamer sequence --start 10 --dry-run`, **Then** the preview numbers the files starting at `010_` and `011_`.
2. **Given** files `cover.png` and `index.png`, **When** the user runs `renamer sequence --placement prefix --separator "-" --dry-run`, **Then** the preview shows names such as `001-cover.png` and `002-index.png` with the separator preserved.

---

### Edge Cases

- Generated filename conflicts with an existing filesystem entry outside the batch: skip the conflicting candidate, continue with the rest, and warn with conflict details.
- Requested width smaller than digits required is automatically expanded with a warning so numbering completes without truncation.
- New files encountered between preview and apply are ignored; only the previewed candidates are renamed.
- Invalid starting number, width, or separator arguments produce a hard error with non-zero exit status; no preview or apply runs until corrected.
- Directories included via `--include-dirs` are left unchanged; numbering applies exclusively to files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST expose a `sequence` command that applies an ordered numbering rule to all candidates within the current scope while preserving the preview-first workflow used by other renamer commands.
- **FR-002**: The command MUST use a deterministic ordering strategy (default: path-sorted traversal after scope filters) so preview and apply yield identical sequences.
- **FR-003**: Users MUST be able to configure the sequence starting value via a `--start` flag (default `1`) accepting positive integers only, with validation errors for invalid input.
- **FR-004**: Users MUST be able to configure the minimum digit width via a `--width` flag (default determined by total candidates) and the tool MUST zero-pad numbers to match the requested width.
- **FR-005**: Users MUST be able to choose number placement (`prefix` or `suffix`, default prefix) and optionally set a separator string plus static number prefix/suffix tokens while preserving file extensions and directory structure.
- **FR-006**: Preview output MUST display original and proposed names, total candidates, total changed, and warnings when numbering would exceed the requested width or create conflicts.
- **FR-007**: Apply MUST record the numbering rule (start, width, placement, separator, ordering) in the `.renamer` ledger, alongside per-file operations, so that undo can faithfully restore original names.
- **FR-008**: Undo MUST revert sequence-based renames in reverse order even if additional files have been added since the apply step, skipping only those already removed.
- **FR-009**: The `sequence` command MUST respect existing scope flags (`--path`, `--recursive`, `--include-dirs`, `--hidden`, `--extensions`, `--dry-run`, `--yes`) with identical semantics to other commands.
- **FR-010**: When numbering would collide with an existing filesystem entry, the CLI MUST skip the conflicting candidate, continue numbering the remaining files, and emit a warning that lists the skipped items; apply MUST still abort if scope filters yield zero candidates.
- **FR-011**: Invalid formatting arguments (negative start, zero/negative width, unsupported separator) MUST trigger a human-readable error, exit with non-zero status, and prevent preview/apply execution.
- **FR-012**: Directories included in scope via `--include-dirs` MUST be preserved without numbering; only files receive sequence numbers while directories remain untouched.

### Key Entities *(include if feature involves data)*

- **SequenceRequest**: Captures user-supplied configuration (start value, width, placement, separator, scope flags, execution mode).
- **SequencePlan**: Represents the ordered list of candidate files with assigned sequence numbers, proposed names, conflicts, and summary counts.
- **SequenceLedgerEntry**: Stores metadata required for undo, including request parameters, execution timestamp, and file rename mappings.

### Assumptions

- Ordering follows the preview list sorted by relative path unless future features introduce additional ordering controls.
- If numbering exceeds the requested width, the command extends the width automatically, surfaces a warning, and continues rather than failing the batch.
- Default placement is prefix with an underscore separator (e.g., `001_name.ext`) unless overridden by flags.
- Scope and ledger behavior mirror existing rename commands; no new traversal modes are introduced.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can number 500 files (preview + apply) in under 120 seconds on a representative workstation without manual intervention.
- **SC-002**: At least 95% of sampled previews match their subsequent apply results exactly during beta testing (no ordering drift or mismatched numbering).
- **SC-003**: 90% of beta participants report that numbering settings (start value, width, placement) meet their formatting needs without external tools.
- **SC-004**: Support requests related to manual numbering workflows decrease by 40% within one release cycle after launch.
