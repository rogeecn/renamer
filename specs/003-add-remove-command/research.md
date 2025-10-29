# Phase 0 Research: Remove Command with Sequential Multi-Pattern Support

## Decision: Sequential removal executed in-memory before filesystem writes
- **Rationale**: Computing the full rename plan in memory guarantees deterministic previews,
  simplifies conflict detection, and avoids partial renames that could increase IO load or leave the
  filesystem inconsistent.
- **Alternatives considered**:
  - *Apply-and-check per token*: rejected due to repeated filesystem mutations and difficulty keeping
    undo history coherent.
  - *Streaming rename per file*: rejected because conflicts can only be detected after all tokens
    apply.

## Decision: Dedicated `internal/remove` package mirroring replace architecture
- **Rationale**: Keeps responsibilities separated (parser, engine, summary) and allows reuse of
  traversal/history helpers. Aligns with Composable Rule Engine principle.
- **Alternatives considered**:
  - *Extending replace package*: rejected to avoid coupling distinct behaviors and tests.
  - *Embedding logic directly in command*: rejected for testability and maintainability reasons.

## Decision: Empty-result handling warns and skips rename
- **Rationale**: Removing multiple tokens could produce empty basenames; skipping prevents creating
  invalid filenames while still informing the user.
- **Alternatives considered**:
  - *Allow empty names*: rejected as unsafe and difficult to undo cleanly on certain filesystems.
  - *Hard fail entire batch*: rejected because unaffected files should still be processed.

## Decision: Ledger metadata records ordered tokens and counts
- **Rationale**: Automation and undo workflows need insight into which tokens were removed and how
  often, mirroring replace’s metadata for consistency.
- **Alternatives considered**:
  - *Only store operations*: insufficient for auditing complex removals.

## Decision: CLI help & quickstart emphasize ordering semantics
- **Rationale**: Sequential behavior is the primary mental model difference from other commands; clear
  documentation reduces support load and user confusion.
- **Alternatives considered**:
  - *Rely on examples alone*: risk of users assuming parallel removal and encountering surprises.
