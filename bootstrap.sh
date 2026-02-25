#!/bin/bash
set -e

# Configuration
HIGHS_VERSION="1.13.1"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        HIGHS_ARCH="x86_64"
        ;;
    aarch64|arm64)
        HIGHS_ARCH="aarch64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

HIGHS_URL="https://github.com/ERGO-Code/HiGHS/releases/download/v${HIGHS_VERSION}/highs-${HIGHS_VERSION}-${HIGHS_ARCH}-linux-gnu-static-apache.tar.gz"
BIN_DIR="./bin"
TMP_DIR="./tmp_highs"

echo "🚀 Bootstrapping Squidgame dependencies (Arch: $HIGHS_ARCH)..."

# Create bin directory
mkdir -p "$BIN_DIR"

# Create temporary directory for extraction
mkdir -p "$TMP_DIR"

echo "📥 Downloading HiGHS v${HIGHS_VERSION}..."
curl -L "$HIGHS_URL" -o "${TMP_DIR}/highs.tar.gz"

echo "📦 Extracting binary..."
tar -xzf "${TMP_DIR}/highs.tar.gz" -C "$TMP_DIR"

# Move the binary to ./bin/highs
if [ -f "${TMP_DIR}/bin/highs" ]; then
    mv "${TMP_DIR}/bin/highs" "$BIN_DIR/highs"
    chmod +x "$BIN_DIR/highs"
    echo "✅ HiGHS installed successfully to ${BIN_DIR}/highs"
else
    echo "❌ Error: Could not find highs binary in the extracted archive."
    exit 1
fi

# Cleanup
echo "🧹 Cleaning up..."
rm -rf "$TMP_DIR"

echo "✨ Bootstrap complete. You are ready to play the game."
./bin/highs --version
