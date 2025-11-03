package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/rogeecn/renamer/internal/ai/prompt"
)

// SaveResponse writes the AI rename response to disk for later editing.
func SaveResponse(path string, resp prompt.RenameResponse) error {
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ai plan: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write ai plan %s: %w", path, err)
	}
	return nil
}

// LoadResponse reads an edited AI rename response from disk.
func LoadResponse(path string) (prompt.RenameResponse, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return prompt.RenameResponse{}, fmt.Errorf("plan file %s not found", path)
		}
		return prompt.RenameResponse{}, fmt.Errorf("read plan file %s: %w", path, err)
	}
	var resp prompt.RenameResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return prompt.RenameResponse{}, fmt.Errorf("parse plan file %s: %w", path, err)
	}
	return resp, nil
}
