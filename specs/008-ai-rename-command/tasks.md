# Tasks: AI-Assisted Rename Command

**Input**: Design documents from `/specs/008-ai-rename-command/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add Go-based Genkit dependency and scaffold AI flow package.

- [x] T001 Ensure Genkit Go module dependency (`github.com/firebase/genkit/go`) is present in `go.mod` / `go.sum`
- [x] T002 Create AI flow package directories in `internal/ai/flow/`
- [x] T003 [P] Add Go test harness scaffold for AI flow in `internal/ai/flow/flow_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Provide shared assets and configuration required by all user stories.

- [x] T004 Author AI rename prompt template with JSON instructions in `internal/ai/flow/prompt.tmpl`
- [x] T005 Implement reusable JSON parsing helpers for Genkit responses in `internal/ai/flow/json.go`
- [x] T006 Implement AI credential loader reading `RENAMER_AI_KEY` in `internal/ai/config.go`
- [x] T007 Register the `ai` Cobra command scaffold in `cmd/root.go`

---

## Phase 3: User Story 1 - Request AI rename plan (Priority: P1) 🎯 MVP

**Goal**: Allow users to preview AI-generated rename suggestions for a scoped set of files.

**Independent Test**: Run `renamer ai --path <dir> --dry-run` and verify the preview table lists original → suggested names without renaming files.

### Tests for User Story 1

- [x] T008 [P] [US1] Add prompt rendering unit test covering file list formatting in `internal/ai/flow/prompt_test.go`
- [x] T009 [P] [US1] Create CLI preview contract test enforcing JSON schema in `tests/contract/ai_command_preview_test.go`

### Implementation for User Story 1

- [x] T010 [US1] Implement `renameFlow` Genkit workflow with JSON-only response in `internal/ai/flow/rename_flow.go`
- [x] T011 [P] [US1] Build Genkit client wrapper and response parser in `internal/ai/client.go`
- [x] T012 [P] [US1] Implement suggestion validation rules (extensions, duplicates, illegal chars) in `internal/ai/validation.go`
- [x] T013 [US1] Map AI suggestions to preview rows with rationale fields in `internal/ai/preview.go`
- [x] T014 [US1] Wire `renamer ai` command to gather scope, invoke AI flow, and render preview in `cmd/ai.go`
- [x] T015 [US1] Document preview usage and flags for `renamer ai` in `docs/cli-flags.md`

---

## Phase 4: User Story 2 - Refine and confirm suggestions (Priority: P2)

**Goal**: Let users iterate on AI guidance, regenerate suggestions, and resolve conflicts before applying changes.

**Independent Test**: Run `renamer ai` twice with updated prompts, confirm regenerated preview replaces the previous batch, and verify conflicting targets block approval with actionable warnings.

### Tests for User Story 2

- [x] T016 [P] [US2] Add integration test for preview regeneration and cancel flow in `tests/integration/ai_preview_regen_test.go`

### Implementation for User Story 2

- [x] T017 [US2] Extend interactive loop in `cmd/ai.go` to support prompt refinement and regeneration commands
- [x] T018 [P] [US2] Enhance conflict and warning annotations for regenerated suggestions in `internal/ai/validation.go`
- [x] T019 [US2] Persist per-session prompt history and guidance notes in `internal/ai/session.go`

---

## Phase 5: User Story 3 - Apply and audit AI renames (Priority: P3)

**Goal**: Execute approved AI rename batches, record them in the ledger, and ensure undo restores originals.

**Independent Test**: Accept an AI preview with `--yes`, verify files are renamed, ledger entry captures AI metadata, and `renamer undo` restores originals.

### Tests for User Story 3

- [X] T020 [P] [US3] Add integration test covering apply + undo lifecycle in `tests/integration/ai_flow_apply_test.go`
- [X] T021 [P] [US3] Add ledger contract test verifying AI metadata persistence in `tests/contract/ai_ledger_entry_test.go`

### Implementation for User Story 3

- [X] T022 [US3] Implement confirm/apply execution path with `--yes` handling in `cmd/ai.go`
- [X] T023 [P] [US3] Append AI batch metadata to ledger entries in `internal/history/ai_entry.go`
- [X] T024 [US3] Ensure undo replay reads AI ledger metadata in `internal/history/undo.go`
- [X] T025 [US3] Display progress and per-file outcomes during apply in `internal/output/progress.go`

---

## Phase 6: Polish & Cross-Cutting

**Purpose**: Final quality improvements, docs, and operational readiness.

- [ ] T026 Add smoke test script invoking `renamer ai` preview/apply flows in `scripts/smoke-test-ai.sh`
- [X] T027 Update top-level documentation with AI command overview and credential requirements in `README.md`

---

## Dependencies

- Complete Phases 1 → 2 before starting user stories.
- User Story order: US1 → US2 → US3 (each builds on prior capabilities).
- Polish tasks run after all user stories are feature-complete.

## Parallel Execution Opportunities

- US1: T011 and T012 can run in parallel after T010 completes.
- US2: T018 can run in parallel with T017 once session loop scaffolding exists.
- US3: T023 and T025 can proceed concurrently after T022 defines apply workflow.

## Implementation Strategy

1. Deliver User Story 1 as the MVP (preview-only experience).
2. Iterate on refinement workflow (User Story 2) to reduce risk of bad suggestions before apply.
3. Add apply + ledger integration (User Story 3) to complete end-to-end flow.
4. Finish with polish tasks to solidify operational readiness.
