# Implementation Plan: Replace Command with Multi-Pattern Support

**Branch**: `002-add-replace-command` | **Date**: 2025-10-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-add-replace-command/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Introduce a `renamer replace` subcommand that accepts multiple literal patterns and replaces them
with a single target string while honoring existing preview → confirm → ledger → undo guarantees.
The feature will extend shared scope flags, reuse traversal/filtering pipelines, and add
automation-friendly validation and help documentation.

## Technical Context

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`  
**Storage**: Local filesystem (no persistent database)  
**Testing**: Go `testing` package with CLI integration/contract tests  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)  
**Project Type**: Single CLI project  
**Performance Goals**: Complete preview + apply for 100 files within 2 minutes  
**Constraints**: Deterministic previews, reversible ledger entries, conflict detection before apply  
**Scale/Scope**: Handles hundreds of files per invocation; patterns limited to user-provided literals

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety).
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger).
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine).
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal).
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship).

**Gate Alignment**:
- Replace command will reuse preview + confirmation pipeline; no direct rename without preview.
- Ledger entries will include replacement mappings to maintain undo guarantees.
- Replacement logic will be implemented as composable rule(s) integrating with existing rename engine.
- Command will rely on shared scope flags (`--path`, `-r`, `-d`, `--hidden`, `-e`) to avoid divergence.
- Cobra command structure and automated tests will cover help/validation/undo parity.

## Project Structure

### Documentation (this feature)

```text
specs/002-add-replace-command/
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
└── list.go (existing CLI commands; replace command will be added here or alongside)

internal/
├── listing/
├── output/
├── traversal/
└── (new replace-specific packages under internal/replace/)

scripts/
└── smoke-test-list.sh

tests/
├── contract/
├── integration/
└── fixtures/

docs/
└── cli-flags.md
```

**Structure Decision**: Single CLI repository with commands under `cmd/` and reusable logic under
`internal/`. Replace-specific logic will live in `internal/replace/` (request parsing, rule engine,
summary). CLI command wiring will reside in `cmd/replace.go`. Tests will follow existing
contract/integration directories.

## Complexity Tracking

No constitution gate violations identified; additional complexity justification not required.
