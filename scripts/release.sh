#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

VERSION="${1:-0.1.0}"
GOOS="${2:-linux}"
GOARCH="${3:-amd64}"
OUTPUT_DIR="release"

echo "==> Building release: v${VERSION} for ${GOOS}/${GOARCH}"
mkdir -p "${OUTPUT_DIR}"

GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -ldflags "-X github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/app.Version=${VERSION} -s -w" \
    -o "${OUTPUT_DIR}/vplc-${VERSION}-${GOOS}-${GOARCH}" \
    ./cmd/vplc

echo "==> Release binary: ${OUTPUT_DIR}/vplc-${VERSION}-${GOOS}-${GOARCH}"
