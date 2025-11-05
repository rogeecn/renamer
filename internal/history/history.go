package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const ledgerFileName = ".renamer"

// Operation records a single rename from source to target.
type Operation struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Entry represents a batch of operations appended to the ledger.
type Entry struct {
	Timestamp  time.Time      `json:"timestamp"`
	Command    string         `json:"command"`
	WorkingDir string         `json:"workingDir"`
	Operations []Operation    `json:"operations"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

func remarshal(value any, target any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// Append writes a new entry to the ledger in newline-delimited JSON format.
func Append(workingDir string, entry Entry) error {
	entry.Timestamp = time.Now().UTC()
	entry.WorkingDir = workingDir

	path := ledgerPath(workingDir)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(entry)
}

// Undo reverts the most recent ledger entry and removes it from the ledger file.
func Undo(workingDir string) (Entry, error) {
	path := ledgerPath(workingDir)

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Entry{}, errors.New("no ledger entries available")
	} else if err != nil {
		return Entry{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	entries := make([]Entry, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(append([]byte(nil), line...), &e); err != nil {
			return Entry{}, err
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, errors.New("no ledger entries available")
	}

	last := entries[len(entries)-1]

	// Revert operations in reverse order.
	for i := len(last.Operations) - 1; i >= 0; i-- {
		op := last.Operations[i]
		source := filepath.Join(workingDir, op.To)
		destination := filepath.Join(workingDir, op.From)
		if err := os.Rename(source, destination); err != nil {
			return Entry{}, err
		}
	}

	// Rewrite ledger without the last entry.
	if len(entries) == 1 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return Entry{}, err
		}
	} else {
		tmp := path + ".tmp"
		output, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return Entry{}, err
		}
		enc := json.NewEncoder(output)
		for _, e := range entries[:len(entries)-1] {
			if err := enc.Encode(e); err != nil {
				output.Close()
				return Entry{}, err
			}
		}
		if err := output.Close(); err != nil {
			return Entry{}, err
		}
		if err := os.Rename(tmp, path); err != nil {
			return Entry{}, err
		}
	}

	return last, nil
}

// ledgerPath returns the absolute path to the ledger file under workingDir.
func ledgerPath(workingDir string) string {
	return filepath.Join(workingDir, ledgerFileName)
}
