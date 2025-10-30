# Tasks: Insert Command for Positional Text Injection

**Input**: Design documents from `/specs/005-add-insert-command/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Contract and integration tests included per spec emphasis on preview determinism, ledger integrity, and Unicode handling.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish initial command scaffolding and directories.

- [X] T001 Create insert package scaffolding `internal/insert/doc.go`
- [X] T002 Add placeholder Cobra command entry point `cmd/insert.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core parsing, summary structures, and helpers needed by all stories.

- [X] T003 Define `InsertRequest` builder and execution mode helpers in `internal/insert/request.go`
- [X] T004 Implement `InsertSummary`, preview entries, and conflict types in `internal/insert/summary.go`
- [X] T005 Build Unicode-aware position parsing and normalization utilities in `internal/insert/positions.go`

**Checkpoint**: Foundation ready — user story implementation can now begin.

---

## Phase 3: User Story 1 – Insert Text at Target Position (Priority: P1) 🎯 MVP

**Goal**: Provide preview + apply flow that inserts text at specified positions with Unicode handling.

**Independent Test**: `renamer insert 3 _tag --dry-run` confirms preview insertion per code point ordering; `--yes` applies and ledger logs metadata.

### Tests

- [X] T006 [P] [US1] Add contract preview/apply coverage in `tests/contract/insert_command_test.go`
- [X] T007 [P] [US1] Add integration flow test for positional insert in `tests/integration/insert_flow_test.go`

### Implementation

- [X] T008 [US1] Implement planning engine to compute proposed names in `internal/insert/engine.go`
- [X] T009 [US1] Render preview output with highlighted segments in `internal/insert/preview.go`
- [X] T010 [US1] Apply filesystem changes with ledger logging in `internal/insert/apply.go`
- [X] T011 [US1] Wire Cobra command to parse args, perform preview/apply in `cmd/insert.go`

**Checkpoint**: User Story 1 functionality testable end-to-end.

---

## Phase 4: User Story 2 – Automation-Friendly Batch Inserts (Priority: P2)

**Goal**: Ensure ledger metadata, undo, and exit codes support automation.

**Independent Test**: `renamer insert $ _ARCHIVE --yes --path ./fixtures` exits `0` with ledger metadata; `renamer undo` restores filenames.

### Tests

- [ ] T012 [P] [US2] Extend contract tests for ledger metadata & exit codes in `tests/contract/insert_ledger_test.go`
- [ ] T013 [P] [US2] Add automation/undo integration scenario in `tests/integration/insert_undo_test.go`

### Implementation

- [ ] T014 [US2] Persist position token and inserted text in ledger metadata via `internal/insert/apply.go`
- [ ] T015 [US2] Enhance undo CLI feedback for insert batches in `cmd/undo.go`
- [ ] T016 [US2] Ensure zero-match runs exit `0` with notice in `cmd/insert.go`

**Checkpoint**: User Stories 1 & 2 independently verifiable.

---

## Phase 5: User Story 3 – Validate Positions and Multilingual Inputs (Priority: P3)

**Goal**: Robust validation, conflict detection, and messaging for out-of-range or conflicting inserts.

**Independent Test**: Invalid positions produce descriptive errors; duplicate targets block apply; Chinese filenames preview correctly.

### Tests

- [X] T017 [P] [US3] Add validation/conflict contract coverage in `tests/contract/insert_validation_test.go`
- [X] T018 [P] [US3] Add conflict-blocking integration scenario in `tests/integration/insert_validation_test.go`

### Implementation

- [X] T019 [US3] Implement parsing + error messaging for position tokens in `internal/insert/parser.go`
- [X] T020 [US3] Detect conflicting targets and report warnings in `internal/insert/conflicts.go`
- [X] T021 [US3] Surface validation failures and conflict gating in `cmd/insert.go`

**Checkpoint**: All user stories function with robust validation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, tooling, and quality improvements.

- [X] T022 Update CLI flags documentation for insert command in `docs/cli-flags.md`
- [X] T023 Add insert smoke test script `scripts/smoke-test-insert.sh`
- [X] T024 Run gofmt and `go test ./...` from repo root `./`

---

## Dependencies & Execution Order

### Phase Dependencies

1. Phase 1 (Setup) → groundwork for new command/package.
2. Phase 2 (Foundational) → required before user story work.
3. Phase 3 (US1) → delivers MVP after foundational tasks.
4. Phase 4 (US2) → builds on US1 for automation support.
5. Phase 5 (US3) → extends validation/conflict handling.
6. Phase 6 (Polish) → final documentation and quality checks.

### User Story Dependencies

- US1 depends on foundational tasks only.
- US2 depends on US1 implementation (ledger/apply logic).
- US3 depends on US1 preview/apply and US2 ledger updates.

### Task Dependencies (selected)

- T008 requires T003–T005.
- T009, T010 depend on T008.
- T011 depends on T008–T010.
- T014 depends on T010.
- T015 depends on T014.
- T019 depends on T003, T005.
- T020 depends on T008, T009.
- T021 depends on T019–T020.

---

## Parallel Execution Examples

- Within US1, tasks T006 and T007 can run in parallel once T011 is in progress.
- Within US2, tests T012/T013 may execute while T014–T016 are implemented.
- Within US3, contract vs integration tests (T017/T018) can proceed concurrently after T021 adjustments.

---

## Implementation Strategy

### MVP (US1)
1. Complete Phases 1–2 foundation.
2. Deliver Phase 3 (US1) to enable core insert functionality.
3. Validate via contract/integration tests (T006/T007) and manual dry-run/apply checks.

### Incremental Delivery
- Phase 4 adds automation/undo guarantees after MVP.
- Phase 5 hardens validation and conflict management.
- Phase 6 completes documentation, smoke coverage, and regression checks.

### Parallel Approach
- One developer handles foundational + US1 engine.
- Another focuses on test coverage and CLI wiring after foundations.
- Additional developer can own US2 automation tasks while US1 finalizes, then US3 validation enhancements.
