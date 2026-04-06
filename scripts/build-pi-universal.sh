#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-$ROOT_DIR/config.pi.json}"
ARM64_OUTPUT="${2:-$ROOT_DIR/dist/vacancy-board-linux-arm64/vacancy-board}"
ARMV6_OUTPUT="${3:-$ROOT_DIR/dist/vacancy-board-linux-armv6/vacancy-board}"
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

mkdir -p "$(dirname "$ARM64_OUTPUT")" "$(dirname "$ARMV6_OUTPUT")"
"$ROOT_DIR/scripts/build-pi.sh" "$CONFIG_PATH" "$ARM64_OUTPUT" >/dev/null

if [[ "$GOOS_VALUE" == "linux" && "$GOARCH_VALUE" == "arm" ]]; then
  GOOS=linux GOARCH=arm GOARM=6 \
    "$ROOT_DIR/scripts/build-current-platform.sh" "$CONFIG_PATH" "$ARMV6_OUTPUT" >/dev/null
else
  CONFIG_IN_CONTAINER="$(container_path "$CONFIG_PATH")"
  OUTPUT_IN_CONTAINER="$(container_path "$ARMV6_OUTPUT")"
  COMMAND="GOOS=linux GOARCH=arm GOARM=6 ./scripts/build-current-platform.sh \"$CONFIG_IN_CONTAINER\" \"$OUTPUT_IN_CONTAINER\""
  PI_DOCKER_PLATFORM=linux/arm/v7 "$ROOT_DIR/scripts/run-in-pi-container.sh" "$COMMAND" >/dev/null
fi

echo "$ARM64_OUTPUT"
echo "$ARMV6_OUTPUT"
