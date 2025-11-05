package ai

import (
	"errors"
	"fmt"
	"os"
)

var apiKeyEnvVars = []string{
	"GOOGLE_API_KEY",
	"GEMINI_API_KEY",
	"RENAMER_AI_KEY",
}

// Credentials encapsulates the values required to authenticate with the AI provider.
type Credentials struct {
	APIKey string
}

// LoadCredentials returns the AI credentials sourced from environment variables.
func LoadCredentials() (Credentials, error) {
	for _, env := range apiKeyEnvVars {
		if key, ok := os.LookupEnv(env); ok && key != "" {
			return Credentials{APIKey: key}, nil
		}
	}
	return Credentials{}, errors.New("AI provider key missing; set GOOGLE_API_KEY (recommended), GEMINI_API_KEY, or RENAMER_AI_KEY")
}

// MaskedCredentials returns a redacted view of the credentials for logging purposes.
func MaskedCredentials(creds Credentials) string {
	if creds.APIKey == "" {
		return "(empty)"
	}

	if len(creds.APIKey) <= 6 {
		return "***"
	}

	return fmt.Sprintf("%s***", creds.APIKey[:3])
}
