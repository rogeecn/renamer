package extension

import "strings"

// NormalizeSourceExtensions returns case-insensitive unique source extensions while preserving
// the first-seen display token for each canonical value. Duplicate entries are surfaced for warnings.
func NormalizeSourceExtensions(inputs []string) (canonical []string, display []string, duplicates []string) {
	seen := make(map[string]struct{})

	for _, raw := range inputs {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		canon := CanonicalExtension(trimmed)
		if _, exists := seen[canon]; exists {
			duplicates = append(duplicates, trimmed)
			continue
		}

		seen[canon] = struct{}{}
		canonical = append(canonical, canon)
		display = append(display, trimmed)
	}

	return canonical, display, duplicates
}

// NormalizeTargetExtension trims surrounding whitespace but preserves caller-provided casing.
func NormalizeTargetExtension(target string) string {
	return strings.TrimSpace(target)
}

// CanonicalExtension is the shared case-folded representation used for comparisons and maps.
func CanonicalExtension(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// ExtensionsEqual reports true when two extensions match case-insensitively.
func ExtensionsEqual(a, b string) bool {
	return CanonicalExtension(a) == CanonicalExtension(b)
}

// IdentifyNoOpSources returns the original tokens that would be no-ops against the target extension.
func IdentifyNoOpSources(original []string, target string) []string {
	if len(original) == 0 {
		return nil
	}

	canonTarget := CanonicalExtension(target)
	var noOps []string
	for _, token := range original {
		if CanonicalExtension(token) == canonTarget {
			noOps = append(noOps, token)
		}
	}
	return noOps
}
