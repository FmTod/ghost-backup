#!/usr/bin/env bash
# Ghost Backup Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/FmTod/ghost-backup/main/install.sh | bash

set -e

# Configuration
REPO="FmTod/ghost-backup"
BINARY_NAME="ghost-backup"
DEFAULT_INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse arguments
INSTALL_DIR="$DEFAULT_INSTALL_DIR"
while [[ $# -gt 0 ]]; do
    case $1 in
        --prefix)
            INSTALL_DIR="$2/bin"
            shift 2
            ;;
        --help)
            echo "Ghost Backup Installation Script"
            echo ""
            echo "Usage: install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --prefix DIR    Install to DIR/bin (default: /usr/local)"
            echo "  --help          Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)
            os="linux"
            ;;
        Darwin*)
            os="darwin"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}/${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        log_error "Failed to fetch latest release version"
        exit 1
    fi

    echo "$version"
}

# Download and install binary
install_binary() {
    local platform=$1
    local version=$2
    local os="${platform%/*}"
    local arch="${platform#*/}"

    local binary_name="${BINARY_NAME}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi

    local download_url="https://github.com/${REPO}/releases/download/${version}/${binary_name}"
    local temp_file=$(mktemp)

    log_info "Downloading ${BINARY_NAME} ${version} for ${os}/${arch}..."

    if ! curl -fL "$download_url" -o "$temp_file"; then
        log_error "Failed to download ${BINARY_NAME}"
        rm -f "$temp_file"
        exit 1
    fi

    # Make binary executable
    chmod +x "$temp_file"

    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating installation directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR" || {
            log_error "Failed to create directory. You may need to run with sudo."
            rm -f "$temp_file"
            exit 1
        }
    fi

    # Install binary
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [ "$os" = "windows" ]; then
        install_path="${install_path}.exe"
    fi

    log_info "Installing to $install_path..."

    if ! mv "$temp_file" "$install_path" 2>/dev/null; then
        log_warn "Installation requires elevated privileges."
        if command -v sudo &> /dev/null; then
            sudo mv "$temp_file" "$install_path" || {
                log_error "Failed to install ${BINARY_NAME}"
                rm -f "$temp_file"
                exit 1
            }
        else
            log_error "sudo not found and installation failed. Please run as administrator."
            rm -f "$temp_file"
            exit 1
        fi
    fi

    log_info "Successfully installed ${BINARY_NAME} to $install_path"
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."

    if command -v "$BINARY_NAME" &> /dev/null; then
        local version=$("$BINARY_NAME" --version 2>&1 || echo "unknown")
        log_info "✓ ${BINARY_NAME} is installed and available in PATH"
        log_info "  Version: $version"
    else
        log_warn "${BINARY_NAME} is installed but not in PATH"
        log_warn "Add $INSTALL_DIR to your PATH:"
        log_warn "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# Main installation flow
main() {
    echo ""
    log_info "Ghost Backup Installation Script"
    echo ""

    # Check for required commands
    for cmd in curl grep sed; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "Required command not found: $cmd"
            exit 1
        fi
    done

    local platform=$(detect_platform)
    log_info "Detected platform: $platform"

    local version=$(get_latest_version)
    log_info "Latest version: $version"

    install_binary "$platform" "$version"
    verify_installation

    echo ""
    log_info "Installation complete! 🎉"
    echo ""
    log_info "Get started with:"
    log_info "  ghost-backup --help"
    log_info "  ghost-backup init"
    echo ""
}

main "$@"

