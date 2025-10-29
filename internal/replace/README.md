# internal/replace

This package hosts the core building blocks for the `renamer replace` command. The modules are
organized as follows:

- `request.go` — CLI input parsing, validation, and normalization of pattern/replacement data.
- `parser.go` — Helpers for token handling (quoting, deduplication, reporting).
- `traversal.go` — Bridges shared traversal utilities with replace-specific filtering logic.
- `engine.go` — Applies pattern replacements to candidate names and detects conflicts.
- `preview.go` / `apply.go` — Orchestrate preview output and apply/ledger integration (added later).
- `summary.go` — Aggregates match counts and conflict details for previews and ledger entries.

Tests will live alongside the package (unit) and in `tests/contract` + `tests/integration`.
