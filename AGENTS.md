# renamer Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-29

## Active Technologies

- Go 1.24 + `spf13/cobra`, `spf13/pflag` (001-list-command-filters)

## Project Structure

```text
src/
tests/
```

## Commands

- `renamer list` — preview rename scope with shared flags before executing changes.
- Persistent scope flags: `--path`, `-r/--recursive`, `-d/--include-dirs`, `--hidden`, `--extensions`.

## Code Style

Go 1.24: Follow standard conventions

## Recent Changes

- 001-list-command-filters: Added Go 1.24 + `spf13/cobra`, `spf13/pflag`
- 001-list-command-filters: Introduced `renamer list` command with shared scope flags and formatters

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
