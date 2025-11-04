package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/joho/godotenv"
)

const (
	configDirEnvVar   = "RENAMER_CONFIG_DIR"
	configFileName    = ".renamer"
	defaultVendorSlug = "openai"

	vendorTokenSuffix = "_TOKEN"

	errTokenNotFoundFmt = "model token %q not found in %s or the process environment"
)

// TokenProvider resolves API tokens for AI models.
type TokenProvider interface {
	ResolveModelToken(model string) (string, error)
}

// TokenStore loads model authentication tokens from ~/.config/.renamer.
type TokenStore struct {
	configDir string

	once   sync.Once
	values map[string]string
	err    error
}

// NewTokenStore constructs a TokenStore rooted at configDir. When configDir is
// empty the default path of `$HOME/.config/.renamer` is used. An environment
// override can be supplied via RENAMER_CONFIG_DIR.
func NewTokenStore(configDir string) (*TokenStore, error) {
	root := configDir
	if root == "" {
		if override := strings.TrimSpace(os.Getenv(configDirEnvVar)); override != "" {
			root = override
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("resolve user home: %w", err)
			}
			root = filepath.Join(home, ".config", configFileName)
		}
	}

	return &TokenStore{
		configDir: root,
		values:    make(map[string]string),
	}, nil
}

// ConfigDir returns the directory the token store reads from.
func (s *TokenStore) ConfigDir() string {
	return s.configDir
}

// ResolveModelToken returns the token for the provided model name. Model names
// are normalized to match the `<VENDOR>_TOKEN` convention documented
// for the CLI. Environment variables take precedence over file-based tokens.
func (s *TokenStore) ResolveModelToken(model string) (string, error) {
	key := ModelTokenKey(model)
	return s.lookup(key)
}

// lookup loads the requested key from either the environment or cached tokens.
func (s *TokenStore) lookup(key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("token key must not be empty")
	}

	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val), nil
	}

	if err := s.ensureLoaded(); err != nil {
		return "", err
	}

	if val, ok := s.values[key]; ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val), nil
	}

	return "", fmt.Errorf(errTokenNotFoundFmt, key, s.configFilePath())
}

func (s *TokenStore) ensureLoaded() error {
	s.once.Do(func() {
		s.err = s.loadConfigFile()
	})
	return s.err
}

func (s *TokenStore) loadConfigFile() error {
	path := s.configFilePath()
	envMap, err := godotenv.Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	for k, v := range envMap {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		s.values[k] = strings.TrimSpace(v)
	}
	return nil
}

func (s *TokenStore) configFilePath() string {
	info, err := os.Stat(s.configDir)
	if err == nil {
		if info.IsDir() {
			return filepath.Join(s.configDir, configFileName)
		}
		return s.configDir
	}
	if strings.HasSuffix(s.configDir, configFileName) {
		return s.configDir
	}
	return filepath.Join(s.configDir, configFileName)
}

// ModelTokenKey derives the vendor token key for the provided model, following
// the `<VENDOR>_TOKEN` convention. When the vendor cannot be inferred the
// default OpenAI slug is returned.
func ModelTokenKey(model string) string {
	slug := vendorSlugFromModel(model)
	if slug == "" {
		slug = defaultVendorSlug
	}
	return strings.ToUpper(slug) + vendorTokenSuffix
}

func vendorSlugFromModel(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if normalized == "" {
		return defaultVendorSlug
	}

	if explicit := explicitVendorPrefix(normalized); explicit != "" {
		return explicit
	}

	for _, mapping := range vendorHintTable {
		for _, hint := range mapping.hints {
			if strings.Contains(normalized, hint) {
				return mapping.vendor
			}
		}
	}

	if firstToken := leadingToken(normalized); firstToken != "" {
		return slugify(firstToken)
	}

	if slug := slugify(normalized); slug != "" {
		return slug
	}

	return defaultVendorSlug
}

func explicitVendorPrefix(value string) string {
	separators := func(r rune) bool {
		switch r {
		case '/', ':', '@':
			return true
		}
		return false
	}
	parts := strings.FieldsFunc(value, separators)
	if len(parts) > 1 {
		if slug := slugify(parts[0]); slug != "" {
			return slug
		}
	}
	return ""
}

func leadingToken(value string) string {
	for i, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if i == 0 {
			return ""
		}
		return value[:i]
	}
	return value
}

var vendorHintTable = []struct {
	vendor string
	hints  []string
}{
	{vendor: "openai", hints: []string{"openai", "gpt", "o1", "chatgpt"}},
	{vendor: "anthropic", hints: []string{"anthropic", "claude"}},
	{vendor: "google", hints: []string{"google", "gemini", "learnlm", "palm"}},
	{vendor: "mistral", hints: []string{"mistral", "mixtral", "ministral"}},
	{vendor: "cohere", hints: []string{"cohere", "command", "r-plus"}},
	{vendor: "moonshot", hints: []string{"moonshot"}},
	{vendor: "zhipu", hints: []string{"zhipu", "glm"}},
	{vendor: "alibaba", hints: []string{"dashscope", "qwen"}},
	{vendor: "baidu", hints: []string{"wenxin", "ernie", "qianfan"}},
	{vendor: "minimax", hints: []string{"minimax", "abab"}},
	{vendor: "bytedance", hints: []string{"doubao", "bytedance"}},
	{vendor: "baichuan", hints: []string{"baichuan"}},
	{vendor: "deepseek", hints: []string{"deepseek"}},
	{vendor: "xai", hints: []string{"grok", "xai"}},
}

func slugify(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(input))

	lastUnderscore := false
	for _, r := range input {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		default:
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	return strings.Trim(b.String(), "_")
}
