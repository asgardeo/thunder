#!/bin/bash
# verify-sample-artifacts.sh
# Verifies that all expected sample app artifacts were created for a specific platform
#
# Usage: ./scripts/verify-sample-artifacts.sh <OS> <ARCH>
#
# Arguments:
#   OS   - Operating system (linux, win, macos)
#   ARCH - Architecture (x64, arm64)
#
# Examples:
#   ./scripts/verify-sample-artifacts.sh linux x64
#   ./scripts/verify-sample-artifacts.sh macos arm64
#
# Exit codes:
#   0 - All sample artifacts found
#   1 - Invalid arguments or one or more artifacts missing

set -e

# Check arguments
OS="${1}"
ARCH="${2}"

if [ -z "$OS" ] || [ -z "$ARCH" ]; then
  echo "❌ Error: Missing required arguments"
  echo ""
  echo "Usage: $0 <OS> <ARCH>"
  echo ""
  echo "Arguments:"
  echo "  OS   - Operating system (linux, win, macos)"
  echo "  ARCH - Architecture (x64, arm64)"
  echo ""
  echo "Examples:"
  echo "  $0 linux x64"
  echo "  $0 macos arm64"
  exit 1
fi

echo "📦 Verifying sample artifacts for $OS/$ARCH..."

# Define expected sample apps
SAMPLE_APPS=("react-vanilla" "react-sdk" "react-api-based")

# Track any missing artifacts
MISSING_COUNT=0

for app in "${SAMPLE_APPS[@]}"; do
  EXPECTED_PATTERN="target/dist/sample-app-${app}-*-$OS-$ARCH.zip"

  # Expand glob once into an array (nullglob ensures empty array if no match)
  shopt -s nullglob
  MATCHED_FILES=($EXPECTED_PATTERN)
  shopt -u nullglob

  if [ ${#MATCHED_FILES[@]} -eq 0 ]; then
    echo "❌ Sample artifact not found: $EXPECTED_PATTERN"
    MISSING_COUNT=$((MISSING_COUNT + 1))
  else
    echo "✅ Found sample artifact: $(basename "${MATCHED_FILES[0]}")"
  fi
done

if [ $MISSING_COUNT -gt 0 ]; then
  echo ""
  echo "❌ $MISSING_COUNT sample artifact(s) missing for $OS/$ARCH!"
  exit 1
fi

echo ""
echo "✅ All sample packages verified for $OS/$ARCH"
exit 0
