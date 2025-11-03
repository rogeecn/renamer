package plan

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogeecn/renamer/internal/ai/prompt"
	"github.com/rogeecn/renamer/internal/history"
)

// ApplyOptions describe the data required to apply an AI rename plan.
type ApplyOptions struct {
	WorkingDir string
	Candidates []Candidate
	Response   prompt.RenameResponse
	Policies   prompt.NamingPolicyConfig
	PromptHash string
}

// Apply executes the AI rename plan and records the outcome in the ledger.
func Apply(ctx context.Context, opts ApplyOptions) (history.Entry, error) {
	entry := history.Entry{Command: "ai"}

	if len(opts.Response.Items) == 0 {
		return entry, errors.New("ai apply: no items to apply")
	}

	candidateMap := make(map[string]Candidate, len(opts.Candidates))
	for _, cand := range opts.Candidates {
		key := strings.ToLower(strings.TrimSpace(cand.OriginalPath))
		candidateMap[key] = cand
	}

	type operation struct {
		sourceRel string
		targetRel string
		sourceAbs string
		targetAbs string
		depth     int
	}

	ops := make([]operation, 0, len(opts.Response.Items))
	seenTargets := make(map[string]string)

	conflicts := make([]Conflict, 0)

	for _, item := range opts.Response.Items {
		key := strings.ToLower(strings.TrimSpace(item.Original))
		cand, ok := candidateMap[key]
		if !ok {
			conflicts = append(conflicts, Conflict{
				OriginalPath: item.Original,
				Issue:        "missing_candidate",
				Details:      "original file not found in current scope",
			})
			continue
		}

		target := strings.TrimSpace(item.Proposed)
		if target == "" {
			conflicts = append(conflicts, Conflict{
				OriginalPath: item.Original,
				Issue:        "empty_target",
				Details:      "proposed name cannot be empty",
			})
			continue
		}

		normalizedTarget := filepath.ToSlash(filepath.Clean(target))
		if strings.HasPrefix(normalizedTarget, "../") {
			conflicts = append(conflicts, Conflict{
				OriginalPath: item.Original,
				Issue:        "unsafe_target",
				Details:      "proposed path escapes the working directory",
			})
			continue
		}

		targetKey := strings.ToLower(normalizedTarget)
		if existing, exists := seenTargets[targetKey]; exists && existing != item.Original {
			conflicts = append(conflicts, Conflict{
				OriginalPath: item.Original,
				Issue:        "duplicate_target",
				Details:      fmt.Sprintf("target %q reused", normalizedTarget),
			})
			continue
		}
		seenTargets[targetKey] = item.Original

		sourceRel := filepath.ToSlash(cand.OriginalPath)
		sourceAbs := filepath.Join(opts.WorkingDir, filepath.FromSlash(sourceRel))
		targetAbs := filepath.Join(opts.WorkingDir, filepath.FromSlash(normalizedTarget))

		if sameFile, err := isSameFile(sourceAbs, targetAbs); err != nil {
			return history.Entry{}, err
		} else if sameFile {
			continue
		}

		if _, err := os.Stat(targetAbs); err == nil {
			conflicts = append(conflicts, Conflict{
				OriginalPath: item.Original,
				Issue:        "target_exists",
				Details:      fmt.Sprintf("target %q already exists", normalizedTarget),
			})
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return history.Entry{}, err
		}

		op := operation{
			sourceRel: sourceRel,
			targetRel: normalizedTarget,
			sourceAbs: sourceAbs,
			targetAbs: targetAbs,
			depth:     cand.Depth,
		}
		ops = append(ops, op)
	}

	if len(conflicts) > 0 {
		return history.Entry{}, ApplyConflictError{Conflicts: conflicts}
	}

	if len(ops) == 0 {
		return entry, nil
	}

	sort.SliceStable(ops, func(i, j int) bool {
		return ops[i].depth > ops[j].depth
	})

	done := make([]history.Operation, 0, len(ops))

	revert := func() error {
		for i := len(done) - 1; i >= 0; i-- {
			op := done[i]
			src := filepath.Join(opts.WorkingDir, filepath.FromSlash(op.To))
			dst := filepath.Join(opts.WorkingDir, filepath.FromSlash(op.From))
			if err := os.Rename(src, dst); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		return nil
	}

	for _, op := range ops {
		if err := ctx.Err(); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		if dir := filepath.Dir(op.targetAbs); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				_ = revert()
				return history.Entry{}, err
			}
		}
		if err := os.Rename(op.sourceAbs, op.targetAbs); err != nil {
			_ = revert()
			return history.Entry{}, err
		}

		done = append(done, history.Operation{
			From: op.sourceRel,
			To:   op.targetRel,
		})
	}

	if len(done) == 0 {
		return entry, nil
	}

	entry.Operations = done

	aiMetadata := history.AIMetadata{
		PromptHash: opts.PromptHash,
		Model:      opts.Response.Model,
		Policies: prompt.NamingPolicyConfig{
			Prefix:            opts.Policies.Prefix,
			Casing:            opts.Policies.Casing,
			AllowSpaces:       opts.Policies.AllowSpaces,
			KeepOriginalOrder: opts.Policies.KeepOriginalOrder,
			ForbiddenTokens:   append([]string(nil), opts.Policies.ForbiddenTokens...),
		},
		BatchSize: len(done),
	}

	if hash, err := ResponseDigest(opts.Response); err == nil {
		aiMetadata.ResponseHash = hash
	}

	entry.AttachAIMetadata(aiMetadata)

	if err := history.Append(opts.WorkingDir, entry); err != nil {
		_ = revert()
		return history.Entry{}, err
	}

	return entry, nil
}

// ApplyConflictError signals that the plan contained conflicts that block apply.
type ApplyConflictError struct {
	Conflicts []Conflict
}

func (e ApplyConflictError) Error() string {
	if len(e.Conflicts) == 0 {
		return "ai apply: conflicts detected"
	}
	return fmt.Sprintf("ai apply: %d conflicts detected", len(e.Conflicts))
}

// ResponseDigest returns a hash of the AI response payload for ledger metadata.
func ResponseDigest(resp prompt.RenameResponse) (string, error) {
	data, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func isSameFile(a, b string) (bool, error) {
	infoA, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	infoB, err := os.Stat(b)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return os.SameFile(infoA, infoB), nil
}
