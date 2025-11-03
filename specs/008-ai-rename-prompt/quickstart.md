# Quickstart: AI-Assisted Rename Prompting

## Prerequisites
- Go 1.24 environment (CLI build/test)
- `*_MODEL_AUTH_TOKEN` stored under `$HOME/.config/.renamer/` (default OpenAI-compatible key)

## Install Dependencies
```bash
# Sync Go modules (includes google/genkit)
go mod tidy
```

## Preview AI Rename Plan
```bash
go run ./cmd/renamer ai \
  --path ./fixtures/batch \
  --sequence-width 3 \
  --sequence-style prefix \
  --naming-casing kebab \
  --banned "promo,ad" \
  --dry-run
```
> CLI invokes the in-process Genkit workflow and renders a preview table with sequential, sanitized names.

## Apply Approved Plan
```bash
go run ./cmd/renamer ai --path ./fixtures/batch --yes
```
> Validates the cached plan, applies filesystem renames, and writes ledger entry with AI metadata.

## Testing
```bash
# Go unit + integration tests (includes Genkit workflow tests)
go test ./...
```

## Troubleshooting
- **Genkit errors**: Run with `--debug-genkit` to emit inline prompt/response traces (written to `~/.renamer/genkit.log`).
- **Validation failures**: Run with `--export-plan out.json` to inspect AI output and manually edit.
- **Rate limits**: Configure `--genkit-model` flag or `GENKIT_MODEL` env variable to select a lighter model.
