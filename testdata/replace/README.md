# Replace Command Fixtures

This directory contains lightweight fixtures for manual and automated testing of the `renamer
replace` command.

## Structure

```
replace/
├── case-sensitivity/
│   ├── draft.txt
│   ├── Draft.txt
│   └── README.md
├── multi-pattern/
│   ├── alpha-beta.txt
│   ├── beta-gamma.log
│   ├── nested/
│   │   └── gamma-delta.md
│   └── README.md
└── hidden-files/
    ├── .draft.tmp
    ├── notes.txt
    └── README.md
```

Each subdirectory contains intentionally blank files used to verify preview, apply, and undo
behaviour. Customize or copy these fixtures as needed for new tests.
