package insert

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ParseInputs validates the position token and insert text before preview/apply.
func ParseInputs(positionToken, insertText string, stemLength int) error {
	if positionToken == "" {
		return errors.New("position token cannot be empty")
	}
	if insertText == "" {
		return errors.New("insert text cannot be empty")
	}
	if !utf8.ValidString(insertText) {
		return errors.New("insert text must be valid UTF-8")
	}
	for _, r := range insertText {
		if r == '/' || r == '\\' {
			return errors.New("insert text must not contain path separators")
		}
		if r < 0x20 {
			return errors.New("insert text must not contain control characters")
		}
	}
	if stemLength >= 0 {
		if _, err := ResolvePosition(positionToken, stemLength); err != nil {
			return err
		}
	}
	return nil
}

// ResolvePositionWithValidation wraps ResolvePosition with explicit range checks.
func ResolvePositionWithValidation(positionToken string, stemLength int) (Position, error) {
	pos, err := ResolvePosition(positionToken, stemLength)
	if err != nil {
		return Position{}, err
	}
	if pos.Index < 0 || pos.Index > stemLength {
		return Position{}, fmt.Errorf("position %s out of range for %d-character stem", positionToken, stemLength)
	}
	return pos, nil
}
