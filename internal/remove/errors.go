package remove

import "errors"

var (
	// ErrNoTokens indicates that no removal tokens were provided.
	ErrNoTokens = errors.New("at least one non-empty token is required")
)
