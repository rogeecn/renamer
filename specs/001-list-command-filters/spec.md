# Feature Specification: Cobra List Command with Global Filters

**Feature Branch**: `001-list-command-filters`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "实现文件列表遍历展示cobra 子命令（list），支持当前系统要求的过滤参数（过滤参数为全局生效，所以应该指定到root command上）。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discover Filtered Files Before Renaming (Priority: P1)

As a command-line user preparing a batch rename, I want to run `renamer list` with the same filters
that the rename command will honor so I can preview the exact files that will be affected.

**Why this priority**: Prevents accidental renames by making scope explicit before any destructive
action.

**Independent Test**: Execute `renamer list -e .jpg|.png` in a sample directory and verify the output
lists only matching files without performing any rename.

**Acceptance Scenarios**:

1. **Given** a directory containing mixed file types, **When** the user runs `renamer list -e .jpg`,
   **Then** the output only includes `.jpg` files and reports the total count.
2. **Given** a directory tree with nested folders, **When** the user runs `renamer list -r`,
   **Then** results include files from subdirectories with each entry showing the relative path.

---

### User Story 2 - Apply Global Filters Consistently (Priority: P2)

As an operator scripting renamer commands, I want filter flags (`-r`, `-d`, `-e`) to be defined on
the root command so they apply consistently to `list`, `preview`, and future subcommands without
redundant configuration.

**Why this priority**: Ensures a single source of truth for scope filters, reducing user error and
documentation complexity.

**Independent Test**: Run `renamer list` and `renamer preview` with the same global flags in a
script, confirming both commands interpret scope identically.

**Acceptance Scenarios**:

1. **Given** the root command defines global filter flags, **When** the user specifies `--extensions
   .mov|.mp4` with `renamer list`, **Then** running `renamer preview` in the same shell session with
   identical flags produces the same candidate set.

---

### User Story 3 - Review Listing Output Comfortably (Priority: P3)

As a user reviewing large directories, I want the `list` output to provide human-readable columns
and an optional machine-friendly format so I can spot issues quickly or pipe results into other
tools.

**Why this priority**: Good ergonomics encourage the list command to become part of every workflow,
increasing safety and adoption.

**Independent Test**: Run `renamer list --format table` and `renamer list --format plain` to confirm
both modes display the same entries in different formats without extra configuration.

**Acceptance Scenarios**:

1. **Given** the list command is executed with default settings, **When** multiple files are
   returned, **Then** the output presents aligned columns containing path, type (file/directory),
   and size.
2. **Given** the user supplies `--format plain`, **When** the command runs, **Then** the output
   emits one path per line suitable for piping into other commands.

---

### Edge Cases

- Directory contains no items after filters are applied; command must exit gracefully with zero
  results and a clear message.
- Filters include duplicate or malformed extensions (e.g., `-e .jpg||.png`); command must reject the
  input and surface a descriptive validation error.
- Listing directories requires read permissions; command must skip unreadable paths, log warnings,
  and continue scanning allowed paths.
- File system contains symbolic links or junctions; traversal must avoid infinite loops and clearly
  mark symlinked entries.
- Large directories (>10k entries); command must stream results without excessive memory usage and
  display progress feedback when execution exceeds a user-friendly threshold.
- Hidden files may be unintentionally included; unless `--hidden` is provided, they must remain
  excluded and the help text should explain how to opt in.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The root Cobra command MUST expose global filter flags for recursion (`-r`), directory
  inclusion (`-d`), and extension filtering (`-e .ext|.ext2`) that apply to all subcommands.
- **FR-002**: The `list` subcommand MUST enumerate files and directories matching the active filters
  within the current (or user-specified) working directory without modifying the filesystem.
- **FR-003**: Listing results MUST be deterministic: entries sorted lexicographically by relative
  path with directories identified distinctly from files.
- **FR-004**: The command MUST support at least two output formats: a human-readable table (default)
  and a plain-text list (`--format plain`) for automation.
- **FR-005**: When filters exclude all entries, the CLI MUST communicate that zero results were
  found and suggest reviewing filter parameters.
- **FR-006**: The `list` subcommand MUST share validation and traversal utilities with preview and
  rename flows to guarantee identical scope resolution across commands.
- **FR-007**: The command MUST return a non-zero exit code when input validation fails and zero when
  execution completes successfully, enabling scripting.
- **FR-008**: Hidden files and directories MUST be excluded by default and only included when users
  explicitly pass a `--hidden` flag.

### Key Entities *(include if feature involves data)*

- **ListingRequest**: Captures active filters (recursion, directory inclusion, extensions, path) and
  desired output format for a listing invocation.
- **ListingEntry**: Represents a single file or directory discovered during traversal, including its
  relative path, type, size (bytes), and depth.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can execute `renamer list` with filters on a directory containing up to 5,000
  entries and receive the first page of results within 2 seconds.
- **SC-002**: In usability testing, 90% of participants correctly predict which files will be renamed
  after reviewing `renamer list` output.
- **SC-003**: Automated regression tests confirm that `list`, `preview`, and `rename` commands return
  identical candidate counts for the same filter combinations in 100% of tested scenarios.
- **SC-004**: Support requests related to "unexpected files being renamed" decrease by 30% in the
  first release cycle following launch.

## Assumptions

- Default output format is a table suitable for terminals with ANSI support; plain text is available
  for scripting without additional flags beyond `--format`.
- Users run commands from the directory they intend to operate on; specifying alternative roots will
  follow existing conventions (e.g., `--path`) if already provided by the tool.
- Existing preview and rename workflows already rely on shared traversal utilities that can be
  extended for the list command.

## Dependencies & Risks

- Requires traversal utilities to handle large directories efficiently; performance optimizations may
  be needed if current implementation does not stream results.
- Depends on existing validation logic for extension filtering and directory scopes; any divergence
  introduces inconsistency between commands.
- Risk of confusing users if help documentation is not updated to emphasize using `list` before
  running rename operations.

## Clarifications

### Session 2025-10-29

- Q: Should the list command include hidden files by default or require an explicit opt-in? → A:
  Exclude hidden files by default; add a `--hidden` flag to include them.
