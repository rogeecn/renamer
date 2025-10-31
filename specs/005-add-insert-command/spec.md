# Feature Specification: Insert Command for Positional Text Injection

**Feature Branch**: `005-add-insert-command`  
**Created**: 2025-10-30  
**Status**: Draft  
**Input**: User description: "实现插入（Insert）字符，支持在文件名指定位置插入指定字符串数据，示例 renamer insert <position> <string>, position:可以包含 ^：开头、$：结尾、 正数：字符位置、使用 N$ 表示倒数第 N 个字符。！！重要：需要考虑中文字符"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Insert Text at Target Position (Priority: P1)

As a power user organizing media files, I want to insert a label into filenames at a specific character position so that I can batch-tag assets without manually editing each name.

**Why this priority**: Provides the core value proposition of the insert command—precise, repeatable filename updates that accelerate organization tasks.

**Independent Test**: In a sample directory with mixed filenames, run `renamer insert 3 _tag --dry-run` and verify the preview shows the marker inserted at the third character (Unicode-aware). Re-run with `--yes` to confirm filesystem changes and ledger entry.

**Acceptance Scenarios**:

1. **Given** files named `项目A报告.docx` and `项目B报告.docx`, **When** the user runs `renamer insert ^ 2025- --dry-run`, **Then** the preview lists `2025-项目A报告.docx` and `2025-项目B报告.docx`.
2. **Given** files named `holiday.jpg` and `trip.jpg`, **When** the user runs `renamer insert 1$ X --yes`, **Then** the apply step inserts `X` before the last character of each base name while preserving extensions.

---

### User Story 2 - Automation-Friendly Batch Inserts (Priority: P2)

As an operator running renamer in CI, I need deterministic exit codes, ledger metadata, and undo support when inserting text so scripted jobs remain auditable and reversible.

**Why this priority**: Ensures production and automation workflows can rely on the new command without risking data loss.

**Independent Test**: Execute `renamer insert $ _ARCHIVE --yes --path ./fixtures`, verify exit code `0`, inspect the latest `.renamer` entry for recorded positions/string, then run `renamer undo` to restore originals.

**Acceptance Scenarios**:

1. **Given** a non-interactive run with `--yes`, **When** insertion completes without conflicts, **Then** exit code is `0` and the ledger stores the original names, position rule, inserted text, and timestamps.
2. **Given** a ledger entry produced by `renamer insert`, **When** `renamer undo` executes, **Then** all affected files revert to their exact previous names even across locales.

---

### User Story 3 - Validate Positions and Multilingual Inputs (Priority: P3)

As a user preparing filenames with multilingual characters, I want validation and preview warnings for invalid positions, overlapping results, or unsupported encodings so I can adjust commands before committing changes.

**Why this priority**: Protects against data corruption, especially with Unicode characters where byte counts differ from visible characters.

**Independent Test**: Run `renamer insert 50 _X --dry-run` on files shorter than 50 code points and confirm the command exits with a non-zero status explaining the out-of-range index; validate that Chinese filenames are handled correctly in previews.

**Acceptance Scenarios**:

1. **Given** an index larger than the filename length, **When** the command runs, **Then** it fails with a descriptive error and no filesystem changes occur.
2. **Given** an insertion that would create duplicate names, **When** preview executes, **Then** conflicts are surfaced and apply is blocked until resolved.

---

### Edge Cases

- What happens when the requested position is outside the filename length (positive or negative)?
- How are filenames handled when inserting before/after Unicode characters or surrogate pairs?
- How are directories or hidden files treated when `--include-dirs` or `--hidden` is omitted or provided?
- What feedback is provided when insertion would produce duplicate targets or empty names?
- How does the command behave when the inserted string is empty, whitespace-only, or contains path separators?
- What occurs when multiple files differ only by case and the insert results in conflicting targets on case-insensitive filesystems?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST provide a dedicated `insert` subcommand accepting a positional argument (`^`, `$`, forward indexes like `3`/`^3`, or backward indexes like `1$`) and the string to insert.
- **FR-002**: Insert positions MUST be interpreted using Unicode code points on the filename stem (excluding extension) so multi-byte characters count as a single position.
- **FR-003**: The command MUST support scope flags (`--path`, `--recursive`, `--include-dirs`, `--hidden`, `--extensions`, `--dry-run`, `--yes`) consistent with existing commands.
- **FR-004**: Preview MUST display original and proposed names, highlighting inserted segments, and block apply when conflicts or invalid positions are detected.
- **FR-005**: Apply MUST update filesystem entries atomically per batch, record operations in the `.renamer` ledger with inserted string, position rule, affected files, and timestamps, and support undo.
- **FR-006**: Validation MUST reject positions outside the allowable range, empty insertion strings (unless explicitly allowed), and inputs containing path separators or control characters.
- **FR-007**: Help output MUST describe position semantics (`^`, `$`, forward indexes, backward suffix tokens such as `N$`), Unicode handling, and examples for both files and directories.
- **FR-008**: Automation runs with `--yes` MUST emit deterministic exit codes (`0` success, non-zero on validation/conflicts) and human-readable messages that can be parsed for errors.

### Key Entities

- **InsertRequest**: Working directory, position token, insertion string, scope flags, dry-run/apply mode.
- **InsertSummary**: Counts of processed items, per-position match details, conflicts, warnings, and preview entries with status (`changed`, `no_change`, `skipped`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users insert text into 500 filenames (preview + apply) in under 2 minutes end-to-end.
- **SC-002**: 95% of beta users correctly apply a positional insert after reading `renamer insert --help` without additional guidance.
- **SC-003**: Automated regression tests confirm insert + undo cycles leave the filesystem unchanged in 100% of scripted scenarios.
- **SC-004**: Support tickets related to manual filename labeling drop by 30% within the first release cycle post-launch.

## Assumptions

- Positions operate on the filename stem (path excluded, extension preserved); inserting at `$` occurs immediately before the extension dot when present.
- Empty insertion strings are treated as invalid to avoid silent no-ops; users must provide at least one visible character.
- Unicode normalization is assumed to be NFC; filenames are treated as sequences of Unicode code points using the runtime’s native string handling.

## Dependencies & Risks

- Requires reuse and possible extension of existing traversal, preview, and ledger infrastructure to accommodate positional operations.
- Accurate Unicode handling depends on the runtime’s Unicode utilities; additional testing may be necessary for combining marks and surrogate pairs.
- Insertion near filesystem path separators must be restricted to avoid creating invalid paths or escape sequences.
