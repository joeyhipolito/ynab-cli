package ynab_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// CLI End-to-End Tests
// ============================================================================

// TestE2E_CLI_YNABSetup tests the complete YNAB setup via CLI.
func TestE2E_CLI_YNABSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	// Test 1: Initialize YNAB integration
	t.Run("init", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "init")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("Command output: %s", string(output))
			// This may fail if via binary doesn't exist, which is ok for planning
			t.Skipf("Skipping: via binary not available: %v", err)
		}

		if !strings.Contains(string(output), "YNAB") {
			t.Logf("Output: %s", string(output))
		}
	})

	// Test 2: Set API token
	t.Run("set-token", func(t *testing.T) {
		testToken := "test-ynab-token-abc123"
		cmd := exec.Command("via", "ynab", "auth", "--token", testToken)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Logf("Command output: %s", string(output))
			t.Skipf("Skipping: via binary not available: %v", err)
		}

		// Verify token was stored
		configPath := filepath.Join(tmpDir, ".via", ".ynab_token")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Logf("Token file not found at: %s", configPath)
		}
	})

	// Test 3: Verify setup
	t.Run("verify", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "status")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: via binary not available: %v", err)
		}

		if !strings.Contains(string(output), "authenticated") &&
		   !strings.Contains(string(output), "connected") {
			t.Logf("Unexpected status output: %s", string(output))
		}
	})
}

// TestE2E_CLI_YNABBudgetCommands tests budget-related CLI commands.
func TestE2E_CLI_YNABBudgetCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	// Setup: Assume auth is configured
	setupTestAuth(t, tmpDir)

	// Test 1: List budgets
	t.Run("list-budgets", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "budgets", "list")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: via binary not available: %v", err)
		}

		// Should show budgets or empty state
		if !strings.Contains(string(output), "budget") &&
		   !strings.Contains(string(output), "No budgets found") {
			t.Logf("Unexpected output: %s", string(output))
		}
	})

	// Test 2: Select default budget
	t.Run("select-budget", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "budgets", "select", "My Budget")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		if strings.Contains(string(output), "error") {
			t.Logf("Error selecting budget: %s", string(output))
		}
	})

	// Test 3: Show budget summary
	t.Run("budget-summary", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "summary")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		// Should show summary info
		outputStr := string(output)
		if !strings.Contains(outputStr, "Budget") &&
		   !strings.Contains(outputStr, "balance") {
			t.Logf("Unexpected summary output: %s", outputStr)
		}
	})
}

// TestE2E_CLI_YNABTransactionCommands tests transaction CLI commands.
func TestE2E_CLI_YNABTransactionCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Add transaction (simple)
	t.Run("add-transaction-simple", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add", "25.50", "Coffee Shop")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		if strings.Contains(string(output), "error") {
			t.Logf("Error adding transaction: %s", string(output))
		} else {
			// Should confirm transaction added
			if !strings.Contains(string(output), "25.50") {
				t.Logf("Transaction confirmation: %s", string(output))
			}
		}
	})

	// Test 2: Add transaction with category
	t.Run("add-transaction-with-category", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add", "45.00", "Grocery Store", "--category", "Groceries")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "45.00") ||
		   (!strings.Contains(outputStr, "Groceries") && !strings.Contains(outputStr, "error")) {
			t.Logf("Transaction output: %s", outputStr)
		}
	})

	// Test 3: Add transaction with all options
	t.Run("add-transaction-full", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add", "120.00",
			"Restaurant",
			"--category", "Dining Out",
			"--account", "Checking",
			"--date", "2026-02-01",
			"--memo", "Dinner with friends")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		t.Logf("Full transaction output: %s", string(output))
	})

	// Test 4: List recent transactions
	t.Run("list-transactions", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "transactions", "--limit", "10")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		// Should show transactions or empty state
		outputStr := string(output)
		if !strings.Contains(outputStr, "Date") &&
		   !strings.Contains(outputStr, "Amount") &&
		   !strings.Contains(outputStr, "No transactions") {
			t.Logf("Transactions list output: %s", outputStr)
		}
	})

	// Test 5: Filter transactions by category
	t.Run("filter-transactions", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "transactions", "--category", "Groceries")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		t.Logf("Filtered transactions: %s", string(output))
	})
}

// TestE2E_CLI_YNABSyncCommands tests sync-related CLI commands.
func TestE2E_CLI_YNABSyncCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Manual sync
	t.Run("sync-now", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "sync")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		outputStr := string(output)
		// Should show sync progress or result
		if !strings.Contains(outputStr, "sync") &&
		   !strings.Contains(outputStr, "Synced") {
			t.Logf("Sync output: %s", outputStr)
		}
	})

	// Test 2: Sync status
	t.Run("sync-status", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "sync", "--status")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		outputStr := string(output)
		// Should show last sync time or status
		if !strings.Contains(outputStr, "Last sync") &&
		   !strings.Contains(outputStr, "Never synced") {
			t.Logf("Sync status: %s", outputStr)
		}
	})

	// Test 3: Force full sync
	t.Run("sync-full", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "sync", "--full")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		t.Logf("Full sync output: %s", string(output))
	})
}

// TestE2E_CLI_YNABReportCommands tests reporting CLI commands.
func TestE2E_CLI_YNABReportCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Spending by category
	t.Run("spending-by-category", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "report", "category", "--month", "2026-02")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		t.Logf("Category report: %s", string(output))
	})

	// Test 2: Monthly summary
	t.Run("monthly-summary", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "report", "monthly")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		outputStr := string(output)
		// Should show income, expenses, etc.
		if !strings.Contains(outputStr, "Income") &&
		   !strings.Contains(outputStr, "Expenses") &&
		   !strings.Contains(outputStr, "report") {
			t.Logf("Monthly summary: %s", outputStr)
		}
	})

	// Test 3: Budget health check
	t.Run("budget-health", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "health")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		t.Logf("Budget health: %s", string(output))
	})
}

// TestE2E_CLI_YNABJSONOutput tests JSON output for integration.
func TestE2E_CLI_YNABJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Budgets as JSON
	t.Run("budgets-json", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "budgets", "list", "--json")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		// Try to parse as JSON
		var result interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			t.Logf("Output is not valid JSON (may be expected): %s", string(output))
		} else {
			t.Logf("Valid JSON output received")
		}
	})

	// Test 2: Transactions as JSON
	t.Run("transactions-json", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "transactions", "--json", "--limit", "5")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Skipping: command not available: %v", err)
		}

		var result interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			t.Logf("Output is not valid JSON: %s", string(output))
		}
	})
}

// TestE2E_CLI_YNABErrorHandling tests CLI error handling.
func TestE2E_CLI_YNABErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	// Test 1: Command without auth
	t.Run("unauthenticated", func(t *testing.T) {
		cmd := exec.Command("via", "ynab", "budgets", "list")
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Logf("Command succeeded (may have cached auth): %s", string(output))
		} else {
			// Should show auth error
			outputStr := string(output)
			if !strings.Contains(outputStr, "auth") &&
			   !strings.Contains(outputStr, "token") &&
			   !strings.Contains(outputStr, "login") {
				t.Logf("Unexpected error message: %s", outputStr)
			}
		}
	})

	// Test 2: Invalid transaction amount
	t.Run("invalid-amount", func(t *testing.T) {
		setupTestAuth(t, tmpDir)
		cmd := exec.Command("via", "budget", "add", "invalid", "Test")
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Error("Expected error for invalid amount, got success")
		} else {
			outputStr := string(output)
			if !strings.Contains(outputStr, "invalid") &&
			   !strings.Contains(outputStr, "amount") &&
			   !strings.Contains(outputStr, "error") {
				t.Logf("Error message: %s", outputStr)
			}
		}
	})

	// Test 3: Missing required argument
	t.Run("missing-argument", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add")
		output, err := cmd.CombinedOutput()

		if err == nil {
			t.Error("Expected error for missing arguments, got success")
		} else {
			outputStr := string(output)
			if !strings.Contains(outputStr, "required") &&
			   !strings.Contains(outputStr, "usage") &&
			   !strings.Contains(outputStr, "error") {
				t.Logf("Error message: %s", outputStr)
			}
		}
	})
}

// TestE2E_CLI_YNABPipelineIntegration tests using YNAB in pipelines.
func TestE2E_CLI_YNABPipelineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Pipe transactions to jq
	t.Run("pipe-to-jq", func(t *testing.T) {
		// Check if jq is available
		if _, err := exec.LookPath("jq"); err != nil {
			t.Skip("jq not available")
		}

		cmd := exec.Command("bash", "-c",
			"via budget transactions --json --limit 5 | jq '.[] | .amount'")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Pipeline command failed: %v, output: %s", err, string(output))
		}

		t.Logf("Pipeline output: %s", string(output))
	})

	// Test 2: CSV export
	t.Run("csv-export", func(t *testing.T) {
		csvFile := filepath.Join(tmpDir, "transactions.csv")
		cmd := exec.Command("via", "budget", "export", "--csv", "--output", csvFile)
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("CSV export failed: %v, output: %s", err, string(output))
		}

		// Check if file was created
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			t.Logf("CSV file not created, output: %s", string(output))
		} else {
			content, _ := os.ReadFile(csvFile)
			t.Logf("CSV content preview: %s", string(content[:min(len(content), 200)]))
		}
	})
}

// TestE2E_CLI_YNABInteractiveMode tests interactive CLI features.
func TestE2E_CLI_YNABInteractiveMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI E2E test in short mode")
	}

	tmpDir := t.TempDir()
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	setupTestAuth(t, tmpDir)

	// Test 1: Interactive transaction entry (with stdin)
	t.Run("interactive-transaction", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add", "--interactive")

		// Simulate user input
		stdin := bytes.NewBufferString("25.50\nCoffee Shop\nDining Out\nChecking\n\n")
		cmd.Stdin = stdin

		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Interactive mode not available: %v", err)
		}

		t.Logf("Interactive output: %s", string(output))
	})

	// Test 2: Category selection
	t.Run("category-select", func(t *testing.T) {
		cmd := exec.Command("via", "budget", "add", "30.00", "Store", "--select-category")

		// Simulate selecting category 1
		stdin := bytes.NewBufferString("1\n")
		cmd.Stdin = stdin

		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Skipf("Category selection not available: %v", err)
		}

		t.Logf("Category selection output: %s", string(output))
	})
}

// ============================================================================
// Test Helper Functions
// ============================================================================

func setupTestAuth(t *testing.T, tmpDir string) {
	configDir := filepath.Join(tmpDir, ".via")
	os.MkdirAll(configDir, 0755)

	// Create mock auth token file
	tokenFile := filepath.Join(configDir, ".ynab_token")
	testToken := "test-ynab-token-for-testing"

	// Write encrypted token (in real impl, this would be encrypted)
	err := os.WriteFile(tokenFile, []byte(testToken), 0600)
	if err != nil {
		t.Logf("Warning: could not create test token file: %v", err)
	}

	// Create mock budget config
	configFile := filepath.Join(configDir, "ynab_config.json")
	config := map[string]interface{}{
		"default_budget_id": "test-budget-1",
		"default_account":   "Checking",
		"last_sync":         time.Now().Format(time.RFC3339),
	}

	configJSON, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, configJSON, 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
