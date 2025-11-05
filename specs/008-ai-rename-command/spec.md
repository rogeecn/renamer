# Feature Specification: AI-Assisted Rename Command

**Feature Branch**: `008-ai-rename-command`  
**Created**: 2025-11-05  
**Status**: Draft  
**Input**: User description: "添加 ai 子命令，使用go genkit 调用ai能力对文件列表进行重命名。"

## Clarifications

### Session 2025-11-05

- Q: How should the CLI handle filename privacy when calling the AI service? → A: Send raw filenames without masking.
- Q: How should the AI provider credential be supplied to the CLI? → A: Read from an environment variable (e.g., `RENAMER_AI_KEY`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Request AI rename plan (Priority: P1)

As a command-line user, I can request AI-generated rename suggestions for a set of files so that I get a consistent naming plan without defining rules manually.

**Why this priority**: This delivers the core value of leveraging AI to save time on naming decisions for large batches.

**Independent Test**: Execute the AI rename command against a sample directory and verify a preview of suggested names is produced without altering files.

**Acceptance Scenarios**:

1. **Given** a directory with mixed file names and optional user instructions, **When** the user runs the AI rename command, **Then** the tool returns a preview mapping each original name to a suggested name.
2. **Given** a scope that includes hidden files when the user enables the corresponding flag, **When** the AI rename command runs, **Then** the preview reflects only the files allowed by the selected scope options.

---

### User Story 2 - Refine and confirm suggestions (Priority: P2)

As a command-line user, I can review, adjust, or regenerate AI suggestions before applying them so that I have control over the final names.

**Why this priority**: Users need confidence and agency to ensure AI suggestions match their intent, reducing the risk of undesired renames.

**Independent Test**: Run the AI rename command, adjust the instruction text, regenerate suggestions, and confirm the tool updates the preview without applying changes until approval.

**Acceptance Scenarios**:

1. **Given** an initial AI preview, **When** the user supplies new guidance or rejects the batch, **Then** the tool allows a regeneration or cancellation without renaming any files.
2. **Given** highlighted conflicts or invalid suggestions in the preview, **When** the user attempts to accept the batch, **Then** the tool blocks execution and instructs the user to resolve the issues.

---

### User Story 3 - Apply and audit AI renames (Priority: P3)

As a command-line user, I can apply approved AI rename suggestions and rely on the existing history and undo mechanisms so that AI-driven batches are traceable and reversible.

**Why this priority**: Preserving auditability and undo aligns AI-driven actions with existing safety guarantees.

**Independent Test**: Accept an AI rename batch, verify files are renamed, the ledger records the operation, and the undo command restores originals.

**Acceptance Scenarios**:

1. **Given** an approved AI rename preview, **When** the user confirms execution, **Then** the files are renamed and the batch details are recorded in the ledger with AI-specific metadata.
2. **Given** an executed AI rename batch, **When** the user runs the undo command, **Then** all affected files return to their original names and the ledger reflects the reversal.

---

### Edge Cases

- AI service fails, times out, or returns an empty response; the command must preserve current filenames and surface actionable error guidance.
- The AI proposes duplicate, conflicting, or filesystem-invalid names; the preview must flag each item and prevent application until resolved.
- The selected scope includes more files than the AI request limit; the command must communicate limits and guide the user to narrow the scope or batch the request.
- The ledger already contains pending batches for the same files; the tool must clarify how the new AI batch interacts with existing history before proceeding.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI must gather the current file scope using existing flags and present the selected files and optional instructions to the AI suggestion service.
- **FR-002**: The system must generate a human-readable preview that pairs each original filename with the AI-proposed name and indicates the rationale or confidence when available.
- **FR-003**: The CLI must block application when the preview contains conflicts, invalid names, or missing suggestions and must explain the required corrective actions.
- **FR-004**: Users must be able to modify guidance and request a new set of AI suggestions without leaving the command until they accept or exit.
- **FR-005**: When users approve a preview, the tool must execute the rename batch, record it in the ledger with the user guidance and AI attribution, and support undo via the existing command.
- **FR-006**: The command must support dry-run mode that exercises the AI interaction and preview without writing to disk, clearly labeling the output as non-destructive.
- **FR-007**: The system must handle AI service errors gracefully by retaining current filenames, logging diagnostic information, and providing retry instructions.

### Key Entities

- **AISuggestionBatch**: Captures the scope summary, user guidance, timestamp, AI provider metadata, and the list of rename suggestions evaluated during a session.
- **RenameSuggestion**: Represents a single proposed change with original name, suggested name, validation status, and optional rationale.
- **UserGuidance**: Stores free-form instructions supplied by the user, including any follow-up refinements applied within the session.

## Assumptions

- AI rename suggestions are generated within existing rate limits; large directories may require the user to split the work manually.
- Users running the AI command have network access and credentials required to reach the AI service.
- Existing ledger and undo mechanisms remain unchanged and can store additional metadata without format migrations.
- AI requests transmit the original filenames without masking; users must avoid including sensitive names when invoking the command.
- The CLI reads AI provider credentials from environment variables (default `RENAMER_AI_KEY`); no interactive credential prompts are provided.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95% of AI rename previews for up to 200 files complete in under 30 seconds from command invocation.
- **SC-002**: 90% of accepted AI rename batches complete without conflicts or manual post-fix adjustments reported by users.
- **SC-003**: 100% of AI-driven rename batches remain fully undoable via the existing undo command.
- **SC-004**: In post-launch surveys, at least 80% of participating users report that AI suggestions improved their rename workflow efficiency.
