#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

echo "==> Verifying modules..."
go mod verify

echo "==> Running vet..."
go vet ./...

echo "==> Building binary..."
mkdir -p dist
go build -o dist/vplc ./cmd/vplc

echo "==> Smoke: version check..."
./dist/vplc --version

echo "==> Smoke: config validation..."
./dist/vplc --validate-config --config configs/default.json

echo "==> Running smoke tests..."
go test -run TestSmoke -v ./tests/smoke/...

echo "All smoke checks passed."
