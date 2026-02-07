package api_test

import (
	"fmt"
	"log"

	"github.com/joeyhipolito/ynab-cli/internal/api"
)

// Example showing how to initialize the client and get budgets
func Example_getBudgets() {
	// Create client (reads YNAB_ACCESS_TOKEN from environment)
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Get all budgets
	budgets, err := client.GetBudgets()
	if err != nil {
		log.Fatal(err)
	}

	for _, budget := range budgets {
		fmt.Printf("Budget: %s (%s)\n", budget.Name, budget.ID)
	}
}

// Example showing how to get categories for a budget
func Example_getCategories() {
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Get categories (empty string uses default budget)
	categoryGroups, err := client.GetCategories("")
	if err != nil {
		log.Fatal(err)
	}

	for _, group := range categoryGroups {
		fmt.Printf("Category Group: %s\n", group.Name)
		for _, category := range group.Categories {
			// Convert milliunits to dollars for display
			balance := float64(category.Balance) / 1000.0
			fmt.Printf("  - %s: $%.2f\n", category.Name, balance)
		}
	}
}

// Example showing how to update a category budget
func Example_updateCategoryBudget() {
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Budget $100.00 for a category
	category, err := client.UpdateCategoryBudget(
		"category-id-here", // Replace with actual category ID
		100000,             // $100.00 in milliunits
		"2026-02-01",       // Month
		"",                 // Use default budget
	)
	if err != nil {
		log.Fatal(err)
	}

	budgeted := float64(category.Budgeted) / 1000.0
	fmt.Printf("Updated %s: budgeted $%.2f\n", category.Name, budgeted)
}

// Example showing how to create a transaction
func Example_createTransaction() {
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Create a $50.00 expense
	transaction, err := client.CreateTransaction(&api.TransactionRequest{
		AccountID:  "account-id-here",    // Replace with actual account ID
		Date:       "2026-02-02",         // ISO date format
		Amount:     -50000,               // Negative for outflow ($50.00)
		PayeeName:  "Coffee Shop",
		CategoryID: "category-id-here",   // Replace with actual category ID
		Memo:       "Morning coffee",
		Cleared:    "uncleared",
		Approved:   true,
	})
	if err != nil {
		log.Fatal(err)
	}

	amount := float64(transaction.Amount) / 1000.0
	fmt.Printf("Created transaction: $%.2f to %s\n", amount, transaction.PayeeName)
}

// Example showing how to get accounts
func Example_getAccounts() {
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	// Get all accounts
	accounts, err := client.GetAccounts("")
	if err != nil {
		log.Fatal(err)
	}

	for _, account := range accounts {
		balance := float64(account.Balance) / 1000.0
		fmt.Printf("Account: %s, Balance: $%.2f\n", account.Name, balance)
	}
}

// Example showing comprehensive error handling
func Example_errorHandling() {
	client, err := api.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	budgets, err := client.GetBudgets()
	if err != nil {
		// Check if it's a YNAB API error
		if ynabErr, ok := err.(*api.YNABError); ok {
			fmt.Printf("YNAB API Error:\n")
			fmt.Printf("  Message: %s\n", ynabErr.Message)
			fmt.Printf("  Status Code: %d\n", ynabErr.StatusCode)
			if ynabErr.ErrorID != "" {
				fmt.Printf("  Error ID: %s\n", ynabErr.ErrorID)
			}
			if ynabErr.Detail != "" {
				fmt.Printf("  Detail: %s\n", ynabErr.Detail)
			}
		} else {
			// Generic error
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	fmt.Printf("Successfully retrieved %d budgets\n", len(budgets))
}
