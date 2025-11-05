package flow

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Suggestion represents a single rename mapping emitted by the Genkit flow.
type Suggestion struct {
	Original  string `json:"original"`
	Suggested string `json:"suggested"`
}

// Output wraps the list of suggestions returned by the flow.
type Output struct {
	Suggestions []Suggestion `json:"suggestions"`
}

var (
	errEmptyResponse      = errors.New("genkit flow returned empty response")
	errMissingSuggestions = errors.New("genkit flow response missing suggestions")
)

// ParseOutput converts the raw JSON payload into a structured Output.
func ParseOutput(raw []byte) (Output, error) {
	if len(raw) == 0 {
		return Output{}, errEmptyResponse
	}

	var out Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return Output{}, fmt.Errorf("failed to decode genkit output: %w", err)
	}

	if len(out.Suggestions) == 0 {
		return Output{}, errMissingSuggestions
	}

	return out, nil
}

// MarshalInput serialises the flow input for logging or replay.
func MarshalInput(input any) ([]byte, error) {
	buf, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to encode genkit input: %w", err)
	}
	return buf, nil
}
