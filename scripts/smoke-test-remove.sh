#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR/nested"
touch "$TMP_DIR/report copy draft.txt"
touch "$TMP_DIR/nested/notes draft.txt"

echo "Previewing removals..."
$BIN "$ROOT_DIR/main.go" remove " copy" " draft" --path "$TMP_DIR" --recursive --dry-run >/dev/null

if [[ ! -f "$TMP_DIR/report copy draft.txt" ]]; then
  echo "preview should not modify files" >&2
  exit 1
fi

echo "Applying removals..."
$BIN "$ROOT_DIR/main.go" remove " copy" " draft" --path "$TMP_DIR" --recursive --yes >/dev/null

if [[ ! -f "$TMP_DIR/report.txt" ]]; then
  echo "expected report.txt to exist after removal" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/nested/notes.txt" ]]; then
  echo "expected nested/notes.txt to exist after removal" >&2
  exit 1
fi

if [[ -f "$TMP_DIR/report copy draft.txt" ]]; then
  echo "source file still exists after apply" >&2
  exit 1
fi

echo "Undoing removals..."
$BIN "$ROOT_DIR/main.go" undo --path "$TMP_DIR" >/dev/null

if [[ ! -f "$TMP_DIR/report copy draft.txt" ]]; then
  echo "undo failed to restore report copy draft.txt" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/nested/notes draft.txt" ]]; then
  echo "undo failed to restore nested/notes draft.txt" >&2
  exit 1
fi

if [[ -f "$TMP_DIR/report.txt" ]]; then
  echo "undo failed to clean up report.txt" >&2
  exit 1
fi

echo "Remove smoke test succeeded."
