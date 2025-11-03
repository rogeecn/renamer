# Changelog

## Unreleased

- Add `renamer sequence` subcommand with configurable numbering (start, width, placement—default prefix—separator, number prefix/suffix) and ledger-backed apply/undo flows.
- Add `renamer remove` subcommand with sequential multi-token deletions, empty-name safeguards, and ledger-backed undo.
- Document remove command ordering semantics, duplicate warnings, and automation guidance.
- Add `renamer replace` subcommand supporting multi-pattern replacements, preview/apply/undo, and scope flags.
- Document quoting guidance, `--dry-run` / `--yes` behavior, and automation scenarios for replace command.
- Add `renamer list` subcommand with shared scope flags and plain/table output formats.
- Document global scope flags and hidden-file behavior.
- Add `renamer ai` subcommand with export/import workflow, policy enforcement flags, prompt hash telemetry, and ledger metadata for applied plans.
