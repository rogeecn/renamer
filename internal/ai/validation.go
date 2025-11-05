package ai

import (
	"fmt"
	"path"
	"strings"

	"github.com/rogeecn/renamer/internal/ai/flow"
)

var invalidCharacters = []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}

// Conflict captures a validation failure for a proposed rename.
type Conflict struct {
	Original  string
	Suggested string
	Reason    string
}

// ValidationResult aggregates conflicts and warnings.
type ValidationResult struct {
	Conflicts []Conflict
	Warnings  []string
}

// ValidateSuggestions enforces rename safety rules before applying suggestions.
func ValidateSuggestions(expected []string, suggestions []flow.Suggestion) ValidationResult {
	result := ValidationResult{}

	expectedSet := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		expectedSet[strings.ToLower(flowToKey(name))] = struct{}{}
	}

	seenTargets := make(map[string]string)

	for _, suggestion := range suggestions {
		key := strings.ToLower(flowToKey(suggestion.Original))
		if _, ok := expectedSet[key]; !ok {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "original file not present in scope",
			})
			continue
		}

		cleaned := strings.TrimSpace(suggestion.Suggested)
		if cleaned == "" {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "suggested name is empty",
			})
			continue
		}

		normalizedOriginal := flowToKey(suggestion.Original)
		normalizedSuggested := flowToKey(cleaned)

		if strings.HasPrefix(normalizedSuggested, "/") {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "suggested name must be relative",
			})
			continue
		}

		if containsParentSegment(normalizedSuggested) {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "suggested name cannot traverse directories",
			})
			continue
		}

		base := path.Base(cleaned)
		if containsInvalidCharacter(base) {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "suggested name contains invalid characters",
			})
			continue
		}

		if !extensionsMatch(suggestion.Original, cleaned) {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "file extension changed",
			})
			continue
		}

		if path.Dir(normalizedOriginal) != path.Dir(normalizedSuggested) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("suggestion for %q moves file to a different directory", suggestion.Original))
		}

		targetKey := strings.ToLower(normalizedSuggested)
		if existing, ok := seenTargets[targetKey]; ok && existing != suggestion.Original {
			result.Conflicts = append(result.Conflicts, Conflict{
				Original:  suggestion.Original,
				Suggested: suggestion.Suggested,
				Reason:    "duplicate target generated",
			})
			continue
		}
		seenTargets[targetKey] = suggestion.Original

		if normalizedOriginal == normalizedSuggested {
			result.Warnings = append(result.Warnings, fmt.Sprintf("suggestion for %q does not change the filename", suggestion.Original))
		}
	}

	if len(suggestions) != len(expected) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("expected %d suggestions but received %d", len(expected), len(suggestions)))
	}

	return result
}

func flowToKey(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
}

func containsInvalidCharacter(value string) bool {
	for _, ch := range invalidCharacters {
		if strings.ContainsRune(value, ch) {
			return true
		}
	}
	return false
}

func extensionsMatch(original, proposed string) bool {
	origExt := strings.ToLower(path.Ext(original))
	propExt := strings.ToLower(path.Ext(proposed))
	return origExt == propExt
}

// SummarizeConflicts renders a human-readable summary of conflicts.
func SummarizeConflicts(conflicts []Conflict) string {
	if len(conflicts) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for _, c := range conflicts {
		builder.WriteString(fmt.Sprintf("%s -> %s (%s); ", c.Original, c.Suggested, c.Reason))
	}
	return strings.TrimSpace(builder.String())
}

// SummarizeWarnings renders warnings as a delimited string.
func SummarizeWarnings(warnings []string) string {
	return strings.Join(warnings, "; ")
}

func containsParentSegment(value string) bool {
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}
	return false
}
