# ynab-cli

A standalone Go CLI for managing your [YNAB](https://www.ynab.com/) (You Need A Budget) finances from the terminal. Zero external dependencies.

## Features

- **Budget tracking** — status, account balances, categories, monthly budgets
- **Transaction management** — add, edit, delete expenses and income
- **Category budgeting** — move money between categories
- **Scheduled transactions** — view recurring/upcoming transactions
- **Account creation** — add new accounts (checking, savings, credit card, etc.)
- **Payee management** — list and filter payees
- **Interactive configuration** — `ynab configure` setup (like `aws configure`)
- **Diagnostics** — built-in `doctor` command for troubleshooting
- **JSON output** — machine-readable format for scripting (`--json`)
- **Cross-platform** — macOS (arm64/amd64) and Linux (amd64/arm64)
- **Pure Go stdlib** — zero external dependencies

## Installation

### Prerequisites

- Go 1.25 or later
- YNAB Personal Access Token ([get one here](https://app.ynab.com/settings/developer))

### Build and Install

```bash
make install        # Build and symlink to ~/bin
ynab configure      # Interactive setup
ynab doctor         # Verify everything works
```

### Build from Source

```bash
make build          # Build for current platform → bin/ynab
make build-all      # Cross-compile for macOS/Linux (arm64/amd64)
```

## Configuration

Config is stored in `~/.ynab/config` (INI format, `chmod 600`).

```bash
ynab configure          # Interactive setup (recommended)
ynab configure show     # Show current config (token masked)
```

### Config file keys

| Key | Description |
|-----|-------------|
| `access_token` | YNAB Personal Access Token |
| `default_budget_id` | Default budget ID for all commands |
| `api_base_url` | API base URL (default: `https://api.youneedabudget.com/v1`) |

### Environment variables (fallback)

| Variable | Description |
|----------|-------------|
| `YNAB_ACCESS_TOKEN` | Access token (used if no config file) |
| `YNAB_DEFAULT_BUDGET_ID` | Default budget ID (used if no config file) |

## Commands

### Viewing data

```bash
ynab status                     # Budget status and metadata
ynab balance                    # All account balances
ynab balance checking           # Filter by account name
ynab budget                     # Current month's budget with categories
ynab categories                 # List all categories with IDs
ynab months                     # List available months
ynab months 2024-06             # Show detail for a specific month
ynab payees                     # List all payees
ynab payees "Coffee"            # Filter payees by name
ynab scheduled                  # List scheduled/recurring transactions
ynab transactions               # List recent transactions
```

### Transaction filters

```bash
ynab transactions --since 2024-01-01
ynab transactions --account "Checking"
ynab transactions --category "Groceries"
```

### Adding transactions

```bash
# Basic expense (positive number = outflow)
ynab add 50 "Coffee Shop" "Dining Out"

# Income (prefix with +)
ynab add +1000 "Paycheck" --account "Checking"

# With all options
ynab add 75.50 "Grocery Store" "Groceries" \
  --account "Credit Card" \
  --date 2024-01-15 \
  --memo "Weekly shopping"
```

### Editing and deleting

```bash
ynab edit <transaction_id> --amount 42 --payee "New Payee" --cleared
ynab delete <transaction_id>
```

### Moving money between categories

```bash
ynab move 100 --from "Dining Out" --to "Groceries"
ynab move 50 --from "Fun Money" --to "Emergency" --month 2024-06
```

### Account management

```bash
ynab add-account "Savings" savings 5000
```

Account types: `checking`, `savings`, `creditCard`, `cash`, `lineOfCredit`, `otherAsset`, `otherLiability`, `mortgage`, `autoLoan`, `studentLoan`, `personalLoan`, `medicalDebt`, `otherDebt`.

### JSON output

All commands support `--json` for scripting:

```bash
ynab status --json
ynab balance --json
ynab budget --json | jq '.category_groups[].categories[] | select(.balance < 0)'
```

## Architecture

```
cmd/ynab-cli/               # Entry point and command routing
internal/
├── api/                     # YNAB API client (retry, rate limiting)
│   ├── client.go            # HTTP client with exponential backoff
│   ├── methods.go           # 18 API endpoint methods
│   ├── types.go             # Budget, Account, Transaction types
│   └── errors.go            # Error handling
├── cmd/                     # Command implementations
│   ├── status.go            # status, balance, budget commands
│   ├── add.go               # Transaction creation
│   ├── edit.go              # Transaction editing
│   ├── delete.go            # Transaction deletion
│   ├── move.go              # Category money movement
│   ├── transactions.go      # Transaction listing
│   ├── configure.go         # Configuration management
│   └── doctor.go            # Diagnostics
├── config/                  # Config file loading/saving
└── transform/               # Currency formatting (milliunits ↔ dollars)
```

### Design decisions

- **Milliunit arithmetic** — all monetary amounts use `int64` milliunits (1000 = $1.00) to avoid floating-point errors
- **Retry with backoff** — exponential backoff (1s, 2s, 4s) with rate-limit (`429`) awareness
- **No CLI framework** — simple string-based command dispatch, no external dependencies
- **Secure config** — config directory `700`, config file `600` permissions

## Development

```bash
make build              # Build for current platform
make install            # Build and install to ~/bin
make test               # Run unit tests
make test-coverage      # Tests with coverage report
make test-e2e           # E2E tests (requires YNAB_ACCESS_TOKEN)
make fmt                # Format code
make vet                # Vet for issues
make lint               # fmt + vet
make clean              # Remove build artifacts
```

## License

MIT License. See [LICENSE](LICENSE) for details.

## Links

- [YNAB API Documentation](https://api.ynab.com/)
- [Get Access Token](https://app.ynab.com/settings/developer)
