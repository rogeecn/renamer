#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${RENAMER_AI_KEY:-}" ]]; then
  echo "RENAMER_AI_KEY must be set" >&2
  exit 1
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/nested"
touch "$tmp/IMG_0001.jpg"
touch "$tmp/nested/video01.mp4"

echo "Previewing AI suggestions..."
renamer ai --path "$tmp" --prompt "Smoke demo" --dry-run <<<'q'

echo "Applying AI suggestions..."
renamer ai --path "$tmp" --prompt "Smoke demo" --yes

echo "Undoing last AI batch..."
renamer undo --path "$tmp"

echo "Smoke test completed."
