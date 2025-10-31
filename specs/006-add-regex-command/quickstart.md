# Quickstart – Regex Command

1. **Preview a capture-group rename before applying.**
   ```bash
   renamer regex '^(\d{4})-(\d{2})_(.*)$' 'Q@2-@1_@3' --dry-run
   ```
   - Converts `2025-01_report.txt` into `Q01-2025_report.txt` in preview mode.
   - Skipped files remain untouched and are labeled in the preview table.

2. **Limit scope with extension and directory flags.**
   ```bash
   renamer regex '^(build)_(\d+)_v(.*)$' 'release-@2-@1-v@3' --path ./artifacts --extensions .zip|.tar.gz --include-dirs --dry-run
   ```
   - Applies only to archives under `./artifacts`, including subdirectories when paired with `-r`.
   - Hidden files remain excluded unless `--hidden` is added.

3. **Apply changes non-interactively for automation.**
   ```bash
   renamer regex '^(feature)-(.*)$' '@2-@1' --yes --path ./staging
   ```
   - `--yes` confirms using the preview plan and writes a ledger entry containing pattern and template metadata.
   - Exit code `0` indicates success; non-zero signals validation or conflict issues.

4. **Undo the last regex batch if results are unexpected.**
   ```bash
   renamer undo --path ./staging
   ```
   - Restores original filenames using the `.renamer` ledger captured during apply.
