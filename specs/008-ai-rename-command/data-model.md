# Data Model – Genkit renameFlow & AI CLI

## Entity: RenameFlowInput
- **Fields**
  - `fileNames []string` — Ordered list of basenames collected from scope traversal.
  - `userPrompt string` — Optional user guidance merged into the prompt template.
- **Validation Rules**
  - Require at least one filename; enforce maximum of 200 per invocation (soft limit before batching).
  - Reject names containing path separators; traversal supplies basenames only.
  - Trim whitespace from `userPrompt`; clamp length (e.g., 1–500 characters) to guard against prompt injection.
- **Relationships**
  - Serialized to JSON and passed into `genkit.Generate()` as the model input payload.
  - Logged with invocation metadata to support replay/debugging.

## Entity: RenameFlowOutput
- **Fields**
  - `suggestions []RenameSuggestion` — AI-produced rename pairs in same order as input list when possible.
- **Validation Rules**
  - `len(suggestions)` MUST equal length of input `fileNames` before approval.
  - Each suggestion MUST pass filename safety checks (see `RenameSuggestion`).
  - JSON payload MUST parse cleanly with no additional top-level properties.
- **Relationships**
  - Returned to the CLI bridge, transformed into preview rows and ledger entries.

## Entity: RenameSuggestion
- **Fields**
  - `original string` — Original basename (must match an item from input list).
  - `suggested string` — Proposed basename with identical extension as `original`.
- **Validation Rules**
  - Preserve extension suffix (text after last `.`); fail if changed or removed.
  - Disallow illegal filesystem characters: `/ \ : * ? " < > |` and control bytes.
  - Enforce case-insensitive uniqueness across all `suggested` values to avoid collisions.
  - Reject empty or whitespace-only suggestions; trim incidental spaces.
- **Relationships**
  - Consumed by preview renderer to display mappings.
  - Persisted in ledger metadata alongside user prompt and model ID.

## Entity: AISuggestionBatch (Go side)
- **Fields**
  - `Scope traversal.ScopeResult` — Snapshot of files selected for AI processing.
  - `Prompt string` — Rendered prompt sent to Genkit (stored for debugging).
  - `ModelID string` — Identifier for the AI model used during generation.
  - `Suggestions []RenameSuggestion` — Parsed results aligned with scope entries.
  - `Warnings []string` — Issues detected during validation (duplicates, unchanged names, limit truncation).
- **Validation Rules**
  - Warnings that correspond to hard failures (duplicate targets, invalid characters) block apply until resolved.
  - Scope result order MUST align with suggestion order to keep preview deterministic.
- **Relationships**
  - Passed into output renderer for table display.
  - Written to ledger with `history.RecordBatch` for undo.

## Entity: FlowInvocationLog
- **Fields**
  - `InvocationID string` — UUID tying output to ledger entry.
  - `Timestamp time.Time` — Invocation time for audit trail.
  - `Duration time.Duration` — Round-trip latency for success criteria tracking.
  - `InputSize int` — Number of filenames processed (used for batching heuristics).
  - `Errors []string` — Captured model or validation errors.
- **Validation Rules**
  - Duration recorded only on successful completions; errors populated otherwise.
- **Relationships**
  - Optional: appended to debug logs or analytics for performance monitoring (non-ledger).
