#!/usr/bin/env bash
set -euo pipefail

REPO_SLUG="${REPO_SLUG:-rensjaspers/bookzo-vacancy-sign}"
APP_NAME="bookzo-vacancy-sign-pi-universal"
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

restore_execute_bits() {
  chmod +x "$CURRENT_LINK/start.sh"
  chmod +x "$CURRENT_LINK"/bin/*
}

pi_binary_in_bundle() {
  case "$(uname -m)" in
  aarch64 | arm64) echo "$CURRENT_LINK/bin/vacancy-board-linux-arm64" ;;
  armv6l | armv7l | armv8l | arm) echo "$CURRENT_LINK/bin/vacancy-board-linux-armv6" ;;
  *) echo "" ;;
  esac
}

sdl_runtime_complete() {
  [[ "$(uname -s)" == "Linux" ]] || return 0
  local bin_path
  bin_path="$(pi_binary_in_bundle)"
  [[ -n "$bin_path" ]] || return 0
  [[ -x "$bin_path" ]] || return 1
  command -v ldd >/dev/null 2>&1 || return 0
  ! ldd "$bin_path" 2>&1 | grep -q "not found"
}

runtime_report() {
  local bin_path
  bin_path="$(pi_binary_in_bundle)"
  [[ -n "$bin_path" && -x "$bin_path" ]] || return 0
  command -v ldd >/dev/null 2>&1 || return 0
  ldd "$bin_path" 2>&1 || true
}

glibc_too_old() {
  grep -q "GLIBC_[0-9.].*not found"
}

glibc_hint() {
  echo "Deze Raspberry Pi OS-installatie is te oud voor deze release-binary." >&2
  echo "De benodigde glibc-versie ontbreekt." >&2
  echo "Gebruik bij voorkeur een nieuwere Raspberry Pi OS image." >&2
  echo "Een herflash is meestal sneller en betrouwbaarder dan upgraden vanaf Jessie." >&2
}

sdl_apt_hint() {
  echo "Vereiste SDL-bibliotheek ontbreekt (nodig om het bord te tonen)." >&2
  echo "Installeer op Debian/Ubuntu/Raspberry Pi OS bijvoorbeeld met:" >&2
  echo "  sudo apt-get update && sudo apt-get install -y libsdl2-2.0-0 libsdl2-ttf-2.0-0" >&2
}

offer_sdl_apt_install() {
  local answer
  [[ -r /dev/tty && -w /dev/tty ]] || return 1
  read -r -p "Wil je dat nu proberen te installeren met apt? [j/N] " answer < /dev/tty >&2 || return 1
  [[ "${answer:-}" == [jJ]* ]]
}

run_sdl_apt_install() {
  sudo apt-get update && sudo apt-get install -y libsdl2-2.0-0 libsdl2-ttf-2.0-0
}

ensure_sdl_runtime_linux() {
  [[ "$(uname -s)" == "Linux" ]] || return 0
  local report
  report="$(runtime_report)"
  if [[ -n "$report" ]] && printf '%s\n' "$report" | glibc_too_old; then
    glibc_hint
    exit 1
  fi
  sdl_runtime_complete && return 0
  sdl_apt_hint
  if offer_sdl_apt_install && command -v sudo >/dev/null 2>&1 && command -v apt-get >/dev/null 2>&1; then
    run_sdl_apt_install || true
  fi
  report="$(runtime_report)"
  if [[ -n "$report" ]] && printf '%s\n' "$report" | glibc_too_old; then
    glibc_hint
    exit 1
  fi
  sdl_runtime_complete && return 0
  echo "Installeer de pakketten hierboven en voer dit script daarna opnieuw uit of start met: bash $CURRENT_LINK/start.sh" >&2
  exit 1
}

activate_release() {
  local target_dir
  target_dir="$RELEASES_DIR/$(release_name)"
  rm -rf "$target_dir"
  mv "$(bundle_dir)" "$target_dir"
  ln -sfn "$target_dir" "$CURRENT_LINK"
  link_config "$target_dir"
  restore_execute_bits
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
ensure_sdl_runtime_linux

if has_placeholder_key; then
  echo "Installatie voltooid." >&2
  echo "Vul eerst je config in: $SHARED_CONFIG" >&2
  echo "Start daarna met: bash $CURRENT_LINK/start.sh" >&2
  exit 0
fi

start_app
