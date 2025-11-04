package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/ai/genkit"
	"github.com/rogeecn/renamer/internal/ai/plan"
	"github.com/rogeecn/renamer/internal/ai/prompt"
	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/output"
)

type aiCommandOptions struct {
	Model string
	Debug bool
}

const aiPlanFilename = "renamer.plan.json"

// newAICommand 构建 `renamer ai` 子命令，仅保留模型选择与调试标志，其他策略交由 AI 自行生成。
func newAICommand() *cobra.Command {
	ops := &aiCommandOptions{}

	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Generate rename plans using the AI workflow",
		Long:  "Invoke the embedded AI workflow to generate, validate, and optionally apply rename plans.",
		Example: strings.TrimSpace(`  # Generate a plan for review in renamer.plan.json
	renamer ai --path ./photos --dry-run

	# Apply the reviewed plan after confirming the preview
	renamer ai --path ./photos --yes`),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := collectAIOptions(cmd, ops)
			return runAICommand(cmd.Context(), cmd, options)
		},
	}

	bindAIFlags(cmd, ops)

	return cmd
}

func bindAIFlags(cmd *cobra.Command, opts *aiCommandOptions) {
	cmd.Flags().
		StringVar(&opts.Model, "genkit-model", genkit.DefaultModelName, fmt.Sprintf("OpenAI-compatible model identifier (default %s)", genkit.DefaultModelName))
	cmd.Flags().BoolVar(&opts.Debug, "debug-genkit", false, "Write Genkit prompt/response traces to the debug log")
}

func collectAIOptions(cmd *cobra.Command, defaults *aiCommandOptions) aiCommandOptions {
	result := aiCommandOptions{
		Model: genkit.DefaultModelName,
		Debug: false,
	}

	if defaults != nil {
		if defaults.Model != "" {
			result.Model = defaults.Model
		}
		result.Debug = defaults.Debug
	}

	if flag := cmd.Flags().Lookup("genkit-model"); flag != nil {
		if value, err := cmd.Flags().GetString("genkit-model"); err == nil && value != "" {
			result.Model = value
		}
	}

	if flag := cmd.Flags().Lookup("debug-genkit"); flag != nil {
		if value, err := cmd.Flags().GetBool("debug-genkit"); err == nil {
			result.Debug = value
		}
	}

	return result
}

// runAICommand 按以下顺序执行 AI 重命名流程：
// 1. 解析作用范围与是否需要立即应用；
// 2. 自动探测工作目录下的 renamer.plan.json，决定是加载人工调整还是生成新计划；
// 3. 收集候选文件并过滤生成过程中的辅助文件；
// 4. 通过 Genkit 工作流调用模型生成方案或读取既有方案；
// 5. 保存/更新本地计划文件，随后校验、渲染预览并输出冲突与告警；
// 6. 在用户确认后执行改名并记录账本。
func runAICommand(ctx context.Context, cmd *cobra.Command, options aiCommandOptions) error {
	scope, err := listing.ScopeFromCmd(cmd)
	if err != nil {
		return err
	}

	applyRequested, err := getBool(cmd, "yes")
	if err != nil {
		return err
	}

	// 探测当前目录下的计划文件，支持人工预处理后再运行。
	planPath := filepath.Join(scope.WorkingDir, aiPlanFilename)
	planExists := false
	if info, err := os.Stat(planPath); err == nil {
		if info.IsDir() {
			return fmt.Errorf("plan file %s is a directory", planPath)
		}
		planExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("plan file %s: %w", planPath, err)
	}

	// 默认策略完全交由提示模板处理，仅保留基础禁止词。
	casing := "kebab"
	bannedTerms := defaultBannedTerms()

	// 收集所有候选文件，剔除计划文件自身避免被改名。
	candidates, err := plan.CollectCandidates(ctx, scope)
	if err != nil {
		return err
	}
	ignoreSet := buildIgnoreSet(scope.WorkingDir, planPath)
	if len(ignoreSet) > 0 {
		candidates = filterIgnoredCandidates(candidates, ignoreSet)
	}
	if len(candidates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No candidates found")
		return nil
	}

	samples := make([]prompt.SampleCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		samples = append(samples, prompt.SampleCandidate{
			RelativePath: candidate.OriginalPath,
			SizeBytes:    candidate.SizeBytes,
			Depth:        candidate.Depth,
		})
	}

	sequence := prompt.SequenceRule{
		Style:     "prefix",
		Width:     3,
		Start:     1,
		Separator: "_",
	}

	policies := prompt.PolicyConfig{
		Prefix:            "",
		Casing:            casing,
		AllowSpaces:       false,
		KeepOriginalOrder: false,
		ForbiddenTokens:   append([]string(nil), bannedTerms...),
	}
	validatorPolicy := prompt.NamingPolicyConfig{
		Prefix:            policies.Prefix,
		Casing:            policies.Casing,
		AllowSpaces:       policies.AllowSpaces,
		KeepOriginalOrder: policies.KeepOriginalOrder,
		ForbiddenTokens:   append([]string(nil), policies.ForbiddenTokens...),
	}

	var response prompt.RenameResponse
	var promptHash string
	var model string

	if planExists {
		// 若检测到已有计划，则优先加载人工编辑的方案继续校验/执行。
		resp, err := plan.LoadResponse(planPath)
		if err != nil {
			return err
		}
		response = resp
		promptHash = strings.TrimSpace(resp.PromptHash)
		model = strings.TrimSpace(resp.Model)
		if model == "" {
			model = options.Model
		}
	} else {
		// 没有计划文件时，调用 Genkit 工作流生成全新方案。
		builder := prompt.NewBuilder()
		promptPayload, err := builder.Build(prompt.BuildInput{
			WorkingDir:  scope.WorkingDir,
			Samples:     samples,
			TotalCount:  len(candidates),
			Sequence:    sequence,
			Policies:    policies,
			BannedTerms: bannedTerms,
			Metadata: map[string]string{
				"cliVersion": "dev",
			},
		})
		if err != nil {
			return err
		}

		instructions := composeInstructions(sequence, policies, bannedTerms)
		client := genkit.NewClient(genkit.ClientOptions{Model: options.Model})
		invocationResult, err := client.Invoke(ctx, genkit.Invocation{
			Instructions: instructions,
			Prompt:       promptPayload,
			Model:        options.Model,
		})
		if err != nil {
			return err
		}
		response = invocationResult.Response
		promptHash = invocationResult.PromptHash
		model = invocationResult.Response.Model
	}

	if promptHash == "" {
		if hash, err := plan.ResponseDigest(response); err == nil {
			promptHash = hash
		}
	}
	if model == "" {
		model = options.Model
	}
	response.PromptHash = promptHash
	response.Model = model

	// 将生成或加载的计划写回本地，便于后续人工审核或复用。
	if err := plan.SaveResponse(planPath, response); err != nil {
		return err
	}
	message := "AI plan saved to %s\n"
	if planExists {
		message = "AI plan updated at %s\n"
	}
	fmt.Fprintf(cmd.ErrOrStderr(), message, planPath)

	originals := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		originals = append(originals, candidate.OriginalPath)
	}

	validator := plan.NewValidator(originals, validatorPolicy, bannedTerms)
	validationResult, err := validator.Validate(response)
	if err != nil {
		var vErr *plan.ValidationError
		if errors.As(err, &vErr) {
			errorWriter := cmd.ErrOrStderr()
			if len(vErr.PolicyViolations) > 0 {
				messages := make([]output.PolicyViolationMessage, 0, len(vErr.PolicyViolations))
				for _, violation := range vErr.PolicyViolations {
					messages = append(messages, output.PolicyViolationMessage{
						Original: violation.Original,
						Proposed: violation.Proposed,
						Rule:     violation.Rule,
						Message:  violation.Message,
					})
				}
				output.WritePolicyViolations(errorWriter, messages)
			}
		}
		return err
	}

	previewPlan, err := plan.MapResponse(plan.MapInput{
		Candidates:    candidates,
		SequenceWidth: sequence.Width,
	}, validationResult)
	if err != nil {
		return err
	}
	previewPlan.PromptHash = promptHash
	if previewPlan.Model == "" {
		previewPlan.Model = model
	}

	// 输出预览表格与告警，帮助用户确认重命名提案。
	if err := renderAIPlan(cmd.OutOrStdout(), previewPlan); err != nil {
		return err
	}

	errorWriter := cmd.ErrOrStderr()
	if len(previewPlan.Conflicts) > 0 {
		for _, conflict := range previewPlan.Conflicts {
			fmt.Fprintf(errorWriter, "Conflict (%s): %s %s\n", conflict.Issue, conflict.OriginalPath, conflict.Details)
		}
	}

	if options.Debug {
		output.WriteAIPlanDebug(errorWriter, promptHash, previewPlan.Warnings)
	} else if len(previewPlan.Warnings) > 0 {
		output.WriteAIPlanDebug(errorWriter, "", previewPlan.Warnings)
	}

	if !applyRequested {
		return nil
	}

	if len(previewPlan.Conflicts) > 0 {
		return fmt.Errorf("cannot apply AI plan while conflicts remain")
	}

	// 在无冲突且用户确认的情况下，按计划执行并记录到账本。
	applyEntry, err := plan.Apply(ctx, plan.ApplyOptions{
		WorkingDir: scope.WorkingDir,
		Candidates: candidates,
		Response:   response,
		Policies:   validatorPolicy,
		PromptHash: promptHash,
	})
	if err != nil {
		var conflictErr plan.ApplyConflictError
		if errors.As(err, &conflictErr) {
			for _, conflict := range conflictErr.Conflicts {
				fmt.Fprintf(
					errorWriter,
					"Apply conflict (%s): %s %s\n",
					conflict.Issue,
					conflict.OriginalPath,
					conflict.Details,
				)
			}
		}
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Applied %d renames. Ledger updated.\n", len(applyEntry.Operations))
	return nil
}

func renderAIPlan(w io.Writer, preview plan.PreviewPlan) error {
	table := output.NewAIPlanTable()
	if err := table.Begin(w); err != nil {
		return err
	}
	for _, entry := range preview.Entries {
		sanitized := "-"
		if len(entry.SanitizedSegments) > 0 {
			joined := strings.Join(entry.SanitizedSegments, " ")
			sanitized = "removed: " + joined
		}
		if entry.Notes != "" {
			if sanitized == "-" {
				sanitized = entry.Notes
			} else {
				sanitized = fmt.Sprintf("%s (%s)", sanitized, entry.Notes)
			}
		}
		row := output.AIPlanRow{
			Sequence:  entry.SequenceLabel,
			Original:  entry.OriginalPath,
			Proposed:  entry.ProposedPath,
			Sanitized: sanitized,
		}
		if err := table.WriteRow(row); err != nil {
			return err
		}
	}
	return table.End(w)
}

func composeInstructions(sequence prompt.SequenceRule, policies prompt.PolicyConfig, bannedTerms []string) string {
	lines := []string{
		"You are an AI assistant that proposes safe file rename plans.",
		"Return JSON matching this schema: {\"items\":[{\"original\":string,\"proposed\":string,\"sequence\":number,\"notes\"?:string}],\"warnings\"?:[string]}.",
		fmt.Sprintf(
			"Use %s numbering with width %d starting at %d and separator %q.",
			sequence.Style,
			sequence.Width,
			sequence.Start,
			sequence.Separator,
		),
		"Preserve original file extensions exactly as provided.",
		fmt.Sprintf("Apply %s casing to filename stems and avoid promotional or banned terms.", policies.Casing),
		"Ensure proposed names are unique and sequences remain contiguous.",
	}
	if policies.Prefix != "" {
		lines = append(
			lines,
			fmt.Sprintf(
				"Every proposed filename must begin with the prefix %q immediately before descriptive text.",
				policies.Prefix,
			),
		)
	}
	if policies.AllowSpaces {
		lines = append(lines, "Spaces in filenames are permitted when they improve clarity.")
	} else {
		lines = append(lines, "Do not include spaces in filenames; use separators consistent with the requested casing style.")
	}
	if policies.KeepOriginalOrder {
		lines = append(lines, "Preserve the original ordering of meaningful words when generating new stems.")
	}
	if len(bannedTerms) > 0 {
		lines = append(
			lines,
			fmt.Sprintf(
				"Never include these banned tokens (case-insensitive) in any proposed filename: %s.",
				strings.Join(bannedTerms, ", "),
			),
		)
	}
	return strings.Join(lines, "\n")
}

func buildIgnoreSet(workingDir string, paths ...string) map[string]struct{} {
	ignore := make(map[string]struct{})
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		abs, err := filepath.Abs(trimmed)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(workingDir, abs)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, "..") {
			continue
		}
		ignore[strings.ToLower(filepath.ToSlash(rel))] = struct{}{}
	}
	return ignore
}

func filterIgnoredCandidates(candidates []plan.Candidate, ignore map[string]struct{}) []plan.Candidate {
	if len(ignore) == 0 {
		return candidates
	}
	filtered := make([]plan.Candidate, 0, len(candidates))
	for _, cand := range candidates {
		if _, skip := ignore[strings.ToLower(cand.OriginalPath)]; skip {
			continue
		}
		filtered = append(filtered, cand)
	}
	return filtered
}

func defaultBannedTerms() []string {
	terms := []string{"promo", "sale", "free", "clickbait", "sponsored"}
	sort.Strings(terms)
	return terms
}

func init() {
	rootCmd.AddCommand(newAICommand())
}
