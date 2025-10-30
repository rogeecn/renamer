# Phase 0 Research – Extension Command

## Decision: Reuse Replace Engine Structure for Extension Normalization
- **Rationale**: The replace command already supports preview/apply workflows, ledger logging, and shared traversal flags. By reusing its service abstractions (scope resolution → rule application → preview), we minimize new surface area while ensuring compliance with preview-first safety.
- **Alternatives considered**: Building a standalone engine dedicated to extensions was rejected because it would duplicate scope traversal, preview formatting, and ledger writing logic, increasing maintenance and divergence risk.

## Decision: Normalize Extensions Using Case-Insensitive Matching
- **Rationale**: Filesystems differ in case sensitivity; normalizing via `strings.EqualFold` (or equivalent) ensures consistent behavior regardless of platform, aligning with the spec’s clarification and reducing surprise for users migrating mixed-case assets.
- **Alternatives considered**: Relying on filesystem semantics was rejected because it would produce divergent behavior (e.g., Linux vs. macOS). Requiring exact-case matches was rejected for being unfriendly to legacy archives with mixed casing.

## Decision: Record Detailed Extension Metadata in Ledger Entries
- **Rationale**: Persisting the original extension list, target extension, and per-file before/after paths in ledger metadata keeps undo operations auditable and allows future analytics (e.g., reporting normalized counts). Existing ledger schema supports additional metadata fields without incompatible changes.
- **Alternatives considered**: Storing only file path mappings was rejected because it obscures which extensions were targeted, hindering debugging. Creating a new ledger file was rejected for complicating undo logic.

## Decision: Extend Cobra CLI with Positional Arguments for Extensions
- **Rationale**: Cobra natural handling of positional args enables `renamer extension [sources...] [target]`. Using argument parsing consistent with replace/remove reduces UX learning curve, and Cobra’s validation hooks simplify enforcing leading-dot requirements.
- **Alternatives considered**: Introducing new flags (e.g., `--sources`, `--target`) was rejected because it diverges from existing command patterns and complicates scripting. Using prompts was rejected due to automation needs.
