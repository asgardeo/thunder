#!/bin/bash
# verify-build-artifacts.sh
# Verifies that all expected Thunder build artifacts were created for all platforms
#
# Usage: ./scripts/verify-build-artifacts.sh
#
# Exit codes:
#   0 - All artifacts found
#   1 - One or more artifacts missing

set -e

echo "✅ Verifying all build artifacts were created..."

# Define expected platforms (GO_OS:GO_ARCH)
PLATFORMS=("windows:amd64" "linux:amd64" "linux:arm64" "darwin:amd64" "darwin:arm64")

# Track any missing artifacts
MISSING_COUNT=0

for platform in "${PLATFORMS[@]}"; do
  OS="${platform%%:*}"
  ARCH="${platform#*:}"

  # Transform OS name to match package naming (same as build.sh)
  PACKAGE_OS="$OS"
  if [ "$OS" = "darwin" ]; then
    PACKAGE_OS="macos"
  elif [ "$OS" = "windows" ]; then
    PACKAGE_OS="win"
  fi

  # Transform ARCH name to match package naming (same as build.sh)
  PACKAGE_ARCH="$ARCH"
  if [ "$ARCH" = "amd64" ]; then
    PACKAGE_ARCH="x64"
  fi

  # Build the expected file pattern with specific architecture
  EXPECTED_PATTERN="target/dist/thunder-*-${PACKAGE_OS}-${PACKAGE_ARCH}.zip"

  # Expand glob once into an array (nullglob ensures empty array if no match)
  shopt -s nullglob
  MATCHED_FILES=($EXPECTED_PATTERN)
  shopt -u nullglob

  # Check if artifact exists
  if [ ${#MATCHED_FILES[@]} -eq 0 ]; then
    echo "❌ Backend artifact not found for $OS/$ARCH: $EXPECTED_PATTERN"
    MISSING_COUNT=$((MISSING_COUNT + 1))
  else
    echo "✅ Found backend artifact for $OS/$ARCH: ${MATCHED_FILES[*]}"
  fi
done

if [ $MISSING_COUNT -gt 0 ]; then
  echo ""
  echo "❌ $MISSING_COUNT artifact(s) missing!"
  exit 1
fi

echo ""
echo "✅ All backend artifacts verified successfully!"
exit 0
