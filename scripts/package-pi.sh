#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-$ROOT_DIR/config.pi.json}"
GOOS_VALUE="${GOOS:-$(go env GOOS)}"
GOARCH_VALUE="${GOARCH:-$(go env GOARCH)}"

container_path() {
  python3 - "$ROOT_DIR" "$1" <<'PY'
import os, sys
root = os.path.realpath(sys.argv[1])
path = os.path.realpath(sys.argv[2])
rel = os.path.relpath(path, root)
if rel.startswith(".."):
    raise SystemExit("Pad moet binnen de repo liggen voor Docker Pi-packaging.")
print("/work/" + rel)
PY
}

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "Config niet gevonden: $CONFIG_PATH" >&2
  echo "Maak eerst config.pi.json aan, bijvoorbeeld vanaf config.pi.example.json." >&2
  exit 1
fi

if [[ "$GOOS_VALUE" != "linux" || "$GOARCH_VALUE" != "arm64" ]]; then
  CONFIG_IN_CONTAINER="$(container_path "$CONFIG_PATH")"
  COMMAND="./scripts/package-current-platform.sh \"$CONFIG_IN_CONTAINER\""
  exec "$ROOT_DIR/scripts/run-in-pi-container.sh" "$COMMAND"
fi

exec "$ROOT_DIR/scripts/package-current-platform.sh" "$CONFIG_PATH"
