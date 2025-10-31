# Regex Command Scenario Fixtures

This directory provides ready-to-run datasets for validating the `renamer regex`
command against realistic workflows. Copy a scenario to a temporary directory
before mutating files so the repository remains clean:

```bash
TMP_DIR=$(mktemp -d)
cp -R testdata/regex/capture-groups/* "$TMP_DIR/"
go run ./main.go regex '^(\w+)-(\d+)$' '@2_@1' --dry-run --path "$TMP_DIR"
```

## Structure

```
regex/
├── capture-groups/
│   ├── alpha-123.log
│   ├── beta-456.log
│   ├── gamma-789.log
│   └── notes.txt
├── automation/
│   ├── build_101_release.tar.gz
│   ├── build_102_hotfix.tar.gz
│   ├── build_103_varchive/
│   │   └── placeholder.txt
│   └── feature-demo_2025-10-01.txt
└── validation/
    ├── duplicate-a-01.txt
    ├── duplicate-b-01.txt
    └── group-miss.txt
```

### Scenario Highlights

- **capture-groups** – Mirrors the quickstart preview example. Run with
  `renamer regex '^(\w+)-(\d+)$' '@2_@1' --dry-run` to verify captured groups
  swap safely while non-matching files remain untouched.
- **automation** – Supports end-to-end `--yes` applies and undo. Use
  `renamer regex '^(feature)-(.*)$' '@2-@1' --yes` to exercise ledger writes
  and `renamer regex '^(build)_(\d+)_v(.*)$' 'release-@2-@1-v@3'` to combine
  extension filtering with directory handling.
- **validation** – Surfaces error cases. Applying
  `renamer regex '^(duplicate)-(.*)-(\d+)$' '@1-@3' --yes` should report a
  duplicate-target conflict, while referencing `@2` with
  `renamer regex '^(.+)$' '@2' --dry-run` raises an undefined group error.

Extend these fixtures as new edge cases or regression scenarios arise.
