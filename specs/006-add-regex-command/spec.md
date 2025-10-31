# Feature Specification: Regex Command for Pattern-Based Renaming

**Feature Branch**: `006-add-regex-command`  
**Created**: 2025-10-30  
**Status**: Draft  
**Input**: User description: "实现 regex 命令，用于使用正则获取指定位置内容后再重新命名，示例 renamer regexp <pattern> @1-@2 实现了获取正则的第一、二位的匹配数据，并进行重新命名"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rename Files Using Captured Groups (Priority: P1)

As a power user organizing datasets, I want to rename files by extracting portions of their names via regular expressions so that I can normalize naming schemes without writing custom scripts.

**Why this priority**: Provides the core value—regex-driven renaming to rearrange captured data quickly across large batches.

**Independent Test**: In a directory with files named `2025-01_report.txt` and `2025-02_report.txt`, run `renamer regex "^(\\d{4})-(\\d{2})_report" "Q@2-@1" --dry-run` and verify the preview shows `Q01-2025.txt` and `Q02-2025.txt`. Re-run with `--yes` to confirm filesystem updates and ledger entry.

**Acceptance Scenarios**:

1. **Given** files `alpha-123.log` and `beta-456.log`, **When** the user runs `renamer regex "^(\\w+)-(\\d+)" "@2_@1" --dry-run`, **Then** the preview lists `123_alpha.log` and `456_beta.log` as proposed names.
2. **Given** files that do not match the pattern, **When** the command runs in preview mode, **Then** unmatched files are listed with a "skipped" status and no filesystem changes occur on apply.

---

### User Story 2 - Automation-Friendly Regex Renames (Priority: P2)

As a DevOps engineer automating release artifact naming, I need deterministic exit codes, ledger metadata, and undo support for regex-based renames so CI pipelines remain auditable and reversible.

**Why this priority**: Ensures the new command can be safely adopted in automation without risking opaque failures.

**Independent Test**: Execute `renamer regex "^build_(\\d+)_(.*)$" "release-@1-@2" --yes --path ./fixtures`, verify exit code `0`, inspect `.renamer` for recorded pattern, replacement template, and affected files, then run `renamer undo` to restore originals.

**Acceptance Scenarios**:

1. **Given** a non-interactive run with `--yes`, **When** all matches succeed without conflicts, **Then** exit code is `0` and the ledger entry records the regex pattern, replacement template, and matching groups per file.
2. **Given** a ledger entry produced by `renamer regex`, **When** `renamer undo` executes, **Then** filenames revert to their previous values even if the original files contained Unicode characters or were renamed by automation.

---

### User Story 3 - Validate Patterns, Placeholders, and Conflicts (Priority: P3)

As a user experimenting with regex templates, I want clear validation and preview feedback for invalid patterns, missing capture groups, or resulting conflicts so I can adjust commands before committing changes.

**Why this priority**: Prevents accidental data loss and reduces trial-and-error when constructing regex commands.

**Independent Test**: Run `renamer regex "^(.*)$" "@2" --dry-run` and confirm the command exits with a descriptive error because placeholder `@2` is undefined; run a scenario where multiple files would map to the same name and ensure apply is blocked.

**Acceptance Scenarios**:

1. **Given** a replacement template referencing an undefined capture group, **When** the command runs, **Then** it exits non-zero with a message explaining the missing group and no files change.
2. **Given** two files whose matches produce identical targets, **When** preview executes, **Then** conflicts are listed and apply refuses to proceed until resolved.

---

### Edge Cases

- How does the command behave when the regex pattern is invalid or cannot compile?
- What is the outcome when no files match the pattern (preview and apply)?
- How are nested or optional groups handled when placeholders reference non-matching groups?
- What happens if the replacement template results in empty filenames or removes extensions?
- How are directories or hidden files treated when scope flags include/exclude them?
- What feedback is provided when resulting names differ only by case on case-insensitive filesystems?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: CLI MUST provide a `regex` subcommand that accepts a required regex pattern and replacement template arguments (e.g., `renamer regex <pattern> <template>`).
- **FR-002**: Replacement templates MUST support numbered capture placeholders (`@1`, `@2`, etc.) corresponding to the regex groups; referencing undefined groups MUST produce a validation error.
- **FR-003**: Pattern matching MUST operate on the filename stem by default while preserving extensions unless the template explicitly alters them.
- **FR-004**: Preview MUST display original names, proposed names, and highlight skipped entries (unmatched, invalid template) prior to apply; apply MUST be blocked when conflicts or validation errors exist.
- **FR-005**: Execution MUST respect shared scope flags (`--path`, `--recursive`, `--include-dirs`, `--hidden`, `--extensions`, `--dry-run`, `--yes`) consistent with other commands.
- **FR-006**: Ledger entries MUST capture the regex pattern, replacement template, and affected files so undo can restore originals deterministically.
- **FR-007**: The command MUST emit deterministic exit codes: `0` for successful apply or no matches, non-zero for validation failures or conflicts.
- **FR-008**: Help output MUST document pattern syntax expectations, placeholder usage, escaping rules, and examples for both files and directories.

### Key Entities

- **RegexRequest**: Working directory, regex pattern, replacement template, scope flags, dry-run/apply settings.
- **RegexSummary**: Counts of matched files, skipped entries, conflicts, warnings, and preview entries with status (`changed`, `skipped`, `no_change`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users rename 500 files via regex (preview + apply) in under 2 minutes end-to-end.
- **SC-002**: 95% of beta testers correctly apply a regex rename after reading `renamer regex --help` without additional guidance.
- **SC-003**: Automated regression tests confirm regex rename + undo cycles leave the filesystem unchanged in 100% of scripted scenarios.
- **SC-004**: Support tickets related to custom regex renaming scripts drop by 30% within the first release cycle post-launch.

## Clarifications

### Session 2025-10-30
- Q: How should literal @ characters be escaped in templates? → A: Use @@ to emit a literal @ while keeping numbered placeholders intact.

## Assumptions

- Regex evaluation uses the runtime’s built-in engine with RE2-compatible syntax; no backtracking-specific constructs (e.g., look-behind) are supported.
- Matching applies to filename stems by default; users can reconstruct extensions via placeholders if required.
- Unmatched files are skipped gracefully and reported in preview; apply exits `0` when all files are skipped.
- Templates treat `@0` as the entire match if referenced; placeholders are case-sensitive and must be preceded by `@`. Use `@@` to emit a literal `@` character.

## Dependencies & Risks

- Requires extending existing traversal, preview, and ledger infrastructure to accommodate regex replacement logic.
- Complex regex patterns may produce unexpected duplicates; conflict detection must guard against accidental overwrites.
- Users may expect advanced regex features (named groups, non-ASCII classes); documentation must clarify supported syntax to prevent confusion.
