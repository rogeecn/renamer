# Remove Command Scenario Fixtures

Use this directory to capture realistic file layouts for the `renamer remove` tests.

## Layout

```
testdata/remove/basic
├── project copy draft.txt
├── Project copy.txt
├── nested
│   ├── foo draft.txt
│   ├── foo draft draft.txt
│   └── hidden
│       └── .draft copy.md
├── collisions
│   ├── alpha draft.txt
│   └── alpha.txt
└── empty-basename
    ├── draft
    └── draft.txt
```

## Scenarios Covered

- Sequential removals (`" copy"`, `" draft"`) collapsing multiple terms.
- Duplicate token handling in nested directories.
- Hidden file interactions requiring the `--hidden` flag.
- Name collisions after removal (should report conflict prior to apply).
- Entries that would become empty names, verifying they are skipped with warnings.

Adjust or extend files as additional edge cases are added to integration tests.
