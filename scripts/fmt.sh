#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "Formatting files:"
    echo "$unformatted" | while read -r f; do echo "  $f"; done
    gofmt -w .
    echo "Done."
else
    echo "All files are already formatted."
fi
