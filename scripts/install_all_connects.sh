#!/bin/bash

set -e

CONNECT_RELEASES_URL="https://api.github.com/repos/skip-mev/connect/releases?per_page=100"

# Determine the system OS + architecture, or get target OS + arch from params
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$#" -eq 2 ]; then
    OS=$1
    ARCH=$2
fi

# Fetch all releases information
echo "Fetching releases information..."
RELEASES_JSON=$(curl -Ls "${CONNECT_RELEASES_URL}")

# Check that jq is installed
if ! command -v jq &>/dev/null; then
    echo "jq is required but not installed. Please install jq."
    exit 1
fi

# Initialize a flag to identify the first (latest) release
FIRST_RELEASE=true

# Process each release
echo "Processing releases..."
echo "${RELEASES_JSON}" | jq -c '.[]' | while read -r RELEASE; do
    # Extract release information
    TAG_NAME=$(echo "${RELEASE}" | jq -r '.tag_name')
    VERSION="${TAG_NAME#v}"  # Remove the 'v' prefix

    echo "Processing version ${VERSION}..."

    # Map architecture to release file name
    case "${ARCH}" in
        x86_64)
            if [ "${OS}" = "darwin" ]; then
                FILE_NAME="connect-${VERSION}-darwin-amd64.tar.gz"
            else
                FILE_NAME="connect-${VERSION}-linux-amd64.tar.gz"
            fi
            ;;
        aarch64|arm64)
            if [ "${OS}" = "darwin" ]; then
                FILE_NAME="connect-${VERSION}-darwin-arm64.tar.gz"
            else
                FILE_NAME="connect-${VERSION}-linux-arm64.tar.gz"
            fi
            ;;
        i386|i686)
            FILE_NAME="connect-${VERSION}-linux-386.tar.gz"
            ;;
        *)
            echo "Unsupported architecture: ${ARCH}"
            exit 1
            ;;
    esac

    # Get download URL for the specific file
    DOWNLOAD_URL=$(echo "${RELEASE}" | jq -r --arg FILE_NAME "${FILE_NAME}" '.assets[] | select(.name == $FILE_NAME) | .browser_download_url')

    if [ -z "${DOWNLOAD_URL}" ] || [ "${DOWNLOAD_URL}" = "null" ]; then
        echo "Failed to find download URL for ${FILE_NAME} in version ${VERSION}"
        continue
    fi

    # Download the release
    echo "Downloading ${FILE_NAME}..."
    curl -LO "${DOWNLOAD_URL}"

    # Create a temporary directory for extraction
    TEMP_DIR=$(mktemp -d)
    echo "Extracting connect binary to ${TEMP_DIR}..."
    tar -xzf "${FILE_NAME}" -C "${TEMP_DIR}"

    # Find the connect binary
    CONNECT_BIN=$(find "${TEMP_DIR}" -type f -name "connect")

    if [ -z "${CONNECT_BIN}" ]; then
        echo "Failed to find connect binary in the extracted files for version ${VERSION}"
        rm -rf "${TEMP_DIR}"
        rm "${FILE_NAME}"
        continue
    fi

    # install the latest release as 'connect'
    if [ "${FIRST_RELEASE}" = true ]; then
        LATEST_BIN="/usr/local/bin/connect"
        FIRST_RELEASE=false

        echo "Installing latest connect to ${LATEST_BIN}..."
        cp "${CONNECT_BIN}" "${LATEST_BIN}"
        chmod +x "${LATEST_BIN}"
    fi

    # install release as connect-<version>
    DEST_BIN="/usr/local/bin/connect-${VERSION}"
    echo "Installing connect to ${DEST_BIN}..."
    mv "${CONNECT_BIN}" "${DEST_BIN}"
    chmod +x "${DEST_BIN}"

    # Clean up
    rm -rf "${TEMP_DIR}"
    rm "${FILE_NAME}"

    echo "Connect ${VERSION} has been installed successfully!"
    echo "-----------------------------------------------"

    # Stop processing after version 2.0.0
    if [ "${VERSION}" = "2.0.0" ]; then
        echo "Reached version 2.0.0. Installation complete."
        break
    fi

done
