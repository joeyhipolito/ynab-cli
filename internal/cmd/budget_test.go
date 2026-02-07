package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/joeyhipolito/ynab-cli/internal/api"
)

// TestBudgetCmd_Integration tests the budget command with real YNAB API (if token is available).
func TestBudgetCmd_Integration(t *testing.T) {
	if os.Getenv("YNAB_ACCESS_TOKEN") == "" {
		t.Skip("Skipping integration test: YNAB_ACCESS_TOKEN not set")
	}

	client, err := api.NewClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test human-readable output
	t.Run("human readable output", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := BudgetCmd(client, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("BudgetCmd failed: %v", err)
		}

		// Should contain budget information
		if !strings.Contains(output, "Budget for") {
			t.Errorf("Expected output to contain 'Budget for'")
		}

		if !strings.Contains(output, "Overall Totals") {
			t.Errorf("Expected output to contain 'Overall Totals'")
		}
	})

	// Test JSON output
	t.Run("json output", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := BudgetCmd(client, true)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("BudgetCmd failed: %v", err)
		}

		// Should be valid JSON
		var result BudgetOutput
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Invalid JSON output: %v", err)
		}

		// Verify JSON structure
		if result.Month == "" {
			t.Error("Expected month to be set")
		}
	})
}

// TestBudgetOutput_JSON tests JSON marshaling of BudgetOutput.
func TestBudgetOutput_JSON(t *testing.T) {
	output := BudgetOutput{
		Month: "2024-01-01",
		CategoryGroups: []CategoryGroup{
			{
				ID:   "group-1",
				Name: "Bills",
				Categories: []CategoryBudget{
					{
						ID:       "cat-1",
						Name:     "Rent",
						Budgeted: 1500000,
						Activity: -1500000,
						Balance:  0,
					},
					{
						ID:       "cat-2",
						Name:     "Utilities",
						Budgeted: 400000,
						Activity: -350000,
						Balance:  50000,
					},
				},
				TotalBudgeted: 1900000,
				TotalActivity: -1850000,
				TotalBalance:  50000,
			},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify JSON is valid and contains expected fields
	var unmarshaled BudgetOutput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled.Month != "2024-01-01" {
		t.Errorf("Expected month '2024-01-01', got '%s'", unmarshaled.Month)
	}

	if len(unmarshaled.CategoryGroups) != 1 {
		t.Errorf("Expected 1 category group, got %d", len(unmarshaled.CategoryGroups))
	}

	if unmarshaled.CategoryGroups[0].Name != "Bills" {
		t.Errorf("Expected 'Bills', got '%s'", unmarshaled.CategoryGroups[0].Name)
	}

	if len(unmarshaled.CategoryGroups[0].Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(unmarshaled.CategoryGroups[0].Categories))
	}
}
