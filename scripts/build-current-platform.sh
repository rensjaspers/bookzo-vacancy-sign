#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-}"
OUTPUT_PATH="${2:-$ROOT_DIR/vacancy-board}"

default_config_path() {
  if [[ -f "$ROOT_DIR/config.json" ]]; then
    echo "$ROOT_DIR/config.json"
    return
  fi
  if [[ -f "$ROOT_DIR/config.pi.json" ]]; then
    echo "$ROOT_DIR/config.pi.json"
    return
  fi
  echo "Geen config gevonden. Gebruik config.json of config.pi.json." >&2
  exit 1
}

ensure_pkg_config_path() {
  if [[ -d /opt/homebrew/lib/pkgconfig ]]; then
    export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
  fi
}

embedded_config_base64() {
  python3 - "$1" <<'PY'
import base64, json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as handle:
    data = json.load(handle)
payload = json.dumps(data, separators=(',', ':')).encode('utf-8')
print(base64.b64encode(payload).decode('ascii'))
PY
}

if [[ -z "$CONFIG_PATH" ]]; then
  CONFIG_PATH="$(default_config_path)"
fi

ensure_pkg_config_path
source "$ROOT_DIR/scripts/cgo-darwin-no-dup-lib-warn.sh"
EMBEDDED_CONFIG_BASE64="$(embedded_config_base64 "$CONFIG_PATH")"
go build \
  -buildvcs=false \
  -tags sdl \
  -ldflags "-X main.embeddedConfigBase64=$EMBEDDED_CONFIG_BASE64" \
  -o "$OUTPUT_PATH" \
  ./cmd/vacancy-board
echo "$OUTPUT_PATH"
