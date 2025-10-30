# Feature Specification: Extension Command for Multi-Extension Normalization

**Feature Branch**: `004-extension-rename`
**Created**: 2025-10-29
**Status**: Draft
**Input**: User description: "实现扩展名修改（Extension）命令，类似于 replace 命令，可以支持把多个扩展名更改为一个指定的扩展名"

## Clarifications

### Session 2025-10-30
- Q: Should extension comparisons treat casing uniformly or follow the host filesystem? → A: Always case-insensitive
- Q: How should hidden files be handled when `--hidden` is omitted? → A: Exclude hidden entries
- Q: What exit behavior should occur when no files match the given extensions? → A: Exit 0 with notice
- Q: How should files that already have the target extension be represented? → A: Preview as no-change; skip on apply

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Normalize Legacy Extensions in Bulk (Priority: P1)

As a power user cleaning project assets, I want a command to replace multiple file extensions with a single standardized extension so that I can align legacy files (e.g., `.jpeg`, `.JPG`) without hand-editing each one.

**Why this priority**: Delivers the core business value—fast extension normalization across large folders.

**Independent Test**: In a sample directory containing `.jpeg`, `.JPG`, and `.png`, run `renamer extension .jpeg .JPG .png .jpg --dry-run`, verify preview shows the new `.jpg` extension for each, then apply with `--yes` and confirm filesystem updates.

**Acceptance Scenarios**:

1. **Given** files with extensions `.jpeg` and `.JPG`, **When** the user runs `renamer extension .jpeg .JPG .jpg`, **Then** the preview lists each file with the `.jpg` extension and apply renames successfully.
2. **Given** nested directories, **When** the user adds `--recursive`, **Then** all matching extensions in subdirectories are normalized while unrelated files remain untouched.

---

### User Story 2 - Automation-Friendly Extension Updates (Priority: P2)

As an operator integrating renamer into CI scripts, I want deterministic exit codes, ledger entries, and undo support when changing extensions so automated jobs remain auditable and recoverable.

**Why this priority**: Ensures enterprise workflows can rely on extension updates without risking data loss.

**Independent Test**: Run `renamer extension .yaml .yml .yml --yes --path ./fixtures`, verify exit code `0`, inspect `.renamer` ledger for recorded operations, and confirm `renamer undo` restores originals.

**Acceptance Scenarios**:

1. **Given** a non-interactive environment with `--yes`, **When** the command completes without conflicts, **Then** exit code is `0` and ledger metadata captures original extension list and target extension.
2. **Given** a ledger entry exists, **When** `renamer undo` runs, **Then** all files revert to their prior extensions even if the command was executed by automation.

---

### User Story 3 - Validate Extension Inputs and Conflicts (Priority: P3)

As a user preparing an extension migration, I want validation and preview warnings for invalid tokens, duplicate target names, and no-op operations so I can adjust before committing changes.

**Why this priority**: Reduces support load from misconfigured commands and protects against accidental overwrites.

**Independent Test**: Run `renamer extension .mp3 .MP3 mp3 --dry-run`, confirm validation fails because tokens must include leading `.`, and run a scenario where resulting filenames collide to ensure conflicts abort the apply step.

**Acceptance Scenarios**:

1. **Given** invalid input (e.g., missing leading `.` or fewer than two arguments), **When** the command executes, **Then** it exits with non-zero status and prints actionable guidance.
2. **Given** two files that would become the same path after extension normalization, **When** the preview runs, **Then** conflicts are listed and the apply step refuses to proceed until resolved.

---

### Edge Cases

- How does the rename plan surface conflicts when multiple files map to the same normalized extension?
- When the target extension already matches some files, they appear in preview with a “no change” indicator and are skipped during apply without raising errors.
- Hidden files and directories remain excluded unless the user supplies `--hidden`.
- When no files match in preview or apply, the command must surface a “no candidates found” notice while completing successfully.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: CLI MUST provide a dedicated `extension` subcommand that accepts one or more source extensions and a single target extension as positional arguments.
- **FR-002**: Preview → confirm workflow MUST mirror existing commands: list original paths, proposed paths, and highlight extension changes before apply.
- **FR-003**: Executions MUST append ledger entries capturing original extension list, target extension, affected files, and timestamps to support undo.
- **FR-004**: Users MUST be able to undo the most recent extension batch via existing undo mechanics without leaving orphaned files.
- **FR-005**: Command MUST respect global scope flags (`--path`, `--recursive`, `--include-dirs`, `--hidden`, `--extensions`, `--dry-run`, `--yes`) consistent with `list` and `replace`, excluding hidden files and directories unless `--hidden` is explicitly supplied.
- **FR-006**: Extension parsing MUST require leading `.` tokens, deduplicate case-insensitively, warn when duplicates or no-op tokens are supplied, and compare file extensions case-insensitively across all platforms.
- **FR-007**: Preview MUST detect target conflicts (two files mapping to the same new path) and block apply until conflicts are resolved.
- **FR-008**: Invalid invocations (e.g., fewer than two arguments, empty tokens after trimming) MUST exit with non-zero status and provide remediation tips.
- **FR-009**: Help output MUST clearly explain argument order, sequential evaluation rules, and interaction with scope flags.
- **FR-010**: When no files match the provided extensions, preview and apply runs MUST emit a clear “no candidates found” message and exit with status `0`.
- **FR-011**: Preview MUST surface already-targeted files with a “no change” marker, and apply MUST skip them while returning success.

### Key Entities

- **ExtensionRequest**: Working directory, ordered source extension list, target extension, scope flags, dry-run/apply settings.
- **ExtensionSummary**: Totals for candidates, changed files, per-extension match counts, conflicts, and warning messages used for preview and ledger metadata.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users normalize 500 files’ extensions (preview + apply) in under 2 minutes end-to-end.
- **SC-002**: 95% of beta testers correctly supply arguments (source extensions + target) after reading `renamer extension --help` without additional guidance.
- **SC-003**: Automated regression tests confirm extension change + undo leave the filesystem unchanged in 100% of scripted scenarios.
- **SC-004**: Support tickets related to inconsistent file extensions drop by 30% within the first release cycle after launch.

## Assumptions

- Command name will be `renamer extension` to align with existing verb-noun conventions.
- Source extensions are literal matches with leading dots; wildcard or regex patterns remain out of scope.
- Target extension must include a leading dot and is applied case-sensitively as provided.
- Existing traversal, summary, and ledger infrastructure can be extended from the replace/remove commands.

## Dependencies & Risks

- Requires new extension-specific packages analogous to replace/remove for parser, engine, summary, and CLI wiring.
- Help/quickstart documentation must be updated to explain argument order and extension validation.
- Potential filename conflicts after normalization must be detected pre-apply to avoid overwriting files.
