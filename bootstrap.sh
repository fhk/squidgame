#!/bin/bash
set -e

# Configuration
HIGHS_VERSION="1.13.1"
HIGHS_URL="https://github.com/ERGO-Code/HiGHS/releases/download/v${HIGHS_VERSION}/highs-${HIGHS_VERSION}-x86_64-linux-gnu-static-apache.tar.gz"
BIN_DIR="./bin"
TMP_DIR="./tmp_highs"

echo "🚀 Bootstrapping Squidgame dependencies..."

# Create bin directory
mkdir -p "$BIN_DIR"

# Create temporary directory for extraction
mkdir -p "$TMP_DIR"

echo "📥 Downloading HiGHS v${HIGHS_VERSION}..."
curl -L "$HIGHS_URL" -o "${TMP_DIR}/highs.tar.gz"

echo "📦 Extracting binary..."
tar -xzf "${TMP_DIR}/highs.tar.gz" -C "$TMP_DIR"

# Move the binary to ./bin/highs
# The tarball structure typically has bin/highs
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
