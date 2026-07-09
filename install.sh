#!/usr/bin/env bash
# Bullang Installer Bootstrap — Linux / macOS
# Downloads Go if needed, clones the installer, compiles and launches it.

set -e

INSTALLER_REPO="https://github.com/The-Bullang-Foundation/bullang-installer"
GO_VERSION="1.22.3"
INSTALL_DIR="$HOME/.bullang-installer"

echo ""
echo "  ____        _ _"
echo " |  _ \      | | |"
echo " | |_) |_   _| | | __ _ _ __   __ _"
echo " |  _ <| | | | | |/ _\` | '_ \ / _\` |"
echo " | |_) | |_| | | | (_| | | | | (_| |"
echo " |____/ \__,_|_|_|\__,_|_| |_|\__, |"
echo "                                __/ |"
echo "                               |___/"
echo ""
echo "  Bullang Ecosystem Installer"
echo "  ─────────────────────────────────────────"
echo ""

# ── Step 1: Install Go if missing ────────────────────────────────────────────

if command -v go &>/dev/null; then
    echo "  ✓ Go is already installed: $(go version)"
else
    echo "  → Installing Go $GO_VERSION..."

    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux)
            case "$ARCH" in
                x86_64)  GOARCH="amd64" ;;
                aarch64) GOARCH="arm64" ;;
                armv6l)  GOARCH="armv6l" ;;
                *)       echo "  ✗ Unsupported architecture: $ARCH"; exit 1 ;;
            esac
            GOTAR="go${GO_VERSION}.linux-${GOARCH}.tar.gz"
            GOURL="https://go.dev/dl/${GOTAR}"
            curl -fsSL "$GOURL" -o "/tmp/${GOTAR}"
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "/tmp/${GOTAR}"
            rm "/tmp/${GOTAR}"
            export PATH="/usr/local/go/bin:$PATH"
            echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$HOME/.profile"
            ;;
        Darwin)
            case "$ARCH" in
                x86_64) GOARCH="amd64" ;;
                arm64)  GOARCH="arm64" ;;
                *)      echo "  ✗ Unsupported architecture: $ARCH"; exit 1 ;;
            esac
            GOTAR="go${GO_VERSION}.darwin-${GOARCH}.tar.gz"
            GOURL="https://go.dev/dl/${GOTAR}"
            curl -fsSL "$GOURL" -o "/tmp/${GOTAR}"
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "/tmp/${GOTAR}"
            rm "/tmp/${GOTAR}"
            export PATH="/usr/local/go/bin:$PATH"
            echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$HOME/.zprofile"
            echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$HOME/.bash_profile"
            ;;
        *)
            echo "  ✗ Unsupported OS: $OS"
            exit 1
            ;;
    esac

    echo "  ✓ Go $GO_VERSION installed."
fi

# ── Step 2: Install git if missing ───────────────────────────────────────────

if ! command -v git &>/dev/null; then
    echo "  → Installing git..."
    OS="$(uname -s)"
    if [ "$OS" = "Linux" ]; then
        if command -v apt-get &>/dev/null; then
            sudo apt-get install -y git
        elif command -v pacman &>/dev/null; then
            sudo pacman -Sy --noconfirm git
        fi
    elif [ "$OS" = "Darwin" ]; then
        xcode-select --install 2>/dev/null || true
    fi
fi

# ── Step 3: Clone the installer ──────────────────────────────────────────────

echo "  → Downloading Bullang installer..."
rm -rf "$INSTALL_DIR"
git clone --depth 1 "$INSTALLER_REPO" "$INSTALL_DIR"

# ── Step 4: Build and run ────────────────────────────────────────────────────

echo "  → Building installer (this may take a moment)..."
cd "$INSTALL_DIR"
go mod tidy
go build -o bullang-installer .

echo "  → Launching installer..."
./bullang-installer
