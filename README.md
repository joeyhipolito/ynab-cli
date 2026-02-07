# ynab

A standalone Go CLI tool for interacting with YNAB (You Need A Budget) API.

## Features

- **Budget tracking**: Status, balances, categories, monthly budget
- **Transactions**: Add expenses and income
- **Configuration**: Interactive setup like `aws configure`
- **Diagnostics**: Built-in `doctor` command for troubleshooting
- **JSON output**: Machine-readable format for scripting (`--json`)
- **Pure Go stdlib**: Zero external dependencies, ~2.3 MB binary

## Installation

### Prerequisites

- Go 1.21 or later
- YNAB Personal Access Token ([get one here](https://app.ynab.com/settings/developer))

### Build and Install

```bash
# Build and install to ~/bin
make install

# Configure (interactive)
ynab configure

# Verify setup
ynab doctor
```

## Configuration

```bash
# Interactive setup (recommended)
ynab configure

# Show current config
ynab configure show

# Troubleshoot
ynab doctor
```

Config is stored in `~/.ynab/config` (INI format, `chmod 600`).
Falls back to `YNAB_ACCESS_TOKEN` environment variable if no config file.

## Commands

```bash
ynab status              # Budget status and metadata
ynab balance             # Account balances
ynab balance checking    # Filter by account name
ynab budget              # Current month's budget with categories
ynab categories          # List all categories with IDs
ynab add 50 "Coffee"     # Add expense transaction
ynab configure           # Interactive setup
ynab doctor              # Validate installation
```

### Add Transaction

```bash
# Basic expense (positive = outflow)
ynab add 50 "Coffee Shop" "Dining Out"

# Income (+ prefix)
ynab add +1000 "Paycheck" --account "Checking"

# With all options
ynab add 75.50 "Grocery Store" "Groceries" \
  --account "Credit Card" \
  --date 2024-01-15 \
  --memo "Weekly shopping"
```

### JSON Output

All commands support `--json` for scripting:

```bash
ynab status --json
ynab balance --json
ynab budget --json | jq '.category_groups[].categories[] | select(.balance < 0)'
```

## Development

### Project Structure

```
ynab-cli/
├── cmd/ynab-cli/           # Main entry point
│   └── main.go
├── internal/
│   ├── api/                # YNAB API client (retry, rate limiting)
│   ├── cmd/                # Command implementations
│   ├── config/             # Config file loading/saving
│   └── transform/          # Currency/date formatting
├── Makefile
└── README.md
```

### Building

```bash
make build          # Build for current platform
make install        # Build and symlink to ~/bin
make test           # Run unit tests
make test-coverage  # Run tests with coverage report
make build-all      # Build for macOS/Linux (arm64/amd64)
make clean          # Remove build artifacts
```

## Architecture

- **Pure Go stdlib**: No external dependencies
- **Clean architecture**: Separate API client, commands, config, and transform layers
- **Config file**: `~/.ynab/config` (INI format, like AWS CLI)
- **Cross-platform**: macOS (darwin/arm64, darwin/amd64) and Linux (linux/amd64, linux/arm64)

## Exit Codes

- `0`: Success
- `1`: Error (missing config, API error, etc.)

## License

Part of the Via project.

## Links

- YNAB API Documentation: https://api.ynab.com/
- Get Access Token: https://app.ynab.com/settings/developer
