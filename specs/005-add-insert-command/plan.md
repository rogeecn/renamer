# Implementation Plan: Insert Command for Positional Text Injection

**Branch**: `005-add-insert-command` | **Date**: 2025-10-30 | **Spec**: `specs/005-add-insert-command/spec.md`
**Input**: Feature specification from `/specs/005-add-insert-command/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Deliver a `renamer insert` subcommand that inserts a specified string into filenames at designated positions (start, end, absolute, relative) with Unicode-aware behavior, deterministic previews, ledger-backed undo, and automation-friendly outputs.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, internal traversal/history/output packages  
**Storage**: Local filesystem + `.renamer` ledger files  
**Testing**: `go test ./...`, contract + integration suites under `tests/`, new smoke script  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows shells)
**Project Type**: Single CLI project (`cmd/`, `internal/`, `tests/`, `scripts/`)  
**Performance Goals**: 500-file insert (preview+apply) completes in <2 minutes end-to-end  
**Constraints**: Unicode-aware insertion points, preview-first safety, ledger reversibility  
**Scale/Scope**: Operates on thousands of filesystem entries per invocation within local directories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety). ✅ Extend insert preview to render original → proposed names with highlighted insertion segments before any apply.
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger). ✅ Reuse ledger append with metadata (position token, inserted string) and ensure `undo` replays operations safely.
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine). ✅ Define an insert rule consuming position tokens, performing Unicode-aware slicing, and integrating with existing traversal pipeline.
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal). ✅ Leverage shared listing/traversal flags so insert respects scope filters and hidden/default exclusions.
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship). ✅ Add Cobra subcommand with consistent flags, help examples, contract/integration tests, and smoke coverage.

*Post-Design Verification (2025-10-30): Research and design artifacts document preview behavior, ledger metadata, Unicode-aware positions, and CLI UX updates — no gate violations detected.*

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
cmd/
├── root.go
├── list.go
├── replace.go
├── remove.go
├── extension.go
└── undo.go

internal/
├── filters/
├── history/
├── listing/
├── output/
├── remove/
├── replace/
├── extension/
└── traversal/

tests/
├── contract/
├── integration/
├── fixtures/
└── unit/

scripts/
├── smoke-test-list.sh
├── smoke-test-replace.sh
├── smoke-test-remove.sh
└── smoke-test-extension.sh
```

**Structure Decision**: Extend the single CLI project by adding new `cmd/insert.go`, `internal/insert/` package, contract/integration coverage under existing `tests/` hierarchy, and an insert smoke script alongside other command scripts.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
