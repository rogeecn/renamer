# Tasks: Remove Command with Sequential Multi-Pattern Support

**Input**: Design documents from `/specs/003-add-remove-command/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare shared fixtures and tooling used across all remove command user stories.

- [X] T001 Create remove command fixture directories (`basic/`, `conflicts/`, `empties/`) with placeholder files and README in `tests/fixtures/remove-samples/`.
- [X] T002 [P] Author baseline smoke script showing preview → apply → undo flow for remove command in `scripts/smoke-test-remove.sh`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core remove command structures required before any user story implementation.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 [P] Define `RemoveRequest` struct plus validation and scope adaptation helpers in `internal/remove/request.go`.
- [X] T004 [P] Implement argument parsing (trimming, minimum token count, parse result object) in `internal/remove/parser.go`.
- [X] T005 [P] Create remove summary types and aggregation helpers for counts/conflicts in `internal/remove/summary.go`.
- [X] T006 Build traversal adapter that reuses listing scope to enumerate candidates for removal in `internal/remove/traversal.go`.

**Checkpoint**: Foundation ready—user story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Remove Unwanted Tokens in One Pass (Priority: P1) 🎯 MVP

**Goal**: Deliver sequential substring removal with preview and apply, covering the core filename cleanup workflow.

**Independent Test**: Run `renamer remove " copy" " draft" --dry-run` against `tests/fixtures/remove-samples/basic`, verify preview ordering, then apply with `--yes` and confirm filesystem changes.

### Tests for User Story 1 ⚠️

- [ ] T007 [P] [US1] Add unit tests covering sequential token application and unchanged cases in `tests/unit/remove_engine_test.go`.
- [ ] T008 [P] [US1] Create contract test validating preview table output and dry-run messaging in `tests/contract/remove_command_preview_test.go`.
- [ ] T009 [P] [US1] Write integration test exercising preview → apply flow with multiple files in `tests/integration/remove_flow_test.go`.

### Implementation for User Story 1

- [ ] T010 [US1] Implement sequential removal engine producing planned operations in `internal/remove/engine.go`.
- [ ] T011 [US1] Build preview pipeline that aggregates summaries, detects conflicts, and streams output in `internal/remove/preview.go`.
- [ ] T012 [US1] Implement apply pipeline executing planned operations without ledger writes in `internal/remove/apply.go`.
- [ ] T013 [US1] Wire new Cobra command in `cmd/remove.go` (with registration in `cmd/root.go`) to drive preview/apply using shared scope flags.

**Checkpoint**: User Story 1 functional end-to-end with preview/apply validated by automated tests.

---

## Phase 4: User Story 2 - Script-Friendly Removal Workflow (Priority: P2)

**Goal**: Ensure automation can run `renamer remove` non-interactively with deterministic exit codes and ledger-backed undo.

**Independent Test**: Execute `renamer remove foo bar --dry-run` followed by `--yes` inside CI fixture, verify exit code 0 on success, ledger metadata persists tokens, and `renamer undo` restores originals.

### Tests for User Story 2 ⚠️

- [ ] T014 [P] [US2] Add contract test asserting ledger entries capture ordered tokens and match counts in `tests/contract/remove_command_ledger_test.go`.
- [ ] T015 [P] [US2] Add integration test covering `--yes` automation path and subsequent undo in `tests/integration/remove_undo_test.go`.

### Implementation for User Story 2

- [ ] T016 [US2] Extend apply pipeline to append ledger entries with ordered tokens and match counts in `internal/remove/apply.go`.
- [ ] T017 [US2] Update `cmd/remove.go` to support non-interactive `--yes` execution, emit automation-oriented messages, and propagate exit codes.

**Checkpoint**: User Story 2 complete—CLI safe for scripting with ledger + undo parity.

---

## Phase 5: User Story 3 - Validate Sequential Removal Inputs (Priority: P3)

**Goal**: Provide clear validation and warnings for duplicate tokens, empty results, and risky removals.

**Independent Test**: Run `renamer remove "Project X" " Project" "X" --dry-run` and confirm duplicate dedupe warnings plus empty-result skips appear in preview output.

### Tests for User Story 3 ⚠️

- [ ] T018 [P] [US3] Add parser validation tests for duplicate tokens and whitespace edge cases in `tests/unit/remove_parser_test.go`.
- [ ] T019 [P] [US3] Add integration test verifying empty-basename warnings and skips in `tests/integration/remove_validation_test.go`.

### Implementation for User Story 3

- [ ] T020 [US3] Implement duplicate token deduplication with ordered warning collection in `internal/remove/parser.go`.
- [ ] T021 [US3] Add empty-basename detection and warning tracking in `internal/remove/summary.go`.
- [ ] T022 [US3] Surface duplicate and empty warnings in CLI output handling within `cmd/remove.go`.

**Checkpoint**: All user stories deliver value; validations prevent risky rename plans.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, tooling, and quality improvements spanning all user stories.

- [ ] T023 [P] Update remove command documentation and sequential behavior guidance in `docs/cli-flags.md`.
- [ ] T024 Record release notes for remove command launch in `docs/CHANGELOG.md`.
- [ ] T025 [P] Finalize `scripts/smoke-test-remove.sh` with assertions and integrate into CI instructions.
- [ ] T026 Add remove command walkthrough to project onboarding materials in `AGENTS.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies—start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion—BLOCKS all user stories.
- **User Stories (Phase 3–5)**: Each depends on Foundational phase; implement in priority order (P1 → P2 → P3) or in parallel once shared blockers clear.
- **Polish (Final Phase)**: Depends on completion of targeted user stories.

### User Story Dependencies

- **US1 (P1)**: Requires Foundational tasks (T003–T006).
- **US2 (P2)**: Requires US1 core command and apply pipeline (T010–T013).
- **US3 (P3)**: Requires parser and summary scaffolding plus US1 preview pipeline (T004–T013).

### Within Each User Story

- Tests (if included) MUST be authored before implementation tasks.
- Engine/traversal logic precedes CLI wiring for predictable integration.
- Command wiring completes only after engine/preview/apply logic is ready.

### Parallel Opportunities

- Setup tasks (T001–T002) can run in parallel.
- Foundational tasks marked [P] (T003–T005) may proceed concurrently after directory scaffolding.
- US1 test tasks (T007–T009) can run in parallel once fixtures exist.
- US2 and US3 test tasks (T014–T019) can execute concurrently after their respective foundations.
- Polish tasks marked [P] (T023, T025) can occur alongside documentation updates.

---

## Parallel Example: User Story 1

```bash
# Parallel test development for US1:
#   - T007: tests/unit/remove_engine_test.go
#   - T008: tests/contract/remove_command_preview_test.go
#   - T009: tests/integration/remove_flow_test.go
#
# Once tests are in place, run them together:
go test ./tests/unit ./tests/contract ./tests/integration -run Remove
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Setup) and Phase 2 (Foundational).
2. Finish Phase 3 (US1) delivering preview/apply with automated coverage.
3. Validate with `go test ./...` and smoke script before moving on.

### Incremental Delivery

1. Deliver US1 (core removal) → release MVP.
2. Add US2 (automation + ledger) → publish update.
3. Enhance with US3 (advanced validation) → finalize release notes.

### Parallel Team Strategy

- After Phase 2, one developer tackles US1 implementation while another starts US2 tests.
- US3 validation enhancements can begin once parser scaffolding lands, overlapping with documentation polish.
- Conclude with Polish phase tasks to align docs, smoke tests, and onboarding materials.
