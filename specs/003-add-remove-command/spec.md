# Feature Specification: Remove Command with Sequential Multi-Pattern Support

**Feature Branch**: `003-add-remove-command`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "添加 移除（Remove）命令，用于删除指定字符、字符串，支持同时删除多个字符串，示例：renamer remove str1 str2 ....，注意：多个移除时后续参数的移除依赖于前一个移除后的结果，移除计算完成前不进行重命名，避免IO负载过高"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remove Unwanted Tokens in One Pass (Priority: P1)

As a CLI user tidying filenames, I want `renamer remove` to delete multiple substrings in order so I
can normalize file names without writing custom scripts.

**Why this priority**: Delivers the core value—batch cleanup of recurring tokens across many files
with predictable results.

**Independent Test**: Run `renamer remove " copy" " draft" --dry-run` in a sample directory,
confirm preview shows the ordered removal effects, then apply with `--yes` and verify names update
accordingly.

**Acceptance Scenarios**:

1. **Given** files `report copy draft.txt` and `notes draft.txt`, **When** the user runs
   `renamer remove " copy" " draft"`, **Then** the preview shows both tokens removed sequentially
   and execution produces `report.txt` and `notes.txt`.
2. **Given** patterns where later removals depend on earlier results (e.g., removing `foo` then `foo-`),
   **When** the command runs, **Then** each removal applies to the output of the previous step before
   computing rename conflicts.

---

### User Story 2 - Script-Friendly Removal Workflow (Priority: P2)

As an operator automating rename tasks, I want deterministic previews, exit codes, and ledger entries
so scripts can run `renamer remove` safely without interactive prompts.

**Why this priority**: Ensures automation pipelines can rely on the same safety guarantees as manual
runs.

**Independent Test**: In a CI script, call `renamer remove ... --dry-run`, assert exit code 0, then
run with `--yes` and verify ledger entry plus `renamer undo` restores originals.

**Acceptance Scenarios**:

1. **Given** a non-interactive context, **When** the user passes `--yes` after a successful preview,
   **Then** the command exits 0 on success and writes a ledger entry capturing tokens removed per file.
2. **Given** invalid input (e.g., fewer than two arguments), **When** the command executes, **Then** it
   exits with non-zero status and instructs the user on correct sequential argument usage.

---

### User Story 3 - Validate Sequential Removal Inputs (Priority: P3)

As a power user managing complex token lists, I want clear validation and guidance for spaces,
duplicate tokens, and results that could produce empty filenames so I can adjust before applying.

**Why this priority**: Prevents surprise failures when tokens overlap or yield empty names.

**Independent Test**: Run `renamer remove "Project X" " Project" "X" --dry-run`, confirm preview
shows the sequential impact and warns if names collapse; invalid quoting should produce actionable
errors.

**Acceptance Scenarios**:

1. **Given** duplicate tokens, **When** the command runs, **Then** duplicates are deduplicated with a
   warning and order preserved for remaining unique tokens.
2. **Given** a removal sequence that would produce an empty basename, **When** the preview runs,
   **Then** the command warns and excludes the rename unless the user overrides in a future version.

---

### Edge Cases

- Sequential removals should operate on the output of prior removals within the same filename.
- Removing tokens may collapse names to empty strings or leave trailing separators; preview must flag
  these cases before apply.
- Resulting names may collide; conflicts must be reported before confirmation.
- Hidden files or directories may need inclusion depending on scope flags (`--hidden`).
- No matches should result in a friendly "No entries matched" message and exit code 0 without ledger
  writes.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST expose `renamer remove <pattern1> [pattern2 ...]` where all tokens are
  literal substrings removed sequentially; final name computation MUST complete before any renames
  occur.
- **FR-002**: Preview → confirm workflow MUST mirror existing commands, listing original and proposed
  names with highlighted removals.
- **FR-003**: Executions MUST append detailed entries to `.renamer` including original names, tokens
  removed (with order), resulting names, and timestamps so undo remains possible.
- **FR-004**: Users MUST be able to undo the most recent remove batch via existing undo mechanics
  without leaving orphaned files.
- **FR-005**: Command MUST respect global scope flags (`--path`, `--recursive`, `--include-dirs`,
  `--hidden`, `--extensions`, `--dry-run`, `--yes`) identical to `list` / `replace` behavior.
- **FR-006**: Preview MUST evaluate all removals first, calculate conflicts, and only then apply
  filesystem operations when confirmed, limiting IO load.
- **FR-007**: Command MUST warn (and skip) renames that would result in empty basenames unless a
  future explicit override flag is provided.
- **FR-008**: Invalid invocations (fewer than two arguments, empty tokens after trimming) MUST fail
  with exit code ≠0 and actionable usage guidance.
- **FR-009**: Help output MUST document sequential behavior, whitespace quoting, and interaction with
  other scope flags.

### Key Entities

- **RemoveRequest**: Captures working directory, scope flags, ordered token list, and preview/apply
  options.
- **RemoveSummary**: Aggregates per-token match counts, per-file outcomes, conflicts, and warnings for
  preview and ledger output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users complete a sequential removal across 100 files (preview + apply) in under 2
  minutes end-to-end.
- **SC-002**: 95% of usability test participants correctly understand that removals execute in the
  provided order after reading `renamer remove --help`.
- **SC-003**: Automated regression tests confirm remove + undo leave the filesystem unchanged in
  100% of scripted scenarios.
- **SC-004**: Support requests related to manual substring cleanup drop by 35% within the first release
  cycle after launch.

## Assumptions

- Removals are literal substring matches; regex or wildcard support is out of scope for this release.
- Default matching is case-sensitive; case-insensitive options can be considered later if needed.
- Delete operations target filenames (and directories when `-d/--include-dirs` is set), not file
  contents.
- Existing traversal, conflict detection, and ledger infrastructure can be extended for the remove
  command.

## Dependencies & Risks

- Requires new remove-specific packages analogous to replace to maintain modularity.
- Help/quickstart documentation must be updated to explain sequential removal behavior.
- Potential filename conflicts or empty results must be detected pre-apply to avoid data loss.
