# Implementation Plan: Remove Command with Sequential Multi-Pattern Support

**Branch**: `003-add-remove-command` | **Date**: 2025-10-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-add-remove-command/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Introduce a `renamer remove` subcommand that deletes multiple literal substrings sequentially while
respecting preview → confirm → ledger → undo guarantees. The command must evaluate all removals in
memory before touching the filesystem to avoid excessive IO and ensure conflict detection happens on
final names.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`  
**Storage**: Local filesystem only (ledger persisted as `.renamer`)  
**Testing**: Go `testing` package with contract, integration, and unit tests  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: Single CLI project  
**Performance Goals**: Preview + apply for 100 files completes within 2 minutes  
**Constraints**: Deterministic previews, reversible ledger entries, conflict/empty-name detection before apply  
**Scale/Scope**: Handles hundreds of files per invocation; token list expected to be small (≤20 literals)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety).
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger).
- Planned remove rules MUST document their inputs, validations, and composing order (Composable Rule Engine).
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal).
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship).

**Gate Alignment**:
- Remove command will reuse preview/confirm pipeline; no rename occurs before preview approval.
- Ledger entries will include ordered tokens removed per file to maintain undo guarantees.
- Removal logic will be implemented as composable rule(s) similar to replace, enabling reuse and testing.
- Command will consume shared scope flags (`--path`, `-r`, `-d`, `--hidden`, `-e`, `--dry-run`, `--yes`).
- Cobra wiring + automated tests will cover help text, sequential behavior warnings, and undo parity.

## Project Structure

### Documentation (this feature)

```text
specs/003-add-remove-command/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── root.go
├── list.go
├── replace.go
├── remove.go          # new CLI command for sequential removal
└── undo.go

internal/
├── listing/
├── replace/
├── remove/           # new package mirroring replace for request/parser/engine/summary
├── output/
├── traversal/
└── history/

scripts/
├── smoke-test-list.sh
├── smoke-test-replace.sh
└── smoke-test-remove.sh  # new end-to-end smoke test

tests/
├── contract/
├── integration/
├── unit/
└── fixtures/
```

**Structure Decision**: Single CLI repository with new remove-specific logic under `internal/remove/`
and CLI wiring in `cmd/remove.go`. Tests follow the existing contract/integration/unit layout and a
new smoke test script will live under `scripts/`.

## Complexity Tracking

No constitution gate violations identified; no additional complexity justification required.
