#!/bin/sh
set -eu

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/bin"
BIN_NAME="${1:-docsfy}"

mkdir -p "${OUT_DIR}"

echo "Building ${BIN_NAME}..."
go build -trimpath -ldflags="-s -w" -o "${OUT_DIR}/${BIN_NAME}" "${ROOT_DIR}/cmd/main.go"

echo "Done: ${OUT_DIR}/${BIN_NAME}"
echo "Embedded resources: web/templates + web/assets"
