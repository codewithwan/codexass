#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
BUILD_DIR="$REPO_ROOT/build"
BINARY_PATH="$BUILD_DIR/codexass"

mkdir -p "$BUILD_DIR"

go build -o "$BINARY_PATH" ./cmd/codexass

printf 'Built: %s\n' "$BINARY_PATH"
