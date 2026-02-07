//go:build integration
// +build integration

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joeyhipolito/ynab-cli/internal/api"
)

func main() {
	token := os.Getenv("YNAB_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("YNAB_ACCESS_TOKEN environment variable is required")
	}

	client, err := api.NewClient(token)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Testing YNAB API Client Integration")
	fmt.Println("====================================")

	// Test GetBudgets
	fmt.Println("\n1. Testing GetBudgets()...")
	budgets, err := client.GetBudgets()
	if err != nil {
		log.Fatalf("GetBudgets failed: %v", err)
	}
	fmt.Printf("✓ Found %d budget(s)\n", len(budgets))
	for i, budget := range budgets {
		fmt.Printf("  %d. %s (ID: %s)\n", i+1, budget.Name, budget.ID)
		if budget.LastModifiedOn != "" {
			fmt.Printf("     Last Modified: %s\n", budget.LastModifiedOn)
		}
	}

	if len(budgets) == 0 {
		fmt.Println("No budgets found. Cannot test GetAccounts.")
		return
	}

	// Test GetDefaultBudgetID
	fmt.Println("\n2. Testing GetDefaultBudgetID()...")
	defaultBudgetID, err := client.GetDefaultBudgetID()
	if err != nil {
		log.Fatalf("GetDefaultBudgetID failed: %v", err)
	}
	fmt.Printf("✓ Default Budget ID: %s\n", defaultBudgetID)

	// Test GetAccounts
	fmt.Println("\n3. Testing GetAccounts()...")
	accounts, err := client.GetAccounts(defaultBudgetID)
	if err != nil {
		log.Fatalf("GetAccounts failed: %v", err)
	}
	fmt.Printf("✓ Found %d account(s)\n", len(accounts))
	for i, account := range accounts {
		status := "Open"
		if account.Closed {
			status = "Closed"
		}
		if account.Deleted {
			status = "Deleted"
		}

		// Convert milliunits to dollars for display
		balance := float64(account.Balance) / 1000.0
		clearedBalance := float64(account.ClearedBalance) / 1000.0

		fmt.Printf("  %d. %s (%s) [%s]\n", i+1, account.Name, account.Type, status)
		fmt.Printf("     Balance: $%.2f\n", balance)
		fmt.Printf("     Cleared Balance: $%.2f\n", clearedBalance)
	}

	// Test GetAccounts with empty budget ID (should use default)
	fmt.Println("\n4. Testing GetAccounts(\"\") - using default budget...")
	accounts2, err := client.GetAccounts("")
	if err != nil {
		log.Fatalf("GetAccounts with empty budgetID failed: %v", err)
	}
	fmt.Printf("✓ Found %d account(s) using default budget\n", len(accounts2))

	fmt.Println("\n====================================")
	fmt.Println("All tests passed!")
}
