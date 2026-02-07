package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joeyhipolito/ynab-cli/internal/api"
)

// createTestServer creates an HTTP test server with mock YNAB API responses
func createTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Mock budgets endpoint
		if r.URL.Path == "/budgets" {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{
				"data": {
					"budgets": [
						{
							"id": "test-budget-1",
							"name": "Test Budget"
						}
					]
				}
			}`)
			return
		}

		// Mock accounts endpoint
		if strings.HasSuffix(r.URL.Path, "/accounts") {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{
				"data": {
					"accounts": [
						{
							"id": "acc-1",
							"name": "Checking Account",
							"type": "checking",
							"on_budget": true,
							"closed": false,
							"balance": 150000000,
							"cleared_balance": 145000000,
							"uncleared_balance": 5000000,
							"deleted": false
						},
						{
							"id": "acc-2",
							"name": "Savings Account",
							"type": "savings",
							"on_budget": true,
							"closed": false,
							"balance": 250000000,
							"cleared_balance": 250000000,
							"uncleared_balance": 0,
							"deleted": false
						},
						{
							"id": "acc-3",
							"name": "Credit Card",
							"type": "creditCard",
							"on_budget": true,
							"closed": false,
							"balance": -50000000,
							"cleared_balance": -45000000,
							"uncleared_balance": -5000000,
							"deleted": false
						},
						{
							"id": "acc-4",
							"name": "Investment Account",
							"type": "otherAsset",
							"on_budget": false,
							"closed": false,
							"balance": 500000000,
							"cleared_balance": 500000000,
							"uncleared_balance": 0,
							"deleted": false
						},
						{
							"id": "acc-5",
							"name": "Old Checking",
							"type": "checking",
							"on_budget": true,
							"closed": true,
							"balance": 0,
							"cleared_balance": 0,
							"uncleared_balance": 0,
							"deleted": false
						},
						{
							"id": "acc-6",
							"name": "Deleted Account",
							"type": "checking",
							"deleted": true
						}
					]
				}
			}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
}

// createTestClient creates a test API client using a test server
func createTestClient(t *testing.T, server *httptest.Server) *api.Client {
	t.Helper()
	client, err := api.NewClient("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Use reflection or a helper to override the base URL
	// Since we can't easily do this without modifying the api package,
	// we'll need to add a SetBaseURL method or similar
	// For now, we'll skip the full integration test and focus on unit tests
	return client
}

func TestBalanceCmd_Integration(t *testing.T) {
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

		err := BalanceCmd(client, "", false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("BalanceCmd failed: %v", err)
		}

		// Should contain account balances
		if !strings.Contains(output, "Account Balances") {
			t.Errorf("Expected output to contain 'Account Balances'")
		}
	})

	// Test JSON output
	t.Run("json output", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := BalanceCmd(client, "", true)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Errorf("BalanceCmd failed: %v", err)
		}

		// Should be valid JSON
		var result BalanceOutput
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Invalid JSON output: %v", err)
		}
	})
}

func TestFormatAccountType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"checking", "Checking"},
		{"savings", "Savings"},
		{"creditCard", "Credit Card"},
		{"cash", "Cash"},
		{"lineOfCredit", "Line of Credit"},
		{"otherAsset", "Other Asset"},
		{"otherLiability", "Other Liability"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := formatAccountType(tt.input)
			if got != tt.want {
				t.Errorf("formatAccountType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBalanceOutput_JSON(t *testing.T) {
	// Test JSON marshaling
	output := BalanceOutput{
		Accounts: []AccountBalance{
			{
				ID:               "acc-1",
				Name:             "Test Account",
				Type:             "checking",
				Balance:          100000,
				ClearedBalance:   90000,
				UnclearedBalance: 10000,
				OnBudget:         true,
				Closed:           false,
			},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify JSON is valid and contains expected fields
	var unmarshaled BalanceOutput
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(unmarshaled.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(unmarshaled.Accounts))
	}

	if unmarshaled.Accounts[0].Name != "Test Account" {
		t.Errorf("Expected 'Test Account', got '%s'", unmarshaled.Accounts[0].Name)
	}
}
