#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."
mkdir -p dist
go build -o dist/vplc ./cmd/vplc
