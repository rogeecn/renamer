# Phase 0 Research: Replace Command with Multi-Pattern Support

## Decision: Literal substring replacement with ordered evaluation
- **Rationale**: Aligns with current rename semantics and keeps user expectations simple for first
  iteration; avoids complexity of regex/glob interactions. Ordered application ensures predictable
  handling of overlapping patterns.
- **Alternatives considered**:
  - *Regex support*: powerful but significantly increases validation surface and user errors.
  - *Simultaneous substitution without order*: risk of ambiguous conflicts when one pattern is subset
    of another.

## Decision: Dedicated replace service under `internal/replace`
- **Rationale**: Keeps responsibilities separated from existing listing module, enabling reusable
  preview + apply logic while encapsulating pattern parsing, summary, and reporting.
- **Alternatives considered**:
  - *Extending existing listing package*: would blur responsibilities between read-only listing and
    mutation workflows.
  - *Embedding in command file*: hinders testability and violates composable rule principle.

## Decision: Pattern delimiter syntax `with`
- **Rationale**: Matches user description and provides a clear boundary between patterns and
  replacement string. Works well with Cobra argument parsing and allows quoting for whitespace.
- **Alternatives considered**:
  - *Using flags for replacement string*: more verbose and inconsistent with provided example.
  - *Special separators like `--`*: less descriptive and increases documentation burden.

## Decision: Conflict detection before apply
- **Rationale**: Maintains Preview-First Safety by ensuring duplicates or invalid filesystem names are
  reported before commit. Reuses existing validation helpers from rename pipeline.
- **Alternatives considered**:
  - *Best-effort renames with partial success*: violates atomic undo expectations.
  - *Skipping conflicting files silently*: unsafe and would erode trust.
