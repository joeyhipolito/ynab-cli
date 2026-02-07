# YNAB API Client

Go implementation of the YNAB API client with automatic retries, rate limit handling, and comprehensive error handling.

## Features

- **Authentication**: Bearer token authentication via `YNAB_ACCESS_TOKEN` environment variable
- **Retry Logic**: Automatic retry with exponential backoff (3 retries max)
- **Rate Limiting**: Automatic handling of 429 responses with `Retry-After` header
- **Error Handling**: Structured error types with detailed error information
- **Type Safety**: Full type definitions for all API responses

## Usage

### Initialize Client

```go
import "github.com/joeyhipolito/via/features/ynab/internal/api"

// From environment variable
client, err := api.NewClient("")
if err != nil {
    log.Fatal(err)
}

// With explicit token
client, err := api.NewClient("your-access-token")
if err != nil {
    log.Fatal(err)
}
```

### Get Budgets

```go
budgets, err := client.GetBudgets()
if err != nil {
    log.Fatal(err)
}

for _, budget := range budgets {
    fmt.Printf("Budget: %s (%s)\n", budget.Name, budget.ID)
}
```

### Get Categories

```go
// Use default budget
categories, err := client.GetCategories("")
if err != nil {
    log.Fatal(err)
}

// Use specific budget
categories, err := client.GetCategories("budget-id")
if err != nil {
    log.Fatal(err)
}

for _, group := range categories {
    fmt.Printf("Group: %s\n", group.Name)
    for _, cat := range group.Categories {
        fmt.Printf("  - %s: %d\n", cat.Name, cat.Balance)
    }
}
```

### Update Category Budget

```go
category, err := client.UpdateCategoryBudget(
    "category-id",    // Category ID
    100000,           // Budgeted amount in milliunits ($100.00)
    "2026-02-01",     // Month (YYYY-MM-DD)
    "",               // Budget ID (empty = use default)
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated category: %s, budgeted: %d\n", category.Name, category.Budgeted)
```

### Create Transaction

```go
txn, err := client.CreateTransaction(&api.TransactionRequest{
    AccountID:  "account-id",
    Date:       "2026-02-02",
    Amount:     -50000,           // Negative for outflow ($50.00)
    PayeeName:  "Coffee Shop",
    CategoryID: "category-id",
    Memo:       "Morning coffee",
    Cleared:    "uncleared",
    Approved:   true,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created transaction: %s\n", txn.ID)
```

### Get Accounts

```go
accounts, err := client.GetAccounts("")
if err != nil {
    log.Fatal(err)
}

for _, account := range accounts {
    fmt.Printf("Account: %s, Balance: %d\n", account.Name, account.Balance)
}
```

## Error Handling

The client returns structured errors with detailed information:

```go
budgets, err := client.GetBudgets()
if err != nil {
    if ynabErr, ok := err.(*api.YNABError); ok {
        fmt.Printf("YNAB Error: %s\n", ynabErr.Message)
        fmt.Printf("Status: %d\n", ynabErr.StatusCode)
        fmt.Printf("Error ID: %s\n", ynabErr.ErrorID)
        fmt.Printf("Detail: %s\n", ynabErr.Detail)
    } else {
        fmt.Printf("General error: %v\n", err)
    }
    return
}
```

## Milliunits

All monetary amounts in the YNAB API are represented in "milliunits" (1/1000th of the currency unit):

- $1.00 = 1000 milliunits
- $100.00 = 100000 milliunits
- $0.01 = 10 milliunits

To convert:
- Dollars to milliunits: `amount * 1000`
- Milliunits to dollars: `milliunits / 1000.0`

## Testing

Run the test suite:

```bash
cd features/ynab/internal/api
go test -v
```

Tests include:
- Client initialization
- Request retry logic
- Rate limit handling
- Error parsing
- All API methods
- Input validation

## Implementation Notes

- **No external dependencies**: Uses only Go standard library
- **Session reuse**: HTTP client with connection pooling
- **Default budget**: Automatically uses first budget if not specified
- **Timeout**: 30-second timeout for all requests
- **User-Agent**: `Via-YNAB-CLI/1.0`
