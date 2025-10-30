# Tasks: Extension Command for Multi-Extension Normalization

**Input**: Design documents from `/specs/004-extension-rename/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Contract and integration tests are included because the specification mandates deterministic previews, ledger-backed undo, and automation safety.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish initial package and CLI scaffolding.

- [X] T001 Create package doc stub for extension engine in `internal/extension/doc.go`
- [X] T002 Register placeholder Cobra subcommand in `cmd/extension.go` and wire it into `cmd/root.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core data structures and utilities required by all user stories.

- [X] T003 Define `ExtensionRequest` parsing helpers in `internal/extension/request.go`
- [X] T004 Define `ExtensionSummary`, preview statuses, and metadata container in `internal/extension/summary.go`
- [X] T005 Implement case-insensitive normalization and dedup helpers in `internal/extension/normalize.go`

**Checkpoint**: Foundation ready — user story implementation can now begin.

---

## Phase 3: User Story 1 – Normalize Legacy Extensions in Bulk (Priority: P1) 🎯 MVP

**Goal**: Provide preview and apply flows that replace multiple source extensions with a single target extension across the scoped filesystem.

**Independent Test**: In a directory containing mixed `.jpeg`, `.JPG`, and `.png` files, run `renamer extension .jpeg .JPG .png .jpg --dry-run` to inspect preview output, then run with `--yes` to confirm filesystem updates.

### Tests for User Story 1

- [X] T010 [P] [US1] Add preview/apply contract coverage in `tests/contract/extension_command_test.go`
- [X] T011 [P] [US1] Add normalization happy-path integration flow in `tests/integration/extension_flow_test.go`

### Implementation for User Story 1

- [X] T006 [US1] Implement scoped candidate discovery and planning in `internal/extension/engine.go`
- [X] T007 [US1] Render deterministic preview entries with change/no-change markers in `internal/extension/preview.go`
- [X] T008 [US1] Apply planned renames with filesystem operations in `internal/extension/apply.go`
- [X] T009 [US1] Wire Cobra command to preview/apply pipeline with scope flags in `cmd/extension.go`

**Checkpoint**: User Story 1 functional and independently testable.

---

## Phase 4: User Story 2 – Automation-Friendly Extension Updates (Priority: P2)

**Goal**: Ensure ledger entries, exit codes, and undo support enable scripted execution without manual intervention.

**Independent Test**: Execute `renamer extension .yaml .yml .yml --yes --path ./fixtures`, verify exit code `0`, inspect `.renamer` ledger for metadata, then run `renamer undo` to restore originals.

### Tests for User Story 2

- [X] T015 [P] [US2] Extend contract tests to verify ledger metadata and exit codes in `tests/contract/extension_ledger_test.go`
- [X] T016 [P] [US2] Add automation/undo integration scenario in `tests/integration/extension_undo_test.go`

### Implementation for User Story 2

- [X] T012 [US2] Persist extension-specific metadata during apply in `internal/extension/apply.go`
- [X] T013 [US2] Ensure undo and CLI output handle extension batches in `cmd/undo.go`
- [X] T014 [US2] Guarantee deterministic exit codes and non-match notices in `cmd/extension.go`

**Checkpoint**: User Stories 1 and 2 functional and independently testable.

---

## Phase 5: User Story 3 – Validate Extension Inputs and Conflicts (Priority: P3)

**Goal**: Provide robust input validation and conflict detection to prevent unsafe applies.

**Independent Test**: Run `renamer extension .mp3 .MP3 mp3 --dry-run` to confirm validation failure for missing dot, and run a collision scenario to verify preview conflicts block apply.

### Tests for User Story 3

- [X] T020 [P] [US3] Add validation and conflict contract coverage in `tests/contract/extension_validation_test.go`
- [X] T021 [P] [US3] Add conflict-blocking integration scenario in `tests/integration/extension_validation_test.go`

### Implementation for User Story 3

- [X] T017 [US3] Implement CLI argument validation and error messaging in `internal/extension/parser.go`
- [X] T018 [US3] Detect conflicting target paths and accumulate preview warnings in `internal/extension/conflicts.go`
- [X] T019 [US3] Surface validation failures and conflict gating in `cmd/extension.go`

**Checkpoint**: All user stories functional with independent validation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, tooling, and quality improvements spanning multiple stories.

- [X] T022 Update CLI flag documentation for extension command in `docs/cli-flags.md`
- [X] T023 Add smoke test script covering extension preview/apply/undo in `scripts/smoke-test-extension.sh`
- [X] T024 Run gofmt and `go test ./...` from repository root `./`

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Setup (Phase 1)** → primes new package and CLI wiring.
2. **Foundational (Phase 2)** → must complete before any user story work.
3. **User Story Phases (3–5)** → execute in priority order (P1 → P2 → P3) once foundational tasks finish.
4. **Polish (Phase 6)** → after desired user stories are complete.

### User Story Dependencies

- **US1 (P1)** → depends on Phase 2 completion; delivers MVP.
- **US2 (P2)** → depends on US1 groundwork (ledger metadata builds atop apply logic).
- **US3 (P3)** → depends on US1 preview/apply pipeline; validation hooks extend existing command.

### Task Dependencies (Selected)

- T006 depends on T003–T005.
- T007, T008 depend on T006.
- T009 depends on T006–T008.
- T012 depends on T008.
- T013, T014 depend on T012.
- T017 depends on T003, T005.
- T018 depends on T006–T007.
- T019 depends on T017–T018.

---

## Parallel Execution Examples

- **Within US1**: After T009, tasks T010 and T011 can run in parallel to add contract and integration coverage.
- **Across Stories**: Once US1 implementation (T006–T009) is complete, US2 test tasks T015 and T016 can proceed in parallel while T012–T014 are under development.
- **Validation Work**: For US3, T020 and T021 can execute in parallel after T019 ensures CLI gating is wired.

---

## Implementation Strategy

### MVP First

1. Complete Phases 1–2 to establish scaffolding and data structures.
2. Deliver Phase 3 (US1) to unlock core extension normalization (MVP).
3. Validate with contract/integration tests (T010, T011) and smoke through Quickstart scenario.

### Incremental Delivery

1. After MVP, implement Phase 4 (US2) to add ledger/undo guarantees for automation.
2. Follow with Phase 5 (US3) to harden validation and conflict handling.
3. Finish with Phase 6 polish tasks for documentation and operational scripts.

### Parallel Approach

1. One developer completes Phases 1–2.
2. Parallelize US1 implementation (engine vs. CLI vs. tests) once foundations are ready.
3. Assign US2 automation tasks to a second developer after US1 apply logic stabilizes.
4. Run US3 validation tasks concurrently with Polish updates once US2 nears completion.
