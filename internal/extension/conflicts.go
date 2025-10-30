package extension

import (
	"errors"
	"fmt"
	"os"
)

type conflictDetector struct {
	planned map[string]string
}

func newConflictDetector() *conflictDetector {
	return &conflictDetector{planned: make(map[string]string)}
}

// evaluateTarget inspects plan collisions and filesystem conflicts. It returns true when the
// rename can proceed, false when it must be skipped, propagating any hard errors encountered.
func (d *conflictDetector) evaluateTarget(summary *ExtensionSummary, candidateRel, targetRel, originalAbs, targetAbs string) (bool, error) {
	if existing, ok := d.planned[targetRel]; ok && existing != candidateRel {
		summary.AddConflict(Conflict{
			OriginalPath: candidateRel,
			ProposedPath: targetRel,
			Reason:       "duplicate_target",
		})
		summary.AddWarning(fmt.Sprintf("skipped %s because %s also maps to %s", candidateRel, existing, targetRel))
		return false, nil
	}

	origInfo, origErr := os.Stat(originalAbs)
	if origErr != nil {
		return false, origErr
	}

	if info, err := os.Stat(targetAbs); err == nil {
		if os.SameFile(info, origInfo) {
			d.planned[targetRel] = candidateRel
			return true, nil
		}

		reason := "existing_file"
		if info.IsDir() {
			reason = "existing_directory"
		}
		summary.AddConflict(Conflict{
			OriginalPath: candidateRel,
			ProposedPath: targetRel,
			Reason:       reason,
		})
		summary.AddWarning(fmt.Sprintf("skipped %s because %s already exists", candidateRel, targetRel))
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	d.planned[targetRel] = candidateRel
	return true, nil
}
