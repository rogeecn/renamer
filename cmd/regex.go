package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/regex"
)

func newRegexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regex <pattern> <template>",
		Short: "Rename files using regex capture groups",
		Long: `Preview and apply filename changes by extracting capture groups from a regular
expression pattern. Placeholders like @1, @2 refer to captured groups; @0 expands to the full match,
and @@ emits a literal @. Undefined placeholders and invalid replacement templates result in
validation errors before any filesystem changes occur.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			dryRun, err := getBool(cmd, "dry-run")
			if err != nil {
				return err
			}
			autoApply, err := getBool(cmd, "yes")
			if err != nil {
				return err
			}
			if dryRun && autoApply {
				return errors.New("--dry-run cannot be combined with --yes; remove one of them")
			}

			request := regex.NewRequest(scope.WorkingDir)
			request.Pattern = args[0]
			request.Template = args[1]
			request.IncludeDirectories = scope.IncludeDirectories
			request.Recursive = scope.Recursive
			request.IncludeHidden = scope.IncludeHidden
			request.Extensions = append([]string(nil), scope.Extensions...)
			request.DryRun = dryRun
			request.AutoConfirm = autoApply

			out := cmd.OutOrStdout()
			summary, planned, err := regex.Preview(cmd.Context(), request, out)
			if err != nil {
				return err
			}

			for _, warning := range summary.Warnings {
				fmt.Fprintf(out, "Warning: %s\n", warning)
			}

			if len(summary.Conflicts) > 0 {
				for _, conflict := range summary.Conflicts {
					fmt.Fprintf(out, "CONFLICT: %s -> %s (%s)\n", conflict.OriginalPath, conflict.ProposedPath, conflict.Reason)
				}
				return errors.New("conflicts detected; aborting")
			}

			if summary.Changed == 0 {
				fmt.Fprintln(out, "No regex renames required.")
				return nil
			}

			if !autoApply {
				fmt.Fprintf(out, "Preview complete: %d matched, %d changed, %d skipped.\n", summary.Matched, summary.Changed, summary.Skipped)
				fmt.Fprintln(out, "Preview complete. Re-run with --yes to apply.")
				return nil
			}

			entry, err := regex.Apply(cmd.Context(), request, planned, summary)
			if err != nil {
				return err
			}

			if len(entry.Operations) == 0 {
				fmt.Fprintln(out, "Nothing to apply; files already matched requested pattern.")
				return nil
			}

			fmt.Fprintf(out, "Applied %d regex renames. Ledger updated.\n", len(entry.Operations))
			return nil
		},
	}

	cmd.Example = `  renamer regex "^(\\w+)-(\\d+)" "@2_@1" --dry-run
  renamer regex "^(build)_(\\d+)_v(.*)$" "release-@2-@1-v@3" --yes --path ./artifacts
  renamer regex "^(.*)$" "release-@1" --dry-run   # fails when placeholders are undefined`

	return cmd
}

func init() {
	rootCmd.AddCommand(newRegexCommand())
}
