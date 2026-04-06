#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG_PATH="${1:-$ROOT_DIR/config.pi.json}"
APP_NAME="hotel-rasch-vacancy-pi-universal"
BUNDLE_DIR="$ROOT_DIR/dist/$APP_NAME"
ZIP_PATH="$BUNDLE_DIR.zip"
ARM64_BUILD="$ROOT_DIR/dist/vacancy-board-linux-arm64/vacancy-board"
ARMV6_BUILD="$ROOT_DIR/dist/vacancy-board-linux-armv6/vacancy-board"

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "Config niet gevonden: $CONFIG_PATH" >&2
  echo "Maak eerst config.pi.json aan, bijvoorbeeld vanaf config.pi.example.json." >&2
  exit 1
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
mkdir -p "$BUNDLE_DIR/bin" "$BUNDLE_DIR/fonts"
"$ROOT_DIR/scripts/build-pi-universal.sh" "$CONFIG_PATH" "$ARM64_BUILD" "$ARMV6_BUILD" >/dev/null
cp "$ARM64_BUILD" "$BUNDLE_DIR/bin/vacancy-board-linux-arm64"
cp "$ARMV6_BUILD" "$BUNDLE_DIR/bin/vacancy-board-linux-armv6"
cp "$ROOT_DIR/deploy/start.sh" "$BUNDLE_DIR/start.sh"
cp "$ROOT_DIR/README.md" "$BUNDLE_DIR/"
cp "$FONT_SOURCE" "$BUNDLE_DIR/fonts/board.ttf"
chmod +x "$BUNDLE_DIR/start.sh" "$BUNDLE_DIR/bin/vacancy-board-linux-arm64" "$BUNDLE_DIR/bin/vacancy-board-linux-armv6"
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
