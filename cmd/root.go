/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
)

var rootCmd = &cobra.Command{
	Use:   "renamer",
	Short: "Safe, scriptable batch renaming utility",
	Long: `Renamer provides preview-first, undoable rename operations for files and directories.
Use subcommands like "preview", "rename", and "list" with shared scope flags to target exactly
the paths you intend to change.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	listing.RegisterScopeFlags(rootCmd.PersistentFlags())
}
