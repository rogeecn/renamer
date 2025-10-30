# renamer Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-29

## Active Technologies
- Local filesystem (no persistent database) (002-add-replace-command)
- Go 1.24 + `spf13/cobra`, `spf13/pflag` (001-list-command-filters)
- Local filesystem only (ledger persisted as `.renamer`) (003-add-remove-command)
- Go 1.24 + `spf13/cobra`, `spf13/pflag`, internal traversal/ledger packages (004-extension-rename)
- Local filesystem + `.renamer` ledger files (004-extension-rename)

## Project Structure

```text
cmd/
internal/
scripts/
tests/
```

## Commands

- `renamer list` — preview rename scope with shared flags before executing changes.
- `renamer replace` — consolidate multiple literal patterns into a single replacement (supports `--dry-run` + `--yes`).
- `renamer remove` — delete ordered substrings from filenames with empty-name protections, duplicate warnings, and undoable ledger entries.
- `renamer undo` — revert the most recent rename/replace batch using ledger entries.
- Persistent scope flags: `--path`, `-r/--recursive`, `-d/--include-dirs`, `--hidden`, `--extensions`, `--yes`, `--dry-run`.

## Code Style

- Go 1.24: follow gofmt (already checked in CI)
- Prefer composable packages under `internal/` for reusable logic
- Keep CLI wiring thin; place business logic in services

## Testing

- `go test ./...`
- Contract tests: `tests/contract/replace_command_test.go`, `tests/contract/remove_command_preview_test.go`, `tests/contract/remove_command_ledger_test.go`
- Integration tests: `tests/integration/replace_flow_test.go`, `tests/integration/remove_flow_test.go`, `tests/integration/remove_undo_test.go`, `tests/integration/remove_validation_test.go`
- Smoke: `scripts/smoke-test-replace.sh`, `scripts/smoke-test-remove.sh`

## Recent Changes
- 004-extension-rename: Added Go 1.24 + `spf13/cobra`, `spf13/pflag`, internal traversal/ledger packages
- 003-add-remove-command: Added sequential `renamer remove` subcommand, automation-friendly ledger metadata, and CLI warnings for duplicates/empty results
- 002-add-replace-command: Added `renamer replace` command, ledger metadata, and automation docs.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
