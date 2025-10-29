# Quickstart: Replace Command with Multi-Pattern Support

## Goal
Demonstrate how to consolidate multiple filename patterns into a single replacement while using the
preview → apply → undo workflow safely.

## Prerequisites
- Go toolchain (>= 1.24) installed for building the CLI locally.
- Sample directory with files containing inconsistent substrings (e.g., `draft`, `Draft`, `DRAFT`).

## Steps

1. **Build the CLI**
   ```bash
   go build -o renamer ./...
   ```

2. **Inspect replace help**
   ```bash
   ./renamer replace --help
   ```
   Review syntax, especially the "final argument is replacement" guidance and quoting rules.

3. **Run a preview with multiple patterns**
   ```bash
   ./renamer replace draft Draft DRAFT final --dry-run
   ```
   Confirm the table shows each occurrence mapped to `final` and the summary lists per-pattern counts.

4. **Apply replacements after review**
   ```bash
   ./renamer replace draft Draft DRAFT final --yes
   ```
   Observe the confirmation summary, then verify file names have been updated.

5. **Undo if necessary**
   ```bash
   ./renamer undo
   ```
   Ensure the ledger entry created by step 4 is reversed and filenames restored.

6. **Handle patterns with spaces**
   ```bash
   ./renamer replace "Project X" "Project-X" ProjectX --dry-run
   ```
   Verify that quoting preserves whitespace and the preview reflects the intended substitution.

7. **Combine with scope filters**
   ```bash
 ./renamer replace tmp temp stable --path ./examples --extensions .log|.txt --recursive
  ```
  Confirm only matching files under `./examples` are listed.

8. **Integrate into automation**
   ```bash
   ./renamer replace draft Draft final --dry-run && \
   ./renamer replace draft Draft final --yes
   ```
   The first command previews changes; the second applies them with exit code `0` when successful.

## Next Steps
- Add contract/integration tests covering edge cases (empty replacement, conflicts).
- Update documentation (`docs/cli-flags.md`, README) with replace command examples.
