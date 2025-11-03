# Research Log

## Cobra Flag Validation For Sequence Command
- **Decision**: Validate `--start`, `--width`, `--placement`, and `--separator` flags using Cobra command `PreRunE` with shared helpers, returning errors for invalid inputs before preview executes.
- **Rationale**: Cobra documentation and community guides recommend using `RunE`/`PreRunE` to surface validation errors with non-zero exit codes, ensuring CLI consistency and enabling tests to cover messaging.
- **Alternatives considered**: Inline validation inside business logic (rejected—mixes CLI parsing with domain rules and complicates contract tests); custom flag types (rejected—adds complexity without additional value).

## Sequence Ordering And Determinism
- **Decision**: Reuse the existing traversal service to produce a stable, path-sorted candidate list and derive sequence numbers from index positions in the preview plan.
- **Rationale**: Internal traversal package already guarantees deterministic ordering and filtering; leveraging it avoids duplicating scope logic and satisfies Preview-First and Scope-Aware principles.
- **Alternatives considered**: Implement ad-hoc sorting inside sequence rule (rejected—risk of diverging from other commands); rely on filesystem iteration order (rejected—non-deterministic across platforms).

## Ledger Metadata Capture
- **Decision**: Extend history ledger entries with a new sequence record type storing sequence parameters and per-file mappings, ensuring undo can skip missing files but restore others.
- **Rationale**: Existing ledger pattern (as used by replace/remove commands) stores rule metadata for undo; following same structure keeps undo consistent and auditable.
- **Alternatives considered**: Store only file rename pairs without parameters (rejected—undo would lack context if future migrations require differentiation); create a separate ledger file (rejected—breaks append-only guarantee).

## Conflict Handling Strategy
- **Decision**: During apply, skip conflicting file targets, log a warning via output package, and continue numbering remaining candidates; conflicts remain in preview so users can resolve them beforehand.
- **Rationale**: Aligns with clarified requirements and minimizes partial ledger entries while informing users; consistent with existing warning infrastructure used by other commands.
- **Alternatives considered**: Abort entire batch on conflict (rejected—user explicitly requested skip behavior); auto-adjust numbers (rejected—violates preview/apply parity).

## Directory Inclusion Policy
- **Decision**: Filter traversal results so directories included via `--include-dirs` are reported but not renamed; numbering applies only to file candidates.
- **Rationale**: Keeps command behavior predictable, avoids confusing two numbering schemes, and respects clarified requirement without altering traversal contract tests.
- **Alternatives considered**: Separate numbering sequences for files vs directories (rejected—adds complexity with little user need); rename directories by default (rejected—breaks clarified guidance).
