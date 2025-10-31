# Implementation Plan: Regex Command for Pattern-Based Renaming

**Branch**: `006-add-regex-command` | **Date**: 2025-10-30 | **Spec**: `specs/006-add-regex-command/spec.md`
**Input**: Feature specification from `/specs/006-add-regex-command/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Deliver a `renamer regex` subcommand that compiles a user-supplied pattern, substitutes numbered capture groups into a replacement template, surfaces deterministic previews, and records ledger metadata so undo and automation workflows remain safe and auditable.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, Go `regexp` (RE2 engine), internal traversal/history/output packages  
**Storage**: Local filesystem and `.renamer` ledger files  
**Testing**: `go test ./...`, contract suites under `tests/contract`, integration flows under `tests/integration`, targeted smoke script  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows shells)  
**Project Type**: Single CLI project (`cmd/`, `internal/`, `tests/`, `scripts/`)  
**Performance Goals**: Preview + apply 500 regex-driven renames in <2 minutes end-to-end  
**Constraints**: Preview-first confirmation, reversible ledger entries, Unicode-safe regex evaluation, conflict detection before apply  
**Scale/Scope**: Expected to operate on thousands of entries per invocation within local directories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety). ✅ Use shared preview renderer to list original → proposed names plus skipped/conflict indicators prior to apply.
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger). ✅ Append ledger entries containing pattern, template, captured groups per file, enabling `renamer undo` to restore originals.
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine). ✅ Build a dedicated regex rule that compiles patterns, validates templates, and plugs into traversal pipeline without altering shared state.
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal). ✅ Reuse traversal filters so regex respects directory, recursion, hidden, and extension flags identically to other commands.
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship). ✅ Add Cobra subcommand with documented flags, examples, help output, and contract/integration coverage for preview/apply/undo flows.

*Post-Design Verification (2025-10-30): Research, data model, contracts, and quickstart documents confirm preview coverage, ledger metadata, regex template validation, and CLI UX updates — no gate violations detected.*

## Project Structure

### Documentation (this feature)

```text
specs/006-add-regex-command/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md   # Generated via /speckit.tasks
```

### Source Code (repository root)

```text
cmd/
├── root.go
├── list.go
├── replace.go
├── remove.go
├── extension.go
├── insert.go
├── regex.go          # NEW
└── undo.go

internal/
├── filters/
├── history/
├── listing/
├── output/
├── remove/
├── replace/
├── extension/
├── insert/
└── regex/            # NEW: pattern compilation, template evaluation, engine, ledger metadata

tests/
├── contract/
├── integration/
├── fixtures/
└── unit/

scripts/
├── smoke-test-list.sh
├── smoke-test-replace.sh
├── smoke-test-remove.sh
├── smoke-test-extension.sh
├── smoke-test-insert.sh
└── smoke-test-regex.sh   # NEW
```

**Structure Decision**: Extend the single CLI project by introducing `cmd/regex.go`, a new `internal/regex` package for rule evaluation, and corresponding contract/integration tests plus a smoke script under existing directories.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
