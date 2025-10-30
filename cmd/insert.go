package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/insert"
	"github.com/rogeecn/renamer/internal/listing"
)

func newInsertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "insert <position> <text>",
		Short: "Insert text into filenames at specified positions",
		Long: `Insert a Unicode string into each candidate filename at a specific position.
Supported positions: "^" (start), "$" (before extension), positive indexes (1-based),
negative indexes counting back from the end.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			req := insert.NewRequest(scope)

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
			req.SetExecutionMode(dryRun, autoApply)

			positionToken := args[0]
			insertText := strings.Join(args[1:], " ")
			req.SetPositionAndText(positionToken, insertText)

			summary, planned, err := insert.Preview(cmd.Context(), req, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			if summary.HasConflicts() {
				return errors.New("conflicts detected; resolve them before applying")
			}

			if dryRun || !autoApply {
				if !autoApply {
					fmt.Fprintln(cmd.OutOrStdout(), "Preview complete. Re-run with --yes to apply.")
				}
				return nil
			}

			if len(planned) == 0 {
				if summary.TotalCandidates == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No candidates found.")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "Nothing to apply; files already reflect requested insert.")
				}
				return nil
			}

			entry, err := insert.Apply(cmd.Context(), req, planned, summary)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Applied %d insert updates. Ledger updated.\n", len(entry.Operations))
			return nil
		},
	}

	cmd.Example = `  renamer insert ^ "[2025] " --dry-run
  renamer insert -1 _FINAL --yes --path ./reports`

	return cmd
}

func init() {
	rootCmd.AddCommand(newInsertCommand())
}
