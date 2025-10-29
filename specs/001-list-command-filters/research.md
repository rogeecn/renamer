# Phase 0 Research: Cobra List Command with Global Filters

## Decision: Promote scope flags to Cobra root command persistent flags
- **Rationale**: Persistent flags defined on the root command automatically apply to all
  subcommands, ensuring a single source of truth for recursion (`-r`), directory inclusion (`-d`),
  and extension filters (`-e`). This prevents divergence between `list`, `preview`, and future
  rename commands.
- **Alternatives considered**:
  - *Duplicate flag definitions per subcommand*: rejected because it risks inconsistent validation
    and requires keeping help text in sync.
  - *Environment variables for shared filters*: rejected because CLI users expect flag-driven scope
    control and env vars complicate scripting.

## Decision: Stream directory traversal using `filepath.WalkDir` with early emission
- **Rationale**: `WalkDir` supports depth-first traversal with built-in symlink detection and allows
  emitting entries as they are encountered, keeping memory usage bounded even for directories with
  >10k items.
- **Alternatives considered**:
  - *Preloading all entries into slices before formatting*: rejected because it inflates memory and
    delays first output, violating the 2-second responsiveness goal.
  - *Shelling out to `find` or OS-specific tools*: rejected due to portability concerns and reduced
    testability inside Go.

## Decision: Provide dual output renderers (table + plain) via shared formatter interface
- **Rationale**: Implementing a small formatter interface allows easy expansion of output modes and
  keeps the list command decoupled from presentation details. A table renderer can rely on
  text/tabwriter, while the plain renderer writes newline-delimited paths to satisfy scripting use
  cases.
- **Alternatives considered**:
  - *Hard-coding formatted strings inside the command*: rejected because it complicates testing and
    future format additions (JSON, CSV).
  - *Introducing a heavy templating library*: rejected as unnecessary overhead for simple CLI
    output.
