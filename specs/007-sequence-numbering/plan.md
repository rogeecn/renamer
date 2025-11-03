# Implementation Plan: Sequence Numbering Command

**Branch**: `001-sequence-numbering` | **Date**: 2025-11-03 | **Spec**: `specs/001-sequence-numbering/spec.md`
**Input**: Feature specification from `/specs/001-sequence-numbering/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Deliver a new `renamer sequence` command that appends deterministic sequence numbers to file candidates following the preview-first workflow. The command respects existing scope flags, supports configuration for start value, width, placement, and separator, records batches in the `.renamer` ledger for undo, and skips conflicting filesystem entries while warning the user.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, internal traversal/history/output packages  
**Storage**: Local filesystem + `.renamer` ledger files  
**Testing**: `go test ./...`, contract + integration suites under `tests/`  
**Target Platform**: Cross-platform CLI (Linux/macOS/Windows shells)  
**Project Type**: CLI application (single Go project)  
**Performance Goals**: Preview + apply 500 files in ≤120s; preview/apply parity ≥95%  
**Constraints**: Deterministic ordering, atomic ledger writes, skip conflicts while warning  
**Scale/Scope**: Operates on batches up to hundreds of files per invocation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow will extend existing preview pipeline to list original → numbered name mappings and highlight skipped conflicts before requiring `--yes`.  
- Undo strategy leverages ledger entries capturing sequence parameters (start, width, placement) and per-file mappings, ensuring reversal mirrors numbering order.  
- Sequence rule will be implemented as a composable transformation module declaring inputs (scope candidates + sequence config), validations, and outputs, reusable across preview/apply.  
- Scope handling continues to consume existing traversal services, honoring `-d`, `-r`, `--extensions`, and leaving directories untouched per clarified requirement while preventing scope escape.  
- CLI UX will wire flags via Cobra, update help text, and add tests covering flag validation, preview output, warning messaging, and undo flow consistency.

**Post-Design Review:** Research and design artifacts confirm all principles remain satisfied; no constitution waivers required.

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
├── renamer/               # Cobra CLI entrypoints and command wiring
internal/
├── traversal/             # Scope resolution and ordering services
├── history/               # Ledger and undo utilities
├── output/                # Preview/summary formatting
└── sequence/              # [to be added] sequence rule implementation
tests/
├── contract/              # CLI contract tests
├── integration/           # Multi-command flow tests
└── smoke/                 # Smoke scripts under scripts/
```

**Structure Decision**: Extend existing single Go CLI project; new logic lives under `internal/sequence`, with command wiring in `cmd/renamer`.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
