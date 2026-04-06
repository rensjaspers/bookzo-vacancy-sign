#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
HOST_UID="$(id -u)"
HOST_GID="$(id -g)"
COMMAND="${1:?Geef een commando mee voor de Pi-container}"
PI_DOCKER_PLATFORM="${PI_DOCKER_PLATFORM:-linux/arm64}"
PLATFORM_TAG="$(echo "$PI_DOCKER_PLATFORM" | tr '/:' '--')"
IMAGE_NAME="bookzo-vacancy-sign-pi-builder:${PLATFORM_TAG}-bookworm"

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is nodig om vanaf macOS een Pi-binary te bouwen." >&2
  exit 1
fi

docker build --platform "$PI_DOCKER_PLATFORM" -t "$IMAGE_NAME" -f "$ROOT_DIR/docker/pi-builder.Dockerfile" "$ROOT_DIR" >/dev/null

docker run --rm --platform "$PI_DOCKER_PLATFORM" \
  -e GOCACHE=/tmp/gocache \
  -e HOST_UID="$HOST_UID" \
  -e HOST_GID="$HOST_GID" \
  -v "$ROOT_DIR:/work" \
  -w /work \
  "$IMAGE_NAME" \
  sh -lc "
    set -e
    export PATH=/usr/local/go/bin:\$PATH
    set +e
    $COMMAND
    status=\$?
    set -e
    chown -R \$HOST_UID:\$HOST_GID /work/dist /work/vacancy-board /work/render-spike 2>/dev/null || true
    exit \$status
  "
