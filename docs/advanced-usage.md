# Advanced Usage

This guide covers advanced features and patterns for using the Cloudflare Auth SDK.

## Table of Contents

- [Custom JWT Expiration](#custom-jwt-expiration)
- [Advanced KV Operations](#advanced-kv-operations)
- [Error Handling Patterns](#error-handling-patterns)
- [Testing](#testing)
- [Best Practices](#best-practices)

## Custom JWT Expiration

You can customize the JWT token expiration time:

```go
client, err := sdk.NewClient(&sdk.ClientOptions{
    APIToken:           os.Getenv("CLOUDFLARE_API_TOKEN"),
    AccountID:          os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
    NamespaceID:        os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
    JWTSecret:          os.Getenv("JWT_SECRET"),
    JWTExpirationHours: 168, // 7 days
})
```

## Advanced KV Operations

### Storing Data with Expiration

```go
// Store data that expires after 1 hour
err := client.KVSet(ctx, "session:12345", []byte("session-data"), &sdk.KVWriteOptions{
    ExpirationTTL: 3600, // seconds
    Metadata:      "session metadata",
})
```

### Storing Data with Absolute Expiration Time

```go
// Expire at specific time
expireAt := time.Now().Add(24 * time.Hour).Unix()
err := client.KVSet(ctx, "cache:item", data, &sdk.KVWriteOptions{
    Expiration: expireAt,
    Metadata:   "cache entry",
})
```

### Listing Keys with Pagination

```go
// List all keys with a specific prefix
var allKeys []sdk.KVKey
cursor := ""

for {
    result, err := client.KVList(ctx, "users:", 1000)
    if err != nil {
        log.Fatal(err)
    }
    
    allKeys = append(allKeys, result...)
    
    // Check if there are more results
    if len(result) < 1000 {
        break
    }
}

fmt.Printf("Total keys found: %d\n", len(allKeys))
```

### Bulk Operations

```go
// Delete multiple keys at once
keysToDelete := []string{
    "temp:key1",
    "temp:key2",
    "temp:key3",
}

err := client.KVDeleteBulk(ctx, keysToDelete)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling Patterns

### Comprehensive Error Handling

```go
user, err := client.GetUserByID(ctx, userID)
if err != nil {
    switch {
    case sdk.IsUserNotFound(err):
        return nil, fmt.Errorf("user %s not found", userID)
    case sdk.IsUnauthorized(err):
        return nil, errors.New("unauthorized access")
    case sdk.IsInternalError(err):
        log.Printf("Internal error: %v", err)
        return nil, errors.New("service temporarily unavailable")
    default:
        log.Printf("Unexpected error: %v", err)
        return nil, err
    }
}
```

### Extracting Error Details

```go
import "errors"

var appErr *sdk.AppError
if errors.As(err, &appErr) {
    log.Printf("Error code: %s", appErr.Code)
    log.Printf("Error message: %s", appErr.Message)
    if appErr.Err != nil {
        log.Printf("Underlying error: %v", appErr.Err)
    }
}
```

### Retry Logic

```go
func registerWithRetry(ctx context.Context, client *sdk.Client, email, password string) (*sdk.User, error) {
    maxRetries := 3
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        user, err := client.Register(ctx, email, password)
        if err == nil {
            return user, nil
        }
        
        // Don't retry on user already exists
        if sdk.IsUserAlreadyExists(err) {
            return nil, err
        }
        
        // Only retry on internal errors
        if !sdk.IsInternalError(err) {
            return nil, err
        }
        
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## Testing

### Mock Client for Testing

```go
// Create a mock client for testing
type mockClient struct {
    RegisterFunc      func(context.Context, string, string) (*sdk.User, error)
    LoginFunc         func(context.Context, string, string) (*sdk.LoginResponse, error)
    ValidateTokenFunc func(context.Context, string) (*sdk.UserInfo, error)
}

func (m *mockClient) Register(ctx context.Context, email, password string) (*sdk.User, error) {
    return m.RegisterFunc(ctx, email, password)
}

// Use in tests
func TestMyFunction(t *testing.T) {
    mock := &mockClient{
        RegisterFunc: func(ctx context.Context, email, password string) (*sdk.User, error) {
            return &sdk.User{
                ID:    "test-id",
                Email: email,
            }, nil
        },
    }
    
    // Test your code with mock
}
```

### Integration Testing

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    client, err := sdk.NewClient(&sdk.ClientOptions{
        APIToken:    os.Getenv("CLOUDFLARE_API_TOKEN"),
        AccountID:   os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
        NamespaceID: os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
        JWTSecret:   "test-secret",
    })
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test registration
    user, err := client.Register(ctx, "test@example.com", "password123")
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
    
    // Cleanup
    defer client.DeleteUser(ctx, user.ID)
}
```

## Best Practices

### 1. Context Management

Always pass context with timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

user, err := client.Register(ctx, email, password)
```

### 2. Password Security

Enforce strong passwords in your application:

```go
func isStrongPassword(password string) error {
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
    
    if !hasUpper || !hasLower || !hasDigit {
        return errors.New("password must contain uppercase, lowercase, and digits")
    }
    
    return nil
}
```

### 3. Token Storage

Store tokens securely:

```go
// Don't log tokens
log.Printf("User logged in: %s", user.Email) // Good
log.Printf("Token: %s", token) // Bad!

// Use secure storage
// - HTTP-only cookies for web apps
// - Secure keychain for mobile apps
// - Never store in localStorage or plain text
```

### 4. Rate Limiting

Implement rate limiting to prevent abuse:

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 requests per second

func handleLogin(w http.ResponseWriter, r *http.Request) {
    if !limiter.Allow() {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Process login
}
```

### 5. Connection Pooling

The SDK uses the Cloudflare API client which handles connection pooling automatically. No additional configuration needed.

### 6. Graceful Shutdown

```go
func main() {
    client, err := sdk.NewClient(options)
    if err != nil {
        log.Fatal(err)
    }
    
    // Setup graceful shutdown
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    
    // Your application logic
    <-ctx.Done()
    log.Println("Shutting down gracefully...")
}
```

### 7. Structured Logging

```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

user, err := client.Register(ctx, email, password)
if err != nil {
    logger.Error("registration failed",
        "email", email,
        "error", err,
    )
    return err
}

logger.Info("user registered",
    "user_id", user.ID,
    "email", user.Email,
)
```

## Performance Tips

1. **Reuse Client**: Create one client instance and reuse it across requests
2. **Context Timeouts**: Always use context with appropriate timeouts
3. **Batch Operations**: Use `KVDeleteBulk` instead of multiple `KVDelete` calls
4. **Caching**: Cache frequently accessed data locally
5. **Connection Pooling**: The SDK handles this automatically

## See Also

- [API Reference](./api-reference.md)
- [Examples](../examples/)
- [Cloudflare Workers KV Documentation](https://developers.cloudflare.com/kv/)
