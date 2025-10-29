# CLI Scope Flags

Renamer shares a consistent set of scope flags across every command that inspects or mutates the
filesystem. Use these options at the root command level so they apply to all subcommands (`list`,
`replace`, future `preview`/`rename`, etc.).

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | `.` | Working directory root for traversal. |
| `-r`, `--recursive` | `false` | Traverse subdirectories depth-first. Symlinked directories are not followed. |
| `-d`, `--include-dirs` | `false` | Limit results to directories only (files and symlinks are suppressed). Directory traversal still occurs even when the flag is absent. |
| `-e`, `--extensions` | *(none)* | Pipe-separated list of file extensions (e.g. `.jpg|.mov`). Tokens must start with a dot, are lowercased internally, and duplicates are ignored. |
| `--hidden` | `false` | Include dot-prefixed files and directories. By default they are excluded from listings and rename previews. |
| `--yes` | `false` | Apply changes without an interactive confirmation prompt (mutating commands only). |
| `--dry-run` | `false` | Force preview-only behavior even when `--yes` is supplied. |
| `--format` | `table` | Command-specific output formatting option. For `list`, use `table` or `plain`. |

## Validation Rules

- Extension tokens that are empty or missing the leading `.` cause validation errors.
- Filters that match zero entries result in a friendly message and exit code `0`.
- Invalid flag combinations (e.g., unsupported `--format` values) cause the command to exit with a non-zero code.
- Recursive traversal honor `--hidden` and skips unreadable directories while logging warnings.

Keep this document updated whenever a new command is introduced or the global scope behavior
changes.

## Replace Command Quick Reference

```bash
renamer replace <pattern1> [pattern2 ...] <replacement> [flags]
```

- The **final positional argument** is the replacement value; all preceding arguments are treated as
  literal patterns (quotes required when a pattern contains spaces).
- Patterns are applied sequentially and replaced with the same value. Duplicate patterns are
  deduplicated automatically and surfaced in the preview summary.
- Empty replacement strings are allowed (effectively deleting each pattern) but the preview warns
  before confirmation.
- Combine with scope flags (`--path`, `-r`, `--include-dirs`, etc.) to target the desired set of
  files/directories.
- Use `--dry-run` to preview in scripts, then `--yes` to apply once satisfied; combining both flags
  exits with an error to prevent accidental automation mistakes.

### Usage Examples

- Preview files recursively: `renamer --recursive preview`
- List JPEGs only: `renamer --extensions .jpg list`
- Replace multiple patterns: `renamer replace draft Draft final --dry-run`
- Include dotfiles: `renamer --hidden --extensions .env list`
