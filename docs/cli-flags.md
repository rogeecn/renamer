# CLI Scope Flags

Renamer shares a consistent set of scope flags across every command that inspects or mutates the
filesystem. Use these options at the root command level so they apply to all subcommands (`list`,
`replace`, `insert`, `remove`, etc.).

| Flag | Default | Description |
|------|---------|-------------|
| `--path` | `.` | Working directory root for traversal. |
| `-r`, `--recursive` | `false` | Traverse subdirectories depth-first. Symlinked directories are not followed. |
| `-d`, `--include-dirs` | `false` | Include directories in results. |
| `-e`, `--extensions` | *(none)* | Pipe-separated list of file extensions (e.g. `.jpg|.mov`). Tokens must start with a dot, are lowercased internally, and duplicates are ignored. |
| `--hidden` | `false` | Include dot-prefixed files and directories. By default they are excluded from listings and rename previews. |
| `--yes` | `false` | Apply changes without interactive confirmation (mutating commands only). |
| `--dry-run` | `false` | Force preview-only behavior even when `--yes` is supplied. |
| `--format` | `table` | Command-specific output formatting option. For `list`, use `table` or `plain`. |

## Regex Command Quick Reference

```bash
renamer regex <pattern> <template> [flags]
```

- Patterns compile with Go’s RE2 engine and are matched against filename stems; invalid expressions fail fast with helpful errors.
- Templates support numbered placeholders (`@0`, `@1`, …) along with escaped `@@` for literal at-signs; undefined captures block the run.
- Preview mode (`--dry-run`, default) renders the rename plan with change/skipped/conflict statuses; apply with `--yes` writes a ledger entry for undo.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) control candidate discovery just like other commands, and conflicts or empty targets exit non-zero.

### Usage Examples

- Preview captured group swapping: `renamer regex "^(\w+)-(\d+)" "@2_@1" --dry-run --path ./samples`
- Limit by extensions and directories: `renamer regex '^(build)_(\d+)_v(.*)$' 'release-@2-@1-v@3' --extensions .zip|.tar.gz --include-dirs --recursive`
- Automation-friendly apply with undo: `renamer regex '^(feature)-(.*)$' '@2-@1' --yes --path ./staging && renamer undo --path ./staging`

## Insert Command Quick Reference

```bash
renamer insert <position> <text> [flags]
```

- Position tokens:
  - `^` inserts at the beginning of the filename. Append a number (`^3` or just `3`) to insert after the third rune of the stem.
  - `$` inserts immediately before the extension dot (or end if no extension). Append a number (`1$`) to count backward from the end of the stem (e.g., `1$` inserts before the final rune).
- Text must be valid UTF-8 without path separators or control characters; Unicode characters are supported.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) limit the candidate set before insertion.
- `--dry-run` previews the plan; rerun with `--yes` to apply the same operations.

### Usage Examples

- Preview adding a prefix: `renamer insert ^ "[2025] " --dry-run`
- Append before extension: `renamer insert $ _ARCHIVE --yes --path ./docs`
- Insert before the final character: `renamer insert 1$ _TAIL --path ./images --dry-run`
- Insert after third character in stem: `renamer insert 3 _tag --path ./images --dry-run`
- Combine with extension filter: `renamer insert ^ "v1_" --extensions .txt|.md`

## Sequence Command Quick Reference

```bash
renamer sequence [flags]
```

- Applies deterministic numbering to filenames using the active scope filters; preview-first by default.
- Default behavior prepends a three-digit number using an underscore separator (e.g. `001_name.ext`).
- Flags:
  - `--start` (default `1`) sets the initial sequence value (must be ≥1).
  - `--width` (optional) enforces minimum digit width with zero padding; the command auto-expands and warns when more digits are required.
  - `--placement` (`suffix` default, `prefix` alternative) controls whether numbers prepend or append the stem.
  - `--separator` customizes the string placed between the stem and number; path separators are rejected.
  - `--number-prefix` / `--number-suffix` add static text directly before or after the digits (use with `--placement prefix` for labelled sequences such as `seq001-file.ext`).
  - Set `--separator ""` to remove the underscore separator when prefixing numbers (e.g. `seq001file.ext`).
- Conflicting targets are skipped with warnings while remaining files continue numbering; directories included via `--include-dirs` are listed but unchanged.

## Remove Command Quick Reference

```bash
renamer remove <token1> [token2 ...] [flags]
```

- Removal tokens are evaluated in the order supplied. Each token deletes literal substrings from the
  current filename before the next token runs; results are previewed before any filesystem changes.
- Duplicate tokens are deduplicated automatically and surfaced as warnings so users can adjust
  scripts without surprises.
- Tokens that collapse a filename to an empty string are skipped with warnings during preview/apply
  to protect against accidental deletion.
- All scope flags (`--path`, `-r`, `-d`, `--hidden`, `-e`) apply, making it easy to target directories,
  recurse, and limit removals by extension.
- Use `--dry-run` for automation previews and combine with `--yes` to apply unattended; conflicting
  combinations (`--dry-run --yes`) exit with an error to uphold preview-first safety.

### Usage Examples

- Preview sequential removals: `renamer remove " copy" " draft" --dry-run`
- Remove tokens recursively: `renamer remove foo foo- --recursive --path ./reports`
- Combine with extension filters: `renamer remove " Project" --extensions .txt|.md --dry-run`

## Extension Command Quick Reference

```bash
renamer extension <source-ext...> <target-ext> [flags]
```

- Provide one or more dot-prefixed source extensions followed by the target extension. Validation
  fails if any token omits the leading dot or repeats the target exactly.
- Source extensions are normalized case-insensitively; duplicates and no-op tokens are surfaced as
  warnings in the preview rather than silently ignored.
- Case variants of the target extension (for example `.JPG` when targeting `.jpg`) remain untouched
  unless you include them in the source list, ensuring casing changes happen only when explicitly
  requested.
- Preview output lists every candidate with `changed`, `no change`, or `skipped` status so scripts
  can detect conflicts before applying. Conflicting targets block apply and exit with a non-zero
  code.
- Scope flags (`--path`, `-r`, `-d`, `--hidden`, `--extensions`) determine which files and
  directories participate. Hidden assets remain excluded unless `--hidden` is supplied.
- `--dry-run` (default) prints the plan without touching the filesystem. Re-run with `--yes` to
  apply; attempting to combine both flags exits with an error. When no files match, the command
  exits `0` after printing “No candidates found.”

### Usage Examples

- Preview normalization: `renamer extension .jpeg .JPG .jpg --dry-run`
- Apply case-folded extension updates: `renamer extension .yaml .yml .yml --yes --path ./configs`
- Include hidden assets recursively: `renamer extension .TMP .tmp --recursive --hidden`

## AI Command Secrets

- AI model authentication tokens are loaded from `$HOME/.config/.renamer/<MODEL>_MODEL_AUTH_TOKEN`. The default model token file is `default_MODEL_AUTH_TOKEN`, but any `--genkit-model` override maps to the same naming scheme.
- Token files must contain only the raw API key with no extra whitespace; restrictive permissions (owner read/write) are recommended to keep credentials private.

### AI Command Flags

- `--genkit-model <id>` overrides the default OpenAI-compatible model used by the embedded Genkit workflow. When omitted, `gpt-4o-mini` is used.
- `--debug-genkit` streams prompt/response telemetry (including prompt hashes and warnings) to stderr so you can archive the exchange for auditing.
- `--export-plan <path>` writes the exact AI response (prompt hash, model, warnings, and proposed items) to a JSON file. The same file can be edited and re-imported to tweak filenames before applying.
- `--import-plan <path>` loads a previously exported or manually curated JSON plan. The CLI re-validates all entries before previewing or applying changes.
- `--naming-casing <style>` enforces a casing policy (`kebab`, `snake`, `camel`, `pascal`, `title`). Banned tokens, prefix rules, and spacing requirements are evaluated against the imported or generated plan.
- `--naming-prefix`, `--naming-allow-spaces`, `--naming-keep-order`, and `--banned` extend the policy envelope that both the prompt and validator obey.
- `--yes` applies the currently loaded plan. Without `--yes`, the command remains in preview mode even when you import a plan.

> Tip: Run `renamer ai --path ./fixtures --dry-run --export-plan plan.json` to capture the initial draft, edit the JSON file, then `renamer ai --path ./fixtures --import-plan plan.json --yes` to apply the curated result.
