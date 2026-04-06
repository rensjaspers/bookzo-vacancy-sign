#!/usr/bin/env bash
set -euo pipefail

REPO_SLUG="${REPO_SLUG:-rensjaspers/bookzo-vacancy-sign}"
APP_NAME="hotel-rasch-vacancy-pi-universal"
INSTALL_ROOT="${INSTALL_ROOT:-$HOME/$APP_NAME}"
RELEASES_DIR="$INSTALL_ROOT/releases"
CURRENT_LINK="$INSTALL_ROOT/current"
SHARED_CONFIG="$INSTALL_ROOT/config.json"
CONFIG_URL="${CONFIG_URL:-}"
VERSION="${VERSION:-latest}"
TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Commando ontbreekt: $1" >&2
    exit 1
  }
}

asset_url() {
  if [[ "$VERSION" == "latest" ]]; then
    echo "https://github.com/$REPO_SLUG/releases/latest/download/$APP_NAME.zip"
    return
  fi
  echo "https://github.com/$REPO_SLUG/releases/download/$VERSION/$APP_NAME.zip"
}

release_name() {
  if [[ "$VERSION" == "latest" ]]; then
    date +"latest-%Y%m%d%H%M%S"
    return
  fi
  echo "$VERSION"
}

download_bundle() {
  curl -fsSL "$(asset_url)" -o "$TMP_DIR/$APP_NAME.zip"
}

extract_bundle() {
  python3 - "$TMP_DIR/$APP_NAME.zip" "$TMP_DIR" <<'PY'
import sys
import zipfile

with zipfile.ZipFile(sys.argv[1]) as archive:
    archive.extractall(sys.argv[2])
PY
}

bundle_dir() {
  echo "$TMP_DIR/$APP_NAME"
}

download_config() {
  [[ -n "$CONFIG_URL" ]] || return
  curl -fsSL "$CONFIG_URL" -o "$SHARED_CONFIG.tmp"
  mv "$SHARED_CONFIG.tmp" "$SHARED_CONFIG"
}

install_example_config() {
  [[ -f "$SHARED_CONFIG" ]] && return
  cp "$(bundle_dir)/config.json" "$SHARED_CONFIG"
}

link_config() {
  rm -f "$1/config.json"
  ln -s "$SHARED_CONFIG" "$1/config.json"
}

has_placeholder_key() {
  python3 - "$SHARED_CONFIG" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as handle:
    data = json.load(handle)
raise SystemExit(0 if (data.get("apiKey") or "").strip() == "replace-me" else 1)
PY
}

activate_release() {
  local target_dir
  target_dir="$RELEASES_DIR/$(release_name)"
  rm -rf "$target_dir"
  mv "$(bundle_dir)" "$target_dir"
  ln -sfn "$target_dir" "$CURRENT_LINK"
  link_config "$target_dir"
}

start_app() {
  exec bash "$CURRENT_LINK/start.sh"
}

require_command bash
require_command curl
require_command python3
mkdir -p "$RELEASES_DIR"
download_bundle
extract_bundle
download_config
install_example_config
activate_release

if has_placeholder_key; then
  echo "Installatie voltooid." >&2
  echo "Vul eerst je config in: $SHARED_CONFIG" >&2
  echo "Start daarna met: bash $CURRENT_LINK/start.sh" >&2
  exit 0
fi

start_app
