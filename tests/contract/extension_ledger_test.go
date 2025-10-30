package contract

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	renamercmd "github.com/rogeecn/renamer/cmd"
	"github.com/rogeecn/renamer/internal/extension"
	"github.com/rogeecn/renamer/internal/listing"
)

func TestExtensionApplyMetadataCaptured(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	writeTestFile(t, filepath.Join(tmp, "clip.jpeg"))
	writeTestFile(t, filepath.Join(tmp, "poster.JPG"))
	writeTestFile(t, filepath.Join(tmp, "flyer.jpg"))

	scope := &listing.ListingRequest{
		WorkingDir:         tmp,
		IncludeDirectories: false,
		Recursive:          false,
		IncludeHidden:      false,
		Extensions:         nil,
		Format:             listing.FormatTable,
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("validate scope: %v", err)
	}

	req := extension.NewRequest(scope)
	req.SetExecutionMode(true, false)

	sources := []string{".jpeg", ".JPG", ".jpg"}
	canonical, display, duplicates := extension.NormalizeSourceExtensions(sources)
	target := extension.NormalizeTargetExtension(".jpg")
	targetCanonical := extension.CanonicalExtension(target)

	filteredCanonical := make([]string, 0, len(canonical))
	filteredDisplay := make([]string, 0, len(display))
	noOps := make([]string, 0)
	for i, canon := range canonical {
		if canon == targetCanonical {
			noOps = append(noOps, display[i])
			continue
		}
		filteredCanonical = append(filteredCanonical, canon)
		filteredDisplay = append(filteredDisplay, display[i])
	}
	req.SetExtensions(filteredCanonical, filteredDisplay, target)
	req.SetWarnings(duplicates, noOps)

	summary, planned, err := extension.Preview(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("preview error: %v", err)
	}
	req.SetExecutionMode(false, true)

	entry, err := extension.Apply(context.Background(), req, planned, summary)
	if err != nil {
		t.Fatalf("apply error: %v", err)
	}

	if entry.Metadata == nil {
		t.Fatalf("expected metadata to be recorded")
	}

	sourcesMeta, ok := entry.Metadata["sourceExtensions"].([]string)
	if !ok || len(sourcesMeta) != len(filteredDisplay) {
		t.Fatalf("sourceExtensions metadata mismatch: %#v", entry.Metadata["sourceExtensions"])
	}
	if sourcesMeta[0] != ".jpeg" {
		t.Fatalf("expected .jpeg in source metadata, got %v", sourcesMeta)
	}

	targetMeta, ok := entry.Metadata["targetExtension"].(string)
	if !ok || targetMeta != target {
		t.Fatalf("targetExtension metadata mismatch: %v", targetMeta)
	}

	if changed, ok := entry.Metadata["totalChanged"].(int); !ok || changed != summary.TotalChanged {
		t.Fatalf("totalChanged metadata mismatch: %v", entry.Metadata["totalChanged"])
	}
	if noChange, ok := entry.Metadata["noChange"].(int); !ok || noChange != summary.NoChange {
		t.Fatalf("noChange metadata mismatch: %v", entry.Metadata["noChange"])
	}

	counts, ok := entry.Metadata["perExtensionCounts"].(map[string]int)
	if !ok {
		t.Fatalf("perExtensionCounts metadata missing: %#v", entry.Metadata["perExtensionCounts"])
	}
	if counts[".jpeg"] == 0 || counts[".jpg"] == 0 {
		t.Fatalf("expected counts for .jpeg and .jpg, got %#v", counts)
	}

	scopeMeta, ok := entry.Metadata["scope"].(map[string]any)
	if !ok {
		t.Fatalf("scope metadata missing: %#v", entry.Metadata["scope"])
	}
	if includeHidden, _ := scopeMeta["includeHidden"].(bool); includeHidden {
		t.Fatalf("includeHidden should be false, got %v", includeHidden)
	}

	warnings, ok := entry.Metadata["warnings"].([]string)
	if !ok || len(warnings) == 0 {
		t.Fatalf("warnings metadata missing: %#v", entry.Metadata["warnings"])
	}
	joined := strings.Join(warnings, " ")
	if !strings.Contains(joined, "duplicate source extension") {
		t.Fatalf("expected duplicate warning in metadata: %v", warnings)
	}

	ledger := filepath.Join(tmp, ".renamer")
	if _, err := os.Stat(ledger); err != nil {
		t.Fatalf("ledger not created: %v", err)
	}

	if err := os.Remove(ledger); err != nil {
		t.Fatalf("cleanup ledger: %v", err)
	}
}

func TestExtensionCommandExitCodes(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	var out bytes.Buffer
	cmd := renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"extension", ".jpeg", ".jpg", "--dry-run", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected dry-run to exit successfully, err=%v output=%s", err, out.String())
	}

	if !strings.Contains(out.String(), "No candidates found.") {
		t.Fatalf("expected no candidates notice, output=%s", out.String())
	}

	writeTestFile(t, filepath.Join(tmp, "clip.jpeg"))

	out.Reset()
	cmd = renamercmd.NewRootCommand()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"extension", ".jpeg", ".jpg", "--yes", "--path", tmp})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected apply to exit successfully, err=%v output=%s", err, out.String())
	}

	if _, err := os.Stat(filepath.Join(tmp, "clip.jpg")); err != nil {
		t.Fatalf("expected clip.jpg after apply: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "clip.jpeg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected clip.jpeg to be renamed, err=%v", err)
	}
}
