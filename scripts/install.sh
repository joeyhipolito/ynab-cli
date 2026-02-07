#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Project root (parent of scripts/)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "🔧 YNAB CLI Installation"
echo "========================"
echo ""

# Check for Go installation
echo "1. Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    echo ""
    echo "Please install Go from https://go.dev/dl/"
    echo "Minimum version: 1.21"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✓ Go ${GO_VERSION} found${NC}"
echo ""

# Check for YNAB_ACCESS_TOKEN
echo "2. Checking YNAB_ACCESS_TOKEN..."
if [ -z "$YNAB_ACCESS_TOKEN" ]; then
    echo -e "${YELLOW}⚠ YNAB_ACCESS_TOKEN not set in environment${NC}"
    echo ""
    echo "You'll need to set this before using ynab-cli:"
    echo "  export YNAB_ACCESS_TOKEN='your-token-here'"
    echo ""
    echo "Get your token from: https://app.ynab.com/settings/developer"
    echo ""
else
    echo -e "${GREEN}✓ YNAB_ACCESS_TOKEN is set${NC}"
    echo ""
fi

# Build the binary
echo "3. Building ynab-cli..."
cd "$PROJECT_ROOT"
if make build; then
    echo -e "${GREEN}✓ Build successful${NC}"
    echo ""
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Create ~/bin if it doesn't exist
echo "4. Setting up installation directory..."
BIN_DIR="$HOME/bin"
if [ ! -d "$BIN_DIR" ]; then
    mkdir -p "$BIN_DIR"
    echo -e "${GREEN}✓ Created $BIN_DIR${NC}"
else
    echo -e "${GREEN}✓ $BIN_DIR exists${NC}"
fi
echo ""

# Install the binary (symlink to allow easy updates)
echo "5. Installing ynab-cli..."
BINARY_PATH="$PROJECT_ROOT/bin/ynab-cli"
INSTALL_PATH="$BIN_DIR/ynab-cli"

if [ -L "$INSTALL_PATH" ] || [ -f "$INSTALL_PATH" ]; then
    echo "   Removing existing installation..."
    rm -f "$INSTALL_PATH"
fi

ln -s "$BINARY_PATH" "$INSTALL_PATH"
echo -e "${GREEN}✓ Symlinked $BINARY_PATH → $INSTALL_PATH${NC}"
echo ""

# Check if ~/bin is in PATH
echo "6. Checking PATH configuration..."
if [[ ":$PATH:" == *":$BIN_DIR:"* ]]; then
    echo -e "${GREEN}✓ $BIN_DIR is in PATH${NC}"
else
    echo -e "${YELLOW}⚠ $BIN_DIR is not in PATH${NC}"
    echo ""
    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "  export PATH=\"\$HOME/bin:\$PATH\""
    echo ""
    echo "Then reload your shell:"
    echo "  source ~/.bashrc  # or ~/.zshrc"
    echo ""
fi

echo ""
echo "=============================="
echo -e "${GREEN}✅ Installation Complete!${NC}"
echo "=============================="
echo ""
echo "Usage:"
echo "  ynab-cli budgets              # List budgets"
echo "  ynab-cli accounts <budget>    # List accounts"
echo "  ynab-cli balance <budget>     # Show balances"
echo "  ynab-cli add <budget> <account> <amount> <memo>"
echo ""
echo "For help:"
echo "  ynab-cli --help"
echo ""

# Verify installation
if command -v ynab-cli &> /dev/null; then
    echo "Quick test:"
    ynab-cli --version
    echo ""
else
    echo -e "${YELLOW}Note: You may need to reload your shell or add ~/bin to PATH${NC}"
    echo ""
fi
