#!/bin/bash
set -e

# check-install.sh - Verify ynab-cli is installed and in PATH
# If not found, automatically run the installer

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_SCRIPT="$SCRIPT_DIR/install.sh"

# Check if ynab-cli is in PATH
if command -v ynab-cli >/dev/null 2>&1; then
    # Found in PATH - verify it works
    if ynab-cli --version >/dev/null 2>&1; then
        # Silent success - caller can use ynab-cli
        exit 0
    else
        # Binary exists but doesn't work - need to reinstall
        echo "⚠️  ynab-cli found but not working - reinstalling..." >&2
    fi
else
    # Not found in PATH
    echo "⚠️  ynab-cli not found in PATH - installing..." >&2
fi

# Run installer
if [ ! -f "$INSTALL_SCRIPT" ]; then
    echo "❌ Error: Install script not found at $INSTALL_SCRIPT" >&2
    exit 1
fi

# Make installer executable if needed
chmod +x "$INSTALL_SCRIPT"

# Run the installer
"$INSTALL_SCRIPT"

# Verify installation succeeded
if command -v ynab-cli >/dev/null 2>&1; then
    echo "✅ ynab-cli installed successfully" >&2
    exit 0
else
    echo "❌ Error: Installation completed but ynab-cli still not in PATH" >&2
    echo "   You may need to restart your shell or run: source ~/.zshrc" >&2
    exit 1
fi
