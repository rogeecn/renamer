package ai

import (
    "context"
    "errors"

    "github.com/rogeecn/renamer/internal/ai/flow"
)

// Runner executes the rename flow and returns structured suggestions.
type Runner func(ctx context.Context, input *flow.RenameFlowInput) (*flow.Output, error)

// Client orchestrates flow invocation for callers such as the CLI command.
type Client struct {
	runner Runner
}

// ClientOption customises the AI client behaviour.
type ClientOption func(*Client)

// WithRunner overrides the flow runner implementation (useful for tests).
func WithRunner(r Runner) ClientOption {
	return func(c *Client) {
		c.runner = r
	}
}

// NewClient constructs a Client with the default Genkit-backed runner.
func NewClient(opts ...ClientOption) *Client {
    client := &Client{}
    client.runner = func(ctx context.Context, input *flow.RenameFlowInput) (*flow.Output, error) {
        creds, err := LoadCredentials()
        if err != nil {
            return nil, err
        }
        return runRenameFlow(ctx, input, creds)
    }
    for _, opt := range opts {
        opt(client)
    }
    return client
}

// Suggest executes the rename flow and returns structured suggestions.
func (c *Client) Suggest(ctx context.Context, input *flow.RenameFlowInput) (*flow.Output, error) {
	if c == nil {
		return nil, ErrClientNotInitialized
	}
	if c.runner == nil {
		return nil, ErrRunnerNotConfigured
	}
	return c.runner(ctx, input)
}

// ErrClientNotInitialized indicates the client receiver was nil.
var ErrClientNotInitialized = errors.New("ai client not initialized")

// ErrRunnerNotConfigured indicates the client runner is missing.
var ErrRunnerNotConfigured = errors.New("ai client runner not configured")
