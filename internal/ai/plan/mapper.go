package plan

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Candidate represents a file considered for AI renaming.
type Candidate struct {
	OriginalPath string
	SizeBytes    int64
	Depth        int
	Extension    string
}

// MapInput configures the mapping behaviour.
type MapInput struct {
	Candidates    []Candidate
	SequenceWidth int
}

// PreviewPlan aggregates entries ready for preview rendering.
type PreviewPlan struct {
	Entries    []PreviewEntry
	Warnings   []string
	PromptHash string
	Model      string
	Conflicts  []Conflict
}

// PreviewEntry is a single row in the preview table.
type PreviewEntry struct {
	Sequence          int
	SequenceLabel     string
	OriginalPath      string
	ProposedPath      string
	SanitizedSegments []string
	Notes             string
}

// MapResponse converts a validated response into a preview plan.
func MapResponse(input MapInput, validation ValidationResult) (PreviewPlan, error) {
	if input.SequenceWidth <= 0 {
		input.SequenceWidth = 3
	}

	itemByOriginal := make(map[string]struct {
		item promptRenameItem
	}, len(validation.Items))
	for _, item := range validation.Items {
		key := normalizePath(item.Original)
		itemByOriginal[key] = struct{ item promptRenameItem }{item: promptRenameItem{
			Original: item.Original,
			Proposed: item.Proposed,
			Sequence: item.Sequence,
			Notes:    item.Notes,
		}}
	}

	entries := make([]PreviewEntry, 0, len(input.Candidates))
	for _, candidate := range input.Candidates {
		key := normalizePath(candidate.OriginalPath)
		entryData, ok := itemByOriginal[key]
		if !ok {
			return PreviewPlan{}, fmt.Errorf("ai plan: missing response for %s", candidate.OriginalPath)
		}

		item := entryData.item
		label := formatSequence(item.Sequence, input.SequenceWidth)
		sanitized := computeSanitizedSegments(candidate.OriginalPath, item.Proposed)

		entries = append(entries, PreviewEntry{
			Sequence:          item.Sequence,
			SequenceLabel:     label,
			OriginalPath:      candidate.OriginalPath,
			ProposedPath:      item.Proposed,
			SanitizedSegments: sanitized,
			Notes:             item.Notes,
		})
	}

	return PreviewPlan{
		Entries:    entries,
		Warnings:   append([]string(nil), validation.Warnings...),
		PromptHash: validation.PromptHash,
		Model:      validation.Model,
		Conflicts:  detectConflicts(validation.Items),
	}, nil
}

type promptRenameItem struct {
	Original string
	Proposed string
	Sequence int
	Notes    string
}

func formatSequence(seq, width int) string {
	if seq <= 0 {
		return ""
	}
	label := fmt.Sprintf("%0*d", width, seq)
	if len(label) < len(fmt.Sprintf("%d", seq)) {
		return fmt.Sprintf("%d", seq)
	}
	return label
}

func normalizePath(path string) string {
	return strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
}

func computeSanitizedSegments(original, proposed string) []string {
	origStem := stem(original)
	propStem := stem(proposed)

	origTokens := tokenize(origStem)
	propTokens := make(map[string]struct{}, len(origTokens))
	for _, token := range tokenize(propStem) {
		propTokens[token] = struct{}{}
	}

	var sanitized []string
	seen := make(map[string]struct{})
	for _, token := range origTokens {
		if _, ok := propTokens[token]; ok {
			continue
		}
		if _, already := seen[token]; already {
			continue
		}
		if isNumericToken(token) {
			continue
		}
		seen[token] = struct{}{}
		sanitized = append(sanitized, token)
	}
	if len(sanitized) == 0 {
		return nil
	}
	sort.Strings(sanitized)
	return sanitized
}

func stem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		return base[:len(base)-len(ext)]
	}
	return base
}

func tokenize(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		if r >= '0' && r <= '9' {
			return false
		}
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			return false
		}
		return true
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		normalized := strings.ToLower(field)
		if normalized == "" {
			continue
		}
		tokens = append(tokens, normalized)
	}
	return tokens
}

func isNumericToken(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
