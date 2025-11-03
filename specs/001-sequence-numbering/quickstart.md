# Quickstart: Sequence Numbering Command

## Prerequisites
- Go 1.24 toolchain installed.
- `renamer` repository cloned and bootstrapped (`go mod tidy` already satisfied in repo).
- Test fixtures available under `tests/` for validation runs.

## Build & Install
```bash
go build -o bin/renamer ./cmd/renamer
```

## Preview Sequence Numbering
```bash
bin/renamer sequence \
  --path ./fixtures/sample-batch \
  --dry-run
```
Outputs a preview table showing `001_`, `002_`, … prefixes based on alphabetical order.

## Customize Formatting
```bash
bin/renamer sequence \
  --path ./fixtures/sample-batch \
  --start 10 \
  --width 4 \
  --number-prefix seq \
  --separator "" \
  --dry-run
```
Produces names such as `seq0010file.ext`. Errors if width/start are invalid.

## Apply Changes
```bash
bin/renamer sequence \
  --path ./fixtures/sample-batch \
  --yes
```
Writes rename results to the `.renamer` ledger while skipping conflicting targets and warning the user.

## Undo Sequence Batch
```bash
bin/renamer undo --path ./fixtures/sample-batch
```
Restores filenames using the most recent ledger entry.

## Run Automated Tests
```bash
go test ./...
tests/integration/remove_flow_test.go    # existing suites ensure regressions are caught
```

## Troubleshooting
- Conflict warnings indicate existing files with the same numbered name; resolve manually or adjust flags.
- Zero candidates cause a 409-style error; adjust scope flags to include desired files.
