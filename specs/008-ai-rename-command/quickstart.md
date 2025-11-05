# Quickstart – AI Rename Command

1. **Preview AI suggestions before applying.**
   ```bash
   renamer ai --path ./photos --prompt "Hawaii vacation album"
   ```
   - Traverses `./photos` (non-recursive by default) and sends the collected basenames to `renameFlow`.
   - Displays a preview table with original → suggested names and any validation warnings.

2. **Adjust scope or guidance and regenerate.**
   ```bash
   renamer ai --path ./photos --recursive --hidden \
     --prompt "Group by location, keep capture order"
   ```
   - `--recursive` includes nested folders; `--hidden` opts in hidden files.
   - Re-running the command with updated guidance regenerates suggestions without modifying files.

3. **Apply suggestions non-interactively when satisfied.**
   ```bash
   renamer ai --path ./photos --prompt "Hawaii vacation" --yes
   ```
   - `--yes` skips the interactive confirmation while still logging the preview.
   - Use `--dry-run` to inspect output programmatically without touching the filesystem.

4. **Undo the most recent AI batch if needed.**
   ```bash
   renamer undo
   ```
   - Restores original filenames using the ledger entry created by the AI command.
