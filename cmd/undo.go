package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/history"
)

func newUndoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the most recent rename or replace batch",
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDir, err := resolveWorkingDir(cmd)
			if err != nil {
				return err
			}

			entry, err := history.Undo(workingDir)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Undo applied: %d operations reversed\n", len(entry.Operations))
			return nil
		},
	}

	return cmd
}

func resolveWorkingDir(cmd *cobra.Command) (string, error) {
	if flag := cmd.Flags().Lookup("path"); flag != nil {
		if value := flag.Value.String(); value != "" {
			return filepath.Abs(value)
		}
	}
	if flag := cmd.InheritedFlags().Lookup("path"); flag != nil {
		if value := flag.Value.String(); value != "" {
			return filepath.Abs(value)
		}
	}
	return os.Getwd()
}

func init() {
	rootCmd.AddCommand(newUndoCommand())
}
