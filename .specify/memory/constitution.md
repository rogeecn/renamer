<!--
Sync Impact Report
Version change: 1.0.0 → 1.1.0
Modified principles:
- Scope-Aware Traversal → Scope-Aware Traversal (extension filters mandated)
Added sections:
- None
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
Follow-up TODOs:
- None
-->

# Renamer CLI Constitution

## Core Principles

### Preview-First Safety
Renamer MUST present a deterministic preview of every pending rename (files and directories) before
applying changes. Users MUST explicitly confirm the preview or abort; unattended destructive modes
are prohibited. The preview MUST surface conflicts and skipped items so the final rename plan
matches user intent. **Rationale:** Safeguarding rename operations prevents accidental data loss and
builds trust in the tool.

### Persistent Undo Ledger
Every confirmed rename batch MUST append an entry to the `.renamer` ledger in the working
directory, capturing original paths, new paths, applied rules, and timestamps. The CLI MUST expose
an undo command that replays the last ledger entry in reverse and refuses to proceed if the ledger
is inconsistent or missing. Ledger writes MUST be atomic to guarantee reversibility. **Rationale:**
An auditable history is essential for recovering from mistakes and for user confidence.

### Composable Rule Engine
Rename logic MUST be expressed as composable, deterministic rules that can be chained without
mutating shared state. Each rule MUST declare its inputs, validations, and postconditions so new
rules can be added without rewriting existing ones. Rule evaluation MUST be tested independently
and in combination to guarantee predictable outcomes across platforms. **Rationale:** A modular rule
engine unlocks flexibility while containing complexity.

### Scope-Aware Traversal
By default the CLI MUST operate on the current directory, renaming files only. The `-d` flag MUST
explicitly include directory renames, and the optional `-r` flag MUST traverse subdirectories
depth-first while avoiding hidden/system paths unless the user opts in. The `-e` flag MUST accept a
`.`-prefixed, `|`-delimited list of extensions (e.g., `-e .jpg|.exe|.mov`) that filters candidates
before preview, execution, and undo. Traversal MUST protect against escaping the requested scope and
MUST report skipped paths or filters that yield no matches. **Rationale:** Clear scope controls keep
operations targeted and safe for diverse directory structures.

### Ergonomic CLI Stewardship
The CLI MUST use Cobra for command structure, flag parsing, and contextual help. Commands MUST
provide consistent flag naming, validation, exit codes, and scriptable output modes for automation.
Tests MUST cover help text, flag behavior, preview output, and undo flows to guarantee a polished
experience. **Rationale:** A dependable CLI experience drives adoption and lowers operational risk.

## Operational Constraints

- Commands operate relative to the invocation directory; alternate roots MUST be supplied via an
  explicit flag and validated before execution.
- The `.renamer` ledger MUST be treated as append-only, stored alongside user-controlled files, and
  ignored by rename scans to avoid self-modification.
- Preview mode MUST remain the default behavior; force-apply pathways require a future governance
  amendment before implementation is permitted.
- Extension filters MUST require `.`-prefixed tokens, reject duplicate/empty values, and clearly
  indicate when filters exclude all candidates so users can adjust before executing.
- File system interactions MUST handle cross-platform path semantics (case sensitivity, Unicode) via
  Go’s standard libraries or vetted wrappers.

## Development Workflow

- Specs, plans, and tasks MUST document how preview, ledger, and traversal guarantees are satisfied
  before implementation begins.
- Each feature implementation MUST add or update automated tests that demonstrate preview accuracy,
  ledger integrity, and undo safety for the affected rules.
- Code reviews MUST verify compliance with every principle and confirm that documentation explains
  new flags, rules, traversal behaviors, and extension filtering semantics.

## Governance

- **Amendments:** Proposals MUST include impacted principles, updated template guidance, and a dry
  run preview demonstrating continued safety before acceptance.
- **Versioning Policy:** This constitution follows semantic versioning. MAJOR increments reflect
  breaking governance changes, MINOR increments add or materially expand principles/sections, and
  PATCH increments capture clarifications.
- **Compliance Reviews:** Before each release, the maintainer MUST confirm that CLI behavior, tests,
  and documentation satisfy the principles and that the `.renamer` ledger remains reversible.

**Version**: 1.1.0 | **Ratified**: 2025-10-29 | **Last Amended**: 2025-10-29
