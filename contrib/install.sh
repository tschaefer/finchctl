#!/bin/sh
set -eu

# Configuration
INSTALL_DIR="/usr/local/bin"
GITHUB_REPO="tschaefer/finchctl"

# Colors for output (disabled if NO_COLOR is set)
if [ "${NO_COLOR:-}" = "1" ] || [ "${NO_COLOR:-}" = "true" ]; then
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
else
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
fi

# Print usage
usage() {
    cat << EOF
Usage: $0 [--help]

Options:
  --help, -h    Show this help message

Environment Variables:
  NO_COLOR=1    Disable colored output

This script will:
  - Download and install the latest binary for your os and architecture
    /usr/local/bin/finchctl
EOF
    exit 0
}

# Print message
info() {
    printf "%b[INFO]%b %s\n" "${GREEN}" "${NC}" "$1"
}

warn() {
    printf "%b[WARN]%b %s\n" "${YELLOW}" "${NC}" "$1"
}

error() {
    printf "%b[ERROR]%b %s\n" "${RED}" "${NC}" "$1" >&2
}

# Check if running as root
check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect operating system
detect_os() {
    os=$(uname -s)
    case "$os" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            error "Supported operating systems: Linux, Darwin (macOS)"
            exit 1
            ;;
    esac
}

# Detect system architecture
detect_arch() {
    arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            error "Supported architectures: x86_64 (amd64), aarch64 (arm64)"
            exit 1
            ;;
    esac
}

# Check required commands
check_dependencies() {
    if ! command -v curl >/dev/null 2>&1; then
        error "Required command 'curl' not found. Please install it first."
        exit 1
    fi
}

# Get latest release download URL
get_latest_release_url() {
    os="$1"
    arch="$2"
    binary_name="finchctl-${os}-${arch}"
    download_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${binary_name}"

    echo "$download_url"
}

# Download and install binary
install_binary() {
    os="$1"
    arch="$2"
    download_url=$(get_latest_release_url "$os" "$arch")

    info "Downloading finchctl binary for ${os}-${arch}..."
    if ! curl -sfL -o /tmp/finchctl "$download_url"; then
        error "Failed to download finchctl binary"
        exit 1
    fi

    # Verify downloaded file is not empty
    if [ ! -s /tmp/finchctl ]; then
        error "Downloaded binary file is empty or does not exist"
        rm -f /tmp/finchctl
        exit 1
    fi

    info "Installing binary to ${INSTALL_DIR}/finchctl..."
    install -m 0755 /tmp/finchctl "${INSTALL_DIR}/finchctl"
    rm -f /tmp/finchctl

    info "Binary installed successfully"
}

# Parse command line arguments
for arg in "$@"; do
    case "$arg" in
        --help|-h)
            usage
            exit 0
            ;;
        *)
            error "Unknown option: $arg"
            echo "Usage: $0 [--help]"
            exit 1
            ;;
    esac
done

# Main installation function
main() {
    echo "======================================"
    echo "  finchctl Installation Script"
    echo "======================================"
    echo

    check_root
    check_dependencies

    os=$(detect_os)
    arch=$(detect_arch)
    info "Detected operating system architecture: ${os}-${arch}"

    install_binary "$os" "$arch"

    echo
    echo "======================================"
    echo "  Installation Complete!"
    echo "======================================"
    echo
    info "finchctl has been installed"
    info "Start manually: finchctl --help"
}

main "$@"
