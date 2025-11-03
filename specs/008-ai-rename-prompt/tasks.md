# Tasks: AI-Assisted Rename Prompting

**Input**: Design documents from `/specs/008-ai-rename-prompt/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish tooling and scaffolding for the embedded Go Genkit workflow.

- [x] T001 Pin Google Genkit Go SDK dependency and run tidy in `go.mod`
- [x] T002 Scaffold `cmd/ai.go` command file with Cobra boilerplate
- [x] T003 Create `internal/ai` package directories (`prompt`, `genkit`, `plan`) in the repository
- [x] T004 Document model token location in `docs/cli-flags.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core plumbing that all user stories rely on.

- [x] T005 Implement env loader for `$HOME/.config/.renamer` in `internal/ai/config/token_store.go` with package `github.com/joho/godotenv`
- [x] T006 Define shared prompt/response structs per spec in `internal/ai/prompt/types.go`
- [x] T007 Implement Go Genkit workflow skeleton with default OpenAI-compatible model in `internal/ai/genkit/workflow.go`
- [x] T008 Build response validator ensuring coverage/uniqueness in `internal/ai/plan/validator.go`
- [x] T009 Add CLI flag parsing for AI options (model override, debug export) in `cmd/ai.go`
- [x] T010 Wire ledger metadata schema for AI batches in `internal/history/history.go`

**Checkpoint**: Genkit workflow callable from CLI with validation scaffolding ready.

---

## Phase 3: User Story 1 - Generate AI Rename Plan (Priority: P1) 🎯 MVP

**Goal**: Produce a previewable AI-generated rename plan with sequential, sanitized filenames.

**Independent Test**: `go run . ai --path <fixtures>` prints numbered preview (`001_...`) without spam terms and logs prompt hash.

### Tests for User Story 1

- [x] T011 [P] [US1] Add contract test covering Genkit prompt/response schema in `tests/contract/ai_prompt_contract_test.go`
- [x] T012 [P] [US1] Add integration test for preview output on sample batch in `tests/integration/ai_preview_flow_test.go`

### Implementation for User Story 1

- [x] T013 [US1] Assemble prompt builder using traversal samples in `internal/ai/prompt/builder.go`
- [x] T014 [US1] Execute Genkit workflow and capture telemetry in `internal/ai/genkit/client.go`
- [x] T015 [US1] Map Genkit response to preview plan entries in `internal/ai/plan/mapper.go`
- [x] T016 [US1] Render preview table with sequence + sanitization notes in `internal/output/table.go`
- [x] T017 [US1] Log prompt hash and response warnings to debug output in `internal/output/plain.go`

**Checkpoint**: `renamer ai --dry-run` fully functional for default policies.

---

## Phase 4: User Story 2 - Enforce Naming Standards (Priority: P2)

**Goal**: Allow operators to specify naming policies that the AI prompt and validator enforce.

**Independent Test**: `renamer ai --naming-casing kebab --prefix proj --dry-run` produces kebab-case names with `proj` prefix and fails invalid responses.

### Tests for User Story 2

- [x] T018 [P] [US2] Contract test ensuring casing/prefix rules reach Genkit input in `tests/contract/ai_policy_contract_test.go`
- [x] T019 [P] [US2] Integration test covering policy violations in `tests/integration/ai_policy_validation_test.go`

### Implementation for User Story 2

- [x] T020 [US2] Extend CLI flags/environment parsing for naming policies in `cmd/ai.go`
- [x] T021 [US2] Inject policy directives into prompt payload in `internal/ai/prompt/builder.go`
- [x] T022 [US2] Enhance validator to enforce casing/prefix/banned tokens in `internal/ai/plan/validator.go`
- [x] T023 [US2] Surface policy failures with actionable messages in `internal/output/plain.go`

**Checkpoint**: Policy-driven prompts and enforcement operational.

---

## Phase 5: User Story 3 - Review, Edit, and Apply Safely (Priority: P3)

**Goal**: Support manual edits, conflict resolution, and safe apply/undo flows.

**Independent Test**: Modify exported plan, revalidate, then `renamer ai --yes` applies changes and ledger records AI metadata; undo restores originals.

### Tests for User Story 3

- [x] T024 [P] [US3] Integration test covering manual edits + apply/undo in `tests/integration/ai_apply_undo_test.go`
- [x] T025 [P] [US3] Contract test ensuring ledger metadata captures prompt/response hashes in `tests/contract/ai_ledger_contract_test.go`

### Implementation for User Story 3

- [x] T026 [US3] Implement plan editing/export/import helpers in `internal/ai/plan/editor.go`
- [x] T027 [US3] Revalidation workflow for edited plans in `internal/ai/plan/validator.go`
- [x] T028 [US3] Conflict detection (duplicate targets, missing sequences) surfaced in preview in `internal/ai/plan/conflicts.go`
- [x] T029 [US3] Apply pipeline recording AI metadata to ledger in `internal/ai/plan/apply.go`
- [x] T030 [US3] Update undo path to respect AI metadata in `cmd/undo.go`

**Checkpoint**: Full review/edit/apply loop complete with undo safety.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T031 [P] Add CLI help and usage examples for `renamer ai` in `cmd/root.go`
- [x] T032 [P] Update end-user documentation in `docs/cli-flags.md`
- [x] T033 [P] Add smoke script exercising AI flow in `scripts/smoke-test-ai.sh`
- [x] T034 [P] Record prompt/response telemetry opt-in in `docs/CHANGELOG.md`

---

## Dependencies

1. Setup → Foundational → US1 → US2 → US3 → Polish
2. User story dependencies: US2 depends on US1; US3 depends on US1 and US2.

## Parallel Execution Examples

- During US1, prompt builder (T013) and Genkit client (T014) can proceed in parallel after foundational tasks.
- US2 policy contract test (T018) can run alongside validator enhancements (T022) once prompt builder updates (T021) start.
- In US3, ledger integration (T029) can progress concurrently with conflict detection (T028).
- Polish tasks (T031–T034) may run in parallel after US3 completes.

## Implementation Strategy

1. Deliver MVP by completing Phases 1–3 (US1) to provide AI-generated preview with validation.
2. Layer policy enforcement (US2) to align output with organizational naming standards.
3. Finish with editing/apply safety (US3) and polish tasks before release.
