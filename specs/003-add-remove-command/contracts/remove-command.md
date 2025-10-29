# CLI Contract: `renamer remove`

## Command Synopsis

```bash
renamer remove <pattern1> [pattern2 ...] [flags]
```

### Global Flags (inherited from root command)
- `--path <dir>` (defaults to current working directory)
- `-r`, `--recursive`
- `-d`, `--include-dirs`
- `--hidden`
- `-e`, `--extensions <.ext|.ext2>`
- `--dry-run`
- `--yes`

## Description
Sequentially removes literal substrings from file and directory names. Every token is applied in the
order provided, and the resulting name is used for subsequent removals before any filesystem rename
occurs.

## Arguments & Flags

| Argument | Required | Description |
|----------|----------|-------------|
| `<pattern...>` | Yes (≥1) | Literal substrings to remove sequentially. Quotes required when tokens contain spaces. |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | `.` | Working directory for traversal (global flag). |
| `-r`, `--recursive` | bool | `false` | Traverse subdirectories (global flag). |
| `-d`, `--include-dirs` | bool | `false` | Include directory names in removal scope (global flag). |
| `--hidden` | bool | `false` | Include hidden files/directories (global flag). |
| `-e`, `--extensions` | string | (none) | Restrict removals to files matching `|`-delimited extensions (global flag). |
| `--dry-run` | bool | `false` | Preview only; print proposed removals without applying (global flag). |
| `--yes` | bool | `false` | Apply changes without interactive prompt (global flag). |

## Exit Codes

| Code | Meaning | Example Trigger |
|------|---------|-----------------|
| `0` | Success | Preview completed or apply executed without conflicts. |
| `2` | Validation error | No patterns provided, empty token after trimming, unreadable directory. |
| `3` | Conflict detected | Target filename already exists after removals. |

## Preview Output
- Lists each impacted item with columns: `PATH`, `TOKENS REMOVED`, `NEW PATH`.
- Summary line: `Total: <candidates> (changed: <count>, empties: <skipped>, conflicts: <count>)`.

## Apply Behavior
- Re-validates preview results, then performs renames in sequence while tracking undo metadata.
- Writes ledger entry containing ordered token list and per-token match counts.
- Aborts without partial renames if conflicts arise between preview and apply.

## Validation Rules
- Tokens are deduplicated case-sensitively but order of first occurrence preserved; duplicates logged
  as warnings.
- Resulting names that collapse to empty strings are skipped with warnings.
- Conflicts (targets already existing) abort the operation; users must resolve manually.

## Examples

```bash
renamer remove " copy" " draft"
renamer remove foo foo- bar --path ./reports --recursive --dry-run
renamer remove "Project X" " X" --extensions .txt|.md --yes
```
