# Implementation Plan: Cobra List Command with Global Filters

**Branch**: `001-list-command-filters` | **Date**: 2025-10-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-list-command-filters/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Introduce a Cobra `list` subcommand that enumerates filesystem entries using the same global filter
flags (`-r`, `-d`, `-e`) shared with preview/rename flows. The command will reuse traversal and
validation utilities to guarantee consistent candidate sets, provide table/plain output formats, and
surface clear messaging for empty results or invalid filters.

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.24  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`  
**Storage**: Local filesystem (read-only listing)  
**Testing**: Go `testing` package with CLI-focused integration tests  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single CLI project  
**Performance Goals**: First page of 5k entries within 2 seconds via streaming output  
**Constraints**: Deterministic ordering, no filesystem mutations, filters shared across commands  
**Scale/Scope**: Operates on directories with tens of thousands of entries per invocation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety).
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger).
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine).
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal).
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship).

**Gate Alignment**:
- Listing command will remain a read-only helper; preview/rename confirmation flow stays unchanged.
- Ledger logic untouched; plan maintains append-only guarantees by reusing existing history modules.
- Filters, traversal, and rule composition will be centralized to avoid divergence between commands.
- Root-level flags (`-r`, `-d`, `-e`) will configure shared traversal services so all subcommands honor identical scope rules.
- Cobra command UX will include consistent help text, validation errors, and integration tests for list/preview parity.

## Project Structure

### Documentation (this feature)

```text
specs/001-list-command-filters/
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
├── root.go          # Cobra root command with global flags
├── list.go          # New list subcommand entry point
└── preview.go       # Existing preview wiring (to be reconciled with shared filters)

internal/
├── traversal/       # Scope walking utilities (files, directories, recursion)
├── filters/         # Extension parsing/validation shared across commands
├── listing/         # New package composing traversal + formatting
└── output/          # Shared renderers for table/plain display

tests/
├── contract/
│   └── list_command_test.go
├── integration/
│   └── list_and_preview_parity_test.go
└── fixtures/        # Sample directory trees for CLI tests
```

**Structure Decision**: Single CLI repository rooted at `cmd/` with supporting packages under
`internal/`. Shared traversal and filter logic will move into dedicated packages to ensure the
global flags are consumed identically by `list`, `preview`, and rename workflows. Tests live under
`tests/` mirroring contract vs integration coverage.

## Complexity Tracking

No constitution gate violations identified; no additional complexity justifications required.
