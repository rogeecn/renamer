package prompt

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const defaultMaxSamples = 10

// SequenceRule captures the numbering instructions forwarded to the AI.
type SequenceRule struct {
	Style     string
	Width     int
	Start     int
	Separator string
}

// PolicyConfig enumerates naming policy directives for the AI prompt.
type PolicyConfig struct {
	Prefix            string
	Casing            string
	AllowSpaces       bool
	KeepOriginalOrder bool
	ForbiddenTokens   []string
}

// SampleCandidate represents a traversal sample considered for inclusion in the prompt.
type SampleCandidate struct {
	RelativePath string
	SizeBytes    int64
	Depth        int
}

// BuildInput aggregates the contextual data required to assemble the AI prompt payload.
type BuildInput struct {
	WorkingDir  string
	Samples     []SampleCandidate
	TotalCount  int
	Sequence    SequenceRule
	Policies    PolicyConfig
	BannedTerms []string
	Metadata    map[string]string
}

// Builder constructs RenamePrompt payloads from traversal context.
type Builder struct {
	maxSamples int
	clock      func() time.Time
}

// Option mutates builder configuration.
type Option func(*Builder)

// WithMaxSamples overrides the number of sampled files emitted in the prompt (default 10).
func WithMaxSamples(n int) Option {
	return func(b *Builder) {
		if n > 0 {
			b.maxSamples = n
		}
	}
}

// WithClock injects a deterministic clock for metadata generation.
func WithClock(clock func() time.Time) Option {
	return func(b *Builder) {
		if clock != nil {
			b.clock = clock
		}
	}
}

// NewBuilder instantiates a Builder with default configuration.
func NewBuilder(opts ...Option) *Builder {
	builder := &Builder{
		maxSamples: defaultMaxSamples,
		clock:      time.Now().UTC,
	}
	for _, opt := range opts {
		opt(builder)
	}
	return builder
}

// Build produces a RenamePrompt populated with traversal context.
func (b *Builder) Build(input BuildInput) (RenamePrompt, error) {
	if strings.TrimSpace(input.WorkingDir) == "" {
		return RenamePrompt{}, errors.New("prompt builder: working directory required")
	}
	if input.TotalCount <= 0 {
		return RenamePrompt{}, errors.New("prompt builder: total count must be positive")
	}
	if strings.TrimSpace(input.Sequence.Style) == "" {
		return RenamePrompt{}, errors.New("prompt builder: sequence style required")
	}
	if input.Sequence.Width <= 0 {
		return RenamePrompt{}, errors.New("prompt builder: sequence width must be positive")
	}
	if input.Sequence.Start <= 0 {
		return RenamePrompt{}, errors.New("prompt builder: sequence start must be positive")
	}
	if strings.TrimSpace(input.Policies.Casing) == "" {
		return RenamePrompt{}, errors.New("prompt builder: naming casing required")
	}

	samples := make([]SampleCandidate, 0, len(input.Samples))
	for _, sample := range input.Samples {
		if strings.TrimSpace(sample.RelativePath) == "" {
			continue
		}
		samples = append(samples, sample)
	}

	sort.Slice(samples, func(i, j int) bool {
		a := strings.ToLower(samples[i].RelativePath)
		b := strings.ToLower(samples[j].RelativePath)
		if a == b {
			return samples[i].RelativePath < samples[j].RelativePath
		}
		return a < b
	})

	max := b.maxSamples
	if max <= 0 || max > len(samples) {
		max = len(samples)
	}

	promptSamples := make([]PromptSample, 0, max)
	for i := 0; i < max; i++ {
		sample := samples[i]
		ext := filepath.Ext(sample.RelativePath)
		promptSamples = append(promptSamples, PromptSample{
			OriginalName: sample.RelativePath,
			Extension:    ext,
			SizeBytes:    sample.SizeBytes,
			PathDepth:    sample.Depth,
		})
	}

	banned := normalizeBannedTerms(input.BannedTerms)

	metadata := make(map[string]string, len(input.Metadata)+1)
	for k, v := range input.Metadata {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		metadata[k] = v
	}
	metadata["generatedAt"] = b.clock().Format(time.RFC3339)

	return RenamePrompt{
		WorkingDir: promptAbs(input.WorkingDir),
		Samples:    promptSamples,
		TotalCount: input.TotalCount,
		SequenceRule: SequenceRuleConfig{
			Style:     input.Sequence.Style,
			Width:     input.Sequence.Width,
			Start:     input.Sequence.Start,
			Separator: input.Sequence.Separator,
		},
		Policies: NamingPolicyConfig{
			Prefix:            input.Policies.Prefix,
			Casing:            input.Policies.Casing,
			AllowSpaces:       input.Policies.AllowSpaces,
			KeepOriginalOrder: input.Policies.KeepOriginalOrder,
			ForbiddenTokens:   append([]string(nil), input.Policies.ForbiddenTokens...),
		},
		BannedTerms: banned,
		Metadata:    metadata,
	}, nil
}

func promptAbs(dir string) string {
	return strings.TrimSpace(dir)
}

func normalizeBannedTerms(values []string) []string {
	unique := make(map[string]struct{})
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if lower == "" {
			continue
		}
		unique[lower] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	terms := make([]string, 0, len(unique))
	for term := range unique {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	return terms
}
