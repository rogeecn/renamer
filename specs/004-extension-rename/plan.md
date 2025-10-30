# Implementation Plan: Extension Command for Multi-Extension Normalization

**Branch**: `004-extension-rename` | **Date**: 2025-10-30 | **Spec**: `specs/004-extension-rename/spec.md`
**Input**: Feature specification from `/specs/004-extension-rename/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Deliver a `renamer extension` subcommand that normalizes one or more source extensions to a single target extension with deterministic previews, ledger-backed undo, and automation-friendly exit codes. The implementation will extend existing replace/remove infrastructure, ensuring case-insensitive extension matching, hidden-file opt-in, conflict detection, and “no change” handling for already-targeted files.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, internal traversal/ledger packages  
**Storage**: Local filesystem + `.renamer` ledger files  
**Testing**: `go test ./...`, smoke + integration scripts under `tests/` and `scripts/`  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows shells)
**Project Type**: Single CLI project (`cmd/`, `internal/`, `tests/`, `scripts/`)  
**Performance Goals**: Normalize 500 files (preview+apply) in <2 minutes end-to-end  
**Constraints**: Preview-first workflow, ledger append-only, hidden files excluded unless `--hidden`  
**Scale/Scope**: Operates on thousands of filesystem entries per invocation within local directory trees

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety). ✅ Extend existing preview engine to list original → target paths, highlighting extension changes and “no change” rows; apply remains gated by `--yes`.
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger). ✅ Record source extension list, target, per-file outcomes in ledger entry and ensure undo replays entries via existing ledger service.
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine). ✅ Define an extension normalization rule that consumes scope matches, validates tokens, deduplicates case-insensitively, and reuses sequencing from replace engine.
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal). ✅ Continue relying on shared traversal component honoring flags; ensure hidden assets stay excluded unless `--hidden`.
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship). ✅ Implement `extension` Cobra command mirroring existing flag sets, update help text, and add contract/integration tests for preview, apply, and undo.

*Post-Design Verification (2025-10-30): Research and design artifacts document preview flow, ledger metadata, rule composition, scope behavior, and Cobra UX updates—no gate violations detected.*

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
└── undo.go

internal/
├── filters/
├── history/
├── listing/
├── output/
├── remove/
├── replace/
├── traversal/
└── extension/        # new package for extension normalization engine

scripts/
├── smoke-test-remove.sh
└── smoke-test-replace.sh

tests/
├── contract/
│   ├── remove_command_ledger_test.go
│   ├── remove_command_preview_test.go
│   ├── replace_command_test.go
│   └── (new) extension_command_test.go
├── integration/
│   ├── remove_flow_test.go
│   ├── remove_undo_test.go
│   ├── remove_validation_test.go
│   └── replace_flow_test.go
└── fixtures/         # shared test inputs
```

**Structure Decision**: Maintain the single CLI project layout, adding an `internal/extension` package plus contract/integration tests mirroring existing replace/remove coverage.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _None_ | — | — |
