#!/bin/sh
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { printf "${GREEN}%s${NC}\n" "$1"; }
warn() { printf "${YELLOW}%s${NC}\n" "$1"; }
fail() { printf "${RED}%s${NC}\n" "$1" >&2; }

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux) OS_NAME="linux" ;;
    Darwin) OS_NAME="darwin" ;;
    *) fail "Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64) ARCH_NAME="amd64" ;;
    aarch64|arm64) ARCH_NAME="arm64" ;;
    *) fail "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY="llmtop-${OS_NAME}-${ARCH_NAME}"
URL="https://github.com/changye/llmtop/releases/download/latest/${BINARY}"

# Determine install directory
if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
else
    if [ -d "$HOME/.local/bin" ]; then
        INSTALL_DIR="$HOME/.local/bin"
    elif [ -d "$HOME/bin" ]; then
        INSTALL_DIR="$HOME/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi
fi

if [ -f "$INSTALL_DIR/llmtop" ]; then
    info "Existing 'llmtop' binary found at $INSTALL_DIR. Upgrading..."
    IS_UPGRADE=true
else
    info "Installing 'llmtop' for the first time."
    IS_UPGRADE=false
fi

info "Downloading from $URL"
TMP_DIR=$(mktemp -d)
curl -sL "$URL" -o "$TMP_DIR/llmtop"
chmod +x "$TMP_DIR/llmtop"

if ! "$TMP_DIR/llmtop" --help >/dev/null 2>"$TMP_DIR/llmtop.err"; then
    fail "Downloaded 'llmtop' binary could not run on this machine."
    if grep 'GLIBC_.*not found' "$TMP_DIR/llmtop.err" >/dev/null 2>&1; then
        fail "This binary requires a newer GLIBC than your system provides."
        warn "Build locally instead:"
        printf "  git clone https://github.com/changye/llmtop.git\n"
        printf "  cd llmtop && make build\n"
    else
        fail "Runtime error:"
        sed 's/^/  /' "$TMP_DIR/llmtop.err" >&2
    fi
    rm -rf "$TMP_DIR"
    exit 1
fi

mv "$TMP_DIR/llmtop" "$INSTALL_DIR/llmtop"
rm -rf "$TMP_DIR"

if [ "$IS_UPGRADE" = true ]; then
    info "Successfully upgraded 'llmtop' to $INSTALL_DIR"
else
    info "Successfully installed 'llmtop' to $INSTALL_DIR"
fi

# Check PATH
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        warn "Warning: '$INSTALL_DIR' is not in your PATH."
        warn "Add this line to ~/.bashrc or ~/.zshrc:"
        printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
        ;;
esac

info "Installation complete. Run 'llmtop' to start."
