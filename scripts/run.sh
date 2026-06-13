#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."
go run ./cmd/vplc --run --config configs/default.json
