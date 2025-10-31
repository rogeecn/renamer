#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN=(go run)
TMP_DIR="$(mktemp -d)"
WORK_DIR="$TMP_DIR/workspace"
PREVIEW_LOG=""
SCOPE_LOG=""
APPLY_LOG=""
UNDO_LOG=""

cleanup() {
  rm -rf "$TMP_DIR"
  for log in "$PREVIEW_LOG" "$SCOPE_LOG" "$APPLY_LOG" "$UNDO_LOG"; do
    [[ -n "${log:-}" ]] && rm -f "$log"
  done
}
trap cleanup EXIT

mkdir -p "$WORK_DIR" "$WORK_DIR/artifacts"

# Seed workspace with fixtures referenced in quickstart docs.
cp -R "$ROOT_DIR/tests/fixtures/regex/baseline/." "$WORK_DIR/"
cp "$ROOT_DIR/tests/fixtures/regex/mixed/feature-demo_2025-10-01.txt" "$WORK_DIR/"
cp "$ROOT_DIR/tests/fixtures/regex/mixed/build_101_release.tar.gz" "$WORK_DIR/artifacts/build_101_vrelease.tar.gz"
cp "$ROOT_DIR/tests/fixtures/regex/mixed/build_102_hotfix.tar.gz" "$WORK_DIR/artifacts/build_102_vhotfix.tar.gz"
mkdir -p "$WORK_DIR/artifacts/build_103_varchive"
printf 'Quarterly summary\n' >"$WORK_DIR/2025-01_report.txt"

PREVIEW_LOG="$(mktemp)"
SCOPE_LOG="$(mktemp)"
APPLY_LOG="$(mktemp)"
UNDO_LOG="$(mktemp)"

run_cli() {
  local log="$1"
  shift
  "${BIN[@]}" "$ROOT_DIR/main.go" "$@" >"$log"
  cat "$log"
}

echo "Quickstart #1: Preview captured group substitution (--dry-run)."
run_cli "$PREVIEW_LOG" regex '^(\\d{4})-(\\d{2})_(.*)$' 'Q@2-@1_@3' --dry-run --path "$WORK_DIR"
if ! grep -q 'Q01-2025_report.txt' "$PREVIEW_LOG"; then
  echo "Expected preview rename for 2025-01_report.txt missing." >&2
  exit 1
fi

echo
echo "Quickstart #2: Scope-limited preview with extensions and directories."
run_cli "$SCOPE_LOG" regex '^(build)_(\\d+)_v(.*)$' 'release-@2-@1-v@3' --dry-run --path "$WORK_DIR/artifacts" --extensions '.zip|.tar.gz' --include-dirs
if ! grep -q 'release-101-build-vrelease.tar.gz' "$SCOPE_LOG"; then
  echo "Expected scoped preview for build artifacts missing." >&2
  exit 1
fi

echo
echo "Quickstart #3: Apply regex rename non-interactively (--yes)."
run_cli "$APPLY_LOG" regex '^(feature)-(.*)$' '@2-@1' --yes --path "$WORK_DIR"
if ! [[ -f "$WORK_DIR/demo_2025-10-01-feature.txt" ]]; then
  echo "Applied rename did not produce demo_2025-10-01-feature.txt." >&2
  exit 1
fi

echo
echo "Quickstart #4: Undo the latest regex batch."
run_cli "$UNDO_LOG" undo --path "$WORK_DIR"
if ! [[ -f "$WORK_DIR/feature-demo_2025-10-01.txt" ]]; then
  echo "Undo did not restore feature-demo_2025-10-01.txt." >&2
  exit 1
fi

echo
echo "Regex smoke test completed successfully."
