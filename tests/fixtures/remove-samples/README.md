# Remove Command Fixtures

Sample directory structures used by remove command tests. Keep filenames ASCII and small.

- `basic/` — General-purpose samples demonstrating sequential token removals.
- `conflicts/` — Files that collide after token removal to exercise conflict handling.
- `empties/` — Names that collapse to empty or near-empty results to validate warnings.

Add directories/files sparingly so tests stay fast and portable.
