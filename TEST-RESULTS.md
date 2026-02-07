# YNAB CLI Integration Test Results

## Test Environment
- **Date**: 2026-02-02
- **Binary Location**: ~/bin/ynab-cli
- **Version**: 1.0.0
- **Platform**: darwin/arm64

## Tests Completed Without Token

### ✓ Binary Installation
- Binary exists at ~/bin/ynab-cli
- Binary is executable
- Binary is in PATH

### ✓ Version Information
```bash
$ ynab-cli --version
ynab-cli version 1.0.0
```

### ✓ Help Information
```bash
$ ynab-cli --help
```
Help text displays correctly with:
- Command descriptions
- Usage examples
- Options documentation
- Environment variables
- Example commands

### ✓ Error Handling
```bash
$ ynab-cli status
Error: YNAB_ACCESS_TOKEN environment variable is required
```
- Proper error message when token is missing
- Exits with code 1
- Clear instruction to user

## Commands Available
Based on `--help` output, the following commands are implemented:
- `status` - Show budget status and metadata
- `balance [filter]` - Show account balances
- `budget` - Show current month's budget with category details
- `categories` - List all categories with their IDs
- `add` - Add a new transaction

## Tests Requiring YNAB_ACCESS_TOKEN

The following tests require a valid YNAB access token:

### 1. Status Command
- [ ] Human-readable output: `ynab-cli status`
- [ ] JSON output: `ynab-cli status --json`

### 2. Balance Command
- [ ] Human-readable output: `ynab-cli balance`
- [ ] JSON output: `ynab-cli balance --json`
- [ ] With filter: `ynab-cli balance checking`

### 3. Budget Command
- [ ] Human-readable output: `ynab-cli budget`
- [ ] JSON output: `ynab-cli budget --json`

### 4. Categories Command
- [ ] Human-readable output: `ynab-cli categories`
- [ ] JSON output: `ynab-cli categories --json`

## Running Full Integration Tests

To run the complete integration test suite:

```bash
# Set your YNAB access token
export YNAB_ACCESS_TOKEN='your-token-here'

# Run the test script
./features/ynab/scripts/test-cli.sh
```

The test script will:
1. Verify token is set
2. Verify binary exists
3. Test all commands with human-readable output
4. Test all commands with --json flag
5. Test balance filtering
6. Verify --help and --version flags

## Expected Output Format

### Human-Readable Format
- Clean, formatted text suitable for terminal display
- Tables with aligned columns
- Currency values formatted with $ and commas
- Dates in readable format

### JSON Format
- Valid JSON output (parseable by jq)
- Structured data with all relevant fields
- Consistent field names across commands
- Suitable for scripting and automation

## Next Steps

To complete integration testing:
1. Obtain YNAB_ACCESS_TOKEN from https://app.ynab.com/settings/developer
2. Export the token: `export YNAB_ACCESS_TOKEN='your-token'`
3. Run: `./features/ynab/scripts/test-cli.sh`
4. Verify all commands produce expected output
5. Test JSON output is valid (use `jq` to parse)
6. Verify filtering works correctly
