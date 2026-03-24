#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-1.0.0}"
FORMULA="packaging/homebrew/gorenel.rb"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

AMD_URL="https://github.com/Bekican/gorenel/releases/download/v${VERSION}/gorenel-client-darwin-amd64"
ARM_URL="https://github.com/Bekican/gorenel/releases/download/v${VERSION}/gorenel-client-darwin-arm64"

curl -fsSL "$AMD_URL" -o "$TMP_DIR/amd64"
curl -fsSL "$ARM_URL" -o "$TMP_DIR/arm64"

AMD_SHA="$(sha256sum "$TMP_DIR/amd64" | awk '{print $1}')"
ARM_SHA="$(sha256sum "$TMP_DIR/arm64" | awk '{print $1}')"

sed -i "s/version \"[0-9.]*\"/version \"${VERSION}\"/" "$FORMULA"
sed -i "s/REPLACE_WITH_DARWIN_AMD64_SHA256/${AMD_SHA}/" "$FORMULA"
sed -i "s/REPLACE_WITH_DARWIN_ARM64_SHA256/${ARM_SHA}/" "$FORMULA"

echo "Formula updated: $FORMULA"
echo "  darwin-amd64: $AMD_SHA"
echo "  darwin-arm64: $ARM_SHA"
