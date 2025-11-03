//go:build tools

package tools

// This file ensures Go modules keep the Genkit dependency pinned even before
// runtime wiring lands.
import (
	_ "github.com/firebase/genkit/go/genkit"
)
