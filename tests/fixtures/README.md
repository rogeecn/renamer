# Test Fixtures

This directory stores sample filesystem layouts used by CLI integration and contract
tests. Keep fixtures small and descriptive so test output remains easy to reason
about.

## Naming Conventions

- Use one subdirectory per scenario (e.g., `basic-mixed-types`, `nested-hidden-files`).
- Include a `README.md` inside complex scenarios to explain intent when necessary.
- Avoid binary assets larger than a few kilobytes; prefer small text placeholders.

## Maintenance Tips

- Regenerate fixture trees with helper scripts instead of manual editing whenever possible.
- Document any platform-specific quirks (case sensitivity, symlinks) alongside the fixture.
- Update this file when adding new conventions or shared assumptions.
