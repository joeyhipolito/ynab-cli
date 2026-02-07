package transform

import (
	"math"
	"testing"
	"time"
)

func TestDollarsToMilliunits(t *testing.T) {
	tests := []struct {
		name     string
		dollars  float64
		expected int64
	}{
		// Basic conversions
		{"zero", 0.0, 0},
		{"one dollar", 1.0, 1000},
		{"one hundred dollars", 100.0, 100000},
		{"one cent", 0.01, 10},
		{"fifty cents", 0.50, 500},

		// Negative amounts (expenses)
		{"negative one dollar", -1.0, -1000},
		{"negative fifty dollars", -50.0, -50000},
		{"negative one cent", -0.01, -10},

		// Fractional milliunits (rounding)
		// Go uses banker's rounding (round half to even)
		{"rounds half up", 1.5055, 1506}, // 1505.5 rounds to 1506 (even)
		{"rounds half down", 1.5045, 1505}, // 1504.5 rounds to 1505 (even)
		{"three decimal places", 1.505, 1505},

		// Large numbers
		{"one thousand dollars", 1000.0, 1000000},
		{"ten thousand dollars", 10000.0, 10000000},
		{"one million dollars", 1000000.0, 1000000000},

		// Small fractional amounts
		{"very small positive", 0.001, 1},
		{"very small negative", -0.001, -1},
		{"sub-milliunit positive", 0.0005, 1}, // Rounds up
		{"sub-milliunit negative", -0.0005, -1}, // Rounds down (away from zero)

		// Real-world examples
		{"grocery bill", 47.32, 47320},
		{"rent", 1250.00, 1250000},
		{"coffee", 4.75, 4750},
		{"refund", -23.50, -23500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DollarsToMilliunits(tt.dollars)
			if result != tt.expected {
				t.Errorf("DollarsToMilliunits(%f) = %d, want %d", tt.dollars, result, tt.expected)
			}
		})
	}
}

func TestMilliunitsToDollars(t *testing.T) {
	tests := []struct {
		name       string
		milliunits int64
		expected   float64
	}{
		// Basic conversions
		{"zero", 0, 0.0},
		{"one thousand milliunits", 1000, 1.0},
		{"one hundred thousand milliunits", 100000, 100.0},
		{"ten milliunits", 10, 0.01},
		{"five hundred milliunits", 500, 0.50},

		// Negative amounts
		{"negative one thousand", -1000, -1.0},
		{"negative fifty thousand", -50000, -50.0},
		{"negative ten", -10, -0.01},

		// Fractional dollars
		{"one thousand five hundred five", 1505, 1.505},
		{"one thousand five hundred four", 1504, 1.504},

		// Large numbers
		{"one million", 1000000, 1000.0},
		{"ten million", 10000000, 10000.0},
		{"one billion", 1000000000, 1000000.0},

		// Single milliunits
		{"one milliunit", 1, 0.001},
		{"negative one milliunit", -1, -0.001},

		// Real-world examples
		{"grocery bill", 47320, 47.32},
		{"rent", 1250000, 1250.0},
		{"coffee", 4750, 4.75},
		{"refund", -23500, -23.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MilliunitsToDollars(tt.milliunits)
			// Use a small epsilon for floating point comparison
			if math.Abs(result-tt.expected) > 0.0000001 {
				t.Errorf("MilliunitsToDollars(%d) = %f, want %f", tt.milliunits, result, tt.expected)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name       string
		milliunits int64
		expected   string
	}{
		// Basic formatting
		{"zero", 0, "$0.00"},
		{"one dollar", 1000, "$1.00"},
		{"one hundred dollars", 100000, "$100.00"},
		{"one cent", 10, "$0.01"},
		{"fifty cents", 500, "$0.50"},

		// Negative amounts
		{"negative one dollar", -1000, "-$1.00"},
		{"negative fifty dollars", -50000, "-$50.00"},
		{"negative one cent", -10, "-$0.01"},

		// Thousands separators
		{"one thousand dollars", 1000000, "$1,000.00"},
		{"ten thousand dollars", 10000000, "$10,000.00"},
		{"one hundred thousand", 100000000, "$100,000.00"},
		{"one million dollars", 1000000000, "$1,000,000.00"},
		{"negative thousands", -1000000, "-$1,000.00"},

		// Fractional cents (rounds)
		{"rounds to 1.50", 1505, "$1.50"},     // 1.505 rounds to 1.50 (displays with 2 decimals)
		{"rounds to 1.50 also", 1504, "$1.50"},     // 1.504 rounds to 1.50 (displays with 2 decimals)

		// Real-world examples
		{"grocery bill", 47320, "$47.32"},
		{"rent", 1250000, "$1,250.00"},
		{"coffee", 4750, "$4.75"},
		{"refund", -23500, "-$23.50"},
		{"salary", 5500000, "$5,500.00"},

		// Edge cases
		{"very small positive", 1, "$0.00"},      // 0.001 rounds to 0.00
		{"very small negative", -1, "-$0.00"},    // -0.001 rounds to -0.00
		{"nine cents", 90, "$0.09"},
		{"ninety-nine cents", 990, "$0.99"},

		// Multiple thousands separators
		{"twelve thousand", 12345678, "$12,345.68"},
		{"two hundred thousand", 234567890, "$234,567.89"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCurrency(tt.milliunits)
			if result != tt.expected {
				t.Errorf("FormatCurrency(%d) = %s, want %s", tt.milliunits, result, tt.expected)
			}
		})
	}
}

// TestRoundTrip verifies that converting dollars to milliunits and back
// preserves the value (within floating point precision).
func TestRoundTrip(t *testing.T) {
	testValues := []float64{
		0.0,
		1.0,
		100.0,
		0.01,
		-50.0,
		1234.56,
		-9876.54,
	}

	for _, original := range testValues {
		t.Run("round trip", func(t *testing.T) {
			milliunits := DollarsToMilliunits(original)
			dollars := MilliunitsToDollars(milliunits)

			// Allow small epsilon for floating point comparison
			if math.Abs(dollars-original) > 0.001 {
				t.Errorf("Round trip failed: %f -> %d -> %f", original, milliunits, dollars)
			}
		})
	}
}

// TestFormatWithThousands tests the internal helper function.
func TestFormatWithThousands(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		decimals int
		expected string
	}{
		{"no thousands", 123.45, 2, "123.45"},
		{"one thousand", 1234.56, 2, "1,234.56"},
		{"ten thousand", 12345.67, 2, "12,345.67"},
		{"hundred thousand", 123456.78, 2, "123,456.78"},
		{"million", 1234567.89, 2, "1,234,567.89"},
		{"no decimals", 1234.0, 0, "1,234"},
		{"three decimals", 1234.567, 3, "1,234.567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWithThousands(tt.value, tt.decimals)
			if result != tt.expected {
				t.Errorf("formatWithThousands(%f, %d) = %s, want %s", tt.value, tt.decimals, result, tt.expected)
			}
		})
	}
}

// TestAddThousandsSeparators tests the internal helper function.
func TestAddThousandsSeparators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single digit", "1", "1"},
		{"two digits", "12", "12"},
		{"three digits", "123", "123"},
		{"four digits", "1234", "1,234"},
		{"five digits", "12345", "12,345"},
		{"six digits", "123456", "123,456"},
		{"seven digits", "1234567", "1,234,567"},
		{"nine digits", "123456789", "123,456,789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addThousandsSeparators(tt.input)
			if result != tt.expected {
				t.Errorf("addThousandsSeparators(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// BenchmarkDollarsToMilliunits measures performance of conversion.
func BenchmarkDollarsToMilliunits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DollarsToMilliunits(1234.56)
	}
}

// BenchmarkMilliunitsToDollars measures performance of conversion.
func BenchmarkMilliunitsToDollars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MilliunitsToDollars(1234560)
	}
}

// BenchmarkFormatCurrency measures performance of formatting.
func BenchmarkFormatCurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatCurrency(1234560)
	}
}

// TestParseMonth tests parsing of YNAB month strings.
func TestParseMonth(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectYear  int
		expectMonth int
		expectError bool
	}{
		// Valid formats
		{"simple month", "2024-01", 2024, 1, false},
		{"december", "2024-12", 2024, 12, false},
		{"with day (ignored)", "2024-06-15", 2024, 6, false},
		{"january 2023", "2023-01", 2023, 1, false},
		{"future year", "2030-03", 2030, 3, false},
		{"past year", "2020-08", 2020, 8, false},

		// Edge cases - valid months
		{"january", "2024-01", 2024, 1, false},
		{"december", "2024-12", 2024, 12, false},

		// Invalid formats
		{"empty string", "", 0, 0, true},
		{"missing month", "2024", 0, 0, true},
		{"invalid format", "2024/01", 0, 0, true},
		{"non-numeric year", "abcd-01", 0, 0, true},
		{"non-numeric month", "2024-ab", 0, 0, true},

		// Invalid month values
		{"month too low", "2024-00", 0, 0, true},
		{"month too high", "2024-13", 0, 0, true},
		{"negative month", "2024--01", 0, 0, true},

		// Real-world examples
		{"current month", "2024-02", 2024, 2, false},
		{"budget month", "2024-07", 2024, 7, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, month, err := ParseMonth(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseMonth(%s) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMonth(%s) unexpected error: %v", tt.input, err)
				return
			}

			if year != tt.expectYear {
				t.Errorf("ParseMonth(%s) year = %d, want %d", tt.input, year, tt.expectYear)
			}

			if month != tt.expectMonth {
				t.Errorf("ParseMonth(%s) month = %d, want %d", tt.input, month, tt.expectMonth)
			}
		})
	}
}

// TestFormatMonth tests formatting of year/month to YNAB month string.
func TestFormatMonth(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    int
		expected string
	}{
		// Basic formatting
		{"january 2024", 2024, 1, "2024-01"},
		{"december 2024", 2024, 12, "2024-12"},
		{"june 2023", 2023, 6, "2023-06"},

		// Zero-padding
		{"single-digit month", 2024, 3, "2024-03"},
		{"double-digit month", 2024, 10, "2024-10"},

		// Edge cases
		{"year 2000", 2000, 1, "2000-01"},
		{"year 9999", 9999, 12, "9999-12"},

		// Real-world examples
		{"current budget month", 2024, 2, "2024-02"},
		{"future planning", 2025, 7, "2025-07"},
		{"past budget", 2020, 4, "2020-04"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMonth(tt.year, tt.month)
			if result != tt.expected {
				t.Errorf("FormatMonth(%d, %d) = %s, want %s", tt.year, tt.month, result, tt.expected)
			}
		})
	}
}

// TestParseDate tests parsing of YNAB date strings.
func TestParseDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectYear  int
		expectMonth time.Month
		expectDay   int
		expectZero  bool // true if we expect a zero time (invalid date)
	}{
		// Valid dates
		{"simple date", "2024-01-15", 2024, time.January, 15, false},
		{"first of month", "2024-06-01", 2024, time.June, 1, false},
		{"end of month", "2024-12-31", 2024, time.December, 31, false},
		{"leap year", "2024-02-29", 2024, time.February, 29, false},
		{"past date", "2020-03-15", 2020, time.March, 15, false},
		{"future date", "2030-08-22", 2030, time.August, 22, false},

		// Edge cases - valid
		{"january first", "2024-01-01", 2024, time.January, 1, false},
		{"december last", "2024-12-31", 2024, time.December, 31, false},

		// Invalid formats (should return zero time)
		{"empty string", "", 0, 0, 0, true},
		{"invalid format", "2024/01/15", 0, 0, 0, true},
		{"missing day", "2024-01", 0, 0, 0, true},
		{"non-numeric", "abcd-ef-gh", 0, 0, 0, true},
		{"invalid day", "2024-02-30", 0, 0, 0, true},
		{"invalid month", "2024-13-01", 0, 0, 0, true},

		// Real-world transaction dates
		{"transaction date 1", "2024-01-15", 2024, time.January, 15, false},
		{"transaction date 2", "2024-07-04", 2024, time.July, 4, false},
		{"transaction date 3", "2023-11-23", 2023, time.November, 23, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDate(tt.input)

			if tt.expectZero {
				if !result.IsZero() {
					t.Errorf("ParseDate(%s) expected zero time, got %v", tt.input, result)
				}
				return
			}

			if result.IsZero() {
				t.Errorf("ParseDate(%s) returned zero time, expected valid date", tt.input)
				return
			}

			if result.Year() != tt.expectYear {
				t.Errorf("ParseDate(%s) year = %d, want %d", tt.input, result.Year(), tt.expectYear)
			}

			if result.Month() != tt.expectMonth {
				t.Errorf("ParseDate(%s) month = %v, want %v", tt.input, result.Month(), tt.expectMonth)
			}

			if result.Day() != tt.expectDay {
				t.Errorf("ParseDate(%s) day = %d, want %d", tt.input, result.Day(), tt.expectDay)
			}
		})
	}
}

// TestFormatDate tests formatting of time.Time to YNAB date string.
func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		// Basic formatting
		{
			"simple date",
			time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
			"2024-01-15",
		},
		{
			"first of month",
			time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
			"2024-06-01",
		},
		{
			"end of month",
			time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			"2024-12-31",
		},

		// Time is ignored (only date matters)
		{
			"with time component",
			time.Date(2024, time.March, 15, 14, 30, 45, 0, time.UTC),
			"2024-03-15",
		},

		// Different timezones (date is formatted as-is)
		{
			"UTC timezone",
			time.Date(2024, time.July, 4, 0, 0, 0, 0, time.UTC),
			"2024-07-04",
		},

		// Edge cases
		{
			"zero time",
			time.Time{},
			"0001-01-01",
		},
		{
			"leap year date",
			time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			"2024-02-29",
		},

		// Real-world examples
		{
			"transaction date",
			time.Date(2024, time.January, 15, 12, 0, 0, 0, time.UTC),
			"2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDate(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMonthRoundTrip verifies that parsing and formatting months preserves values.
func TestMonthRoundTrip(t *testing.T) {
	testCases := []struct {
		year  int
		month int
	}{
		{2024, 1},
		{2024, 12},
		{2023, 6},
		{2025, 3},
		{2020, 11},
	}

	for _, tc := range testCases {
		t.Run("round trip", func(t *testing.T) {
			// Format to string
			formatted := FormatMonth(tc.year, tc.month)

			// Parse back
			year, month, err := ParseMonth(formatted)
			if err != nil {
				t.Errorf("Round trip failed: FormatMonth(%d, %d) -> ParseMonth(%s) error: %v",
					tc.year, tc.month, formatted, err)
				return
			}

			if year != tc.year || month != tc.month {
				t.Errorf("Round trip failed: (%d, %d) -> %s -> (%d, %d)",
					tc.year, tc.month, formatted, year, month)
			}
		})
	}
}

// TestDateRoundTrip verifies that parsing and formatting dates preserves values.
func TestDateRoundTrip(t *testing.T) {
	testDates := []time.Time{
		time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.March, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC), // Leap year
	}

	for _, original := range testDates {
		t.Run("round trip", func(t *testing.T) {
			// Format to string
			formatted := FormatDate(original)

			// Parse back
			parsed := ParseDate(formatted)
			if parsed.IsZero() {
				t.Errorf("Round trip failed: FormatDate(%v) -> ParseDate(%s) returned zero time",
					original, formatted)
				return
			}

			// Compare date components only (ignore time/timezone)
			if parsed.Year() != original.Year() ||
				parsed.Month() != original.Month() ||
				parsed.Day() != original.Day() {
				t.Errorf("Round trip failed: %v -> %s -> %v",
					original, formatted, parsed)
			}
		})
	}
}

// BenchmarkParseMonth measures performance of month parsing.
func BenchmarkParseMonth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseMonth("2024-06")
	}
}

// BenchmarkFormatMonth measures performance of month formatting.
func BenchmarkFormatMonth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FormatMonth(2024, 6)
	}
}

// BenchmarkParseDate measures performance of date parsing.
func BenchmarkParseDate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate("2024-06-15")
	}
}

// BenchmarkFormatDate measures performance of date formatting.
func BenchmarkFormatDate(b *testing.B) {
	date := time.Date(2024, time.June, 15, 0, 0, 0, 0, time.UTC)
	for i := 0; i < b.N; i++ {
		FormatDate(date)
	}
}
