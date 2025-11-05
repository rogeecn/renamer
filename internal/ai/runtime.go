package ai

import (
	"context"
	"os"
	"sync"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"google.golang.org/genai"

	"github.com/rogeecn/renamer/internal/ai/flow"
)

var (
	runtimeOnce sync.Once
	runtimeErr  error
	runtimeFlow *core.Flow[*flow.RenameFlowInput, *flow.Output, struct{}]
)

func ensureRuntime(creds Credentials) error {
	runtimeOnce.Do(func() {
		ctx := context.Background()
		geminiBase := os.Getenv("GOOGLE_GEMINI_BASE_URL")
		vertexBase := os.Getenv("GOOGLE_VERTEX_BASE_URL")
		if geminiBase != "" || vertexBase != "" {
			genai.SetDefaultBaseURLs(genai.BaseURLParameters{
				GeminiURL: geminiBase,
				VertexURL: vertexBase,
			})
		}

		plugin := &googlegenai.GoogleAI{APIKey: creds.APIKey}

		g := genkit.Init(ctx,
			genkit.WithPlugins(plugin),
			genkit.WithDefaultModel(defaultModelID),
		)

		runtimeFlow = flow.Define(g)
	})
	return runtimeErr
}

func runRenameFlow(ctx context.Context, input *flow.RenameFlowInput, creds Credentials) (*flow.Output, error) {
	if err := ensureRuntime(creds); err != nil {
		return nil, err
	}
	if runtimeFlow == nil {
		return nil, runtimeErr
	}
	return runtimeFlow.Run(ctx, input)
}
