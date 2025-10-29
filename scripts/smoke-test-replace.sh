#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/nested"
touch "$TMP_DIR/foo_draft.txt"
touch "$TMP_DIR/nested/bar_draft.txt"

echo "Previewing replacements..."
$BIN "$ROOT_DIR/main.go" replace draft Draft final --path "$TMP_DIR" --recursive --dry-run >/dev/null

echo "Applying replacements..."
$BIN "$ROOT_DIR/main.go" replace draft Draft final --path "$TMP_DIR" --recursive --yes >/dev/null

if [[ ! -f "$TMP_DIR/foo_final.txt" ]]; then
  echo "expected foo_final.txt to exist" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/nested/bar_final.txt" ]]; then
  echo "expected bar_final.txt to exist" >&2
  exit 1
fi

echo "Undoing replacements..."
$BIN "$ROOT_DIR/main.go" undo --path "$TMP_DIR" >/dev/null

if [[ ! -f "$TMP_DIR/foo_draft.txt" ]]; then
  echo "undo failed to restore foo_draft.txt" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/nested/bar_draft.txt" ]]; then
  echo "undo failed to restore nested/bar_draft.txt" >&2
  exit 1
fi

echo "Smoke test succeeded."
