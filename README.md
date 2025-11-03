# renamer

`renamer` is a preview-first, undoable batch renaming CLI for local file systems. It wraps a traversal engine, scoped filters, and a newline-delimited `.renamer` ledger so you can trial changes safely before committing them.

## Highlights

- Preview every operation with conflict detection before touching the filesystem.
- Share a single set of scope flags (`--path`, `--recursive`, `--include-dirs`, etc.) across all commands.
- Apply literal replacements, substring removals, extension normalization, positional inserts, or regex-driven renames.
- Persist every batch into a `.renamer` ledger and undo the most recent change set at any time.
- Output results in table or plain text formats to fit automation and scripting workflows.

## Installation

Prerequisites: Go 1.24+

```bash
git clone https://github.com/rogeecn/renamer.git
cd renamer
go install ./...
```

The `renamer` binary will be placed in your `GOBIN` (defaults to `$GOPATH/bin`).

## Usage

The CLI follows the pattern:

```bash
renamer [global scope flags] <command> [command args]
```

### Shared scope flags

All subcommands accept these persistent flags:

- `--path <dir>`: Working directory (defaults to current directory).
- `-r, --recursive`: Traverse subdirectories.
- `-d, --include-dirs`: Include directories in results.
- `--hidden`: Include hidden files and directories.
- `-e, --extensions ".jpg|.png"`: Pipe-delimited list of extensions to target.
- `--dry-run`: Force preview-only mode.
- `--yes`: Confirm changes without prompting (mutating commands only).

### Commands

- `renamer list [--format table|plain] [--max-depth N]` — Preview the files and directories that match the active scope.
- `renamer replace <pattern...> <replacement>` — Replace multiple literal tokens in sequence. Shows duplicates and conflict warnings, then applies when `--yes` is present.
- `renamer remove <pattern...>` — Strip ordered substrings from names with empty-name protection and duplicate detection.
- `renamer extension <source-ext...> <target-ext>` — Normalize heterogeneous extensions to a single target while keeping a ledger entry for undo.
- `renamer insert <position> <text>` — Insert text at symbolic (`^`, `$`) offsets, count forward with numbers (`3` or `^3`), or backward with suffix tokens like `1$`.
- `renamer sequence [flags]` — Append or prepend zero-padded sequence numbers with configurable start, width, placement (default prefix), separator, and static number prefix/suffix options.
- `renamer regex <pattern> <template>` — Rename via RE2 capture groups using placeholders like `@1`, `@2`, `@0`, or escape literal `@` as `@@`.
- `renamer undo` — Revert the most recent mutating command recorded in the ledger.

### Example workflow

```bash
# Preview JPEG->jpg normalization for a project tree
renamer extension .jpeg .JPG .jpg --recursive --dry-run

# Apply once the preview looks correct
renamer extension .jpeg .JPG .jpg --recursive --yes

# Undo the batch if something looks off
renamer undo
```

## Ledger and undo

Every mutating command appends a newline-delimited JSON entry to `.renamer` in the working directory, capturing the command, metadata, and operations. `renamer undo` reads the ledger backwards, renames entries to their previous paths, and rewrites the ledger to keep history consistent. Delete the ledger file if you want a fresh history.

## Development

- Run the unit, integration, and contract tests with `go test ./...`.
- Contract tests live under `tests/contract`, while integration workflows sit in `tests/integration`.
- Smoke test scripts (`scripts/smoke-test-*.sh`) exercise end-to-end flows against sample fixtures in `testdata/`.
- CI enforces `gofmt`; ensure files remain formatted after edits.

Contributions generally go through the composable packages under `internal/`, keeping `cmd/` focused on CLI wiring. See `docs/` and `specs/` for feature plans and architecture notes.
