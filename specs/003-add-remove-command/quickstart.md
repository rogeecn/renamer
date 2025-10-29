# Quickstart: Remove Command with Sequential Multi-Pattern Support

## Goal
Demonstrate how to delete multiple substrings sequentially from filenames while using the
preview → apply → undo workflow safely.

## Prerequisites
- Go toolchain (>= 1.24) installed for building the CLI locally.
- Sample directory containing files with recurring tokens (e.g., `draft`, `copy`).

## Steps

1. **Build the CLI**
   ```bash
   go build -o renamer ./...
   ```

2. **Inspect remove help**
   ```bash
   ./renamer remove --help
   ```
   Pay attention to sequential behavior: tokens execute in the order provided.

3. **Run a preview with multiple tokens**
   ```bash
   ./renamer remove " copy" " draft" --path ./samples --dry-run
   ```
   Confirm the output table shows each token removed in order and the summary reflects changed files.

4. **Apply removals after review**
   ```bash
   ./renamer remove " copy" " draft" --path ./samples --yes
   ```
   Verify filenames no longer contain the tokens and a ledger entry is created.

5. **Undo if necessary**
   ```bash
   ./renamer undo --path ./samples
   ```
   Ensure filenames return to their original state.

6. **Handle empty-result warnings**
   ```bash
   ./renamer remove "project" "project-" --path ./samples --dry-run
   ```
   Expect the preview to warn and skip items that would collapse to empty names.

7. **Integrate into automation**
   ```bash
   ./renamer remove foo bar --dry-run && \
   ./renamer remove foo bar --yes --path ./automation
   ```
   Use non-zero exit codes to detect invalid input in scripts.

## Next Steps
- Add contract tests covering sequential removals and empty-name warnings.
- Extend documentation (CLI reference, changelog) with remove command examples.
