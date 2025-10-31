# CLI Scope Flags

Renamer shares a consistent set of scope flags across every command that inspects or mutates the
filesystem. Use these options at the root command level so they apply to all subcommands (`list`,
`replace`, `insert`, `remove`, etc.).

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | `.` | Working directory root for traversal. |
| `-r`, `--recursive` | `false` | Traverse subdirectories depth-first. Symlinked directories are not followed. |
| `-d`, `--include-dirs` | `false` | Include directories in results. |
| `-e`, `--extensions` | *(none)* | Pipe-separated list of file extensions (e.g. `.jpg|.mov`). Tokens must start with a dot, are lowercased internally, and duplicates are ignored. |
| `--hidden` | `false` | Include dot-prefixed files and directories. By default they are excluded from listings and rename previews. |
| `--yes` | `false` | Apply changes without interactive confirmation (mutating commands only). |
| `--dry-run` | `false` | Force preview-only behavior even when `--yes` is supplied. |
| `--format` | `table` | Command-specific output formatting option. For `list`, use `table` or `plain`. |

## Regex Command Quick Reference

```bash
renamer regex <pattern> <template> [flags]
```

- Patterns compile with Go’s RE2 engine and are matched against filename stems; invalid expressions fail fast with helpful errors.
- Templates support numbered placeholders (`@0`, `@1`, …) along with escaped `@@` for literal at-signs; undefined captures block the run.
- Preview mode (`--dry-run`, default) renders the rename plan with change/skipped/conflict statuses; apply with `--yes` writes a ledger entry for undo.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) control candidate discovery just like other commands, and conflicts or empty targets exit non-zero.

### Usage Examples

- Preview captured group swapping: `renamer regex "^(\w+)-(\d+)" "@2_@1" --dry-run --path ./samples`
- Limit by extensions and directories: `renamer regex '^(build)_(\d+)_v(.*)$' 'release-@2-@1-v@3' --extensions .zip|.tar.gz --include-dirs --recursive`
- Automation-friendly apply with undo: `renamer regex '^(feature)-(.*)$' '@2-@1' --yes --path ./staging && renamer undo --path ./staging`

## Insert Command Quick Reference

```bash
renamer insert <position> <text> [flags]
```

- Position tokens:
  - `^` inserts at the beginning of the filename.
  - `$` inserts immediately before the extension dot (or end if no extension).
  - Positive integers (1-based) count forward from the start of the stem.
  - Negative integers count backward from the end of the stem (e.g., `-1` inserts before the last rune).
- Text must be valid UTF-8 without path separators or control characters; Unicode characters are supported.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) limit the candidate set before insertion.
- `--dry-run` previews the plan; rerun with `--yes` to apply the same operations.

### Usage Examples

- Preview adding a prefix: `renamer insert ^ "[2025] " --dry-run`
- Append before extension: `renamer insert $ _ARCHIVE --yes --path ./docs`
- Insert after third character in stem: `renamer insert 3 _tag --path ./images --dry-run`
- Combine with extension filter: `renamer insert ^ "v1_" --extensions .txt|.md`

## Remove Command Quick Reference

```bash
renamer remove <token1> [token2 ...] [flags]
```

- Removal tokens are evaluated in the order supplied. Each token deletes literal substrings from the
  current filename before the next token runs; results are previewed before any filesystem changes.
- Duplicate tokens are deduplicated automatically and surfaced as warnings so users can adjust
  scripts without surprises.
- Tokens that collapse a filename to an empty string are skipped with warnings during preview/apply
  to protect against accidental deletion.
- All scope flags (`--path`, `-r`, `-d`, `--hidden`, `-e`) apply, making it easy to target directories,
  recurse, and limit removals by extension.
- Use `--dry-run` for automation previews and combine with `--yes` to apply unattended; conflicting
  combinations (`--dry-run --yes`) exit with an error to uphold preview-first safety.

### Usage Examples

- Preview sequential removals: `renamer remove " copy" " draft" --dry-run`
- Remove tokens recursively: `renamer remove foo foo- --recursive --path ./reports`
- Combine with extension filters: `renamer remove " Project" --extensions .txt|.md --dry-run`

## Extension Command Quick Reference

```bash
renamer extension <source-ext...> <target-ext> [flags]
```

- Provide one or more dot-prefixed source extensions followed by the target extension. Validation
  fails if any token omits the leading dot or repeats the target exactly.
- Source extensions are normalized case-insensitively; duplicates and no-op tokens are surfaced as
  warnings in the preview rather than silently ignored.
- Preview output lists every candidate with `changed`, `no change`, or `skipped` status so scripts
  can detect conflicts before applying. Conflicting targets block apply and exit with a non-zero
  code.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) determine which files and
  directories participate. Hidden assets remain excluded unless `--hidden` is supplied.
- `--dry-run` (default) prints the plan without touching the filesystem. Re-run with `--yes` to
  apply; attempting to combine both flags exits with an error. When no files match, the command
  exits `0` after printing “No candidates found.”

### Usage Examples

- Preview normalization: `renamer extension .jpeg .JPG .jpg --dry-run`
- Apply case-folded extension updates: `renamer extension .yaml .yml .yml --yes --path ./configs`
- Include hidden assets recursively: `renamer extension .TMP .tmp --recursive --hidden`
