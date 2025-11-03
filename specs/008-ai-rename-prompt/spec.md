# Feature Specification: AI-Assisted Rename Prompting

**Feature Branch**: `008-ai-rename-prompt`  
**Created**: 2025-11-03  
**Status**: Draft  
**Input**: User description: "实现使用AI重命名的功能，把当前的文件列表给AI使用AI进行重新命令，AI返回重命令规则后解析AI的规则进行本地重命名操作。你需要帮我考虑如何建立合适的prompt给AI实现重命名。要求：1、带序列号；2、文件名规则统一；3、文件名中去除广告推广等垃圾信息；4、如果还有其它合适规则你来适量添加。"

## Clarifications

### Session 2025-11-03

- Q: Which AI model should the CLI use by default? → A: Default to an OpenAI-compatible model with override flag/env.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate AI Rename Plan (Priority: P1)

As a file curator preparing bulk renames, I want the CLI to compile a clean prompt from my current file list, send it to the AI service, and receive a structured rename plan that I can preview with sequence numbers before applying changes.

**Why this priority**: This delivers the core value—automating consistent, sequential renames without manual rule crafting—unlocking the feature for everyday workflows.

**Independent Test**: With a mixed set of filenames, invoke the new `renamer ai` flow, confirm the prompt includes the sampled names and instructions, and verify the returned plan previews sequential, uniform, sanitized filenames ready for apply.

**Acceptance Scenarios**:

1. **Given** a directory of 25 assorted media files, **When** the user runs the AI rename preview, **Then** the prompt sent to the AI lists representative samples, prescribes ordering, and the returned plan previews filenames numbered `001_...` to `025_...` without junk text.
2. **Given** an AI response that follows the documented schema, **When** the CLI parses it, **Then** each planned rename appears in the standard preview table with sequence numbers and consistent formatting.

---

### User Story 2 - Enforce Naming Standards (Priority: P2)

As a brand manager, I want to configure naming guidelines (e.g., project label, casing style) that the AI prompt reinforces so that resulting filenames stay uniform across batches.

**Why this priority**: Allowing users to express naming policy increases trust and ensures AI output aligns with organizational standards.

**Independent Test**: Run the command with options specifying kebab-case and a prefix token, then confirm the generated prompt includes those rules and the AI response reflects them in the preview.

**Acceptance Scenarios**:

1. **Given** a user-provided naming policy, **When** the AI prompt is generated, **Then** it explicitly lists casing, prefix, and separator requirements.
2. **Given** the AI response violates the declared casing rule, **When** the CLI validates the response, **Then** the run aborts with a descriptive error explaining which filenames failed the policy check.

---

### User Story 3 - Review, Edit, and Apply Safely (Priority: P3)

As a cautious operator, I want to review the AI plan, make manual adjustments if needed, and only apply changes once I am satisfied that collisions, restricted terms, and missing numbers are resolved.

**Why this priority**: Safety controls maintain confidence in AI-driven workflows and reduce post-apply cleanup.

**Independent Test**: After generating an AI plan, edit a subset of proposed names, re-run validation, and ensure the final apply step records the batch with undo support and flags any remaining conflicts.

**Acceptance Scenarios**:

1. **Given** an AI plan with two conflicting targets, **When** the operator attempts to apply without resolving them, **Then** the CLI blocks execution and enumerates the conflicting entries.
2. **Given** the operator edits AI output to change a sequence token, **When** the plan is revalidated, **Then** the tool reorders numbers and confirms the ledger entry captures the final applied mapping.

---

### Edge Cases

- AI response omits some files from the original list or introduces unfamiliar entries.
- Returned filenames exceed allowed length or include forbidden characters for the host OS.
- AI-generated numbers skip or duplicate sequence values.
- The AI service is unavailable, times out, or returns malformed JSON.
- Users request rules that contradict each other (e.g., camelCase and kebab-case simultaneously).
- Sanitization removes all meaningful characters, resulting in empty or duplicate stems.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST collect the active scope (paths, filters, counts) and compose an AI prompt that includes representative filenames plus the required renaming rules (sequence numbering, uniform formatting, spam removal, additional heuristics).
- **FR-002**: The prompt MUST instruct the AI to respond in a documented, parseable structure (e.g., JSON with original and proposed names) and to preserve file extensions; the CLI MUST default to an OpenAI-compatible model (override via flag/env).
- **FR-003**: The system MUST validate the AI response, ensuring every scoped file has a proposed rename, sequence numbers are continuous, names are unique, and disallowed content is removed before previewing changes.
- **FR-004**: Users MUST be able to supply optional naming policy inputs (project tag, casing preference, separator choice, forbidden words) that the prompt reflects and the validator enforces.
- **FR-005**: The preview MUST display AI-proposed names with sequence numbers, highlight sanitized segments, and surface any entries needing manual edits before the apply step is allowed.
- **FR-006**: The CLI MUST allow users to edit or regenerate portions of the AI plan, re-run validation, and only enable apply once all issues are resolved.
- **FR-007**: Apply MUST record the final mappings and the AI prompt/response metadata in the `.renamer` ledger so undo can restore original names and provide auditability.
- **FR-008**: The workflow MUST handle AI communication failures gracefully by surfacing clear errors and leaving existing files untouched.
- **FR-009**: The system MUST prevent AI output from introducing prohibited terms, promotional phrases, or user-defined banned tokens into the resulting filenames.
- **FR-010**: The CLI MUST support dry-run mode for AI interactions, allowing prompt/response review without executing filesystem changes.

### Key Entities *(include if feature involves data)*

- **AiRenamePrompt**: Captures scope summary, sample filenames, mandatory rules, optional user policies, and guardrails sent to the AI.
- **AiRenameResponse**: Structured data returned by the AI containing proposed filenames, rationale, and any warnings or unresolved cases.
- **AiRenamePlan**: Aggregated representation of validated rename operations, including sequence ordering, sanitization notes, and conflict markers used for preview/apply.
- **AiRenameLedgerMetadata**: Audit payload storing prompt hash, response hash, applied policy parameters, and timestamp for undo traceability.

### Assumptions

- Users provide access to an AI endpoint capable of following structured prompts and returning JSON within size limits; default secret tokens (`*_MODEL_AUTH_TOKEN`) are stored under `$HOME/.config/.renamer/`.
- File extensions must remain unchanged; only stems are rewritten.
- Default numbering uses three-digit, zero-padded prefixes unless the user specifies a different width or format.
- The CLI environment already authenticates outbound AI requests; this feature focuses on prompt content and result handling.
- Promotional or spam phrases are identified via a maintainable stop-word list augmented by user-provided banned terms.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 95% of AI-generated rename plans pass validation without requiring manual edits on the first attempt for test batches of up to 1,000 files.
- **SC-002**: Operators can review AI preview results and apply approved renames within 5 minutes for a 200-file batch, including validation and optional edits.
- **SC-003**: 100% of applied AI-driven rename batches produce ledger entries with complete prompt/response metadata, enabling successful undo in under 60 seconds.
- **SC-004**: User satisfaction surveys report at least 85% agreement that AI-assisted renaming produced clearer, uniform filenames compared to prior manual methods within one release cycle.
