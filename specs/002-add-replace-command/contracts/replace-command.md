# CLI Contract: `renamer replace`

## Command Synopsis

```bash
renamer replace <pattern1> [pattern2 ...] <replacement> [flags]
```

### Global Flags (inherited from root command)
- `--path <dir>` (defaults to current working directory)
- `-r`, `--recursive`
- `-d`, `--include-dirs`
- `--hidden`
- `-e`, `--extensions <.ext|.ext2>`
- `--yes` — apply without interactive confirmation (used by all mutating commands)
- `--dry-run` — force preview-only run even if `--yes` is supplied

## Description
Batch-replace literal substrings across filenames and directories using shared traversal and preview
infrastructure. All arguments except the final token are treated as patterns; the last argument is
the replacement value.

## Arguments & Flags

| Argument | Required | Description |
|----------|----------|-------------|
| `<pattern...>` | Yes (≥2) | Literal substrings to be replaced. Quotes required when containing spaces. |
| `<replacement>` | Yes | Final positional argument; literal replacement applied to each pattern. |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | `.` | Working directory for traversal (global flag). |
| `-r`, `--recursive` | bool | `false` | Traverse subdirectories depth-first (global flag). |
| `-d`, `--include-dirs` | bool | `false` | Include directory names in replacement scope (global flag). |
| `--hidden` | bool | `false` | Include hidden files/directories (global flag). |
| `-e`, `--extensions` | string | (none) | Restrict replacements to files matching `|`-delimited extensions (global flag). |
| `--yes` | bool | `false` | Skip confirmation prompt and apply immediately after successful preview (global flag). |
| `--dry-run` | bool | `false` | Force preview-only run even if `--yes` is provided (global flag). |

## Exit Codes

| Code | Meaning | Example Trigger |
|------|---------|-----------------|
| `0` | Success | Preview or apply completed without conflicts. |
| `2` | Validation error | Fewer than two arguments supplied or unreadable directory. |
| `3` | Conflict detected | Target filename already exists; user must resolve before retry. |

## Preview Output
- Lists each impacted file with columns: `PATH`, `MATCHED PATTERN`, `NEW PATH`.
- Summary line: `Total: <files> (patterns: pattern1=#, pattern2=#, conflicts=#)`.

## Apply Behavior
- Re-validates preview results before writing changes.
- Writes ledger entry capturing old/new names, patterns, replacement string, timestamp.
- On conflict, aborts without partial renames.

## Validation Rules
- Minimum two unique patterns required (at least one pattern plus replacement).
- Patterns and replacement treated as UTF-8 literals; no regex expansion.
- Duplicate patterns deduplicated with warning in preview summary.
- Replacement applied to every occurrence within file/dir name.
- Empty replacement allowed but requires confirmation message: "Replacement string is empty; affected substrings will be removed."

## Examples

```bash
renamer replace draft Draft DRAFT final
renamer replace "Project X" "Project-X" ProjectX --extensions .txt|.md
renamer replace tmp temp temp-backup stable --hidden --recursive --yes
```
