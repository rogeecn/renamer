package insert

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Position represents a resolved rune index relative to the filename stem.
type Position struct {
	Index int // zero-based index where insertion should occur
}

// ResolvePosition interprets a position token (`^`, `$`, forward indexes, suffix offsets like `N$`, or legacy negative values) against the stem length.
func ResolvePosition(token string, stemLength int) (Position, error) {
	if token == "" {
		return Position{}, errors.New("position token cannot be empty")
	}

	switch token {
	case "^":
		return Position{Index: 0}, nil
	case "$":
		return Position{Index: stemLength}, nil
	}

	if strings.HasPrefix(token, "^") {
		trimmed := token[1:]
		if trimmed == "" {
			return Position{Index: 0}, nil
		}
		value, err := parseInt(trimmed)
		if err != nil {
			return Position{}, fmt.Errorf("invalid position token %q: %w", token, err)
		}
		if value < 0 {
			return Position{}, fmt.Errorf("invalid position token %q: cannot be negative", token)
		}
		if value > stemLength {
			return Position{}, fmt.Errorf("position %d out of range for %d-character stem", value, stemLength)
		}
		return Position{Index: value}, nil
	}

	if strings.HasSuffix(token, "$") {
		trimmed := token[:len(token)-1]
		if trimmed == "" {
			return Position{Index: stemLength}, nil
		}
		value, err := parseInt(trimmed)
		if err != nil {
			return Position{}, fmt.Errorf("invalid position token %q: %w", token, err)
		}
		if value < 0 {
			return Position{}, fmt.Errorf("invalid position token %q: cannot be negative", token)
		}
		if value > stemLength {
			return Position{}, fmt.Errorf("position %d out of range for %d-character stem", value, stemLength)
		}
		return Position{Index: stemLength - value}, nil
	}

	// Try parsing as integer (positive or negative).
	idx, err := parseInt(token)
	if err != nil {
		return Position{}, fmt.Errorf("invalid position token %q: %w", token, err)
	}

	if idx > 0 {
		if idx > stemLength {
			return Position{}, fmt.Errorf("position %d out of range for %d-character stem", idx, stemLength)
		}
		return Position{Index: idx}, nil
	}

	// Negative index counts backward from end (e.g., -1 inserts before last rune).
	offset := stemLength + idx
	if offset < 0 || offset > stemLength {
		return Position{}, fmt.Errorf("position %d out of range for %d-character stem", idx, stemLength)
	}
	return Position{Index: offset}, nil
}

// CountRunes returns the number of Unicode code points in name.
func CountRunes(name string) int {
	return utf8.RuneCountInString(name)
}

// parseInt is declared separately for test stubbing.
var parseInt = func(token string) (int, error) {
	var sign int = 1
	switch token[0] {
	case '+':
		token = token[1:]
	case '-':
		sign = -1
		token = token[1:]
	}
	if token == "" {
		return 0, errors.New("missing digits")
	}
	var value int
	for _, r := range token {
		if r < '0' || r > '9' {
			return 0, errors.New("non-numeric character in token")
		}
		value = value*10 + int(r-'0')
	}
	return sign * value, nil
}
