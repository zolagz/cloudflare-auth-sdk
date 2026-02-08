# Getting Started with Cloudflare Auth SDK

This guide will help you get started with the Cloudflare Auth SDK for Go.

## Installation

```bash
go get github.com/zolagz/cloudflare-auth-sdk
```

## Prerequisites

Before using the SDK, you need:

1. **Cloudflare Account** with API access
2. **Workers KV Namespace** created in your Cloudflare account
3. **API Token** with Workers KV edit permissions
4. **JWT Secret** for signing authentication tokens

## Configuration

The SDK requires the following configuration:

```go
import sdk "github.com/zolagz/cloudflare-auth-sdk"

client, err := sdk.NewClient(&sdk.ClientOptions{
    APIToken:           "your-cloudflare-api-token",
    AccountID:          "your-cloudflare-account-id",
    NamespaceID:        "your-kv-namespace-id",
    JWTSecret:          "your-jwt-secret-key",
    JWTExpirationHours: 24, // Optional, defaults to 24 hours
})
if err != nil {
    log.Fatal(err)
}
```

### Using Environment Variables

For production use, it's recommended to use environment variables:

```go
client, err := sdk.NewClient(&sdk.ClientOptions{
    APIToken:           os.Getenv("CLOUDFLARE_API_TOKEN"),
    AccountID:          os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
    NamespaceID:        os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
    JWTSecret:          os.Getenv("JWT_SECRET"),
    JWTExpirationHours: 24,
})
```

## Basic Usage

### User Registration

```go
ctx := context.Background()

user, err := client.Register(ctx, "user@example.com", "SecurePassword123!")
if err != nil {
    if sdk.IsUserAlreadyExists(err) {
        // Handle existing user
    } else {
        log.Fatal(err)
    }
}

fmt.Printf("User registered: %s (ID: %s)\n", user.Email, user.ID)
```

### User Login

```go
loginResp, err := client.Login(ctx, "user@example.com", "SecurePassword123!")
if err != nil {
    if sdk.IsInvalidCredentials(err) {
        // Handle invalid credentials
    } else {
        log.Fatal(err)
    }
}

// Use the token for subsequent requests
token := loginResp.Token
```

### Token Validation

```go
user, err := client.ValidateToken(ctx, token)
if err != nil {
    if sdk.IsUnauthorized(err) {
        // Handle invalid/expired token
    } else {
        log.Fatal(err)
    }
}

fmt.Printf("Token is valid for user: %s\n", user.Email)
```

### Get User Information

```go
user, err := client.GetUserByID(ctx, userID)
if err != nil {
    if sdk.IsUserNotFound(err) {
        // Handle user not found
    } else {
        log.Fatal(err)
    }
}
```

### Delete User

```go
err := client.DeleteUser(ctx, userID)
if err != nil {
    log.Fatal(err)
}
```

## Error Handling

The SDK provides helper functions for common error types:

```go
if sdk.IsUserNotFound(err) {
    // User doesn't exist
}

if sdk.IsUserAlreadyExists(err) {
    // User already registered
}

if sdk.IsInvalidCredentials(err) {
    // Wrong password
}

if sdk.IsUnauthorized(err) {
    // Invalid or expired token
}

if sdk.IsInternalError(err) {
    // Server-side error
}
```

## Key-Value Operations

The SDK also supports direct KV operations:

```go
// Store data
err := client.KVSet(ctx, "mykey", []byte("myvalue"), nil)

// Retrieve data
value, err := client.KVGet(ctx, "mykey")

// Delete data
err := client.KVDelete(ctx, "mykey")

// List keys with prefix
keys, err := client.KVList(ctx, "prefix:", 100)
```

See [Advanced Usage](./advanced-usage.md) for more KV operations.

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    sdk "github.com/zolagz/cloudflare-auth-sdk"
)

func main() {
    // Create client
    client, err := sdk.NewClient(&sdk.ClientOptions{
        APIToken:           os.Getenv("CLOUDFLARE_API_TOKEN"),
        AccountID:          os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
        NamespaceID:        os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
        JWTSecret:          os.Getenv("JWT_SECRET"),
        JWTExpirationHours: 24,
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Register user
    user, err := client.Register(ctx, "user@example.com", "password123")
    if err != nil {
        log.Fatal(err)
    }

    // Login
    loginResp, err := client.Login(ctx, "user@example.com", "password123")
    if err != nil {
        log.Fatal(err)
    }

    // Validate token
    validUser, err := client.ValidateToken(ctx, loginResp.Token)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Success! User: %s\n", validUser.Email)
}
```

## Next Steps

- Read the [API Reference](./api-reference.md) for detailed documentation
- Explore [Advanced Usage](./advanced-usage.md) for complex scenarios
- Check the [examples](../examples/) directory for more code samples
