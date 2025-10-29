package remove

import (
	"fmt"

	"github.com/rogeecn/renamer/internal/listing"
)

// Request encapsulates the options required for remove operations.
// It mirrors the listing scope so preview/apply flows stay consistent.
type Request struct {
	WorkingDir         string
	Tokens             []string
	IncludeDirectories bool
	Recursive          bool
	IncludeHidden      bool
	Extensions         []string
}

// FromListing builds a Request from the shared listing scope plus ordered tokens.
func FromListing(scope *listing.ListingRequest, tokens []string) (*Request, error) {
	if scope == nil {
		return nil, fmt.Errorf("scope must not be nil")
	}
	req := &Request{
		WorkingDir:         scope.WorkingDir,
		IncludeDirectories: scope.IncludeDirectories,
		Recursive:          scope.Recursive,
		IncludeHidden:      scope.IncludeHidden,
		Extensions:         append([]string(nil), scope.Extensions...),
		Tokens:             append([]string(nil), tokens...),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

// Validate ensures the request has the required data before traversal happens.
func (r *Request) Validate() error {
	if r.WorkingDir == "" {
		return fmt.Errorf("working directory must be provided")
	}
	if len(r.Tokens) == 0 {
		return fmt.Errorf("at least one removal token is required")
	}
	for i, token := range r.Tokens {
		if token == "" {
			return fmt.Errorf("token at position %d is empty after trimming", i)
		}
	}
	return nil
}
