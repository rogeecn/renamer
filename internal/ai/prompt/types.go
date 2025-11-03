package prompt

// RenamePrompt captures the structured payload sent to the Genkit workflow.
type RenamePrompt struct {
	WorkingDir   string             `json:"workingDir"`
	Samples      []PromptSample     `json:"samples"`
	TotalCount   int                `json:"totalCount"`
	SequenceRule SequenceRuleConfig `json:"sequenceRule"`
	Policies     NamingPolicyConfig `json:"policies"`
	BannedTerms  []string           `json:"bannedTerms,omitempty"`
	Metadata     map[string]string  `json:"metadata,omitempty"`
}

// PromptSample represents a sampled file from the traversal scope.
type PromptSample struct {
	OriginalName string `json:"originalName"`
	Extension    string `json:"extension"`
	SizeBytes    int64  `json:"sizeBytes"`
	PathDepth    int    `json:"pathDepth"`
}

// SequenceRuleConfig captures numbering directives for the AI prompt.
type SequenceRuleConfig struct {
	Style     string `json:"style"`
	Width     int    `json:"width"`
	Start     int    `json:"start"`
	Separator string `json:"separator"`
}

// NamingPolicyConfig enumerates naming policies forwarded to the AI.
type NamingPolicyConfig struct {
	Prefix            string   `json:"prefix,omitempty"`
	Casing            string   `json:"casing"`
	AllowSpaces       bool     `json:"allowSpaces,omitempty"`
	KeepOriginalOrder bool     `json:"keepOriginalOrder,omitempty"`
	ForbiddenTokens   []string `json:"forbiddenTokens,omitempty"`
}

// RenameResponse is the structured payload expected from the AI model.
type RenameResponse struct {
	Items      []RenameItem `json:"items"`
	Warnings   []string     `json:"warnings,omitempty"`
	PromptHash string       `json:"promptHash,omitempty"`
	Model      string       `json:"model,omitempty"`
}

// RenameItem maps an original path to the AI-proposed rename.
type RenameItem struct {
	Original string `json:"original"`
	Proposed string `json:"proposed"`
	Sequence int    `json:"sequence"`
	Notes    string `json:"notes,omitempty"`
}
