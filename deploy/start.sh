#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ARCH="$(uname -m)"

case "$ARCH" in
  aarch64|arm64)
    BINARY_PATH="$ROOT_DIR/bin/vacancy-board-linux-arm64"
    ;;
  armv6l|armv7l|armv8l|arm)
    BINARY_PATH="$ROOT_DIR/bin/vacancy-board-linux-armv6"
    ;;
  *)
    echo "Niet-ondersteunde architectuur: $ARCH" >&2
    echo "Gebruik arm64 of 32-bit ARM (armv6/armv7)." >&2
    exit 1
    ;;
esac

if [[ ! -x "$BINARY_PATH" ]]; then
  echo "Binary niet gevonden of niet uitvoerbaar: $BINARY_PATH" >&2
  exit 1
fi

if command -v ldd >/dev/null 2>&1; then
  if ldd "$BINARY_PATH" 2>&1 | grep -q "not found"; then
    echo "Vereiste SDL libraries ontbreken op deze Raspberry Pi." >&2
    echo "Installeer minstens: libsdl2-2.0-0 en libsdl2-ttf-2.0-0" >&2
    exit 1
  fi
fi

cd "$ROOT_DIR"
exec "$BINARY_PATH" -config "$ROOT_DIR/config.json"
