package sequence

import (
	"fmt"
	"strconv"
)

// formatNumber zero-pads the provided value using the requested width and returns
// both the padded number string and the width that was ultimately used.
func formatNumber(value, requestedWidth int) (string, int) {
	if value < 0 {
		value = 0
	}
	digitCount := len(strconv.Itoa(value))
	width := requestedWidth
	if width <= 0 || width < digitCount {
		width = digitCount
	}
	return fmt.Sprintf("%0*d", width, value), width
}
