# Implementation Plan: AI-Assisted Rename Prompting

**Branch**: `008-ai-rename-prompt` | **Date**: 2025-11-03 | **Spec**: `specs/008-ai-rename-prompt/spec.md`
**Input**: Feature specification from `/specs/008-ai-rename-prompt/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Introduce a `renamer ai` command that embeds a Google Genkit (Go SDK) workflow inside the CLI execution path. The command collects scope metadata, calls the Genkit pipeline in-process (defaulting to an OpenAI-compatible model), validates the response for sequential, uniform, sanitized filenames, allows operator edits, and records final mappings in the undo ledger while managing `*_MODEL_AUTH_TOKEN` secrets under `$HOME/.config/.renamer/`.

## Technical Context

**Language/Version**: Go 1.24 (CLI + Google Genkit Go SDK)  
**Primary Dependencies**: `spf13/cobra`, internal traversal/history/output packages, `github.com/google/genkit/go` (with OpenAI-compatible connectors), OpenAI-compatible HTTP client for fallbacks  
**Storage**: Local filesystem plus `.renamer` append-only ledger; auth tokens cached under `$HOME/.config/.renamer/`  
**Testing**: `go test ./...` for CLI logic, `npm test` (Vitest) for Genkit prompt workflows, contract/integration suites under `tests/`  
**Target Platform**: Cross-platform CLI (macOS, Linux, Windows shells) executing in-process Genkit workflows  
**Project Type**: Single CLI project with integrated Go Genkit module  
**Performance Goals**: Generate validated rename plan for up to 1,000 files in ≤ 30 seconds round-trip  
**Constraints**: Genkit workflow must initialize quickly per invocation; AI requests limited to 2 MB payload; ensure user-provided banned terms removed  
**Scale/Scope**: Typical batches 1–1,000 files; shared Genkit pipeline available for future AI features

## Constitution Check

- Preview flow continues to render deterministic before/after tables and block apply until confirmed, satisfying Preview-First Safety.
- Undo path records final mappings plus AI prompt/response metadata in `.renamer`, preserving Persistent Undo Ledger guarantees.
- AI rename integration becomes a composable rule module (`internal/ai`) that declares inputs (prompt spec), validations, and postconditions while orchestrating the Go Genkit workflow inline, aligning with Composable Rule Engine.
- Scope handling reuses existing traversal services (`internal/traversal`) so filters (`--path`, `-r`, `-d`, `--extensions`) remain enforced per Scope-Aware Traversal.
- Cobra wiring (`cmd/ai.go`) follows existing CLI standards with help text, flag validation, tests, meeting Ergonomic CLI Stewardship.

## Project Structure

### Documentation (this feature)

```text
specs/008-ai-rename-prompt/
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
├── ai.go                # Genkit-powered command wiring
├── root.go

internal/
├── ai/
│   ├── prompt/          # Prompt assembly, policy enforcement
│   ├── genkit/          # Go Genkit workflow definitions and model connectors
│   └── plan/            # Response validation & editing utilities
├── history/
├── output/
├── traversal/
└── sequence/

tests/
├── contract/
├── integration/
└── unit/
```

tests/
├── contract/
├── integration/
└── unit/

**Structure Decision**: Extend existing CLI layout by adding an `internal/ai` package that houses Go Genkit workflows invoked directly from `cmd/ai.go`; existing test directories cover the new command with contract/integration suites.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| Direct Go Genkit integration | First-class Go SDK keeps execution inline and satisfies CLI-only requirement | Manual REST integration would lose Genkit workflows (retriers, evaluators) and require bespoke prompt templating |
