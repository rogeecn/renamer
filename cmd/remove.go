package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/remove"
)

// NewRemoveCommand constructs the remove CLI command; exported for testing.
func NewRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <pattern1> [pattern2 ...]",
		Short: "Remove literal substrings sequentially from names",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := remove.ParseArgs(args)
			if err != nil {
				return err
			}

			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			req, err := remove.FromListing(scope, parsed.Tokens)
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

			out := cmd.OutOrStdout()

			summary, planned, err := remove.Preview(cmd.Context(), req, parsed, out)
			if err != nil {
				return err
			}

			for _, empty := range summary.Empties {
				fmt.Fprintf(out, "Warning: %s would become empty; skipping\n", empty)
			}

			for _, dup := range summary.SortedDuplicates() {
				fmt.Fprintf(out, "Warning: token %q provided multiple times\n", dup)
			}

			if len(summary.Conflicts) > 0 {
				for _, conflict := range summary.Conflicts {
					fmt.Fprintf(out, "CONFLICT: %s -> %s (%s)\n", conflict.OriginalPath, conflict.ProposedPath, conflict.Reason)
				}
				return errors.New("conflicts detected; aborting")
			}

			if summary.ChangedCount == 0 {
				fmt.Fprintln(out, "No removals required")
				return nil
			}

			fmt.Fprintf(out, "Planned removals: %d entries updated across %d candidates\n", summary.ChangedCount, summary.TotalCandidates)
			for _, pair := range summary.SortedTokenMatches() {
				fmt.Fprintf(out, "  %s -> %d occurrences\n", pair.Token, pair.Count)
			}

			if dryRun || !autoApply {
				fmt.Fprintln(out, "Preview complete. Re-run with --yes to apply.")
				return nil
			}

			entry, err := remove.Apply(cmd.Context(), req, planned, summary, parsed.Tokens)
			if err != nil {
				return err
			}

			if len(entry.Operations) == 0 {
				fmt.Fprintln(out, "Nothing to apply; preview already up to date.")
				return nil
			}

			fmt.Fprintf(out, "Applied %d removals. Ledger updated.\n", len(entry.Operations))
			return nil
		},
	}

	cmd.Example = "  renamer remove \" copy\" \" draft\" --dry-run\n  renamer remove foo bar --yes --path ./docs"

	return cmd
}

func init() {
	rootCmd.AddCommand(NewRemoveCommand())
}
