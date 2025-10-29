#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run ."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/nested"
touch "$TMP_DIR/root.txt"
touch "$TMP_DIR/nested/child.jpg"

LIST_OUTPUT="$($BIN list --path "$TMP_DIR" --recursive --format plain)"
LIST_TOTAL="$(printf '%s' "$LIST_OUTPUT" | awk '/^Total:/ {print $2}')"

if $BIN preview --help >/dev/null 2>&1; then
  PREVIEW_OUTPUT="$($BIN preview --path "$TMP_DIR" --recursive)"
  PREVIEW_TOTAL="$(printf '%s' "$PREVIEW_OUTPUT" | awk '/^Total:/ {print $2}')"

  if [[ "$LIST_TOTAL" != "$PREVIEW_TOTAL" ]]; then
    echo "Mismatch between list and preview candidate counts" >&2
    exit 1
  fi
else
  echo "Preview command not available; parity check skipped." >&2
fi

echo "Smoke test completed. List total: $LIST_TOTAL"
