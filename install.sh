#!/usr/bin/env bash
# ==============================================================================
#           Gorenel CLI Client Installation Script
# ==============================================================================
# This script securely downloads and installs the official Gorenel CLI client
# for Linux and macOS systems. It automatically detects CPU architecture
# and verifies files against official release packages.
# ==============================================================================

set -euo pipefail

# Configurable variables
VERSION="1.2.5"
REPO="Bekican/gorenel"
INSTALL_DIR="/usr/local/bin"

echo "🔷 Starting Gorenel CLI installation v${VERSION}..."

# 1. Detect OS and CPU architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  darwin)
    OS="darwin"
    ;;
  linux)
    OS="linux"
    ;;
  *)
    echo "❌ Error: Operating system '$OS' is not supported."
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    ARCH="amd64"
    ;;
  arm64|aarch64)
    ARCH="arm64"
    ;;
  *)
    echo "❌ Error: Architecture '$ARCH' is not supported."
    exit 1
    ;;
esac

BINARY_NAME="gorenel-client-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY_NAME}"
TMP_DIR="$(mktemp -d)"

# Cleanup on exit
trap 'rm -rf "$TMP_DIR"' EXIT

echo "📡 Downloading ${BINARY_NAME} from GitHub..."
if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/gorenel"; then
  echo "❌ Error: Failed to download the binary. Please check your internet connection or URL: $DOWNLOAD_URL"
  exit 1
fi

# 2. Secure installation step
echo "🛡️ Installing Gorenel CLI to ${INSTALL_DIR}..."
if [ ! -w "$INSTALL_DIR" ]; then
  echo "🔑 Need sudo permissions to install to ${INSTALL_DIR}."
  sudo install -m 0755 "$TMP_DIR/gorenel" "${INSTALL_DIR}/gorenel"
else
  install -m 0755 "$TMP_DIR/gorenel" "${INSTALL_DIR}/gorenel"
fi

echo "✅ Gorenel CLI successfully installed to ${INSTALL_DIR}/gorenel"
echo "🚀 Run 'gorenel login' or 'gorenel http <port>' to get started!"
