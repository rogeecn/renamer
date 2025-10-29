package integration

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rogeecn/renamer/internal/listing"
)

func TestScopeFlagsProduceConsistentRequests(t *testing.T) {
	root := &cobra.Command{Use: "renamer"}
	listing.RegisterScopeFlags(root.PersistentFlags())

	listCmd := &cobra.Command{Use: "list"}
	previewCmd := &cobra.Command{Use: "preview"}

	root.AddCommand(listCmd, previewCmd)

	tmp := t.TempDir()

	mustSet := func(name, value string) {
		if err := root.PersistentFlags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}

	mustSet("path", tmp)
	mustSet("recursive", "true")
	mustSet("include-dirs", "true")
	mustSet("hidden", "true")
	mustSet("extensions", ".jpg|.png")

	reqList, err := listing.ScopeFromCmd(listCmd)
	if err != nil {
		t.Fatalf("list request: %v", err)
	}

	reqPreview, err := listing.ScopeFromCmd(previewCmd)
	if err != nil {
		t.Fatalf("preview request: %v", err)
	}

	if reqList.WorkingDir != reqPreview.WorkingDir {
		t.Fatalf("working dir mismatch: %s vs %s", reqList.WorkingDir, reqPreview.WorkingDir)
	}
	if reqList.Recursive != reqPreview.Recursive {
		t.Fatalf("recursive mismatch")
	}
	if reqList.IncludeDirectories != reqPreview.IncludeDirectories {
		t.Fatalf("include-dirs mismatch")
	}
	if reqList.IncludeHidden != reqPreview.IncludeHidden {
		t.Fatalf("hidden mismatch")
	}
	if len(reqList.Extensions) != len(reqPreview.Extensions) {
		t.Fatalf("extension length mismatch: %d vs %d", len(reqList.Extensions), len(reqPreview.Extensions))
	}
	for i := range reqList.Extensions {
		if reqList.Extensions[i] != reqPreview.Extensions[i] {
			t.Fatalf("extension mismatch at %d: %s vs %s", i, reqList.Extensions[i], reqPreview.Extensions[i])
		}
	}

	if filepath.Clean(reqList.WorkingDir) != reqList.WorkingDir {
		t.Fatalf("expected cleaned working dir, got %s", reqList.WorkingDir)
	}
}
