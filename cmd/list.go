package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/output"
)

func newListCommand() *cobra.Command {
	var (
		format   string
		maxDepth int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display rename candidates matched by the active filters",
		Long:  "Enumerate files and directories using the same filters that the rename command will honor.",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			req.Format = format
			req.MaxDepth = maxDepth

			formatter, err := output.NewFormatter(req.Format)
			if err != nil {
				return err
			}

			service := listing.NewService()
			summary, err := service.List(cmd.Context(), req, formatter, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			if summary.Total() == 0 {
				_, err = fmt.Fprintln(cmd.OutOrStdout(), listing.EmptyResultMessage(req))
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", listing.FormatTable, "Output format: table or plain")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum recursion depth (0 = unlimited)")

	return cmd
}

func init() {
	rootCmd.AddCommand(newListCommand())
}
