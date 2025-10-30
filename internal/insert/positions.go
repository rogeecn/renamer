package insert

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// Position represents a resolved rune index relative to the filename stem.
type Position struct {
	Index int // zero-based index where insertion should occur
}

// ResolvePosition interprets a position token (`^`, `$`, positive, negative) against the stem length.
func ResolvePosition(token string, stemLength int) (Position, error) {
	switch token {
	case "^":
		return Position{Index: 0}, nil
	case "$":
		return Position{Index: stemLength}, nil
	}

	if token == "" {
		return Position{}, errors.New("position token cannot be empty")
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
