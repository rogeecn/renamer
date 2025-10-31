# Tasks: Regex Command for Pattern-Based Renaming

**Input**: Design documents from `/specs/006-add-regex-command/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Include targeted contract and integration coverage where scenarios demand automated verification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare fixtures and support tooling required across all stories.

- [X] T001 Create regex test fixtures (`tests/fixtures/regex/`) with sample filenames covering digits, words, and Unicode cases.
- [X] T002 [P] Scaffold `scripts/smoke-test-regex.sh` mirroring quickstart scenarios for preview/apply/undo.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish reusable package skeletons and command registration that all stories build upon.

- [X] T003 Create `internal/regex` package scaffolding (request.go, summary.go, doc.go) matching data-model entities.
- [X] T004 [P] Register a stub `regex` Cobra command in `cmd/regex.go` with flag definitions aligned to shared scope options.

**Checkpoint**: Foundation ready – user story implementation can now begin.

---

## Phase 3: User Story 1 - Rename Files Using Captured Groups (Priority: P1) 🎯 MVP

**Goal**: Allow users to preview regex-based renames that substitute captured groups into templates while preserving extensions.

**Independent Test**: Run `renamer regex "^(\w+)-(\d+)" "@2_@1" --dry-run` against fixtures and verify preview outputs `123_alpha.log`, `456_beta.log` without modifying the filesystem.

### Tests for User Story 1

- [X] T005 [P] [US1] Add preview contract test for capture groups in `tests/contract/regex_command_test.go`.
- [X] T006 [P] [US1] Add integration preview flow test covering dry-run confirmation in `tests/integration/regex_flow_test.go`.

### Implementation for User Story 1

- [X] T007 [P] [US1] Implement template parser handling `@n` and `@@` tokens in `internal/regex/template.go`.
- [X] T008 [P] [US1] Implement regex engine applying capture groups to candidate names in `internal/regex/engine.go`.
- [X] T009 [US1] Build preview planner producing `RegexSummary` entries in `internal/regex/preview.go`.
- [X] T010 [US1] Wire Cobra command to preview/apply planner with scope options in `cmd/regex.go`.

**Checkpoint**: User Story 1 preview capability ready for validation.

---

## Phase 4: User Story 2 - Automation-Friendly Regex Renames (Priority: P2)

**Goal**: Deliver deterministic apply flows with ledger metadata and undo support suitable for CI automation.

**Independent Test**: Execute `renamer regex "^build_(\d+)_(.*)$" "release-@1-@2" --yes --path ./tests/fixtures/regex` and verify exit code `0`, ledger metadata, and successful `renamer undo` restoration.

### Tests for User Story 2

- [X] T011 [P] [US2] Add ledger contract test capturing pattern/template metadata in `tests/contract/regex_ledger_test.go`.
- [X] T012 [P] [US2] Add integration undo flow test for regex entries in `tests/integration/regex_undo_test.go`.

### Implementation for User Story 2

- [X] T013 [P] [US2] Implement apply handler persisting ledger entries in `internal/regex/apply.go`.
- [X] T014 [US2] Ensure `cmd/regex.go` honors `--yes` automation semantics and deterministic exit codes.
- [X] T015 [US2] Extend undo recognition for regex batches in `internal/history/history.go` and shared output messaging.

**Checkpoint**: Automation-focused workflows (apply + undo) validated.

---

## Phase 5: User Story 3 - Validate Patterns, Placeholders, and Conflicts (Priority: P3)

**Goal**: Provide clear feedback for invalid patterns or template conflicts to prevent destructive applies.

**Independent Test**: Run `renamer regex "^(.*)$" "@2" --dry-run` and confirm an error about undefined capture groups; attempt a rename producing duplicate targets and confirm apply is blocked.

### Tests for User Story 3

- [X] T016 [P] [US3] Add validation contract tests for invalid patterns/placeholders in `tests/contract/regex_validation_test.go`.
- [X] T017 [P] [US3] Add integration conflict test ensuring duplicate targets block apply in `tests/integration/regex_conflict_test.go`.

### Implementation for User Story 3

- [X] T018 [P] [US3] Implement validation for undefined groups and empty results in `internal/regex/validate.go`.
- [X] T019 [US3] Extend conflict detection to flag duplicate or empty proposals in `internal/regex/preview.go`.
- [X] T020 [US3] Enhance CLI error messaging and help examples in `cmd/regex.go`.

**Checkpoint**: Validation safeguards complete; regex command safe for experimentation.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final documentation, tooling, and quality passes.

- [X] T021 [P] Update CLI documentation with regex command details in `docs/cli-flags.md`.
- [X] T022 [P] Finalize `scripts/smoke-test-regex.sh` to exercise quickstart scenarios and ledger undo.
- [X] T023 Run `gofmt` and `go test ./...` to verify formatting and regression coverage.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)** → prerequisite for foundational work.
- **Foundational (Phase 2)** → must complete before User Stories begin.
- **User Stories (Phase 3–5)** → execute sequentially by priority or in parallel once dependencies satisfied.
- **Polish (Phase 6)** → runs after desired user stories ship.

### User Story Dependencies

- **US1** depends on Foundational package scaffolding (T003–T004).
- **US2** depends on US1 preview/apply wiring.
- **US3** depends on US1 preview engine and US2 apply infrastructure to validate.

### Parallel Opportunities

- Tasks marked `[P]` operate on distinct files and can proceed concurrently once their prerequisites are met.
- Different user stories can progress in parallel after their dependencies complete, provided shared files (`cmd/regex.go`, `internal/regex/preview.go`) are coordinated sequentially.

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1–2 to establish scaffolding.
2. Implement US1 preview workflow (T005–T010) and validate independently.
3. Ship preview-only capability if automation support can follow later.

### Incremental Delivery

1. Deliver US1 preview/apply basics.
2. Layer US2 automation + ledger features.
3. Add US3 validation/conflict safeguards.
4. Conclude with polish tasks for docs, smoke script, and regression suite.

### Parallel Team Strategy

- Developer A focuses on template/engine internals (T007–T008) while Developer B builds tests (T005–T006).
- After US1, split automation work: ledger implementation (T013) and undo validation tests (T012) run concurrently.
- Validation tasks (T016–T020) can be parallelized between CLI messaging and conflict handling once US2 merges.

---

## Notes

- Keep task granularity small enough for independent completion while documenting file paths for each change.
- Tests should fail before implementation to confirm coverage.
- Mark tasks complete (`[X]`) in this document as work progresses.
