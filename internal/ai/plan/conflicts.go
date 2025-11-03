package plan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rogeecn/renamer/internal/ai/prompt"
)

// Conflict describes an issue detected in an AI rename plan.
type Conflict struct {
	OriginalPath string
	Issue        string
	Details      string
}

func detectConflicts(items []prompt.RenameItem) []Conflict {
	conflicts := make([]Conflict, 0)

	if len(items) == 0 {
		return conflicts
	}

	targets := make(map[string][]prompt.RenameItem)
	sequences := make([]int, 0, len(items))

	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.Proposed))
		if key != "" {
			targets[key] = append(targets[key], item)
		}
		if item.Sequence > 0 {
			sequences = append(sequences, item.Sequence)
		}
	}

	for _, entries := range targets {
		if len(entries) <= 1 {
			continue
		}
		for _, entry := range entries {
			conflicts = append(conflicts, Conflict{
				OriginalPath: entry.Original,
				Issue:        "duplicate_target",
				Details:      fmt.Sprintf("target %q is used by multiple entries", entries[0].Proposed),
			})
		}
	}

	if len(sequences) > 0 {
		sort.Ints(sequences)
		expected := 1
		for _, seq := range sequences {
			if seq != expected {
				conflicts = append(conflicts, Conflict{
					Issue:   "sequence_gap",
					Details: fmt.Sprintf("expected sequence %d but found %d", expected, seq),
				})
				expected = seq
			}
			expected++
		}
	}

	return conflicts
}
