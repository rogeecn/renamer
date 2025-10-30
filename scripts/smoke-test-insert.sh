#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR"
cat <<'FILE' > "$TMP_DIR/holiday.jpg"
photo
FILE
cat <<'FILE' > "$TMP_DIR/trip.jpg"
travel
FILE

echo "Previewing insert..."
$BIN "$ROOT_DIR/main.go" insert 2 _tag --dry-run --path "$TMP_DIR" >/tmp/insert-preview.log
cat /tmp/insert-preview.log

echo "Applying insert..."
$BIN "$ROOT_DIR/main.go" insert 2 _tag --yes --path "$TMP_DIR" >/tmp/insert-apply.log
cat /tmp/insert-apply.log

echo "Undoing insert..."
$BIN "$ROOT_DIR/main.go" undo --path "$TMP_DIR" >/tmp/insert-undo.log
cat /tmp/insert-undo.log

echo "Insert smoke test completed successfully." 
