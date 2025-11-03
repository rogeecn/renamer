package genkit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	genaigo "github.com/firebase/genkit/go/ai"
	"github.com/openai/openai-go/option"

	aiconfig "github.com/rogeecn/renamer/internal/ai/config"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

// WorkflowRunner executes a Genkit request and returns the structured response.
type WorkflowRunner interface {
	Run(ctx context.Context, req Request) (Result, error)
}

// WorkflowFactory constructs workflow runners.
type WorkflowFactory func(ctx context.Context, opts Options) (WorkflowRunner, error)

var (
	factoryMu      sync.RWMutex
	defaultFactory = func(ctx context.Context, opts Options) (WorkflowRunner, error) {
		return NewWorkflow(ctx, opts)
	}
	currentFactory WorkflowFactory = defaultFactory
)

// OverrideWorkflowFactory allows tests to supply custom workflow implementations.
func OverrideWorkflowFactory(factory WorkflowFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	if factory == nil {
		currentFactory = defaultFactory
		return
	}
	currentFactory = factory
}

// ResetWorkflowFactory restores the default workflow constructor.
func ResetWorkflowFactory() {
	OverrideWorkflowFactory(nil)
}

func getWorkflowFactory() WorkflowFactory {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	return currentFactory
}

// ClientOptions configure the Genkit client.
type ClientOptions struct {
	Model          string
	TokenProvider  aiconfig.TokenProvider
	RequestOptions []option.RequestOption
}

// Client orchestrates prompt execution against the configured workflow.
type Client struct {
	model          string
	tokenProvider  aiconfig.TokenProvider
	requestOptions []option.RequestOption
}

// NewClient constructs a client with optional overrides.
func NewClient(opts ClientOptions) *Client {
	model := strings.TrimSpace(opts.Model)
	if model == "" {
		model = DefaultModelName
	}
	return &Client{
		model:          model,
		tokenProvider:  opts.TokenProvider,
		requestOptions: append([]option.RequestOption(nil), opts.RequestOptions...),
	}
}

// Invocation describes a single Genkit call.
type Invocation struct {
	Instructions string
	Prompt       prompt.RenamePrompt
	Model        string
}

// InvocationResult carries the parsed response alongside telemetry.
type InvocationResult struct {
	PromptHash    string
	Model         string
	Response      prompt.RenameResponse
	ModelResponse *genaigo.ModelResponse
	PromptJSON    []byte
}

// Invoke executes the workflow and returns the structured response.
func (c *Client) Invoke(ctx context.Context, inv Invocation) (InvocationResult, error) {
	model := strings.TrimSpace(inv.Model)
	if model == "" {
		model = c.model
	}
	if model == "" {
		model = DefaultModelName
	}

	payload, err := json.Marshal(inv.Prompt)
	if err != nil {
		return InvocationResult{}, fmt.Errorf("marshal prompt payload: %w", err)
	}

	factory := getWorkflowFactory()
	runner, err := factory(ctx, Options{
		Model:          model,
		TokenProvider:  c.tokenProvider,
		RequestOptions: c.requestOptions,
	})
	if err != nil {
		return InvocationResult{}, err
	}

	result, err := runner.Run(ctx, Request{
		Instructions: inv.Instructions,
		Payload:      inv.Prompt,
	})
	if err != nil {
		return InvocationResult{}, err
	}

	if strings.TrimSpace(result.Response.Model) == "" {
		result.Response.Model = model
	}

	promptHash := hashPrompt(inv.Instructions, payload)
	if strings.TrimSpace(result.Response.PromptHash) == "" {
		result.Response.PromptHash = promptHash
	}

	return InvocationResult{
		PromptHash:    promptHash,
		Model:         result.Response.Model,
		Response:      result.Response,
		ModelResponse: result.ModelResponse,
		PromptJSON:    payload,
	}, nil
}

func hashPrompt(instructions string, payload []byte) string {
	hasher := sha256.New()
	hasher.Write([]byte(strings.TrimSpace(instructions)))
	hasher.Write([]byte{'\n'})
	hasher.Write(payload)
	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum)
}
