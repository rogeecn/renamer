# Regex Command Fixtures

These fixtures support contract and integration testing for the `renamer regex` command. Each
subdirectory contains representative filenames used across preview, apply, conflict, and
validation scenarios.

- `baseline/` — ASCII word + digit combinations (e.g., `alpha-123.log`) used to validate basic
  capture group substitution.
- `unicode/` — Multilingual filenames to verify RE2 Unicode handling and ledger persistence.
- `mixed/` — Build-style artifacts with underscores/dashes for automation-style rename flows.
- `case-fold/` — Differing only by case to simulate case-insensitive duplicate conflicts.

Tests should copy these directories to temporary working paths before mutation to keep fixtures
idempotent.
