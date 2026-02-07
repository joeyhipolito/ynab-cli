package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/joeyhipolito/ynab-cli/internal/transform"
)

func TestAmountParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "positive expense (default)",
			input:    "50.00",
			expected: -50000, // Should be negative (expense)
		},
		{
			name:     "explicit negative expense",
			input:    "-50.00",
			expected: -50000,
		},
		{
			name:     "explicit positive income",
			input:    "+100.00",
			expected: 100000,
		},
		{
			name:     "integer amount",
			input:    "25",
			expected: -25000, // Should be negative (expense)
		},
		{
			name:     "cents",
			input:    "0.50",
			expected: -500,
		},
		{
			name:     "large amount",
			input:    "1234.56",
			expected: -1234560,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the parsing logic in AddCmd
			amountFloat, err := parseAmountString(tt.input)
			if err != nil {
				t.Fatalf("failed to parse amount: %v", err)
			}

			amountMilliunits := transform.DollarsToMilliunits(amountFloat)

			// Default to expense (negative) if positive amount is given
			// unless it's explicitly marked with +
			if amountMilliunits > 0 && tt.input[0] != '+' {
				amountMilliunits = -amountMilliunits
			}

			if amountMilliunits != tt.expected {
				t.Errorf("expected %d milliunits, got %d", tt.expected, amountMilliunits)
			}
		})
	}
}

func TestFindAccountLogic(t *testing.T) {
	// Test case-insensitive matching logic
	accounts := []struct {
		name string
	}{
		{"Checking"},
		{"Savings Account"},
		{"Credit Card"},
	}

	tests := []struct {
		name      string
		search    string
		shouldErr bool
		expected  string
	}{
		{
			name:     "exact match",
			search:   "Checking",
			expected: "Checking",
		},
		{
			name:     "case insensitive",
			search:   "checking",
			expected: "Checking",
		},
		{
			name:     "partial match",
			search:   "savings",
			expected: "Savings Account",
		},
		{
			name:      "ambiguous match",
			search:    "c",
			shouldErr: true, // Matches "Checking" and "Credit Card"
		},
		{
			name:      "no match",
			search:    "investment",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the matching logic
			matches := []string{}
			searchLower := strings.ToLower(tt.search)

			// Exact match
			for _, acc := range accounts {
				if strings.ToLower(acc.name) == searchLower {
					matches = []string{acc.name}
					break
				}
			}

			// Partial match if no exact match
			if len(matches) == 0 {
				for _, acc := range accounts {
					if strings.Contains(strings.ToLower(acc.name), searchLower) {
						matches = append(matches, acc.name)
					}
				}
			}

			if tt.shouldErr {
				if len(matches) == 1 {
					t.Errorf("expected error but got match: %s", matches[0])
				}
			} else {
				if len(matches) != 1 {
					t.Errorf("expected single match, got %d matches", len(matches))
				} else if matches[0] != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, matches[0])
				}
			}
		})
	}
}

// Helper functions for tests

func parseAmountString(s string) (float64, error) {
	// Remove leading + if present for parsing
	if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	return strconv.ParseFloat(s, 64)
}
