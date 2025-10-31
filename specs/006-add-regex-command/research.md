# Phase 0 Research – Regex Command

## Decision: Reuse Traversal, Preview, and Ledger Pipelines for Regex Rule
- **Rationale**: Existing replace/remove/extension commands already walk the filesystem, apply scope filters, and feed preview + ledger writers. Plugging a regex rule into this pipeline guarantees consistent conflict detection, skipped reporting, and undo safety without reimplementing traversal safeguards.
- **Alternatives considered**: Building a standalone regex walker was rejected because it would duplicate scope logic and risk violating Scope-Aware Traversal. Embedding regex into replace internals was rejected to keep literal and regex behaviors independent and easier to test.

## Decision: Compile Patterns with Go `regexp` (RE2) and Cache Group Metadata
- **Rationale**: Go’s standard library provides RE2-backed regex compilation with deterministic performance and Unicode safety. Capturing the compiled expression once per invocation lets us pre-count capture groups, validate templates, and apply matches efficiently across many files.
- **Alternatives considered**: Using third-party regex engines (PCRE) was rejected due to external dependencies and potential catastrophic backtracking. Recompiling the pattern per file was rejected for performance reasons.

## Decision: Validate and Render Templates via Placeholder Tokens (`@0`, `@1`, …, `@@`)
- **Rationale**: Parsing the template into literal and placeholder segments ensures undefined group references surface as validation errors before preview/apply, while optional groups that fail to match substitute with empty strings. Doubling `@` (i.e., `@@`) yields a literal `@`, aligning with the clarification already captured in the specification.
- **Alternatives considered**: Allowing implicit zero-value substitution for undefined groups was rejected because it hides mistakes. Relying on `fmt.Sprintf`-style formatting was rejected since it lacks direct mapping to numbered capture groups and complicates escaping rules.

## Decision: Ledger Metadata Includes Pattern, Template, and Match Snapshots
- **Rationale**: Persisting the regex pattern, replacement template, scope flags, and per-file capture arrays alongside old/new paths enables precise undo and supports automation auditing. This mirrors expectations set for other commands and satisfies the Persistent Undo Ledger principle.
- **Alternatives considered**: Logging only before/after filenames was rejected because undo would lose context if filenames changed again outside the tool. Capturing full file contents was rejected as unnecessary and intrusive.

## Decision: Block Apply When Template Yields Conflicts or Empty Targets
- **Rationale**: Conflict detection will reuse existing duplicate/overwrite checks but extend them to treat empty or whitespace-only proposals as invalid. Apply exits non-zero when conflicts remain, protecting against accidental data loss or invalid filenames.
- **Alternatives considered**: Auto-resolving conflicts by suffixing counters was rejected because it introduces nondeterministic results and complicates undo. Allowing empty targets was rejected for safety and compatibility reasons.
