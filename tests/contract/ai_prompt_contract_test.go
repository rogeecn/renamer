package contract

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	aiprompt "github.com/rogeecn/renamer/internal/ai/prompt"
)

func TestRenamePromptSchemaAlignment(t *testing.T) {
	builder := aiprompt.NewBuilder(
		aiprompt.WithClock(func() time.Time {
			return time.Date(2025, 11, 3, 15, 4, 5, 0, time.UTC)
		}),
		aiprompt.WithMaxSamples(2),
	)

	input := aiprompt.BuildInput{
		WorkingDir: "/tmp/workspace",
		TotalCount: 3,
		Sequence: aiprompt.SequenceRule{
			Style:     "prefix",
			Width:     3,
			Start:     1,
			Separator: "_",
		},
		Policies: aiprompt.PolicyConfig{
			Casing: "kebab",
		},
		BannedTerms: []string{"Promo", " ", "promo", "ads"},
		Samples: []aiprompt.SampleCandidate{
			{
				RelativePath: "promo SALE 01.JPG",
				SizeBytes:    2048,
				Depth:       0,
			},
			{
				RelativePath: filepath.ToSlash(filepath.Join("nested", "Report FINAL.PDF")),
				SizeBytes:    1024,
				Depth:       1,
			},
			{
				RelativePath: "notes.txt",
				SizeBytes:    128,
				Depth:       0,
			},
		},
		Metadata: map[string]string{
			"cliVersion": "test-build",
		},
	}

	promptPayload, err := builder.Build(input)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}

	if promptPayload.WorkingDir != input.WorkingDir {
		t.Fatalf("expected working dir %q, got %q", input.WorkingDir, promptPayload.WorkingDir)
	}

	if promptPayload.TotalCount != input.TotalCount {
		t.Fatalf("expected total count %d, got %d", input.TotalCount, promptPayload.TotalCount)
	}

	if len(promptPayload.Samples) != 2 {
		t.Fatalf("expected 2 samples after max cap, got %d", len(promptPayload.Samples))
	}

	first := promptPayload.Samples[0]
	if first.OriginalName != "nested/Report FINAL.PDF" {
		t.Fatalf("unexpected first sample name: %q", first.OriginalName)
	}
	if first.Extension != ".PDF" {
		t.Fatalf("expected extension to remain case-sensitive, got %q", first.Extension)
	}
	if first.SizeBytes != 1024 {
		t.Fatalf("expected size 1024, got %d", first.SizeBytes)
	}
	if first.PathDepth != 1 {
		t.Fatalf("expected depth 1, got %d", first.PathDepth)
	}

	seq := promptPayload.SequenceRule
	if seq.Style != "prefix" || seq.Width != 3 || seq.Start != 1 || seq.Separator != "_" {
		t.Fatalf("sequence rule mismatch: %#v", seq)
	}

	if promptPayload.Policies.Casing != "kebab" {
		t.Fatalf("expected casing kebab, got %q", promptPayload.Policies.Casing)
	}

	expectedTerms := []string{"ads", "promo"}
	if len(promptPayload.BannedTerms) != len(expectedTerms) {
		t.Fatalf("expected %d banned terms, got %d", len(expectedTerms), len(promptPayload.BannedTerms))
	}
	for i, term := range expectedTerms {
		if promptPayload.BannedTerms[i] != term {
			t.Fatalf("banned term at %d mismatch: expected %q got %q", i, term, promptPayload.BannedTerms[i])
		}
	}

	if promptPayload.Metadata["cliVersion"] != "test-build" {
		t.Fatalf("metadata cliVersion mismatch: %s", promptPayload.Metadata["cliVersion"])
	}
	if promptPayload.Metadata["generatedAt"] != "2025-11-03T15:04:05Z" {
		t.Fatalf("expected generatedAt timestamp preserved, got %q", promptPayload.Metadata["generatedAt"])
	}

	raw, err := json.Marshal(promptPayload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal round-trip error: %v", err)
	}

	for _, key := range []string{"workingDir", "samples", "totalCount", "sequenceRule", "policies"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("prompt JSON missing key %q", key)
		}
	}
}
