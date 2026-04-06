#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-$ROOT_DIR/config.pi.json}"
OUTPUT_PATH="${2:-$ROOT_DIR/dist/vacancy-board-linux-arm64/vacancy-board}"
GOOS_VALUE="${GOOS:-$(go env GOOS)}"
GOARCH_VALUE="${GOARCH:-$(go env GOARCH)}"

container_path() {
  python3 - "$ROOT_DIR" "$1" <<'PY'
import os, sys
root = os.path.realpath(sys.argv[1])
path = os.path.realpath(sys.argv[2])
rel = os.path.relpath(path, root)
if rel.startswith(".."):
    raise SystemExit("Pad moet binnen de repo liggen voor Docker Pi-builds.")
print("/work/" + rel)
PY
}

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "Config niet gevonden: $CONFIG_PATH" >&2
  echo "Maak eerst config.pi.json aan, bijvoorbeeld vanaf config.pi.example.json." >&2
  exit 1
fi

mkdir -p "$(dirname "$OUTPUT_PATH")"

if [[ "$GOOS_VALUE" == "linux" && "$GOARCH_VALUE" == "arm64" ]]; then
  exec "$ROOT_DIR/scripts/build-current-platform.sh" "$CONFIG_PATH" "$OUTPUT_PATH"
fi

CONFIG_IN_CONTAINER="$(container_path "$CONFIG_PATH")"
OUTPUT_IN_CONTAINER="$(container_path "$OUTPUT_PATH")"
COMMAND="./scripts/build-current-platform.sh \"$CONFIG_IN_CONTAINER\" \"$OUTPUT_IN_CONTAINER\""
exec "$ROOT_DIR/scripts/run-in-pi-container.sh" "$COMMAND"
