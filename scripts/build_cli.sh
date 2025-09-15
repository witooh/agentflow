#!/usr/bin/env bash
# Build script for AgentFlow CLI (no Docker). Produces binaries in ./dist
set -euo pipefail

# Configuration
APP_NAME="agentflow"
PKG="./cmd/agentflow"
DIST_DIR="dist"

# Default targets (GOOS/GOARCH)
TARGETS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

usage() {
  cat <<EOF
Build AgentFlow CLI binaries

Usage:
  $(basename "$0") [--current] [--target OS/ARCH ...] [--clean]

Options:
  --current        Build only for current platform
  --target         Add an extra GOOS/GOARCH target (can be repeated)
  --clean          Remove ./dist before building
  -h, --help       Show this help

Examples:
  ./scripts/build_cli.sh --current
  ./scripts/build_cli.sh --target linux/386 --target freebsd/amd64
EOF
}

CLEAN=0
CURRENT=0
EXTRA_TARGETS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --clean)
      CLEAN=1; shift ;;
    --current)
      CURRENT=1; shift ;;
    --target)
      [[ $# -ge 2 ]] || { echo "--target requires value like OS/ARCH" >&2; exit 1; }
      EXTRA_TARGETS+=("$2"); shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *) echo "Unknown arg: $1" >&2; usage; exit 1 ;;
  esac
done

if [[ $CLEAN -eq 1 && -d "$DIST_DIR" ]]; then
  rm -rf "$DIST_DIR"
fi
mkdir -p "$DIST_DIR"

if [[ $CURRENT -eq 1 ]]; then
  # Build for current machine only
  EXT=""
  if [[ "${OS:-$(uname -s)}" == "Windows_NT" ]]; then EXT=".exe"; fi
  CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "${DIST_DIR}/${APP_NAME}${EXT}" "$PKG"
  echo "Built ${DIST_DIR}/${APP_NAME}${EXT}"
  exit 0
fi

# Merge default targets with any extras
ALL_TARGETS=("${TARGETS[@]}")
if [[ ${#EXTRA_TARGETS[@]} -gt 0 ]]; then
  ALL_TARGETS+=("${EXTRA_TARGETS[@]}")
fi

for t in "${ALL_TARGETS[@]}"; do
  OS_NAME="${t%%/*}"
  ARCH="${t##*/}"
  EXT=""
  OUT_NAME="${APP_NAME}_${OS_NAME}_${ARCH}"
  if [[ "$OS_NAME" == "windows" ]]; then EXT=".exe"; fi
  echo "Building $OUT_NAME..."
  GOOS="$OS_NAME" GOARCH="$ARCH" CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o "${DIST_DIR}/${OUT_NAME}${EXT}" "$PKG"
  echo "Built ${DIST_DIR}/${OUT_NAME}${EXT}"
done

echo "All builds completed. Binaries are in ./${DIST_DIR}"
