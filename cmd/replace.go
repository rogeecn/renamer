package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/replace"
)

// NewReplaceCommand constructs the replace CLI command; exported for testing.
func NewReplaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace <pattern1> [pattern2 ...] <replacement>",
		Short: "Replace multiple literals in file and directory names",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			parseResult, err := replace.ParseArgs(args)
			if err != nil {
				return err
			}

			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			req := &replace.ReplaceRequest{
				WorkingDir:         scope.WorkingDir,
				Patterns:           parseResult.Patterns,
				Replacement:        parseResult.Replacement,
				IncludeDirectories: scope.IncludeDirectories,
				Recursive:          scope.Recursive,
				IncludeHidden:      scope.IncludeHidden,
				Extensions:         scope.Extensions,
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

			out := cmd.OutOrStdout()
			summary, planned, err := replace.Preview(cmd.Context(), req, parseResult, out)
			if err != nil {
				return err
			}

			for _, dup := range summary.SortedDuplicates() {
				fmt.Fprintf(out, "Warning: pattern %q provided multiple times\n", dup)
			}

			if len(summary.Conflicts) > 0 {
				for _, conflict := range summary.Conflicts {
					fmt.Fprintf(out, "CONFLICT: %s -> %s (%s)\n", conflict.OriginalPath, conflict.ProposedPath, conflict.Reason)
				}
				return errors.New("conflicts detected; aborting")
			}

			if summary.ChangedCount == 0 {
				fmt.Fprintln(out, "No replacements required")
				return nil
			}

			fmt.Fprintf(out, "Planned replacements: %d entries updated across %d candidates\n", summary.ChangedCount, summary.TotalCandidates)
			for pattern, count := range summary.PatternMatches {
				fmt.Fprintf(out, "  %s -> %d occurrences\n", pattern, count)
			}

			if dryRun || !autoApply {
				fmt.Fprintln(out, "Preview complete. Re-run with --yes to apply.")
				return nil
			}

			entry, err := replace.Apply(context.Background(), req, planned, summary)
			if err != nil {
				return err
			}

			if len(entry.Operations) == 0 {
				fmt.Fprintln(out, "Nothing to apply; preview already up to date.")
				return nil
			}

			fmt.Fprintf(out, "Applied %d replacements. Ledger updated.\n", len(entry.Operations))
			return nil
		},
	}

	cmd.Example = `  renamer replace draft Draft final --dry-run
  renamer replace "Project X" "Project-X" ProjectX --yes --path ./docs`

	return cmd
}

func getBool(cmd *cobra.Command, name string) (bool, error) {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return cmd.Flags().GetBool(name)
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return cmd.InheritedFlags().GetBool(name)
	}
	return false, fmt.Errorf("flag %s not defined", name)
}

func init() {
	rootCmd.AddCommand(NewReplaceCommand())
}
