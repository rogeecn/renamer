package ai

import (
	"context"
	"strings"

	flowpkg "github.com/rogeecn/renamer/internal/ai/flow"
)

const defaultModelID = "googleai/gemini-1.5-flash"

// Session tracks prompt history and guidance notes for a single AI preview loop.
type Session struct {
	files             []string
	client            *Client
	prompts           []string
	notes             []string
	model             string
	sequenceSeparator string

	lastOutput     *flowpkg.Output
	lastValidation ValidationResult
}

// NewSession builds a session with the provided scope, initial prompt, and client.
func NewSession(files []string, initialPrompt string, sequenceSeparator string, client *Client) *Session {
	prompts := []string{strings.TrimSpace(initialPrompt)}
	if prompts[0] == "" {
		prompts[0] = ""
	}

	if client == nil {
		client = NewClient()
	}

	sep := strings.TrimSpace(sequenceSeparator)
	if sep == "" {
		sep = "."
	}

	return &Session{
		files:             append([]string(nil), files...),
		client:            client,
		prompts:           prompts,
		notes:             make([]string, 0),
		model:             defaultModelID,
		sequenceSeparator: sep,
	}
}

// Generate executes the flow and returns structured suggestions with validation.
func (s *Session) Generate(ctx context.Context) (*flowpkg.Output, ValidationResult, error) {
	prompt := s.CurrentPrompt()
	input := &flowpkg.RenameFlowInput{
		FileNames:         append([]string(nil), s.files...),
		UserPrompt:        prompt,
		SequenceSeparator: s.sequenceSeparator,
	}

	output, err := s.client.Suggest(ctx, input)
	if err != nil {
		return nil, ValidationResult{}, err
	}

	validation := ValidateSuggestions(s.files, output.Suggestions)
	s.lastOutput = output
	s.lastValidation = validation
	return output, validation, nil
}

// CurrentPrompt returns the most recent prompt in the session.
func (s *Session) CurrentPrompt() string {
	if len(s.prompts) == 0 {
		return ""
	}
	return s.prompts[len(s.prompts)-1]
}

// UpdatePrompt records a new prompt and adds a note for auditing.
func (s *Session) UpdatePrompt(prompt string) {
	trimmed := strings.TrimSpace(prompt)
	s.prompts = append(s.prompts, trimmed)
	s.notes = append(s.notes, "prompt updated")
}

// RecordRegeneration appends an audit note for regenerations.
func (s *Session) RecordRegeneration() {
	s.notes = append(s.notes, "regenerated suggestions")
}

// RecordAcceptance stores an audit note for accepted previews.
func (s *Session) RecordAcceptance() {
	s.notes = append(s.notes, "accepted preview")
}

// PromptHistory returns a copy of the recorded prompts.
func (s *Session) PromptHistory() []string {
	history := make([]string, len(s.prompts))
	copy(history, s.prompts)
	return history
}

// Notes returns audit notes collected during the session.
func (s *Session) Notes() []string {
	copied := make([]string, len(s.notes))
	copy(copied, s.notes)
	return copied
}

// Files returns the original scoped filenames.
func (s *Session) Files() []string {
	copied := make([]string, len(s.files))
	copy(copied, s.files)
	return copied
}

// SequenceSeparator returns the configured sequence separator.
func (s *Session) SequenceSeparator() string {
	return s.sequenceSeparator
}

// LastOutput returns the most recent flow output.
func (s *Session) LastOutput() *flowpkg.Output {
	return s.lastOutput
}

// LastValidation returns the validation result for the most recent output.
func (s *Session) LastValidation() ValidationResult {
	return s.lastValidation
}

// Model returns the model identifier associated with the session.
func (s *Session) Model() string {
	if s.model == "" {
		return defaultModelID
	}
	return s.model
}
