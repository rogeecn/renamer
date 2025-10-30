#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

touch "$TMP_DIR/photo.jpeg"
touch "$TMP_DIR/poster.JPG"
touch "$TMP_DIR/logo.jpg"

echo "Previewing extension normalization..."
$BIN "$ROOT_DIR/main.go" extension .jpeg .JPG .jpg --path "$TMP_DIR" --dry-run >/dev/null

echo "Applying extension normalization..."
$BIN "$ROOT_DIR/main.go" extension .jpeg .JPG .jpg --path "$TMP_DIR" --yes >/dev/null

if [[ ! -f "$TMP_DIR/photo.jpg" ]]; then
  echo "expected photo.jpg to exist" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/poster.jpg" ]]; then
  echo "expected poster.jpg to exist" >&2
  exit 1
fi

echo "Undoing extension normalization..."
$BIN "$ROOT_DIR/main.go" undo --path "$TMP_DIR" >/dev/null

if [[ ! -f "$TMP_DIR/photo.jpeg" ]]; then
  echo "undo failed to restore photo.jpeg" >&2
  exit 1
fi

if [[ ! -f "$TMP_DIR/poster.JPG" ]]; then
  echo "undo failed to restore poster.JPG" >&2
  exit 1
fi

echo "Extension smoke test succeeded."
