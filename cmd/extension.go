package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/extension"
	"github.com/rogeecn/renamer/internal/listing"
)

// NewExtensionCommand constructs the extension CLI command; exported for testing.
func NewExtensionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "extension <source-ext...> <target-ext>",
		Short:        "Normalize multiple file extensions to a single target extension",
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			req := extension.NewRequest(scope)

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

			parsed, err := extension.ParseArgs(args)
			if err != nil {
				return fmt.Errorf("invalid extension arguments: %w", err)
			}

			req.SetExtensions(parsed.SourcesCanonical, parsed.SourcesDisplay, parsed.Target)
			req.SetWarnings(parsed.Duplicates, parsed.NoOps)

			summary, planned, err := extension.Preview(cmd.Context(), req, cmd.OutOrStdout())
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
					fmt.Fprintln(cmd.OutOrStdout(), "Nothing to apply; extensions already normalized.")
				}
				return nil
			}

			entry, err := extension.Apply(cmd.Context(), req, planned, summary)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Applied %d extension updates. Ledger updated.\n", len(entry.Operations))
			return nil
		},
	}

	cmd.Example = `  renamer extension .jpeg .JPG .jpg --dry-run
  renamer extension .yaml .yml .yml --yes --recursive`

	return cmd
}

func init() {
	rootCmd.AddCommand(NewExtensionCommand())
}
