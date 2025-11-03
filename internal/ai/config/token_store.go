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
	defaultConfigRoot = ".renamer"

	modelTokenSuffix = "_MODEL_AUTH_TOKEN"

	defaultEnvFile      = ".env"
	secondaryEnvFile    = "tokens.env"
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
			root = filepath.Join(home, ".config", defaultConfigRoot)
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
// are normalized to match the `<slug>_MODEL_AUTH_TOKEN` convention documented
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

	path := filepath.Join(s.configDir, key)
	raw, err := os.ReadFile(path)
	if err == nil {
		value := strings.TrimSpace(string(raw))
		if value != "" {
			s.values[key] = value
			return value, nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("read token file %s: %w", path, err)
	}

	return "", fmt.Errorf(errTokenNotFoundFmt, key, s.configDir)
}

func (s *TokenStore) ensureLoaded() error {
	s.once.Do(func() {
		s.err = s.loadEnvFiles()
		if s.err != nil {
			return
		}
		s.err = s.scanTokenFiles()
	})
	return s.err
}

func (s *TokenStore) loadEnvFiles() error {
	candidates := []string{
		filepath.Join(s.configDir, defaultEnvFile),
		filepath.Join(s.configDir, secondaryEnvFile),
	}

	for _, path := range candidates {
		envMap, err := godotenv.Read(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
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
	}
	return nil
}

func (s *TokenStore) scanTokenFiles() error {
	entries, err := os.ReadDir(s.configDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("scan %s: %w", s.configDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(s.configDir, name)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		data := strings.TrimSpace(string(content))
		if data == "" {
			continue
		}

		if parsed, perr := godotenv.Unmarshal(data); perr == nil && len(parsed) > 0 {
			for k, v := range parsed {
				if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
					continue
				}
				s.values[k] = strings.TrimSpace(v)
			}
			continue
		}

		s.values[name] = data
	}

	return nil
}

// ModelTokenKey derives the token filename/environment variable for the given
// model name following the `<slug>_MODEL_AUTH_TOKEN` convention. When model is
// empty the default slug `default` is used.
func ModelTokenKey(model string) string {
	slug := slugify(model)
	if slug == "" {
		slug = "default"
	}
	return slug + modelTokenSuffix
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
