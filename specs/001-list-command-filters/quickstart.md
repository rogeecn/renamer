# Quickstart: Cobra List Command with Global Filters

## Goal
Learn how to preview rename scope safely by using the new `renamer list` subcommand with global
filter flags that apply across all commands.

## Prerequisites
- Go toolchain installed (>= 1.24) for building the CLI locally.
- Sample directory containing mixed file types to exercise filters.

## Steps

1. **Build the CLI**
   ```bash
   go build -o renamer ./...
   ```

2. **Inspect available commands and global flags**
   ```bash
   ./renamer --help
   ./renamer list --help
   ```
   Confirm `-r`, `-d`, and `-e` are listed as global flags.

3. **List JPEG assets recursively**
   ```bash
   ./renamer list -r -e .jpg
   ```
   Verify output shows a table with relative paths, types, and sizes. The summary line should report
   total entries found.

4. **Produce automation-friendly output**
   ```bash
   ./renamer list --format plain -e .mov|.mp4 > media-files.txt
   ```
   Inspect `media-files.txt` to confirm one path per line.

5. **Validate filter parity with preview**
 ```bash
  ./renamer list -r -d -e .txt|.md
  ./renamer preview -r -d -e .txt|.md
  ```
  Ensure both commands report the same number of directories, since `-d` suppresses files.

6. **Handle empty results gracefully**
 ```bash
  ./renamer list -e .doesnotexist
  ```
  Expect a friendly message explaining that no entries matched the filters.

7. **Inspect hidden files when needed**
   ```bash
   ./renamer list --hidden -e .env
   ```
   Hidden entries are excluded by default, so `--hidden` opts in explicitly when you need to audit dotfiles.

## Next Steps
- Integrate the list command into scripts that currently run `preview` to guard against unintended
  rename scopes.
- Extend automated tests to cover new filters plus list/preview parity checks.
