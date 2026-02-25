#!/bin/bash
set -e

# Configuration
HIGHS_VERSION="1.13.1"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        HIGHS_ARCH="x86_64"
        DUCKDB_ARCH="amd64"
        ;;
    aarch64|arm64)
        HIGHS_ARCH="aarch64"
        DUCKDB_ARCH="aarch64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

HIGHS_URL="https://github.com/ERGO-Code/HiGHS/releases/download/v${HIGHS_VERSION}/highs-${HIGHS_VERSION}-${HIGHS_ARCH}-linux-gnu-static-apache.tar.gz"
DUCKDB_URL="https://github.com/duckdb/duckdb/releases/download/v1.2.0/duckdb_cli-linux-${DUCKDB_ARCH}.zip"
BIN_DIR="./bin"
TMP_DIR="./tmp_deps"

echo "🚀 Bootstrapping Squidgame dependencies (Arch: $ARCH)..."

# Create bin directory
mkdir -p "$BIN_DIR"

# Create temporary directory for extraction
mkdir -p "$TMP_DIR"

echo "📥 Downloading HiGHS v${HIGHS_VERSION}..."
curl -L "$HIGHS_URL" -o "${TMP_DIR}/highs.tar.gz"

echo "📦 Extracting HiGHS..."
tar -xzf "${TMP_DIR}/highs.tar.gz" -C "$TMP_DIR"
if [ -f "${TMP_DIR}/bin/highs" ]; then
    mv "${TMP_DIR}/bin/highs" "$BIN_DIR/highs"
    chmod +x "$BIN_DIR/highs"
    echo "✅ HiGHS installed successfully"
fi

echo "📥 Downloading DuckDB v1.2.0..."
curl -L "$DUCKDB_URL" -o "${TMP_DIR}/duckdb.zip"

echo "📦 Extracting DuckDB..."
unzip -o "${TMP_DIR}/duckdb.zip" -d "$BIN_DIR"
chmod +x "$BIN_DIR/duckdb"
echo "✅ DuckDB installed successfully"

# Cleanup
echo "🧹 Cleaning up..."
rm -rf "$TMP_DIR"

echo "✨ Bootstrap complete. You are ready to play the game."
./bin/highs --version
./bin/duckdb --version
