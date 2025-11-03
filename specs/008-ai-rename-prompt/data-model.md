# Data Model: AI-Assisted Rename Prompting

## AiRenamePrompt
- **Fields**
  - `WorkingDir` (string; absolute path)
  - `SampleFiles` ([]PromptSample)
  - `TotalCount` (int)
  - `SequenceRule` (SequenceRuleConfig)
  - `NamingPolicies` (NamingPolicyConfig)
  - `BannedTerms` ([]string)
  - `Metadata` (map[string]string; includes CLI version, timestamp)
- **Relationships**: Built from traversal summary; sent to Genkit workflow.
- **Validations**: Require ≥1 sample, forbid empty banned terms, limit payload to ≤2 MB serialized.

### PromptSample
- **Fields**
  - `OriginalName` (string)
  - `SizeBytes` (int64)
  - `Extension` (string)
  - `PathDepth` (int)

### SequenceRuleConfig
- **Fields**
  - `Style` (enum: `prefix`, `suffix`)
  - `Width` (int)
  - `Start` (int)
  - `Separator` (string)

### NamingPolicyConfig
- **Fields**
  - `Prefix` (string)
  - `Casing` (enum: `kebab`, `snake`, `camel`, `pascal`, `title`)
  - `AllowSpaces` (bool)
  - `KeepOriginalOrder` (bool)
  - `ForbiddenTokens` ([]string)

## AiRenameResponse
- **Fields**
  - `Items` ([]AiRenameItem)
  - `Warnings` ([]string)
  - `PromptHash` (string)
  - `Model` (string)
- **Relationships**: Parsed from Genkit output; feeds validation pipeline.
- **Validations**: Items length must equal scoped candidate count; `Original` names must match traversal list; `Proposed` must be unique.

### AiRenameItem
- **Fields**
  - `Original` (string; relative path)
  - `Proposed` (string; sanitized stem + extension)
  - `Sequence` (int)
  - `Notes` (string; optional reasoning)

## AiRenamePlan
- **Fields**
  - `Candidates` ([]PlanEntry)
  - `Conflicts` ([]PlanConflict)
  - `Policies` (NamingPolicyConfig)
  - `SequenceAppliedWidth` (int)
  - `Warnings` ([]string)
- **Relationships**: Created post-validation for preview/apply; persisted to ledger metadata.
- **Validations**: Enforce contiguous sequences, ensure sanitized stems non-empty, confirm banned tokens absent.

### PlanEntry
- **Fields**
  - `OriginalPath` (string)
  - `ProposedPath` (string)
  - `Sequence` (int)
  - `Status` (enum: `pending`, `edited`, `skipped`, `unchanged`)
  - `SanitizedSegments` ([]string)

### PlanConflict
- **Fields**
  - `OriginalPath` (string)
  - `Issue` (enum: `duplicate`, `collision`, `policy_violation`, `missing_sequence`)
  - `Details` (string)

## AiRenameLedgerMetadata
- **Fields**
  - `PromptHash` (string)
  - `ResponseHash` (string)
  - `Model` (string)
  - `Policies` (NamingPolicyConfig)
  - `BatchSize` (int)
  - `AppliedAt` (time.Time)
- **Relationships**: Stored under `Entry.Metadata["ai"]` when applying rename batch; consumed by undo for auditing.

## GenkitWorkflowConfig
- **Fields**
  - `Endpoint` (string; local or remote URL)
  - `Timeout` (Duration)
  - `RetryPolicy` (RetryPolicy)
  - `ApiKeyRef` (string; environment variable name mapping to `$HOME/.config/.renamer/{name}` token file)

### RetryPolicy
- **Fields**
  - `MaxAttempts` (int)
  - `BackoffInitial` (Duration)
  - `BackoffMultiplier` (float64)
  - `FallbackModel` (string; optional)
