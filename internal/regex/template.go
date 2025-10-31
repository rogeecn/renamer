package regex

import (
	"fmt"
	"strconv"
	"strings"
)

type templateSegment struct {
	literal string
	group   int
}

const literalSegment = -1

// template represents a parsed replacement template with capture placeholders.
type template struct {
	segments []templateSegment
}

// parseTemplate converts a string containing literal text, numbered placeholders (@0, @1, ...),
// and escaped @@ sequences into a template structure. It returns the template, the highest
// placeholder index encountered, or an error when syntax is invalid.
func parseTemplate(input string) (template, int, error) {
	segments := make([]templateSegment, 0)
	var literal strings.Builder
	maxGroup := 0

	i := 0
	for i < len(input) {
		ch := input[i]
		if ch != '@' {
			literal.WriteByte(ch)
			i++
			continue
		}

		// Flush any buffered literal before handling placeholder/escape.
		flushLiteral := func() {
			if literal.Len() == 0 {
				return
			}
			segments = append(segments, templateSegment{literal: literal.String(), group: literalSegment})
			literal.Reset()
		}

		if i+1 >= len(input) {
			return template{}, 0, fmt.Errorf("dangling @ at end of template")
		}

		next := input[i+1]
		if next == '@' {
			flushLiteral()
			literal.WriteByte('@')
			i += 2
			continue
		}

		j := i + 1
		for j < len(input) && input[j] >= '0' && input[j] <= '9' {
			j++
		}
		if j == i+1 {
			return template{}, 0, fmt.Errorf("invalid placeholder at offset %d", i)
		}

		indexStr := input[i+1 : j]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return template{}, 0, fmt.Errorf("invalid placeholder index @%s", indexStr)
		}

		flushLiteral()
		segments = append(segments, templateSegment{group: index})
		if index > maxGroup {
			maxGroup = index
		}

		i = j
	}

	if literal.Len() > 0 {
		segments = append(segments, templateSegment{literal: literal.String(), group: literalSegment})
	}

	return template{segments: segments}, maxGroup, nil
}

// render produces the output string for a given set of submatches. The slice must contain the
// full match at index 0 followed by capture groups. Missing groups (e.g., optional matches)
// expand to empty strings. Referencing a group index beyond the available matches returns an error.
func (t template) render(submatches []string) (string, error) {
	var builder strings.Builder

	for _, segment := range t.segments {
		if segment.group == literalSegment {
			builder.WriteString(segment.literal)
			continue
		}

		if segment.group >= len(submatches) {
			return "", ErrUndefinedPlaceholder{Index: segment.group}
		}

		builder.WriteString(submatches[segment.group])
	}

	return builder.String(), nil
}
