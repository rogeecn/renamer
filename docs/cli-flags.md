# CLI Scope Flags

Renamer shares a consistent set of scope flags across every command that inspects or mutates the
filesystem. Use these options at the root command level so they apply to `list`, `preview`, and
`rename` alike.

| Flag | Default | Description |
|------|---------|-------------|
| `-r`, `--recursive` | `false` | Traverse subdirectories depth-first. Symlinked directories are not followed. |
| `-d`, `--include-dirs` | `false` | Limit results to directories only (files and symlinks are suppressed). Directory traversal still occurs even when the flag is absent. |
| `-e`, `--extensions` | *(none)* | Pipe-separated list of file extensions (e.g. `.jpg|.mov`). Tokens must start with a dot, are lowercased internally, and duplicates are ignored. |
| `--hidden` | `false` | Include dot-prefixed files and directories. By default they are excluded from listings and rename previews. |
| `--format` | `table` | Command-specific output formatting option. For `list`, use `table` or `plain`. |

## Validation Rules

- Extension tokens that are empty or missing the leading `.` cause validation errors.
- Filters that match zero entries result in a friendly message and exit code `0`.
- Invalid flag combinations (e.g., unsupported `--format` values) cause the command to exit with a non-zero code.
- Recursive traversal honor `--hidden` and skips unreadable directories while logging warnings.

Keep this document updated whenever a new command is introduced or the global scope behavior
changes.

### Usage Examples

- Preview files recursively: `renamer --recursive preview`
- List JPEGs only: `renamer --extensions .jpg list`
- Include dotfiles: `renamer --hidden --extensions .env list`
