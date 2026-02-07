# YNAB API Error Handling

This document describes the error handling and retry logic in the YNAB API client.

## Error Types

### YNABError

The `YNABError` type represents all errors returned by the YNAB API.

```go
type YNABError struct {
    Message    string  // Human-readable error message
    StatusCode int     // HTTP status code
    ErrorID    string  // YNAB error identifier (optional)
    Detail     string  // Additional error details (optional)
}
```

### Error Categories

The `YNABError` type provides helper methods to categorize errors:

- **`IsAuthError()`** - HTTP 401 (Unauthorized)
- **`IsRateLimitError()`** - HTTP 429 (Too Many Requests)
- **`IsServerError()`** - HTTP 5xx (Server errors)
- **`IsNotFoundError()`** - HTTP 404 (Not Found)
- **`IsBadRequestError()`** - HTTP 400 (Bad Request)
- **`IsRetryable()`** - Returns true for rate limit and server errors

## HTTP Status Code Handling

| Status Code | Behavior | Retries |
|-------------|----------|---------|
| **401 Unauthorized** | Returns auth error immediately | No |
| **400 Bad Request** | Returns error immediately | No |
| **404 Not Found** | Returns error immediately | No |
| **429 Rate Limit** | Waits for Retry-After, then retries | Yes |
| **5xx Server Error** | Retries with exponential backoff | Yes |

## Retry Logic

### Exponential Backoff

The client implements exponential backoff for retryable errors:

```
Attempt 1: Immediate
Attempt 2: Wait 1 second
Attempt 3: Wait 2 seconds
Attempt 4: Wait 4 seconds
```

**Configuration:**
- `MaxRetries = 3` (4 total attempts including initial)
- `InitialBackoff = 1 second`
- Backoff doubles after each retry

### Rate Limit Handling

For HTTP 429 (Rate Limit) errors:

1. Reads `Retry-After` header (defaults to 60 seconds if not present)
2. Waits for the specified duration
3. Retries the request
4. Does NOT count against exponential backoff counter

### Network Errors

Network errors (connection failures, timeouts) are retried with exponential backoff.

## Usage Examples

### Basic Error Handling

```go
client, err := api.NewClient("")
if err != nil {
    log.Fatal(err)
}

budgets, err := client.GetBudgets()
if err != nil {
    if api.IsAuthError(err) {
        log.Fatal("Authentication failed. Check YNAB_ACCESS_TOKEN")
    }
    log.Printf("Error: %v", err)
}
```

### Checking Error Types

```go
_, err := client.GetBudget("invalid-id", 0)
if err != nil {
    var ynabErr *api.YNABError
    if errors.As(err, &ynabErr) {
        switch {
        case ynabErr.IsAuthError():
            fmt.Println("Invalid credentials")
        case ynabErr.IsNotFoundError():
            fmt.Println("Budget not found")
        case ynabErr.IsRateLimitError():
            fmt.Println("Rate limited - try again later")
        case ynabErr.IsServerError():
            fmt.Println("YNAB server error - retry automatically")
        default:
            fmt.Printf("Error: %v\n", ynabErr)
        }
    }
}
```

### Using Helper Functions

```go
budgets, err := client.GetBudgets()
if err != nil {
    if api.IsAuthError(err) {
        return fmt.Errorf("authentication failed: %w", err)
    }
    if api.IsRateLimitError(err) {
        // Will retry automatically, but you can add custom logic
        log.Println("Rate limited - client will retry automatically")
    }
    return err
}
```

## Error Response Format

The YNAB API returns errors in this format:

```json
{
  "error": {
    "id": "error_id",
    "name": "Error Name",
    "detail": "Detailed error message"
  }
}
```

The client parses this and populates the `YNABError` fields accordingly.

## Testing

Comprehensive tests are provided in:

- `errors_test.go` - Error type behavior tests
- `retry_test.go` - Retry logic and backoff tests

Run tests:

```bash
go test -v ./internal/api -run Error
go test -v ./internal/api -run Retry
```

## Best Practices

1. **Always check for auth errors first** - these indicate configuration issues
2. **Don't retry manually** - the client handles retries automatically
3. **Use helper functions** - `IsAuthError()`, `IsRateLimitError()`, etc.
4. **Log retryable errors** - they indicate transient issues
5. **Handle non-retryable errors** - return these to the caller immediately

## Error Messages

Error messages follow this format:

```
[YNAB] <Message> (HTTP <StatusCode>) [Error ID: <ErrorID>]
Details: <Detail>
```

Example:

```
[YNAB] Unauthorized (HTTP 401) [Error ID: auth_123]
Details: Invalid or missing access token
```

## Implementation Notes

- **Thread-safe**: The retry logic is safe for concurrent use
- **Context-aware**: Uses request context for cancellation (future enhancement)
- **Idempotent**: Safe to retry GET requests; POST/PATCH are retried carefully
- **Observable**: All retry attempts can be logged by adding middleware
