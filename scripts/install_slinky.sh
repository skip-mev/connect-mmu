#!/bin/bash

set -e

SLINKY_RELEASES="https://api.github.com/repos/skip-mev/connect/releases/174832995/assets"

# Determine the system OS + architecture, or get target OS + arch from params
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$#" -eq 2 ]; then
    OS=$1
    ARCH=$2
fi

echo "Fetching release information..."
RELEASE_INFO=$(curl -Ls ${SLINKY_RELEASES})
VERSION=1.0.12
# Map architecture to release file name
case "${ARCH}" in
    x86_64)
        if [ "${OS}" = "darwin" ]; then
            FILE_NAME="slinky-${VERSION}-darwin-amd64.tar.gz"
        else
            FILE_NAME="slinky-${VERSION}-linux-amd64.tar.gz"
        fi
        ;;
    aarch64|arm64)
        if [ "${OS}" = "darwin" ]; then
            FILE_NAME="slinky-${VERSION}-darwin-arm64.tar.gz"
        else
            FILE_NAME="slinky-${VERSION}-linux-arm64.tar.gz"
        fi
        ;;
    i386|i686)
        FILE_NAME="slinky-${VERSION}-linux-386.tar.gz"
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

# Get download URL for the specific file
DOWNLOAD_URL=$(echo "${RELEASE_INFO}" | grep -o "\"browser_download_url\": \"[^\"]*${FILE_NAME}\"" | cut -d'"' -f4)

if [ -z "${DOWNLOAD_URL}" ]; then
    echo "Failed to find download URL for ${FILE_NAME}"
    exit 1
fi

# Download the release
echo "Downloading ${FILE_NAME}..."
curl -LO "${DOWNLOAD_URL}"

# Create a temporary directory for extraction
TEMP_DIR=$(mktemp -d)
echo "Extracting slinky binary to ${TEMP_DIR}..."
tar -xzf "${FILE_NAME}" -C "${TEMP_DIR}"

# Find the slinky binary
SLINKY_BIN=$(find "${TEMP_DIR}" -type f -name "slinky")

if [ -z "${SLINKY_BIN}" ]; then
    echo "Failed to find slinky binary in the extracted files"
    rm -rf "${TEMP_DIR}"
    rm "${FILE_NAME}"
    exit 1
fi

# Move the binary to /usr/local/bin
echo "Installing slinky to /usr/local/bin..."
mv "${SLINKY_BIN}" /usr/local/bin/

# Make it executable
chmod +x /usr/local/bin/slinky

# Clean up
rm -rf "${TEMP_DIR}"
rm "${FILE_NAME}"

echo "Slinky ${VERSION} has been installed successfully!"
