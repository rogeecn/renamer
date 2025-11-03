# Research Log

## Genkit Orchestration Strategy
- **Decision**: Integrate the official Google Genkit Go SDK directly in `internal/ai/genkit`, executing workflows synchronously inside the `renamer ai` command without external services.
- **Rationale**: Satisfies the CLI-only constraint, keeps deployment as a single Go binary, and leverages Genkit guardrails, evaluators, and prompt templating.
- **Alternatives considered**: Spawning a Node.js runner (rejected—adds extra runtime and violates updated requirement); Plain REST client to foundation models (rejected—would require rebuilding Genkit safety features manually).

## Prompt Composition Template
- **Decision**: Define typed Go structs mirroring the prompt schema (scope summary, sample filenames, naming policies, banned tokens) and marshal them to JSON for Genkit inputs.
- **Rationale**: Strong typing prevents malformed prompts and aligns with Genkit Go helpers for variable interpolation and logging.
- **Alternatives considered**: Free-form natural language prompts (rejected—harder to validate); YAML serialization (rejected—JSON is the Genkit default and reduces dependency footprint).

## Response Schema & Validation
- **Decision**: Genkit workflow returns Go structs (`RenameItem`, `Warnings`) serialized as JSON; the CLI validates coverage, uniqueness, banned-term removal, and sequential numbering before building the plan.
- **Rationale**: Mirrors existing rename planner data types, enabling reuse of preview/output logic and providing transparent audit metadata.
- **Alternatives considered**: Returning only ordered filenames (rejected—insufficient context for debugging); CSV output (rejected—lossy and awkward for nested metadata).

## Offline & Failure Handling
- **Decision**: Use Genkit middleware for retry/backoff and error classification; if the workflow fails or produces invalid data, the CLI aborts gracefully, surfaces the issue, and can export the prompt/response for manual inspection.
- **Rationale**: Maintains Preview-First safety by never applying partial results and keeps error handling contained within the CLI execution.
- **Alternatives considered**: Persistent background daemon (rejected—contradicts inline execution); Automatic fallback to legacy sequential numbering (rejected—changes user intent, can be added later as an explicit option).
