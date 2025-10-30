# Extension Command Test Data

This directory contains sample files for manual and automated validation of the
`renamer extension` workflow. To avoid mutating source fixtures directly, copy
the `sample/` folder to a temporary location before running commands that
perform real filesystem changes.

Example usage:

```bash
TMP_DIR=$(mktemp -d)
cp -R testdata/extension/sample/* "$TMP_DIR/"
go run ./main.go extension .jpeg .JPG .jpg --path "$TMP_DIR" --dry-run
```

The sample set includes:

- Mixed-case `.jpeg`/`.JPG` extensions
- Nested directories (`sample/nested`)
- An already-normalized `.jpg` file
- A hidden file to verify `--hidden` opt-in behavior
