package output

import "fmt"

// Format identifiers mirrored from listing package to avoid import cycle.
const (
	FormatTable = "table"
	FormatPlain = "plain"
)

// NewFormatter selects the appropriate renderer based on format key.
func NewFormatter(format string) (Formatter, error) {
	switch format {
	case FormatPlain:
		return NewPlainFormatter(), nil
	case FormatTable, "":
		return NewTableFormatter(), nil
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}
