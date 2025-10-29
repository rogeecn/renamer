package listing

import (
	"fmt"
	"os"

	"github.com/rogeecn/renamer/internal/filters"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagPath        = "path"
	flagRecursive   = "recursive"
	flagIncludeDirs = "include-dirs"
	flagHidden      = "hidden"
	flagExtensions  = "extensions"
)

// RegisterScopeFlags defines persistent flags that scope listing, preview, and rename operations.
func RegisterScopeFlags(flags *pflag.FlagSet) {
    flags.String(flagPath, "", "Directory to inspect (defaults to current working directory)")
    flags.BoolP(flagRecursive, "r", false, "Traverse subdirectories")
    flags.BoolP(flagIncludeDirs, "d", false, "Include directories in results")
    flags.Bool(flagHidden, false, "Include hidden files and directories")
    flags.StringP(flagExtensions, "e", "", "Pipe-delimited list of extensions to include (e.g. .jpg|.png)")
}

// ScopeFromCmd builds a ListingRequest populated from scope flags on the provided command.
func ScopeFromCmd(cmd *cobra.Command) (*ListingRequest, error) {
	path, err := getStringFlag(cmd, flagPath)
	if err != nil {
		return nil, err
	}
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = cwd
	}

	recursive, err := getBoolFlag(cmd, flagRecursive)
	if err != nil {
		return nil, err
	}

	includeDirs, err := getBoolFlag(cmd, flagIncludeDirs)
	if err != nil {
		return nil, err
	}

	includeHidden, err := getBoolFlag(cmd, flagHidden)
	if err != nil {
		return nil, err
	}

	extRaw, err := getStringFlag(cmd, flagExtensions)
	if err != nil {
		return nil, err
	}

	extensions, err := filters.ParseExtensions(extRaw)
	if err != nil {
		return nil, err
	}

	req := &ListingRequest{
		WorkingDir:         path,
		IncludeDirectories: includeDirs,
		Recursive:          recursive,
		IncludeHidden:      includeHidden,
		Extensions:         extensions,
		Format:             FormatTable,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func getStringFlag(cmd *cobra.Command, name string) (string, error) {
	if f := cmd.Flags().Lookup(name); f != nil {
		return cmd.Flags().GetString(name)
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil {
		return cmd.InheritedFlags().GetString(name)
	}
	return "", fmt.Errorf("flag %s not defined", name)
}

func getBoolFlag(cmd *cobra.Command, name string) (bool, error) {
	if f := cmd.Flags().Lookup(name); f != nil {
		return cmd.Flags().GetBool(name)
	}
	if f := cmd.InheritedFlags().Lookup(name); f != nil {
		return cmd.InheritedFlags().GetBool(name)
	}
	return false, fmt.Errorf("flag %s not defined", name)
}
