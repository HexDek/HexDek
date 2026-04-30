#!/usr/bin/env bash
# Convenience launcher for the hexdek-server.
# Run from the hexdek/ root.
set -euo pipefail

cd "$(dirname "$0")/.."

go run ./cmd/hexdek-server "$@"
