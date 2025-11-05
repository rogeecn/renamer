# Phase 0 Research – Genkit renameFlow

## Decision: Enforce JSON-Only Responses via Prompt + Guardrails
- **Rationale**: The CLI must parse deterministic structures. Embedding an explicit JSON schema example, restating illegal character rules, and wrapping the Genkit call with `OutputJSON()` (or equivalent) reduces hallucinated prose and aligns with ledger needs.
- **Alternatives considered**: Post-processing free-form text was rejected because it increases parsing failures and weakens auditability. Relaxing constraints to “JSON preferred” was rejected to avoid brittle regex extraction.

## Decision: Keep Prompt Template as External File with Go Template Variables
- **Rationale**: Storing the prompt under `internal/ai/flow/prompt.tmpl` keeps localization and iteration separate from code. Using Go-style templating enables the flow to substitute the file list and user prompt consistently while making it easier to unit test rendered prompts.
- **Alternatives considered**: Hardcoding prompt strings inside the Go flow was rejected due to limited reuse and poor readability; using Markdown-based prompts was rejected because the model might echo formatting in its response.

## Decision: Invoke Genkit Flow In-Process via Go SDK
- **Rationale**: The spec emphasizes local filesystem workflows without network services. Using the Genkit Go SDK keeps execution in-process, avoids packaging a separate runtime, and fits CLI invocation patterns.
- **Alternatives considered**: Hosting a long-lived HTTP service was rejected because it complicates installation and violates the local-only assumption. Spawning an external Node process was rejected due to additional toolchain requirements.

## Decision: Validate Suggestions Against Existing Filename Rules Before Apply
- **Rationale**: Even with JSON enforcement, the model could suggest duplicates, rename directories, or remove extensions. Reusing internal validation logic ensures suggestions honor filesystem invariants and matches ledger expectations before touching disk.
- **Alternatives considered**: Trusting AI output without local validation was rejected due to risk of destructive renames. Silently auto-correcting invalid names was rejected because it obscures AI behavior from users.

## Decision: Align Testing with Contract + Golden Prompt Fixtures
- **Rationale**: Contract tests with fixed model responses (via canned JSON) allow deterministic verification, while golden prompt fixtures ensure template rendering matches expectations. This combo offers coverage without depending on live AI calls in CI.
- **Alternatives considered**: Live integration tests hitting the model were rejected due to cost, flakiness, and determinism concerns. Pure unit tests without prompt verification were rejected because prompt regressions directly impact model quality.
