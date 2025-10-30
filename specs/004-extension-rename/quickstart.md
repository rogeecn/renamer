# Quickstart – Extension Normalization Command

1. **Preview extension changes.**
   ```bash
   renamer extension .jpeg .JPG .jpg --dry-run
   ```
   - Shows original → proposed paths.
   - Highlights conflicts or “no change” rows for files already ending in `.jpg`.

2. **Include nested directories or hidden assets when needed.**
   ```bash
   renamer extension .yaml .yml .yml --recursive --hidden --dry-run
   ```
   - `--recursive` traverses subdirectories.
   - `--hidden` opt-in keeps hidden files in scope.

3. **Apply changes after confirming preview.**
   ```bash
   renamer extension .jpeg .JPG .jpg --yes
   ```
   - `--yes` auto-confirms preview results.
   - Command exits `0` even if no files matched (prints “no candidates found”).

4. **Undo the most recent batch if needed.**
   ```bash
   renamer undo
   ```
   - Restores original extensions using `.renamer` ledger entry.
