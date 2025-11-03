#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="go run"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "$TMP_DIR"
touch "$TMP_DIR/raw_demo 01.txt"
touch "$TMP_DIR/raw_demo 02.txt"

PLAN_JSON="$TMP_DIR/ai-plan.json"

echo "Generating initial AI plan preview..."
$BIN "$ROOT_DIR/main.go" ai --path "$TMP_DIR" --dry-run --export-plan "$PLAN_JSON" >/dev/null

if [[ ! -s "$PLAN_JSON" ]]; then
  echo "expected plan export at $PLAN_JSON" >&2
  exit 1
fi

echo "Editing exported plan for deterministic names..."
python3 - <<'PY'
import json, pathlib
path = pathlib.Path("$PLAN_JSON")
plan = json.loads(path.read_text())
for idx, item in enumerate(plan.get("items", []), start=1):
    item["proposed"] = f"{idx:03d}_final_demo.txt"
plan["warnings"] = plan.get("warnings", []) + ["edited in smoke script"]
path.write_text(json.dumps(plan, indent=2) + "\n")
PY

echo "Validating edited plan (dry run)..."
$BIN "$ROOT_DIR/main.go" ai --path "$TMP_DIR" --dry-run --import-plan "$PLAN_JSON" >/dev/null

echo "Applying edited plan..."
$BIN "$ROOT_DIR/main.go" ai --path "$TMP_DIR" --import-plan "$PLAN_JSON" --yes >/dev/null

if [[ ! -f "$TMP_DIR/001_final_demo.txt" ]]; then
  echo "expected 001_final_demo.txt to exist" >&2
  exit 1
fi
if [[ ! -f "$TMP_DIR/002_final_demo.txt" ]]; then
  echo "expected 002_final_demo.txt to exist" >&2
  exit 1
fi

echo "Undoing AI plan application..."
$BIN "$ROOT_DIR/main.go" undo --path "$TMP_DIR" >/dev/null

if [[ ! -f "$TMP_DIR/raw_demo 01.txt" ]]; then
  echo "undo failed to restore raw_demo 01.txt" >&2
  exit 1
fi
if [[ ! -f "$TMP_DIR/raw_demo 02.txt" ]]; then
  echo "undo failed to restore raw_demo 02.txt" >&2
  exit 1
fi

echo "AI smoke test succeeded."
