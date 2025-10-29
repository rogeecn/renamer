# Replace Command Fixtures

Use this directory to store sample file trees referenced by replace command tests. Keep fixtures
minimal and platform-neutral (ASCII names, small files). Create subdirectories per scenario, e.g.:

- `basic/` — Simple files demonstrating multi-pattern replacements.
- `conflicts/` — Cases that intentionally trigger name collisions for validation.

File contents are typically irrelevant; create empty files unless a test requires data.
