package history

// BuildAIMetadata constructs ledger metadata for AI-driven rename batches.
func BuildAIMetadata(prompt string, promptHistory []string, notes []string, model string, warnings []string) map[string]any {
    data := map[string]any{
        "prompt":   prompt,
        "model":    model,
        "flow":     "renameFlow",
        "warnings": warnings,
    }

    if len(promptHistory) > 0 {
        data["promptHistory"] = append([]string(nil), promptHistory...)
    }

    if len(notes) > 0 {
        data["notes"] = append([]string(nil), notes...)
    }

    return data
}
