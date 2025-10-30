# Phase 0 Research – Insert Command

## Decision: Reuse Traversal + Preview Pipeline for Insert Rule
- **Rationale**: Existing replace/remove/extension commands share a scope walker and preview formatter that already respect `--path`, `-r`, `--hidden`, and ledger logging. Extending this infrastructure minimizes duplication and guarantees constitutional compliance for preview-first safety.
- **Alternatives considered**: Building a standalone walker for insert was rejected because it risks divergence in conflict handling and ledger metadata. Hooking insert into `replace` internals was rejected to keep rule responsibilities separated.

## Decision: Interpret Positions Using Unicode Code Points on Filename Stems
- **Rationale**: Go’s rune indexing treats each Unicode code point as one element, aligning with user expectations for multilingual filenames. Operating on the stem (excluding extension) keeps behavior consistent with common batch-renaming tools.
- **Alternatives considered**: Byte-based offsets were rejected because multi-byte characters would break user expectations. Treating the full filename including extension was rejected to avoid forcing users to re-add extensions manually.

## Decision: Ledger Metadata Includes Position Token and Inserted Text
- **Rationale**: Storing the position directive (`^`, `$`, positive, negative) and inserted string enables precise undo, auditing, and potential future analytics. This mirrors how replace/remove log the rule context.
- **Alternatives considered**: Logging only before/after paths was rejected because it obscures the applied rule and complicates debugging automated runs.

## Decision: Block Apply on Duplicate Targets or Invalid Positions
- **Rationale**: Preventing collisions and out-of-range indices prior to file mutations preserves data integrity and meets preview-first guarantees. Existing conflict detection helpers can be adapted for insert.
- **Alternatives considered**: Allowing apply with overwrites was rejected due to high data-loss risk. Auto-truncating positions was rejected because silent fallback leads to inconsistent results across files.
