package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/ai"
	"github.com/rogeecn/renamer/internal/listing"
	"github.com/rogeecn/renamer/internal/traversal"
)

const maxAIFileCount = 200

func newAICommand() *cobra.Command {
	var prompt string
	var sequenceSeparator string

	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Generate AI-assisted rename suggestions",
		Long:  "Preview rename suggestions proposed by the integrated AI assistant before applying changes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := listing.ScopeFromCmd(cmd)
			if err != nil {
				return err
			}

			autoApply, err := lookupBool(cmd, "yes")
			if err != nil {
				return err
			}
			dryRun, err := lookupBool(cmd, "dry-run")
			if err != nil {
				return err
			}
			if dryRun && autoApply {
				return errors.New("--dry-run cannot be combined with --yes; remove one of them")
			}

			files, err := collectScopeEntries(cmd.Context(), scope)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No files matched the current scope.")
				return nil
			}

			if len(files) > maxAIFileCount {
				return fmt.Errorf("scope contains %d files; reduce to %d or fewer before running ai preview", len(files), maxAIFileCount)
			}

			if sequenceSeparator == "" {
				sequenceSeparator = "."
			}

			client := ai.NewClient()
			session := ai.NewSession(files, prompt, sequenceSeparator, client)

			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			for {
				output, validation, err := session.Generate(cmd.Context())
				if err != nil {
					return err
				}

				if err := ai.PrintPreview(out, output.Suggestions, validation); err != nil {
					return err
				}

				printSessionSummary(out, session)

				if len(validation.Conflicts) > 0 {
					fmt.Fprintln(out, "Conflicts detected. Adjust guidance or scope before proceeding.")
				}

				if autoApply {
					if len(validation.Conflicts) > 0 {
						return errors.New("preview contains conflicts; refine the prompt or scope before using --yes")
					}
					session.RecordAcceptance()
					entry, err := ai.Apply(cmd.Context(), scope.WorkingDir, output.Suggestions, validation, ai.ApplyMetadata{
						Prompt:            session.CurrentPrompt(),
						PromptHistory:     session.PromptHistory(),
						Notes:             session.Notes(),
						Model:             session.Model(),
						SequenceSeparator: session.SequenceSeparator(),
					}, out)
					if err != nil {
						return err
					}
					fmt.Fprintf(out, "Applied %d rename(s). Ledger updated.\n", len(entry.Operations))
					return nil
				}

				action, err := readSessionAction(reader, out, len(validation.Conflicts) == 0)
				if err != nil {
					return err
				}

				switch action {
				case actionQuit:
					fmt.Fprintln(out, "Session ended without applying changes.")
					return nil
				case actionAccept:
					if len(validation.Conflicts) > 0 {
						fmt.Fprintln(out, "Cannot accept preview while conflicts remain. Resolve them first.")
						continue
					}
					if dryRun {
						fmt.Fprintln(out, "Dry-run mode active; no changes were applied.")
						return nil
					}
					applyNow, err := confirmApply(reader, out)
					if err != nil {
						return err
					}
					if !applyNow {
						fmt.Fprintln(out, "Preview accepted without applying changes.")
						return nil
					}
					session.RecordAcceptance()
					entry, err := ai.Apply(cmd.Context(), scope.WorkingDir, output.Suggestions, validation, ai.ApplyMetadata{
						Prompt:            session.CurrentPrompt(),
						PromptHistory:     session.PromptHistory(),
						Notes:             session.Notes(),
						Model:             session.Model(),
						SequenceSeparator: session.SequenceSeparator(),
					}, out)
					if err != nil {
						return err
					}
					fmt.Fprintf(out, "Applied %d rename(s). Ledger updated.\n", len(entry.Operations))
					return nil
				case actionRegenerate:
					session.RecordRegeneration()
					continue
				case actionEdit:
					newPrompt, err := readPrompt(reader, out, session.CurrentPrompt())
					if err != nil {
						return err
					}
					session.UpdatePrompt(newPrompt)
					continue
				}
			}
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Optional guidance for the AI suggestion engine")
	cmd.Flags().StringVar(&sequenceSeparator, "sequence-separator", ".", "Separator inserted between sequence number and generated name")

	return cmd
}

func collectScopeEntries(ctx context.Context, req *listing.ListingRequest) ([]string, error) {
	walker := traversal.NewWalker()
	extensions := make(map[string]struct{}, len(req.Extensions))
	for _, ext := range req.Extensions {
		extensions[strings.ToLower(ext)] = struct{}{}
	}

	var files []string
	err := walker.Walk(req.WorkingDir, req.Recursive, req.IncludeDirectories, req.IncludeHidden, req.MaxDepth, func(relPath string, entry fs.DirEntry, depth int) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.IsDir() {
			if !req.IncludeDirectories {
				return nil
			}
		} else {
			if len(extensions) > 0 {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if _, ok := extensions[ext]; !ok {
					return nil
				}
			}
		}

		relSlash := filepath.ToSlash(relPath)
		if relSlash == "." {
			return nil
		}
		files = append(files, relSlash)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

const (
	actionAccept     = "accept"
	actionRegenerate = "regenerate"
	actionEdit       = "edit"
	actionQuit       = "quit"
)

func readSessionAction(reader *bufio.Reader, out io.Writer, canAccept bool) (string, error) {
	prompt := "Choose action: [Enter] finish, (e) edit prompt, (r) regenerate, (q) quit: "
	if !canAccept {
		prompt = "Choose action: (e) edit prompt, (r) regenerate, (q) quit: "
	}
	fmt.Fprint(out, prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	choice := strings.TrimSpace(strings.ToLower(line))

	if choice == "" {
		if !canAccept {
			return actionRegenerate, nil
		}
		return actionAccept, nil
	}

	switch choice {
	case "e", "edit":
		return actionEdit, nil
	case "r", "regen", "regenerate":
		return actionRegenerate, nil
	case "q", "quit", "exit":
		return actionQuit, nil
	case "accept", "a":
		if canAccept {
			return actionAccept, nil
		}
	}

	fmt.Fprintln(out, "Unrecognised choice; please try again.")
	return readSessionAction(reader, out, canAccept)
}

func readPrompt(reader *bufio.Reader, out io.Writer, current string) (string, error) {
	fmt.Fprintf(out, "Enter new prompt (leave blank to keep %q): ", current)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return current, nil
	}
	return trimmed, nil
}

func printSessionSummary(w io.Writer, session *ai.Session) {
	history := session.PromptHistory()
	fmt.Fprintf(w, "Current prompt: %q\n", session.CurrentPrompt())
	if len(history) > 1 {
		fmt.Fprintf(w, "Prompt history (%d entries): %s\n", len(history), strings.Join(history, " -> "))
	}
	if notes := session.Notes(); len(notes) > 0 {
		fmt.Fprintf(w, "Notes: %s\n", strings.Join(notes, "; "))
	}
}

func confirmApply(reader *bufio.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Apply these changes now? (y/N): ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	choice := strings.TrimSpace(strings.ToLower(line))
	switch choice {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func lookupBool(cmd *cobra.Command, name string) (bool, error) {
	if flag := cmd.Flags().Lookup(name); flag != nil {
		return cmd.Flags().GetBool(name)
	}
	if flag := cmd.InheritedFlags().Lookup(name); flag != nil {
		return cmd.InheritedFlags().GetBool(name)
	}
	return false, nil
}

func init() {
	rootCmd.AddCommand(newAICommand())
}
