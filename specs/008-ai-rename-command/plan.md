# Implementation Plan: AI-Assisted Rename Command

**Branch**: `008-ai-rename-command` | **Date**: 2025-11-05 | **Spec**: `specs/008-ai-rename-command/spec.md`
**Input**: Feature specification from `/specs/008-ai-rename-command/spec.md`

**Note**: This plan grounds the `/speckit.plan` prompt “Genkit Flow 设计 (Genkit Flow Design)” by detailing how the CLI and Genkit workflow collaborate to deliver structured AI rename suggestions.

## Summary

Design and implement a `renameFlow` Genkit workflow that produces deterministic, JSON-formatted rename suggestions and wire it into a new `renamer ai` CLI path. The plan covers prompt templating, JSON validation, scope handling parity with existing commands, preview/confirmation UX, ledger integration, and fallback/error flows to keep AI-generated batches auditable and undoable.

## Technical Context

**Language/Version**: Go 1.24 (CLI + Genkit workflow)  
**Primary Dependencies**: `spf13/cobra`, `spf13/pflag`, internal traversal/history/output packages, `github.com/firebase/genkit/go`, OpenAI-compatible provider bridge  
**Storage**: Local filesystem plus append-only `.renamer` ledger  
**Testing**: `go test ./...` including flow unit tests for prompt/render/validation, contract + integration tests under `tests/`  
**Target Platform**: Cross-platform CLI executed from local shells; Genkit workflow runs in-process via Go bindings  
**Project Type**: Single Go CLI project with additional internal AI packages  
**Performance Goals**: Generate rename suggestions for ≤200 files within 30 seconds end-to-end (per SC-001)  
**Constraints**: Preview-first safety, undoable ledger entries, scope parity with existing commands, deterministic JSON responses, offline fallback excluded (network required)  
**Scale/Scope**: Handles hundreds of files per invocation, with potential thousands when batched; assumes human-in-the-loop confirmation

## Constitution Check

- Preview flow MUST show deterministic rename mappings and require explicit confirmation (Preview-First Safety). ✅ `renamer ai` reuses preview renderer to display AI suggestions, blocks apply until `--yes` or interactive confirmation, and supports `--dry-run`.
- Undo strategy MUST describe how the `.renamer` ledger entry is written and reversed (Persistent Undo Ledger). ✅ Accepted batches append AI metadata (prompt, model, rationale) to ledger entries; undo replays via existing ledger service with no schema break.
- Planned rename rules MUST document their inputs, validations, and composing order (Composable Rule Engine). ✅ `renameFlow` enforces rename suggestion structure (original, suggested), keeps extensions intact, and CLI validates conflicts before applying.
- Scope handling MUST cover files vs directories (`-d`), recursion (`-r`), and extension filtering via `-e` without escaping the requested path (Scope-Aware Traversal). ✅ CLI gathers scope using shared traversal component, honoring existing flags before passing filenames to Genkit.
- CLI UX plan MUST confirm Cobra usage, flag naming, help text, and automated tests for preview/undo flows (Ergonomic CLI Stewardship). ✅ New `ai` command extends Cobra root with existing persistent flags, adds prompt/model overrides, and includes contract + integration coverage for preview/apply/undo.

## Project Structure

### Documentation (this feature)

```text
specs/008-ai-rename-command/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── spec.md
```

### Source Code (repository root)

```text
cmd/
├── root.go
├── ai.go                # new Cobra command wiring + RunE
├── list.go
├── replace.go
├── remove.go
└── undo.go

internal/
├── ai/
│   ├── flow/
│   │   ├── rename_flow.go        # Genkit flow definition using Go SDK
│   │   └── prompt.tmpl           # prompt template with rules/formatting
│   ├── client.go                # wraps Genkit invocation + response handling
│   ├── preview.go               # maps RenameSuggestion -> preview rows
│   ├── validation.go            # conflict + filename safety checks
│   └── session.go               # manages user guidance refinements
├── traversal/
├── output/
├── history/
└── ...

tests/
├── contract/
│   └── ai_command_preview_test.go   # ensures JSON contract adherence
├── integration/
│   └── ai_flow_apply_test.go        # preview, confirm, undo happy path
└── fixtures/
    └── ai/
        └── sample_photos/          # test assets for AI rename flows

scripts/
└── smoke-test-ai.sh                 # optional future smoke harness (planned)
```

**Structure Decision**: Implement the Genkit `renameFlow` directly within Go (`internal/ai/flow`) while reusing shared traversal/output pipelines through the new `ai` command. Tests mirror existing command coverage with contract and integration suites.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| _None_ | — | — |
