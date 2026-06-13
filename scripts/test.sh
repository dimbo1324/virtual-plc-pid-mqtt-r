#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."
find . -type f -name '*.go' -exec gofmt -w {} +
go vet ./...
go test ./...
