package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/sequence"
)

func newSequenceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Append or prepend sequential numbers to filenames",
		Long: `Preview and apply numbered renames across the active scope. Sequence numbers
respect deterministic traversal order, support configurable start offsets, and record every
batch in the .renamer ledger for undo.`,
		Args: cobra.NoArgs,
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

			start, err := cmd.Flags().GetInt("start")
			if err != nil {
				return err
			}
			width, err := cmd.Flags().GetInt("width")
			if err != nil {
				return err
			}
			placement, err := cmd.Flags().GetString("placement")
			if err != nil {
				return err
			}
			separator, err := cmd.Flags().GetString("separator")
			if err != nil {
				return err
			}
			numberPrefix, err := cmd.Flags().GetString("number-prefix")
			if err != nil {
				return err
			}
			numberSuffix, err := cmd.Flags().GetString("number-suffix")
			if err != nil {
				return err
			}

			opts := sequence.DefaultOptions()
			opts.WorkingDir = scope.WorkingDir
			opts.IncludeDirectories = scope.IncludeDirectories
			opts.IncludeHidden = scope.IncludeHidden
			opts.Recursive = scope.Recursive
			opts.Extensions = append([]string(nil), scope.Extensions...)
			opts.DryRun = dryRun
			opts.AutoApply = autoApply
			if start != 0 {
				opts.Start = start
			}
			if cmd.Flags().Changed("width") {
				opts.Width = width
				opts.WidthSet = true
			}
			if placement != "" {
				opts.Placement = sequence.Placement(strings.ToLower(placement))
			}
			if separator != "" {
				opts.Separator = separator
			}
			opts.NumberPrefix = numberPrefix
			opts.NumberSuffix = numberSuffix

			out := cmd.OutOrStdout()

			plan, err := sequence.Preview(cmd.Context(), opts, out)
			if err != nil {
				return err
			}

			for _, candidate := range plan.Candidates {
				status := "UNCHANGED"
				switch candidate.Status {
				case sequence.CandidatePending:
					status = "CHANGE"
				case sequence.CandidateSkipped:
					status = "SKIP"
				}
				fmt.Fprintf(out, "%s: %s -> %s\n", status, candidate.OriginalPath, candidate.ProposedPath)
			}

			for _, conflict := range plan.SkippedConflicts {
				fmt.Fprintf(out, "Warning: %s skipped due to %s (target %s)\n", conflict.OriginalPath, conflict.Reason, conflict.ConflictingPath)
			}
			for _, warning := range plan.Summary.Warnings {
				fmt.Fprintf(out, "Warning: %s\n", warning)
			}

			if plan.Summary.TotalCandidates == 0 {
				fmt.Fprintln(out, "No candidates found.")
				return nil
			}

			fmt.Fprintf(out, "Preview: %d candidates, %d renames, %d skipped (width %d).\n",
				plan.Summary.TotalCandidates,
				plan.Summary.RenamedCount,
				plan.Summary.SkippedCount,
				plan.Summary.AppliedWidth,
			)

			if dryRun || !autoApply {
				if !autoApply {
					fmt.Fprintln(out, "Preview complete. Re-run with --yes to apply.")
				}
				return nil
			}

			entry, err := sequence.Apply(cmd.Context(), opts, plan)
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Applied %d sequence updates. Ledger updated.\n", len(entry.Operations))
			if plan.Summary.SkippedCount > 0 {
				fmt.Fprintf(out, "%d candidates were skipped due to conflicts.\n", plan.Summary.SkippedCount)
			}
			return nil
		},
	}

	cmd.Flags().Int("start", 1, "Starting sequence value (>=1)")
	cmd.Flags().Int("width", 0, "Minimum digit width for zero padding (defaults to 3 digits, auto-expands as needed)")
	cmd.Flags().String("placement", string(sequence.PlacementPrefix), "Placement for the sequence number: prefix or suffix")
	cmd.Flags().String("separator", "_", "Separator between the filename and sequence label and the original name")
	cmd.Flags().String("number-prefix", "", "Static text placed immediately before the sequence digits")
	cmd.Flags().String("number-suffix", "", "Static text placed immediately after the sequence digits")

	return cmd
}

func init() {
	rootCmd.AddCommand(newSequenceCommand())
}
