#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-}"

"$ROOT_DIR/scripts/build-current-platform.sh" "$CONFIG_PATH" "$ROOT_DIR/vacancy-board" >/dev/null
exec "$ROOT_DIR/vacancy-board"
