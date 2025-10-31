package regex

import (
	"fmt"
	"regexp"
)

// Engine encapsulates a compiled regex pattern and parsed template for reuse across candidates.
type Engine struct {
	re     *regexp.Regexp
	tmpl   template
	groups int
}

// NewEngine compiles the regex pattern and template into a reusable Engine instance.
func NewEngine(pattern, tmpl string) (*Engine, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	parsed, maxGroup, err := parseTemplate(tmpl)
	if err != nil {
		return nil, err
	}

	if maxGroup > re.NumSubexp() {
		return nil, ErrTemplateGroupOutOfRange{Group: maxGroup, Available: re.NumSubexp()}
	}

	return &Engine{
		re:     re,
		tmpl:   parsed,
		groups: re.NumSubexp(),
	}, nil
}

// Apply evaluates the regex against input and renders the replacement when it matches.
// When no match occurs, matched is false without error.
func (e *Engine) Apply(input string) (output string, matchGroups []string, matched bool, err error) {
	submatches := e.re.FindStringSubmatch(input)
	if submatches == nil {
		return "", nil, false, nil
	}

	rendered, err := e.tmpl.render(submatches)
	if err != nil {
		return "", nil, false, err
	}

	// Exclude the full match from the recorded match group slice.
	groups := make([]string, 0, len(submatches)-1)
	if len(submatches) > 1 {
		groups = append(groups, submatches[1:]...)
	}

	return rendered, groups, true, nil
}

// ErrTemplateGroupOutOfRange indicates that the template references a capture group that the regex
// does not provide.
type ErrTemplateGroupOutOfRange struct {
	Group     int
	Available int
}

func (e ErrTemplateGroupOutOfRange) Error() string {
	return fmt.Sprintf("template references @%d but pattern only defines %d groups", e.Group, e.Available)
}
