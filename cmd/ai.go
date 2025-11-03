package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	Model             string
	Debug             bool
	ExportPath        string
	ImportPath        string
	Casing            string
	Prefix            string
	AllowSpaces       bool
	KeepOriginalOrder bool
	BannedTokens      []string
}

func newAICommand() *cobra.Command {
	ops := &aiCommandOptions{}

	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Generate rename plans using the AI workflow",
		Long:  "Invoke the embedded AI workflow to generate, validate, and optionally apply rename plans.",
		Example: strings.TrimSpace(`  # Preview an AI plan and export the raw response for edits
	renamer ai --path ./photos --dry-run --export-plan plan.json

	# Import an edited plan and validate it without applying changes
	renamer ai --path ./photos --dry-run --import-plan plan.json

	# Apply an edited plan after validation passes
	renamer ai --path ./photos --import-plan plan.json --yes`),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := collectAIOptions(cmd, ops)
			return runAICommand(cmd.Context(), cmd, options)
		},
	}

	bindAIFlags(cmd, ops)

	return cmd
}

func bindAIFlags(cmd *cobra.Command, opts *aiCommandOptions) {
	cmd.Flags().StringVar(&opts.Model, "genkit-model", genkit.DefaultModelName, fmt.Sprintf("OpenAI-compatible model identifier (default %s)", genkit.DefaultModelName))
	cmd.Flags().BoolVar(&opts.Debug, "debug-genkit", false, "Write Genkit prompt/response traces to the debug log")
	cmd.Flags().StringVar(&opts.ExportPath, "export-plan", "", "Export the raw AI plan JSON to the provided file path")
	cmd.Flags().StringVar(&opts.ImportPath, "import-plan", "", "Import an edited AI plan JSON for validation or apply")
	cmd.Flags().StringVar(&opts.Casing, "naming-casing", "kebab", "Casing style for AI-generated filenames (kebab, snake, camel, pascal, title)")
	cmd.Flags().StringVar(&opts.Prefix, "naming-prefix", "", "Static prefix AI proposals must include (alias: --prefix)")
	cmd.Flags().StringVar(&opts.Prefix, "prefix", "", "Alias for --naming-prefix")
	cmd.Flags().BoolVar(&opts.AllowSpaces, "naming-allow-spaces", false, "Permit spaces in AI-generated filenames")
	cmd.Flags().BoolVar(&opts.KeepOriginalOrder, "naming-keep-order", false, "Instruct AI to preserve original ordering of descriptive terms")
	cmd.Flags().StringSliceVar(&opts.BannedTokens, "banned", nil, "Comma-separated list of additional banned tokens (repeat flag to add more)")
}

func collectAIOptions(cmd *cobra.Command, defaults *aiCommandOptions) aiCommandOptions {
	result := aiCommandOptions{
		Model:      genkit.DefaultModelName,
		Debug:      false,
		ExportPath: "",
		Casing:     "kebab",
	}

	if defaults != nil {
		if defaults.Model != "" {
			result.Model = defaults.Model
		}
		result.Debug = defaults.Debug
		result.ExportPath = defaults.ExportPath
		if defaults.Casing != "" {
			result.Casing = defaults.Casing
		}
		result.Prefix = defaults.Prefix
		result.AllowSpaces = defaults.AllowSpaces
		result.KeepOriginalOrder = defaults.KeepOriginalOrder
		if len(defaults.BannedTokens) > 0 {
			result.BannedTokens = append([]string(nil), defaults.BannedTokens...)
		}
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

	if flag := cmd.Flags().Lookup("export-plan"); flag != nil {
		if value, err := cmd.Flags().GetString("export-plan"); err == nil && value != "" {
			result.ExportPath = value
		}
	}

	if flag := cmd.Flags().Lookup("import-plan"); flag != nil {
		if value, err := cmd.Flags().GetString("import-plan"); err == nil && value != "" {
			result.ImportPath = value
		}
	}

	if flag := cmd.Flags().Lookup("naming-casing"); flag != nil {
		if value, err := cmd.Flags().GetString("naming-casing"); err == nil && value != "" {
			result.Casing = value
		}
	}

	if flag := cmd.Flags().Lookup("naming-prefix"); flag != nil {
		if value, err := cmd.Flags().GetString("naming-prefix"); err == nil {
			result.Prefix = value
		}
	}
	if flag := cmd.Flags().Lookup("prefix"); flag != nil && flag.Changed {
		if value, err := cmd.Flags().GetString("prefix"); err == nil {
			result.Prefix = value
		}
	}

	if flag := cmd.Flags().Lookup("naming-allow-spaces"); flag != nil {
		if value, err := cmd.Flags().GetBool("naming-allow-spaces"); err == nil {
			result.AllowSpaces = value
		}
	}

	if flag := cmd.Flags().Lookup("naming-keep-order"); flag != nil {
		if value, err := cmd.Flags().GetBool("naming-keep-order"); err == nil {
			result.KeepOriginalOrder = value
		}
	}

	if flag := cmd.Flags().Lookup("banned"); flag != nil {
		if value, err := cmd.Flags().GetStringSlice("banned"); err == nil && len(value) > 0 {
			result.BannedTokens = append([]string(nil), value...)
		}
	}

	return result
}

func runAICommand(ctx context.Context, cmd *cobra.Command, options aiCommandOptions) error {
	scope, err := listing.ScopeFromCmd(cmd)
	if err != nil {
		return err
	}

	applyRequested, err := getBool(cmd, "yes")
	if err != nil {
		return err
	}

	options.ImportPath = strings.TrimSpace(options.ImportPath)

	casing, err := normalizeCasing(options.Casing)
	if err != nil {
		return err
	}
	options.Casing = casing
	prefix := strings.TrimSpace(options.Prefix)
	userBanned := sanitizeTokenSlice(options.BannedTokens)
	bannedTerms := mergeBannedTerms(defaultBannedTerms(), userBanned)

	candidates, err := plan.CollectCandidates(ctx, scope)
	if err != nil {
		return err
	}
	ignoreSet := buildIgnoreSet(scope.WorkingDir, options.ExportPath, options.ImportPath)
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
		Prefix:            prefix,
		Casing:            options.Casing,
		AllowSpaces:       options.AllowSpaces,
		KeepOriginalOrder: options.KeepOriginalOrder,
		ForbiddenTokens:   append([]string(nil), userBanned...),
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

	if options.ImportPath != "" {
		resp, err := plan.LoadResponse(options.ImportPath)
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

		if options.ExportPath != "" {
			if err := plan.SaveResponse(options.ExportPath, response); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "AI plan exported to %s\n", options.ExportPath)
		}
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

	if options.ImportPath == "" && options.ExportPath != "" {
		// Plan already exported earlier.
	} else if options.ImportPath != "" && options.ExportPath != "" {
		if err := plan.SaveResponse(options.ExportPath, response); err != nil {
			return err
		}
		fmt.Fprintf(errorWriter, "AI plan exported to %s\n", options.ExportPath)
	}

	if !applyRequested {
		return nil
	}

	if len(previewPlan.Conflicts) > 0 {
		return fmt.Errorf("cannot apply AI plan while conflicts remain")
	}

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
				fmt.Fprintf(errorWriter, "Apply conflict (%s): %s %s\n", conflict.Issue, conflict.OriginalPath, conflict.Details)
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
		fmt.Sprintf("Use %s numbering with width %d starting at %d and separator %q.", sequence.Style, sequence.Width, sequence.Start, sequence.Separator),
		"Preserve original file extensions exactly as provided.",
		fmt.Sprintf("Apply %s casing to filename stems and avoid promotional or banned terms.", policies.Casing),
		"Ensure proposed names are unique and sequences remain contiguous.",
	}
	if policies.Prefix != "" {
		lines = append(lines, fmt.Sprintf("Every proposed filename must begin with the prefix %q immediately before descriptive text.", policies.Prefix))
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
		lines = append(lines, fmt.Sprintf("Never include these banned tokens (case-insensitive) in any proposed filename: %s.", strings.Join(bannedTerms, ", ")))
	}
	return strings.Join(lines, "\n")
}

func normalizeCasing(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "kebab", nil
	}
	lower := strings.ToLower(trimmed)
	supported := map[string]string{
		"kebab":  "kebab",
		"snake":  "snake",
		"camel":  "camel",
		"pascal": "pascal",
		"title":  "title",
	}
	if normalized, ok := supported[lower]; ok {
		return normalized, nil
	}
	return "", fmt.Errorf("unsupported naming casing %q (allowed: kebab, snake, camel, pascal, title)", value)
}

func sanitizeTokenSlice(values []string) []string {
	unique := make(map[string]struct{})
	for _, raw := range values {
		for _, part := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed == "" {
				continue
			}
			lower := strings.ToLower(trimmed)
			if lower == "" {
				continue
			}
			unique[lower] = struct{}{}
		}
	}
	if len(unique) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(unique))
	for token := range unique {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	return tokens
}

func mergeBannedTerms(base, extra []string) []string {
	unique := make(map[string]struct{})
	for _, token := range base {
		lower := strings.ToLower(strings.TrimSpace(token))
		if lower == "" {
			continue
		}
		unique[lower] = struct{}{}
	}
	for _, token := range extra {
		lower := strings.ToLower(strings.TrimSpace(token))
		if lower == "" {
			continue
		}
		unique[lower] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for token := range unique {
		result = append(result, token)
	}
	sort.Strings(result)
	return result
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
