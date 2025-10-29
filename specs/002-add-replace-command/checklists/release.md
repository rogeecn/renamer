# Release Readiness Checklist: Replace Command with Multi-Pattern Support

**Purpose**: Verify readiness for release / polish phase  
**Created**: 2025-10-29  
**Feature**: [spec.md](../spec.md)

## Documentation

- [x] Quickstart reflects latest syntax and automation workflow
- [x] CLI reference (`docs/cli-flags.md`) includes replace command usage and warnings
- [x] AGENTS.md updated with replace command summary
- [x] CHANGELOG entry drafted for replace command

## Quality Gates

- [x] `go test ./...` passing locally
- [x] Smoke test script for replace + undo exists and runs
- [x] Ledger metadata includes pattern counts and is asserted in tests
- [x] Empty replacement path warns users in preview

## Operational Readiness

- [x] `--dry-run` and `--yes` are mutually exclusive and error when combined
- [x] Undo command reverses replace operations via ledger entry
- [x] Scope flags behave identically across list/replace commands

## Notes

- Resolve outstanding checklist items prior to tagging release.
