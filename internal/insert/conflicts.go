package insert

import (
	"fmt"
)

// ConflictDetector tracks proposed targets to detect duplicates.
type ConflictDetector struct {
	planned map[string]string // proposedRelative -> originalRelative
}

// NewConflictDetector creates an empty detector.
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{planned: make(map[string]string)}
}

// Register validates the proposed target and returns an error string if conflict occurred.
func (d *ConflictDetector) Register(original, proposed string) (string, bool) {
	if proposed == "" {
		return "", false
	}
	if existing, ok := d.planned[proposed]; ok && existing != original {
		return fmt.Sprintf("duplicate target with %s", existing), false
	}
	d.planned[proposed] = original
	return "", true
}

// Forget removes a planned target (used if operation skipped).
func (d *ConflictDetector) Forget(proposed string) {
	delete(d.planned, proposed)
}
