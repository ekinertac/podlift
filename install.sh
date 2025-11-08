#!/bin/sh
# podlift installation script
# Usage: 
#   curl -sSL https://podlift.sh/install.sh | sh
#   
#   # Install to user directory (no sudo):
#   curl -sSL https://podlift.sh/install.sh | INSTALL_DIR="$HOME/.local/bin" sh

set -e

# Configuration
REPO="ekinertac/podlift"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="podlift"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Linux*)
            OS="linux"
            ;;
        Darwin*)
            OS="darwin"
            ;;
        *)
            echo "${RED}Error: Unsupported operating system: $OS${NC}"
            echo "podlift supports Linux and macOS only."
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo "${RED}Error: Unsupported architecture: $ARCH${NC}"
            echo "podlift supports amd64 and arm64 only."
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}_${ARCH}"
}

# Get the latest release version
get_latest_version() {
    echo "${BLUE}Fetching latest version...${NC}"
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        echo "${RED}Error: curl or wget is required${NC}"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        echo "${RED}Error: Could not fetch latest version${NC}"
        exit 1
    fi
    
    echo "${GREEN}Latest version: $VERSION${NC}"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}_${PLATFORM}"
    TMP_FILE="/tmp/${BINARY_NAME}_${PLATFORM}"
    
    echo "${BLUE}Downloading from: $DOWNLOAD_URL${NC}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL "$DOWNLOAD_URL" -o "$TMP_FILE"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"
    fi
    
    if [ ! -f "$TMP_FILE" ]; then
        echo "${RED}Error: Failed to download binary${NC}"
        exit 1
    fi
    
    chmod +x "$TMP_FILE"
}

# Install binary
install_binary() {
    echo "${BLUE}Installing to $INSTALL_DIR...${NC}"
    
    # Create directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
            echo "${GREEN}Created directory: $INSTALL_DIR${NC}"
        else
            echo "${YELLOW}Creating directory with sudo...${NC}"
            sudo mkdir -p "$INSTALL_DIR"
        fi
    fi
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        echo "${YELLOW}Requesting sudo access to write to $INSTALL_DIR${NC}"
        echo "${YELLOW}Tip: To install without sudo, use: INSTALL_DIR=\"\$HOME/.local/bin\" sh${NC}"
        sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    if [ $? -ne 0 ]; then
        echo "${RED}Error: Failed to install binary${NC}"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        echo "${YELLOW}Warning: $BINARY_NAME not found in PATH${NC}"
        echo ""
        echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
        echo "Then run: source ~/.zshrc  (or restart your terminal)"
        return 1
    fi
    
    INSTALLED_VERSION=$("$BINARY_NAME" version 2>/dev/null | head -1 || echo "unknown")
    echo "${GREEN}Successfully installed: $INSTALLED_VERSION${NC}"
    return 0
}

# Cleanup
cleanup() {
    if [ -f "$TMP_FILE" ]; then
        rm -f "$TMP_FILE"
    fi
}

# Main installation flow
main() {
    echo ""
    echo "${BLUE}╔═══════════════════════════════════╗${NC}"
    echo "${BLUE}║                                   ║${NC}"
    echo "${BLUE}║  podlift installation script      ║${NC}"
    echo "${BLUE}║                                   ║${NC}"
    echo "${BLUE}╚═══════════════════════════════════╝${NC}"
    echo ""
    
    detect_platform
    echo "${GREEN}Platform detected: $PLATFORM${NC}"
    echo ""
    
    get_latest_version
    download_binary
    install_binary
    
    echo ""
    if verify_installation; then
        echo ""
        echo "${GREEN}🎉 Installation complete!${NC}"
        echo ""
        echo "Get started:"
        echo "  ${BLUE}podlift init${NC}      # Initialize configuration"
        echo "  ${BLUE}podlift setup${NC}     # Prepare servers"
        echo "  ${BLUE}podlift deploy${NC}    # Deploy your app"
        echo ""
        echo "Documentation: ${BLUE}https://github.com/${REPO}${NC}"
    fi
    
    cleanup
}

# Trap errors and cleanup
trap cleanup EXIT

# Run installation
main

