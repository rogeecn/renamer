package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Undo applied: %d operations reversed\n", len(entry.Operations))

			if entry.Command == "extension" && entry.Metadata != nil {
				if target, ok := entry.Metadata["targetExtension"].(string); ok && target != "" {
					fmt.Fprintf(out, "Restored extensions to %s\n", target)
				}
				if sources, ok := entry.Metadata["sourceExtensions"].([]string); ok && len(sources) > 0 {
					fmt.Fprintf(out, "Previous sources: %s\n", strings.Join(sources, ", "))
				}
			}

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
