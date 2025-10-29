# CLI Contract: `renamer list`

## Command Synopsis

```bash
renamer list [--path <dir>] [-r] [-d] [-e .ext|.ext2] [--format table|plain]
```

## Description
Enumerates filesystem entries that match the active global filters without applying any rename
operations. Output supports human-friendly table view and automation-friendly plain mode. Results
mirror the candidate set used by `renamer preview` and `renamer rename`.

## Arguments & Flags

| Flag | Type | Default | Required | Applies To | Description |
|------|------|---------|----------|------------|-------------|
| `--path` | string | current directory | No | root command | Directory to operate on; must exist and be readable |
| `-r`, `--recursive` | bool | `false` | No | root command | Enable depth-first traversal through subdirectories |
| `-d`, `--include-dirs` | bool | `false` | No | root command | Include directories in output alongside files |
| `-e`, `--extensions` | string | (none) | No | root command | `|`-delimited list of `.`-prefixed extensions used to filter files |
| `--format` | enum (`table`, `plain`) | `table` | No | list subcommand | Controls output rendering style |
| `--limit` | int | 0 (no limit) | No | list subcommand | Optional cap on number of entries returned; 0 means unlimited |
| `--no-progress` | bool | `false` | No | list subcommand | Suppress progress indicators for scripting |

## Exit Codes

| Code | Meaning | Example Trigger |
|------|---------|-----------------|
| `0` | Success | Listing completed even if zero entries matched |
| `2` | Validation error | Duplicate/empty extension token, unreadable path |
| `3` | Traversal failure | I/O error during directory walk |

## Output Formats

### Table (default)
```
PATH                       TYPE        SIZE
photos/2024/event.jpg      file        3.2 MB
photos/2024               directory   —
```

### Plain (`--format plain`)
```
photos/2024/event.jpg
photos/2024
```

Both formats MUST include a trailing summary line:
```
Total: 42 entries (files: 38, directories: 4)
```

## Validation Rules
- Extensions MUST begin with `.` and be deduplicated case-insensitively.
- When no entries remain after filtering, the command prints `No entries matched the provided
  filters.` and exits with code `0`.
- Symlinks MUST be reported with type `symlink` and NOT followed recursively unless future scope
  explicitly enables it.
