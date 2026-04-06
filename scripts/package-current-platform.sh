#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-}"
APP_NAME="hotel-rasch-vacancy"
GOOS_VALUE="${GOOS:-$(go env GOOS)}"
GOARCH_VALUE="${GOARCH:-$(go env GOARCH)}"
BUNDLE_DIR="$ROOT_DIR/dist/$APP_NAME-$GOOS_VALUE-$GOARCH_VALUE"
ZIP_PATH="$BUNDLE_DIR.zip"

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

if [[ -z "$CONFIG_PATH" ]]; then
  CONFIG_PATH="$(default_config_path)"
fi

FONT_SOURCE="$(python3 - "$CONFIG_PATH" <<'PY'
import json, os, subprocess, sys
with open(sys.argv[1], 'r', encoding='utf-8') as handle:
    data = json.load(handle)
font_path = (data.get('fontPath') or '').strip()
candidates = [
    '/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf',
    '/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf',
    '/System/Library/Fonts/Supplemental/Arial.ttf',
    '/System/Library/Fonts/Supplemental/Arial Bold.ttf',
    '/System/Library/Fonts/Supplemental/Helvetica.ttc',
    '/Library/Fonts/Arial.ttf',
]
if font_path and os.path.exists(font_path):
    print(font_path)
    raise SystemExit(0)
for candidate in candidates:
    if os.path.exists(candidate):
        print(candidate)
        raise SystemExit(0)
try:
    result = subprocess.run(
        [
            'mdfind',
            'kMDItemContentTypeTree == "public.truetype-ttf-font" || kMDItemContentTypeTree == "public.opentype-font"',
        ],
        check=False,
        capture_output=True,
        text=True,
    )
    for line in result.stdout.splitlines():
        if os.path.exists(line):
            print(line)
            raise SystemExit(0)
except FileNotFoundError:
    pass
raise SystemExit('Geen bruikbaar font gevonden voor packaging.')
PY
)"

rm -rf "$BUNDLE_DIR" "$ZIP_PATH"
mkdir -p "$BUNDLE_DIR/fonts"
"$ROOT_DIR/scripts/build-current-platform.sh" "$CONFIG_PATH" "$BUNDLE_DIR/vacancy-board" >/dev/null
cp "$ROOT_DIR/README.md" "$BUNDLE_DIR/"
cp "$FONT_SOURCE" "$BUNDLE_DIR/fonts/board.ttf"
python3 - "$CONFIG_PATH" "$BUNDLE_DIR/config.json" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as handle:
    data = json.load(handle)
data.pop('fontPath', None)
with open(sys.argv[2], 'w', encoding='utf-8') as handle:
    json.dump(data, handle, indent=2)
    handle.write('\n')
PY
(cd "$ROOT_DIR/dist" && zip -rq "$(basename "$ZIP_PATH")" "$(basename "$BUNDLE_DIR")")
echo "$ZIP_PATH"
