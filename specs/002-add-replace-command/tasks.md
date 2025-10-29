# Tasks: Replace Command with Multi-Pattern Support

**Input**: Design documents from `/specs/002-add-replace-command/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are optional; include them only where they support the user story’s independent validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare project for replace command implementation

- [X] T001 Audit root command for shared flags; document expected additions in `cmd/root.go`
- [X] T002 Create replace package scaffold in `internal/replace/README.md` with intended module layout
- [X] T003 Add sample fixtures directory for replacement tests at `tests/fixtures/replace-samples/README.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core utilities required by every user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Implement `ReplaceRequest` struct and validators in `internal/replace/request.go`
- [X] T005 [P] Implement pattern parser handling quoting/deduplication in `internal/replace/parser.go`
- [X] T006 [P] Extend traversal utilities to emit replace candidates in `internal/replace/traversal.go`
- [X] T007 [P] Implement replacement engine with overlap handling in `internal/replace/engine.go`
- [X] T008 Define summary metrics and conflict structs in `internal/replace/summary.go`
- [X] T009 Document replace syntax and flags draft in `docs/cli-flags.md`

**Checkpoint**: Foundation ready—user story implementation can begin

---

## Phase 3: User Story 1 - Normalize Naming with One Command (Priority: P1) 🎯 MVP

**Goal**: Deliver multi-pattern replacement with preview + apply guarantees

**Independent Test**: `renamer replace foo bar Baz --dry-run` followed by `--yes`; verify preview shows replacements, apply updates files, and `renamer undo` restores originals.

### Tests for User Story 1 (OPTIONAL - included for confidence)

- [X] T010 [P] [US1] Contract test for preview summary counts in `tests/contract/replace_command_test.go`
- [X] T011 [P] [US1] Integration test covering multi-pattern apply + undo in `tests/integration/replace_flow_test.go`

### Implementation for User Story 1

- [X] T012 [US1] Implement CLI command wiring in `cmd/replace.go` using shared scope flags
- [X] T013 [US1] Implement preview rendering and summary output in `internal/replace/preview.go`
- [X] T014 [US1] Hook replace engine into ledger/undo pipeline in `internal/replace/apply.go`
- [X] T015 [US1] Add documentation examples to `docs/cli-flags.md` and quickstart

**Checkpoint**: User Story 1 delivers functional `renamer replace` with preview/apply/undo

---

## Phase 4: User Story 2 - Script-Friendly Replacement Workflows (Priority: P2)

**Goal**: Ensure automation-friendly behaviors (exit codes, non-interactive usage)

**Independent Test**: Scripted run invoking `renamer replace ... --dry-run` then `--yes`; expect consistent exit codes and ledger entry

### Implementation for User Story 2

- [X] T016 [US2] Implement non-interactive flag validation (`--yes` + positional args) in `cmd/replace.go`
- [X] T017 [US2] Add ledger metadata (pattern counts) for automation in `internal/replace/summary.go`
- [X] T018 [US2] Expand integration test to assert exit codes for invalid input in `tests/integration/replace_flow_test.go`
- [X] T019 [US2] Update quickstart section with automation guidance in `specs/002-add-replace-command/quickstart.md`

**Checkpoint**: Scripts can rely on deterministic exit codes and ledger data

---

## Phase 5: User Story 3 - Validate Complex Pattern Input (Priority: P3)

**Goal**: Provide resilient handling for whitespace/special-character patterns and user guidance

**Independent Test**: `renamer replace "Project X" "Project-X" ProjectX --dry-run` plus invalid quoting to verify errors

### Implementation for User Story 3

- [X] T020 [P] [US3] Implement quoting guidance and warnings in `cmd/replace.go`
- [X] T021 [P] [US3] Add parser coverage for whitespace patterns in `tests/unit/replace_parser_test.go`
- [X] T022 [US3] Surface duplicate pattern warnings in preview summary in `internal/replace/preview.go`
- [X] T023 [US3] Document advanced pattern examples in `docs/cli-flags.md`

**Checkpoint**: Power users receive clear guidance and validation feedback

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and release readiness

- [X] T024 Update agent guidance with replace command details in `AGENTS.md`
- [X] T025 Add changelog entry describing new replace command in `docs/CHANGELOG.md`
- [X] T026 Create smoke test script covering replace + undo in `scripts/smoke-test-replace.sh`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No prerequisites
- **Foundational (Phase 2)**: Depends on setup completion; blocks all user stories
- **User Story 1 (Phase 3)**: Depends on foundational phase; MVP delivery
- **User Story 2 (Phase 4)**: Depends on User Story 1 (automation builds on base command)
- **User Story 3 (Phase 5)**: Depends on User Story 1 (parser/extensions) and shares foundation
- **Polish (Phase N)**: Runs after user stories complete

### User Story Dependencies

- **US1**: Requires foundational tasks
- **US2**: Requires US1 implementation + ledger integration
- **US3**: Requires US1 parser and preview infrastructure

### Within Each User Story

- Tests (if included) should be authored before implementation tasks
- Parser/engine updates precede CLI wiring
- Documentation updates finalize after behavior stabilizes

### Parallel Opportunities

- Foundational parser/engine/summary tasks (T005–T008) can progress in parallel after T004
- US1 tests (T010–T011) can run alongside command wiring (T012–T014)
- US3 parser coverage (T021) can proceed independently while warnings (T022) integrate with preview

---

## Parallel Example: User Story 1

```bash
# Terminal 1: Write contract test and run in watch mode
go test ./tests/contract -run TestReplaceCommandPreview

# Terminal 2: Implement preview renderer
$EDITOR internal/replace/preview.go
```

---

## Implementation Strategy

### MVP First (User Story 1)
1. Complete Setup and Foundational phases
2. Implement US1 tasks (T010–T015)
3. Ensure preview/apply/undo works end-to-end

### Incremental Delivery
1. Ship US1 (core replace command)
2. Layer US2 (automation-friendly exit codes and ledger metadata)
3. Add US3 (advanced pattern validation)
4. Execute polish tasks for documentation and smoke tests

### Parallel Team Strategy
- Engineer A: Parser/engine/summary foundational work
- Engineer B: CLI command wiring + tests (US1)
- Engineer C: Automation behaviors and documentation (US2/Polish)
- After US1, shift Engineers to handle US3 enhancements and polish concurrently
