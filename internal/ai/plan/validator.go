package plan

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/rogeecn/renamer/internal/ai/prompt"
)

// Validator checks the AI response for completeness and uniqueness rules.
type Validator struct {
	expected    []string
	expectedSet map[string]struct{}
	policies    prompt.NamingPolicyConfig
	bannedSet   map[string]struct{}
}

// ValidationResult captures the successfully decoded response data.
type ValidationResult struct {
	Items      []prompt.RenameItem
	Warnings   []string
	PromptHash string
	Model      string
}

// InvalidItem describes a single response entry that failed validation.
type InvalidItem struct {
	Index    int
	Original string
	Proposed string
	Reason   string
}

// ValidationError aggregates the issues discovered during validation.
type ValidationError struct {
	Result              ValidationResult
	MissingOriginals    []string
	UnexpectedOriginals []string
	DuplicateOriginals  map[string]int
	DuplicateProposed   map[string][]string
	InvalidItems        []InvalidItem
	PolicyViolations    []PolicyViolation
}

// PolicyViolation captures a single naming-policy breach.
type PolicyViolation struct {
	Original string
	Proposed string
	Rule     string
	Message  string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return ""
	}

	parts := make([]string, 0, 5)
	if len(e.MissingOriginals) > 0 {
		parts = append(parts, fmt.Sprintf("missing %d originals", len(e.MissingOriginals)))
	}
	if len(e.UnexpectedOriginals) > 0 {
		parts = append(parts, fmt.Sprintf("unexpected %d originals", len(e.UnexpectedOriginals)))
	}
	if len(e.DuplicateOriginals) > 0 {
		parts = append(parts, fmt.Sprintf("%d duplicate originals", len(e.DuplicateOriginals)))
	}
	if len(e.DuplicateProposed) > 0 {
		parts = append(parts, fmt.Sprintf("%d duplicate proposed names", len(e.DuplicateProposed)))
	}
	if len(e.InvalidItems) > 0 {
		parts = append(parts, fmt.Sprintf("%d invalid items", len(e.InvalidItems)))
	}
	if len(e.PolicyViolations) > 0 {
		parts = append(parts, fmt.Sprintf("%d policy violations", len(e.PolicyViolations)))
	}

	summary := strings.Join(parts, ", ")
	if summary == "" {
		summary = "response validation failed"
	}
	return fmt.Sprintf("ai response validation failed: %s", summary)
}

// HasIssues indicates whether the validation error captured any rule breaks.
func (e *ValidationError) HasIssues() bool {
	if e == nil {
		return false
	}
	return len(e.MissingOriginals) > 0 ||
		len(e.UnexpectedOriginals) > 0 ||
		len(e.DuplicateOriginals) > 0 ||
		len(e.DuplicateProposed) > 0 ||
		len(e.InvalidItems) > 0 ||
		len(e.PolicyViolations) > 0
}

// NewValidator constructs a validator for the supplied original filenames. Any
// whitespace-only entries are discarded. Duplicate originals are collapsed to
// ensure consistent coverage checks.
func NewValidator(originals []string, policies prompt.NamingPolicyConfig, bannedTerms []string) Validator {
	expectedSet := make(map[string]struct{}, len(originals))
	deduped := make([]string, 0, len(originals))
	for _, original := range originals {
		trimmed := strings.TrimSpace(original)
		if trimmed == "" {
			continue
		}
		if _, exists := expectedSet[trimmed]; exists {
			continue
		}
		expectedSet[trimmed] = struct{}{}
		deduped = append(deduped, trimmed)
	}

	bannedSet := make(map[string]struct{})
	for _, token := range bannedTerms {
		lower := strings.ToLower(strings.TrimSpace(token))
		if lower == "" {
			continue
		}
		bannedSet[lower] = struct{}{}
	}

	policies.Casing = strings.ToLower(strings.TrimSpace(policies.Casing))
	policies.Prefix = strings.TrimSpace(policies.Prefix)
	policies.ForbiddenTokens = append([]string(nil), policies.ForbiddenTokens...)

	return Validator{
		expected:    deduped,
		expectedSet: expectedSet,
		policies:    policies,
		bannedSet:   bannedSet,
	}
}

// Validate ensures the AI response covers each expected original exactly once
// and that the proposed filenames are unique.
func (v Validator) Validate(resp prompt.RenameResponse) (ValidationResult, error) {
	result := ValidationResult{
		Items:      cloneItems(resp.Items),
		Warnings:   append([]string(nil), resp.Warnings...),
		PromptHash: resp.PromptHash,
		Model:      resp.Model,
	}

	if len(resp.Items) == 0 {
		err := &ValidationError{
			Result:           result,
			MissingOriginals: append([]string(nil), v.expected...),
		}
		return result, err
	}

	seenOriginals := make(map[string]int, len(resp.Items))
	seenProposed := make(map[string][]string, len(resp.Items))
	unexpectedSet := map[string]struct{}{}

	invalidItems := make([]InvalidItem, 0)
	policyViolations := make([]PolicyViolation, 0)

	for idx, item := range resp.Items {
		original := strings.TrimSpace(item.Original)
		proposed := strings.TrimSpace(item.Proposed)

		if original == "" {
			invalidItems = append(invalidItems, InvalidItem{
				Index:    idx,
				Original: item.Original,
				Proposed: item.Proposed,
				Reason:   "original is empty",
			})
		} else {
			seenOriginals[original]++
			if _, ok := v.expectedSet[original]; !ok {
				unexpectedSet[original] = struct{}{}
			}
		}

		if proposed == "" {
			invalidItems = append(invalidItems, InvalidItem{
				Index:    idx,
				Original: item.Original,
				Proposed: item.Proposed,
				Reason:   "proposed is empty",
			})
		} else {
			seenProposed[proposed] = append(seenProposed[proposed], original)
		}

		policyViolations = append(policyViolations, v.evaluatePolicies(item)...)
	}

	missing := make([]string, 0)
	for _, original := range v.expected {
		if seenOriginals[original] == 0 {
			missing = append(missing, original)
		}
	}

	duplicateOriginals := make(map[string]int)
	for original, count := range seenOriginals {
		if count > 1 {
			duplicateOriginals[original] = count
		}
	}

	duplicateProposed := make(map[string][]string)
	for proposed, sources := range seenProposed {
		if len(sources) > 1 {
			filtered := make([]string, 0, len(sources))
			for _, src := range sources {
				if strings.TrimSpace(src) != "" {
					filtered = append(filtered, src)
				}
			}
			if len(filtered) > 1 {
				duplicateProposed[proposed] = filtered
			}
		}
	}

	unexpected := orderedKeys(unexpectedSet)

	if len(missing) == 0 &&
		len(unexpected) == 0 &&
		len(duplicateOriginals) == 0 &&
		len(duplicateProposed) == 0 &&
		len(invalidItems) == 0 &&
		len(policyViolations) == 0 {
		return result, nil
	}

	err := &ValidationError{
		Result:              result,
		MissingOriginals:    missing,
		UnexpectedOriginals: unexpected,
		DuplicateOriginals:  duplicateOriginals,
		DuplicateProposed:   duplicateProposed,
		InvalidItems:        invalidItems,
		PolicyViolations:    policyViolations,
	}

	return result, err
}

// Expectation returns a copy of the expected originals tracked by the validator.
func (v Validator) Expectation() []string {
	return append([]string(nil), v.expected...)
}

func cloneItems(items []prompt.RenameItem) []prompt.RenameItem {
	if len(items) == 0 {
		return nil
	}
	cp := make([]prompt.RenameItem, len(items))
	copy(cp, items)
	return cp
}

func orderedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (v Validator) evaluatePolicies(item prompt.RenameItem) []PolicyViolation {
	violations := make([]PolicyViolation, 0)
	proposed := strings.TrimSpace(item.Proposed)
	if proposed == "" {
		return violations
	}
	base := filepath.Base(proposed)
	stem := base
	if ext := filepath.Ext(base); ext != "" {
		stem = base[:len(base)-len(ext)]
	}
	stemLower := strings.ToLower(stem)

	if v.policies.Prefix != "" {
		prefixLower := strings.ToLower(v.policies.Prefix)
		if !strings.HasPrefix(stemLower, prefixLower) {
			violations = append(violations, PolicyViolation{
				Original: item.Original,
				Proposed: item.Proposed,
				Rule:     "prefix",
				Message:  fmt.Sprintf("expected prefix %q", v.policies.Prefix),
			})
		}
	}

	if !v.policies.AllowSpaces && strings.Contains(stem, " ") {
		violations = append(violations, PolicyViolation{
			Original: item.Original,
			Proposed: item.Proposed,
			Rule:     "spaces",
			Message:  "spaces are not allowed",
		})
	}

	if v.policies.Casing != "" {
		if ok, message := matchesCasing(stem, v.policies); !ok {
			violations = append(violations, PolicyViolation{
				Original: item.Original,
				Proposed: item.Proposed,
				Rule:     "casing",
				Message:  message,
			})
		}
	}

	if len(v.bannedSet) > 0 {
		tokens := tokenize(stemLower)
		for _, token := range tokens {
			if _, ok := v.bannedSet[token]; ok {
				violations = append(violations, PolicyViolation{
					Original: item.Original,
					Proposed: item.Proposed,
					Rule:     "banned",
					Message:  fmt.Sprintf("contains banned token %q", token),
				})
				break
			}
		}
	}

	return violations
}

func matchesCasing(stem string, policies prompt.NamingPolicyConfig) (bool, string) {
	core := coreStem(stem, policies.Prefix)
	switch policies.Casing {
	case "kebab":
		if strings.Contains(core, " ") {
			return false, "expected kebab-case (no spaces)"
		}
		if strings.ContainsAny(core, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			return false, "expected kebab-case (use lowercase letters)"
		}
		return true, ""
	case "snake":
		if strings.Contains(core, " ") {
			return false, "expected snake_case (no spaces)"
		}
		if strings.ContainsAny(core, "ABCDEFGHIJKLMNOPQRSTUVWXYZ-") {
			return false, "expected snake_case (lowercase letters with underscores)"
		}
		return true, ""
	case "camel":
		if strings.ContainsAny(core, " -_") {
			return false, "expected camelCase (no separators)"
		}
		runes := []rune(core)
		if len(runes) == 0 {
			return false, "expected camelCase descriptive text"
		}
		if !unicode.IsLower(runes[0]) {
			return false, "expected camelCase (first letter lowercase)"
		}
		return true, ""
	case "pascal":
		if strings.ContainsAny(core, " -_") {
			return false, "expected PascalCase (no separators)"
		}
		runes := []rune(core)
		if len(runes) == 0 {
			return false, "expected PascalCase descriptive text"
		}
		if !unicode.IsUpper(runes[0]) {
			return false, "expected PascalCase (first letter uppercase)"
		}
		return true, ""
	case "title":
		words := strings.Fields(strings.ReplaceAll(core, "-", " "))
		if len(words) == 0 {
			return false, "expected Title Case words"
		}
		for _, word := range words {
			runes := []rune(word)
			if len(runes) == 0 {
				continue
			}
			if !unicode.IsUpper(runes[0]) {
				return false, "expected Title Case (capitalize each word)"
			}
		}
		return true, ""
	default:
		return true, ""
	}
}

func coreStem(stem, prefix string) string {
	trimmed := stem
	if prefix != "" {
		lowerStem := strings.ToLower(trimmed)
		lowerPrefix := strings.ToLower(prefix)
		if strings.HasPrefix(lowerStem, lowerPrefix) {
			trimmed = trimmed[len(prefix):]
			trimmed = strings.TrimLeft(trimmed, "-_ ")
		}
	}
	i := 0
	runes := []rune(trimmed)
	for i < len(runes) {
		r := runes[i]
		if unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' {
			i++
			continue
		}
		break
	}
	return string(runes[i:])
}
