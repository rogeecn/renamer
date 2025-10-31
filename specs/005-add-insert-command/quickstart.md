# Quickstart – Insert Command

1. **Preview an insertion at the start of filenames.**
   ```bash
   renamer insert ^ "[2025] " --dry-run
   ```
   - Shows the prepended tag without applying changes.
   - Useful for tagging archival folders.

2. **Insert text near the end while preserving extensions.**
   ```bash
   renamer insert 1$ "_FINAL" --path ./reports --dry-run
   ```
   - `1$` places the string before the last character of the stem.
   - Combine with `--extensions` to limit to specific file types.

3. **Commit changes once preview looks correct.**
   ```bash
   renamer insert 3 "_QA" --yes --path ./builds
   ```
   - `--yes` auto-confirms using the last preview.
   - Ledger entry records the position token and inserted text.

4. **Undo the most recent insert batch if needed.**
   ```bash
   renamer undo --path ./builds
   ```
   - Restores original names using ledger metadata.
