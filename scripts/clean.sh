#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

DRY_RUN=false
ALL=false

for arg in "$@"; do
    case "$arg" in
        --dry-run) DRY_RUN=true ;;
        --all)     ALL=true ;;
    esac
done

remove_target() {
    if [ -e "$1" ]; then
        if $DRY_RUN; then
            echo "  [dry-run] would remove: $1"
        else
            rm -rf -- "$1"
            echo "  removed: $1"
        fi
    fi
}

remove_runtime_files() {
    dir="$1"
    find "$dir" -maxdepth 1 -type f ! -name '.gitkeep' 2>/dev/null | while read -r f; do
        if $DRY_RUN; then
            echo "  [dry-run] would remove: $f"
        else
            rm -f -- "$f"
            echo "  removed: $f"
        fi
    done || true
}

if $DRY_RUN; then echo "Dry-run mode: no files will be deleted."; fi

echo "Cleaning build artifacts..."
remove_target dist
remove_target bin

echo "Cleaning runtime data..."
remove_runtime_files data
remove_runtime_files logs

if $ALL; then
    echo "Cleaning Go test cache..."
    if $DRY_RUN; then
        echo "  [dry-run] would run: go clean -testcache"
    else
        go clean -testcache
    fi

    echo "Cleaning Go build cache..."
    if $DRY_RUN; then
        echo "  [dry-run] would run: go clean -cache"
    else
        go clean -cache
    fi
fi

echo "Clean complete."
