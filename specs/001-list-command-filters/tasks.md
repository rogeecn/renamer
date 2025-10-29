# Tasks: Cobra List Command with Global Filters

**Input**: Design documents from `/specs/001-list-command-filters/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are optional; include them only where they support the user story’s independent validation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare Cobra project for new subcommand and tests

- [X] T001 Normalize root command metadata and remove placeholder toggle flag in `cmd/root.go`
- [X] T002 Scaffold listing service package with stub struct in `internal/listing/service.go`
- [X] T003 Add fixture guidance for sample directory trees in `tests/fixtures/README.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core utilities shared across list, preview, and rename flows

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Define `ListingRequest`/`ListingEntry` structs with validators in `internal/listing/types.go`
- [X] T005 [P] Implement extension filter parser/normalizer in `internal/filters/extensions.go`
- [X] T006 [P] Implement streaming traversal walker with symlink guard in `internal/traversal/walker.go`
- [X] T007 [P] Declare formatter interface and summary helpers in `internal/output/formatter.go`
- [X] T008 Document global filter contract expectations in `docs/cli-flags.md`

**Checkpoint**: Foundation ready—user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Discover Filtered Files Before Renaming (Priority: P1) 🎯 MVP

**Goal**: Provide a read-only `renamer list` command that mirrors rename scope

**Independent Test**: Run `renamer list -e .jpg|.png` against fixtures and verify results/summary without filesystem changes

### Tests for User Story 1 (OPTIONAL - included for confidence)

- [X] T009 [P] [US1] Add contract test for filtered listing summary in `tests/contract/list_command_test.go`
- [X] T010 [P] [US1] Add integration test covering recursive listing in `tests/integration/list_recursive_test.go`

### Implementation for User Story 1

- [X] T011 [US1] Implement listing pipeline combining traversal, filters, and summary in `internal/listing/service.go`
- [X] T012 [US1] Implement zero-result messaging helper in `internal/listing/summary.go`
- [X] T013 [US1] Add Cobra `list` command entry point in `cmd/list.go`
- [X] T014 [US1] Register `list` command and write help text in `cmd/root.go`
- [X] T015 [US1] Update quickstart usage section for `renamer list` workflow in `specs/001-list-command-filters/quickstart.md`

**Checkpoint**: User Story 1 delivers a safe, filter-aware listing command

---

## Phase 4: User Story 2 - Apply Global Filters Consistently (Priority: P2)

**Goal**: Ensure scope flags (`-r`, `-d`, `-e`) live on the root command and hydrate shared request builders

**Independent Test**: Execute `renamer list` twice—once via command parsing, once via helper—and confirm identical candidate counts

### Implementation for User Story 2

- [X] T016 [P] [US2] Promote scope flags to persistent flags on the root command in `cmd/root.go`
- [X] T017 [US2] Create shared flag extraction helper returning `ListingRequest` in `internal/listing/options.go`
- [X] T018 [US2] Refactor `cmd/list.go` to consume shared helper and root-level flags
- [X] T019 [P] [US2] Add integration test validating flag parity in `tests/integration/global_flag_parity_test.go`
- [X] T020 [US2] Expand CLI flag documentation for global usage patterns in `docs/cli-flags.md`

**Checkpoint**: User Story 2 guarantees consistent scope interpretation across commands

---

## Phase 5: User Story 3 - Review Listing Output Comfortably (Priority: P3)

**Goal**: Offer table and plain output modes for human review and scripting

**Independent Test**: Run `renamer list --format table` vs `--format plain` and ensure entries match across formats

### Implementation for User Story 3

- [X] T021 [P] [US3] Implement table renderer using `text/tabwriter` in `internal/output/table.go`
- [X] T022 [P] [US3] Implement plain renderer emitting newline-delimited paths in `internal/output/plain.go`
- [X] T023 [US3] Wire format selection into listing service dispatcher in `internal/listing/service.go`
- [X] T024 [US3] Extend contract tests to verify format output parity in `tests/contract/list_command_test.go`
- [X] T025 [US3] Update quickstart to demonstrate `--format` options in `specs/001-list-command-filters/quickstart.md`

**Checkpoint**: User Story 3 delivers ergonomic output formats for all audiences

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, documentation, and release notes

- [X] T026 Update agent guidance with list command summary in `AGENTS.md`
- [X] T027 Add release note entry describing new list command in `docs/CHANGELOG.md`
- [X] T028 Create smoke-test script exercising list + preview parity in `scripts/smoke-test-list.sh`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Complete before foundational utilities to ensure project scaffolding exists.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories because shared utilities power every command.
- **User Story 1 (Phase 3)**: Depends on Foundational; delivers MVP list command.
- **User Story 2 (Phase 4)**: Depends on User Story 1 (reuses command structure) and Foundational.
- **User Story 3 (Phase 5)**: Depends on User Story 1 (listing pipeline) and Foundational (formatter interface).
- **Polish (Final Phase)**: Runs after desired user stories finish.

### User Story Dependencies

- **User Story 1 (P1)**: Requires Foundational utilities; no other story prerequisites.
- **User Story 2 (P2)**: Requires User Story 1 to expose list command behaviors before sharing flags.
- **User Story 3 (P3)**: Requires User Story 1 to deliver base listing plus Foundational formatter interface.

### Within Each User Story

- Tests marked [P] should be authored before or alongside implementation; ensure they fail prior to feature work.
- Service implementations depend on validated request structs and traversal utilities.
- CLI command wiring depends on service implementation and helper functions.
- Documentation tasks finalize once command behavior is stable.

### Parallel Opportunities

- Foundational tasks T005–T007 operate on different packages and can proceed in parallel after T004 validates data structures.
- User Story 1 tests (T009, T010) can be developed in parallel before implementation tasks T011–T014.
- User Story 3 renderers (T021, T022) can be implemented concurrently, converging at T023 for integration.

---

## Parallel Example: User Story 1

```bash
# In one terminal: author contract test
go test ./tests/contract -run TestListCommandFilters

# In another terminal: implement listing service
go test ./internal/listing -run TestService
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Setup + Foundational phases.
2. Implement User Story 1 tasks (T009–T015) to deliver a working `renamer list`.
3. Validate with contract/integration tests and quickstart instructions.

### Incremental Delivery

1. Deliver MVP (User Story 1).
2. Add User Story 2 to guarantee global flag consistency.
3. Add User Story 3 to enhance output ergonomics.
4. Finalize polish tasks for documentation and smoke testing.

### Parallel Team Strategy

- Developer A handles Foundational tasks (T004–T007) while Developer B updates documentation (T008).
- After Foundational checkpoint, Developer A implements listing service (T011–T014), Developer B authors tests (T009–T010).
- Once MVP ships, Developer A tackles global flag refactor (T016–T018) while Developer B extends integration tests (T019) and docs (T020).
- For User Story 3, split renderer work (T021 vs T022) before merging at T023.
