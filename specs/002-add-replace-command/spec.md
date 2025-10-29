# Feature Specification: Replace Command with Multi-Pattern Support

**Feature Branch**: `002-add-replace-command`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "添加 替换（replace）命令，用于替换指定的字符串，支持同时对多个字符串进行替换为一个 示例： renamer replace pattern1 pattern2 with space pattern3 <replacement>"

## Clarifications

### Session 2025-10-29

- Q: What syntax separates patterns from the replacement value? → A: The final positional argument is the replacement; no delimiter keyword is used.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Normalize Naming with One Command (Priority: P1)

As a CLI user cleaning messy filenames, I want to run `renamer replace` with multiple source
patterns so I can normalize all variants to a single replacement in one pass.

**Why this priority**: Unlocks the main value proposition—batch consolidation of inconsistent
substrings without writing custom scripts.

**Independent Test**: Execute `renamer list` to confirm scope, then `renamer replace foo bar Baz`
and verify preview shows every `foo` and `bar` occurrence replaced with `Baz` before confirming.

**Acceptance Scenarios**:

1. **Given** files containing `draft`, `Draft`, and `DRAFT`, **When** the user runs `renamer replace draft Draft DRAFT final`, **Then** the preview lists every filename replaced with `final`, and execution applies the same mapping.
2. **Given** filenames where patterns appear multiple times, **When** the user runs `renamer replace beta gamma release`, **Then** each occurrence within a single filename is substituted, and the summary highlights total replacements per file.

---

### User Story 2 - Script-Friendly Replacement Workflows (Priority: P2)

As an operator embedding renamer in automation, I want deterministic exit codes and reusable
profiles for replace operations so scripts can preview, apply, and undo safely.

**Why this priority**: Ensures pipelines can rely on the command without interactive ambiguity.

**Independent Test**: In a CI script, run `renamer replace ... --dry-run` (preview), assert exit code
0, then run with `--yes` and check ledger entry plus `renamer undo` restores originals.

**Acceptance Scenarios**:

1. **Given** a non-interactive environment, **When** the user passes `--yes` to apply replacements after a successful preview, **Then** the command exits 0 on success and writes a ledger entry capturing all mappings.
2. **Given** an invalid argument (e.g., only one token supplied), **When** the command executes, **Then** it exits with a non-zero code and explains that the final positional argument becomes the replacement value.

---

### User Story 3 - Validate Complex Pattern Input (Priority: P3)

As a power user handling patterns with spaces or special characters, I want clear syntax guidance
and validation feedback so I can craft replacements confidently.

**Why this priority**: Prevents user frustration when patterns require quoting or escaping.

**Independent Test**: Run `renamer replace "Project X" "Project-X" ProjectX` and confirm preview
resolves both variants, while invalid quoting produces actionable guidance.

**Acceptance Scenarios**:

1. **Given** a pattern containing whitespace, **When** it is provided in quotes, **Then** the preview reflects the intended substring and documentation shows identical syntax describing the final argument as replacement.
2. **Given** duplicate patterns in the same command, **When** the command runs, **Then** duplicates are deduplicated with a warning so replacements remain idempotent.

---

### Edge Cases

- Replacement produces duplicate target filenames; command must surface conflicts before confirmation.
- Patterns overlap within the same filename (e.g., `aaa` replaced via `aa`); order of application must be deterministic and documented.
- Case-sensitivity opt-in: default literal matching is case-sensitive, but a future `--ignore-case` option is deferred; command must warn that matches are case-sensitive.
- Replacement string empty (aka deletion); preview must show empty substitution clearly and require explicit confirmation.
- No files match any pattern; command exits 0 with "No entries matched" messaging and no ledger entry.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST expose `renamer replace <pattern...> <replacement>` and treat all but the final token as literal patterns (quotes support whitespace).
- **FR-002**: Replace operations MUST integrate with the existing preview → confirm workflow; previews list old and new names with highlighted substrings.
- **FR-003**: Executions MUST append detailed entries to `.renamer` including original names, replacements performed, patterns supplied, and timestamps so undo remains possible.
- **FR-004**: Users MUST be able to undo the most recent replace batch via existing undo mechanics without orphaned files.
- **FR-005**: Command MUST respect global scope flags (`--recursive`, `--include-dirs`, `--hidden`, `--extensions`, `--path`) identical to `list` / `preview` behavior.
- **FR-006**: Command MUST accept two or more patterns and collapse them into a single replacement string applied to every filename in scope.
- **FR-007**: Command MUST preserve idempotence by deduplicating patterns and reporting the number of substitutions per pattern in summary output.
- **FR-008**: Invalid invocations (fewer than two arguments, empty pattern list after deduplication, missing replacement) MUST fail with exit code ≠0 and actionable usage guidance.
- **FR-009**: CLI help MUST document quoting rules for patterns containing spaces or shell special characters.

### Key Entities *(include if feature involves data)*

- **ReplaceRequest**: Captures working directory, scope flags, ordered pattern list, replacement string, and preview/apply options.
- **ReplaceSummary**: Aggregates per-pattern match counts, per-file changes, and conflict warnings for preview and ledger output.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users complete a multi-pattern replacement across 100 files with preview + apply in under 2 minutes end-to-end.
- **SC-002**: 95% of usability test participants interpret the final positional replacement argument correctly after reading `renamer replace --help`.
- **SC-003**: Automated regression tests confirm replace + undo leave the filesystem unchanged in 100% of scripted scenarios.
- **SC-004**: Support tickets related to manual string normalization drop by 40% within the first release cycle after launch.

## Assumptions

- Patterns are treated as literal substrings; regex support is out of scope for this increment.
- Case-sensitive matching aligns with current rename semantics; advanced matching can be added later.
- Replacement applies to filenames (and directories when `-d`/`--include-dirs` is set), not file contents.
- Existing infrastructure for preview, ledger, and undo can be extended without architectural changes.

## Dependencies & Risks

- Requires updates to shared traversal/filter utilities so replace respects global scope flags.
- Help/quickstart documentation must be updated alongside the command to prevent misuse.
- Potential filename conflicts must be detected pre-apply to avoid data loss; conflict handling relies on existing validation pipeline.
