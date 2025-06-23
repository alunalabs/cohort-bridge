#!/bin/bash

# Cohort Tokenize CLI Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/auroradata-ai/cohort-bridge/main/scripts/install.sh | bash

set -e

# Configuration
REPO="auroradata-ai/cohort-bridge"
BINARY_NAME="cohort-tokenize"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if running as root for system installation
check_permissions() {
    if [ "$EUID" -ne 0 ] && [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        print_warning "Installing to system directory requires root privileges"
        print_info "Run with sudo or set INSTALL_DIR environment variable"
        print_info "Example: INSTALL_DIR=~/.local/bin $0"
        exit 1
    fi
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) print_error "Unsupported architecture: $ARCH" ;;
    esac
    
    case $OS in
        linux) PLATFORM="linux-$ARCH" ;;
        darwin) PLATFORM="darwin-$ARCH" ;;
        *) print_error "Unsupported OS: $OS" ;;
    esac
}

# Get latest release version
get_latest_version() {
    print_info "Fetching latest release information..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        print_warning "Could not fetch latest version, using 'latest'"
        VERSION="latest"
    else
        print_info "Latest version: $VERSION"
    fi
}

# Download and install binary
install_binary() {
    print_info "Downloading $BINARY_NAME for $PLATFORM..."
    
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY_NAME-$PLATFORM.tar.gz"
    TEMP_DIR=$(mktemp -d)
    
    cd "$TEMP_DIR"
    
    if ! curl -sL "$DOWNLOAD_URL" | tar xz; then
        print_error "Failed to download or extract binary"
    fi
    
    # Make executable
    chmod +x "$BINARY_NAME-$PLATFORM"
    
    # Install to target directory
    print_info "Installing to $INSTALL_DIR..."
    mkdir -p "$INSTALL_DIR"
    mv "$BINARY_NAME-$PLATFORM" "$INSTALL_DIR/$BINARY_NAME"
    
    # Cleanup
    rm -rf "$TEMP_DIR"
    
    print_success "Successfully installed $BINARY_NAME to $INSTALL_DIR"
}

# Install from source (fallback)
install_from_source() {
    print_info "Installing from source..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
    fi
    
    print_info "Installing via 'go install'..."
    go install "github.com/$REPO/cmd/$BINARY_NAME@latest"
    
    print_success "Successfully installed $BINARY_NAME via go install"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        VERSION_OUTPUT=$($BINARY_NAME -version 2>&1 | head -n1)
        print_success "Installation verified: $VERSION_OUTPUT"
        
        print_info "Try running: $BINARY_NAME -help"
    else
        print_warning "Binary installed but not in PATH"
        print_info "Add $INSTALL_DIR to your PATH or run: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Main installation flow
main() {
    echo "🔐 Cohort Tokenize CLI Installer"
    echo "================================"
    echo
    
    # Check for custom install directory
    if [ -n "$INSTALL_DIR_OVERRIDE" ]; then
        INSTALL_DIR="$INSTALL_DIR_OVERRIDE"
        print_info "Using custom install directory: $INSTALL_DIR"
    fi
    
    # Check permissions for system install
    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        check_permissions
    fi
    
    # Detect platform
    detect_platform
    print_info "Detected platform: $PLATFORM"
    
    # Try binary installation first
    if get_latest_version && install_binary; then
        verify_installation
    else
        print_warning "Binary installation failed, trying source installation..."
        install_from_source
        verify_installation
    fi
    
    echo
    print_success "Installation complete!"
    print_info "Documentation: https://github.com/$REPO/blob/main/INSTALL.md"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--install-dir DIR] [--help]"
            echo
            echo "Options:"
            echo "  --install-dir DIR    Install to custom directory (default: /usr/local/bin)"
            echo "  --help              Show this help message"
            echo
            echo "Environment variables:"
            echo "  INSTALL_DIR         Override default install directory"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            ;;
    esac
done

# Run main installation
main 