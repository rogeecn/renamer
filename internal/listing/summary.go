package listing

import "strings"

// EmptyResultMessage returns a contextual message when no entries match.
func EmptyResultMessage(req *ListingRequest) string {
	if req == nil {
		return "No entries matched the provided filters."
	}

	if len(req.Extensions) > 0 {
		return "No entries matched extensions: " + strings.Join(req.Extensions, ", ")
	}

	if req.IncludeHidden {
		return "No entries matched the provided filters (including hidden files)."
	}

	return "No entries matched the provided filters."
}
