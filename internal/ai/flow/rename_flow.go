package flow

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

//go:embed prompt.tmpl
var promptTemplateSource string

var (
	promptTemplate = template.Must(template.New("renameFlowPrompt").Parse(promptTemplateSource))
)

// RenameFlowInput mirrors the JSON payload passed into the Genkit flow.
type RenameFlowInput struct {
 FileNames  []string `json:"fileNames"`
 UserPrompt string   `json:"userPrompt"`
 SequenceSeparator string `json:"sequenceSeparator,omitempty"`
}

// Validate ensures the flow input is well formed.
func (in *RenameFlowInput) Validate() error {
	if in == nil {
		return errors.New("rename flow input cannot be nil")
	}
	if len(in.FileNames) == 0 {
		return errors.New("no file names provided to rename flow")
	}
	if len(in.FileNames) > 200 {
		return fmt.Errorf("rename flow supports up to 200 files per invocation (received %d)", len(in.FileNames))
	}
	normalized := make([]string, len(in.FileNames))
	for i, name := range in.FileNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return fmt.Errorf("file name at index %d is empty", i)
		}
		normalized[i] = toSlash(trimmed)
	}
	// Ensure no duplicates to simplify downstream validation.
 if dup := firstDuplicate(normalized); dup != "" {
  return fmt.Errorf("duplicate file name %q detected in flow input", dup)
 }
 in.FileNames = normalized

 sep := strings.TrimSpace(in.SequenceSeparator)
 if sep == "" {
  sep = "."
 }
 if strings.ContainsAny(sep, "/\\") {
  return fmt.Errorf("sequence separator %q cannot contain path separators", sep)
 }
 if strings.ContainsAny(sep, "\n\r") {
  return errors.New("sequence separator cannot contain newline characters")
 }
 in.SequenceSeparator = sep
 return nil
}

// RenderPrompt materialises the prompt template for the provided input.
func RenderPrompt(input RenameFlowInput) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}

	var builder strings.Builder
	if err := promptTemplate.Execute(&builder, input); err != nil {
		return "", fmt.Errorf("render rename prompt: %w", err)
	}
	return builder.String(), nil
}

// Define registers the rename flow on the supplied Genkit instance.
func Define(g *genkit.Genkit) *core.Flow[*RenameFlowInput, *Output, struct{}] {
	if g == nil {
		panic("genkit instance cannot be nil")
	}
	return genkit.DefineFlow(g, "renameFlow", flowFn)
}

func flowFn(ctx context.Context, input *RenameFlowInput) (*Output, error) {
 if err := input.Validate(); err != nil {
  return nil, err
 }

 prefix := slugify(input.UserPrompt)
 suggestions := make([]Suggestion, 0, len(input.FileNames))
 dirCounters := make(map[string]int)

 for _, name := range input.FileNames {
  suggestion := deterministicSuggestion(name, prefix, dirCounters, input.SequenceSeparator)
  suggestions = append(suggestions, Suggestion{
   Original:  name,
   Suggested: suggestion,
  })
 }

	sort.SliceStable(suggestions, func(i, j int) bool {
		return suggestions[i].Original < suggestions[j].Original
	})

	return &Output{Suggestions: suggestions}, nil
}

func deterministicSuggestion(rel string, promptPrefix string, dirCounters map[string]int, separator string) string {
 rel = toSlash(rel)
 dir := path.Dir(rel)
 if dir == "." {
  dir = ""
 }

	base := path.Base(rel)
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext)

	sanitizedName := slugify(name)

 candidate := sanitizedName
 if promptPrefix != "" {
  switch {
  case candidate == "":
   candidate = promptPrefix
  default:
			candidate = fmt.Sprintf("%s-%s", promptPrefix, candidate)
		}
	}

	if candidate == "" {
		candidate = "renamed"
	}

	counterKey := dir
 dirCounters[counterKey]++
 seq := dirCounters[counterKey]

 sep := separator
 if sep == "" {
  sep = "."
 }
 numbered := fmt.Sprintf("%02d%s%s", seq, sep, candidate)
 proposed := numbered + ext
 if dir != "" {
  return path.Join(dir, proposed)
 }
	return proposed
}

func slugify(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(value))
	lastHyphen := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastHyphen = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if !lastHyphen && b.Len() > 0 {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	return result
}

func toSlash(pathStr string) string {
	return strings.ReplaceAll(pathStr, "\\", "/")
}

func firstDuplicate(values []string) string {
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		lower := strings.ToLower(v)
		if _, exists := seen[lower]; exists {
			return v
		}
		seen[lower] = struct{}{}
	}
	return ""
}
