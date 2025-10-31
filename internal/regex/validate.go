package regex

import "fmt"

// ValidateTemplate ensures the parsed template does not reference capture groups beyond the
// pattern's capabilities and returns a descriptive error for CLI presentation.
func ValidateTemplate(engine *Engine, tmpl template) error {
	if engine == nil {
		return fmt.Errorf("internal error: regex engine not initialized")
	}
	max := 0
	for _, segment := range tmpl.segments {
		if segment.group > max {
			max = segment.group
		}
	}
	if max > engine.groups {
		return ErrTemplateGroupOutOfRange{Group: max, Available: engine.groups}
	}
	return nil
}

// ErrUndefinedPlaceholder indicates that the template references a group with no match result.
type ErrUndefinedPlaceholder struct {
	Index int
}

func (e ErrUndefinedPlaceholder) Error() string {
	return fmt.Sprintf("template references @%d but the pattern did not produce that group", e.Index)
}
