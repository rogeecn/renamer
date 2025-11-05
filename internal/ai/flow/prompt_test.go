package flow_test

import (
	"strings"
	"testing"

	"github.com/rogeecn/renamer/internal/ai/flow"
)

func TestRenderPromptIncludesFilesAndPrompt(t *testing.T) {
	input := flow.RenameFlowInput{
		FileNames:  []string{"IMG_0001.jpg", "albums/Day 1.png"},
		UserPrompt: "按地点重新命名",
	}

	rendered, err := flow.RenderPrompt(input)
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}

	for _, expected := range []string{"IMG_0001.jpg", "albums/Day 1.png"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("prompt missing filename %q: %s", expected, rendered)
		}
	}

	if !strings.Contains(rendered, "按地点重新命名") {
		t.Fatalf("prompt missing user guidance: %s", rendered)
	}

	if !strings.Contains(rendered, "suggestions") {
		t.Fatalf("prompt missing JSON structure guidance: %s", rendered)
	}
}
