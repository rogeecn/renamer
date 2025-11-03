# Tasks: Sequence Numbering Command

**Input**: Design documents from `/specs/001-sequence-numbering/`
**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish scaffolding required by all user stories.

- [X] T001 Create sequence package documentation stub in `internal/sequence/doc.go`
- [X] T002 Seed sample fixtures for numbering scenarios in `testdata/sequence/basic/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared components that every sequence story depends on.

- [X] T003 Define sequence options struct with default values in `internal/sequence/options.go`
- [X] T004 Implement zero-padding formatter helper in `internal/sequence/format.go`
- [X] T005 Introduce plan and summary data structures in `internal/sequence/plan.go`

**Checkpoint**: Base package compiles with shared types ready for story work.

---

## Phase 3: User Story 1 - Add Sequential Indices to Batch (Priority: P1) 🎯 MVP

**Goal**: Append auto-incremented suffixes (e.g., `_001`) to scoped files with deterministic ordering and ledger persistence.

**Independent Test**: `renamer sequence --dry-run --path <dir>` on three files shows `_001`, `_002`, `_003`; rerun with `--yes` updates ledger.

### Tests for User Story 1

- [X] T006 [P] [US1] Add preview contract test covering default numbering in `tests/contract/sequence_preview_test.go`
- [X] T007 [P] [US1] Add integration flow test verifying preview/apply parity in `tests/integration/sequence_flow_test.go`

### Implementation for User Story 1

- [X] T008 [US1] Implement candidate traversal adapter using listing scope in `internal/sequence/traversal.go`
- [X] T009 [US1] Generate preview plan with conflict detection in `internal/sequence/preview.go`
- [X] T010 [US1] Apply renames and record sequence metadata in `internal/sequence/apply.go`
- [X] T011 [US1] Wire Cobra sequence command execution in `cmd/sequence.go`
- [X] T012 [US1] Register sequence command on the root command in `cmd/root.go`

**Checkpoint**: Sequence preview/apply for default suffix behavior is fully testable and undoable.

---

## Phase 4: User Story 2 - Control Number Formatting (Priority: P2)

**Goal**: Allow explicit width flag with zero padding and warning when auto-expanding.

**Independent Test**: `renamer sequence --width 4 --dry-run` shows `_0001` suffixes; omitting width auto-expands on demand.

### Tests for User Story 2

- [X] T013 [P] [US2] Add contract test for explicit width padding in `tests/contract/sequence_width_test.go`
- [X] T014 [P] [US2] Add integration test validating width flag and warnings in `tests/integration/sequence_width_test.go`

### Implementation for User Story 2

- [X] T015 [US2] Extend options validation to handle width flag rules in `internal/sequence/options.go`
- [X] T016 [US2] Update preview planner to enforce configured width and warnings in `internal/sequence/preview.go`
- [X] T017 [US2] Parse and bind `--width` flag within Cobra command in `cmd/sequence.go`

**Checkpoint**: Users can control sequence width with deterministic zero-padding.

---

## Phase 5: User Story 3 - Configure Starting Number and Placement (Priority: P3)

**Goal**: Support custom start offsets plus prefix/suffix placement with configurable separator.

**Independent Test**: `renamer sequence --start 10 --placement prefix --separator "-" --dry-run` produces `0010-file.ext` entries.

### Tests for User Story 3

- [X] T018 [P] [US3] Add contract test for start and placement variants in `tests/contract/sequence_placement_test.go`
- [X] T019 [P] [US3] Add integration test for start offset with undo coverage in `tests/integration/sequence_start_test.go`

### Implementation for User Story 3

- [X] T020 [US3] Validate start, placement, and separator flags in `internal/sequence/options.go`
- [X] T021 [US3] Update preview generation to honor prefix/suffix placement and separators in `internal/sequence/preview.go`
- [X] T022 [US3] Persist placement and separator metadata during apply in `internal/sequence/apply.go`
- [X] T023 [US3] Wire `--start`, `--placement`, and `--separator` flags in `cmd/sequence.go`

**Checkpoint**: Placement and numbering customization scenarios fully supported with ledger fidelity.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T024 [P] Document sequence command flags in `docs/cli-flags.md`
- [X] T025 [P] Log sequence feature addition in `docs/CHANGELOG.md`
- [X] T026 [P] Update command overview with sequence entry in `README.md`

---

## Dependencies

- Setup (Phase 1) → Foundational (Phase 2) → US1 (Phase 3) → US2 (Phase 4) → US3 (Phase 5) → Polish (Phase 6)
- User Story dependencies: `US1` completion unlocks `US2`; `US2` completion unlocks `US3`.

## Parallel Execution Opportunities

- Contract and integration test authoring tasks (T006, T007, T013, T014, T018, T019) can run concurrently with implementation once shared scaffolding is ready.
- Documentation polish tasks (T024–T026) can be executed in parallel after all story implementations stabilize.

## Implementation Strategy

1. Deliver MVP by completing Phase 1–3 (US1), enabling default sequence numbering with undo.
2. Iterate with formatting controls (Phase 4) to broaden usability while maintaining preview/apply parity.
3. Finish with placement customization (Phase 5) and polish tasks (Phase 6) before release.
