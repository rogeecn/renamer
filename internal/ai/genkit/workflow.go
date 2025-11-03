package genkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/firebase/genkit/go/ai"
	gogenkit "github.com/firebase/genkit/go/genkit"
	oai "github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/openai/openai-go/option"

	aiconfig "github.com/rogeecn/renamer/internal/ai/config"
	"github.com/rogeecn/renamer/internal/ai/prompt"
)

const (
	defaultModelName = "gpt-4o-mini"
	// DefaultModelName exposes the default model identifier used by the CLI.
	DefaultModelName = defaultModelName
)

var (
	// ErrMissingToken indicates the workflow could not locate a model token.
	ErrMissingToken = errors.New("genkit workflow: model token not available")
	// ErrMissingInstructions indicates that no system instructions were provided for a run.
	ErrMissingInstructions = errors.New("genkit workflow: instructions are required")
)

// DataGenerator executes the Genkit request and decodes the structured response.
type DataGenerator func(ctx context.Context, g *gogenkit.Genkit, opts ...ai.GenerateOption) (*prompt.RenameResponse, *ai.ModelResponse, error)

// Options configure a Workflow instance.
type Options struct {
	Model          string
	TokenProvider  aiconfig.TokenProvider
	RequestOptions []option.RequestOption
	Generator      DataGenerator
}

// Request captures the input necessary to execute the Genkit workflow.
type Request struct {
	Instructions string
	Payload      prompt.RenamePrompt
}

// Result bundles the typed response together with the raw Genkit metadata.
type Result struct {
	Response      prompt.RenameResponse
	ModelResponse *ai.ModelResponse
}

// Workflow orchestrates execution of the Genkit rename pipeline.
type Workflow struct {
	modelName string
	genkit    *gogenkit.Genkit
	model     ai.Model
	generate  DataGenerator
}

// NewWorkflow instantiates a Genkit workflow for the preferred model. When no
// model is provided it defaults to gpt-4o-mini. The workflow requires a token
// provider capable of resolving `<model>_MODEL_AUTH_TOKEN` secrets.
func NewWorkflow(ctx context.Context, opts Options) (*Workflow, error) {
	modelName := strings.TrimSpace(opts.Model)
	if modelName == "" {
		modelName = defaultModelName
	}

	token, err := resolveToken(opts.TokenProvider, modelName)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("%w for %q", ErrMissingToken, modelName)
	}

	plugin := &oai.OpenAI{
		APIKey: token,
		Opts:   opts.RequestOptions,
	}

	g := gogenkit.Init(ctx, gogenkit.WithPlugins(plugin))
	model := plugin.Model(g, modelName)

	generator := opts.Generator
	if generator == nil {
		generator = func(ctx context.Context, g *gogenkit.Genkit, opts ...ai.GenerateOption) (*prompt.RenameResponse, *ai.ModelResponse, error) {
			return gogenkit.GenerateData[prompt.RenameResponse](ctx, g, opts...)
		}
	}

	return &Workflow{
		modelName: modelName,
		genkit:    g,
		model:     model,
		generate:  generator,
	}, nil
}

// Run executes the workflow with the provided request and decodes the response
// into the shared RenameResponse structure.
func (w *Workflow) Run(ctx context.Context, req Request) (Result, error) {
	if w == nil {
		return Result{}, errors.New("genkit workflow: nil receiver")
	}
	if strings.TrimSpace(req.Instructions) == "" {
		return Result{}, ErrMissingInstructions
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		return Result{}, fmt.Errorf("marshal workflow payload: %w", err)
	}

	options := []ai.GenerateOption{
		ai.WithModel(w.model),
		ai.WithSystem(req.Instructions),
		ai.WithPrompt(string(payload)),
	}

	response, raw, err := w.generate(ctx, w.genkit, options...)
	if err != nil {
		return Result{}, fmt.Errorf("genkit generate: %w", err)
	}

	return Result{
		Response:      deref(response),
		ModelResponse: raw,
	}, nil
}

func resolveToken(provider aiconfig.TokenProvider, model string) (string, error) {
	if provider != nil {
		if token, err := provider.ResolveModelToken(model); err == nil && strings.TrimSpace(token) != "" {
			return token, nil
		} else if err != nil {
			return "", fmt.Errorf("resolve model token: %w", err)
		}
	}

	if direct := strings.TrimSpace(os.Getenv(aiconfig.ModelTokenKey(model))); direct != "" {
		return direct, nil
	}

	store, err := aiconfig.NewTokenStore("")
	if err != nil {
		return "", err
	}

	token, err := store.ResolveModelToken(model)
	if err != nil {
		return "", err
	}
	return token, nil
}

func deref(resp *prompt.RenameResponse) prompt.RenameResponse {
	if resp == nil {
		return prompt.RenameResponse{}
	}
	return *resp
}
