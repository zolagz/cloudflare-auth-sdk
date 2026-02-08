# Cloudflare Auth SDK

A professional Go SDK for user authentication using Cloudflare Workers KV storage, featuring JWT-based authentication, secure password hashing, and comprehensive KV operations.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## ✨ Features

- 🔐 Complete authentication workflow (registration, login, token validation)
- 🔑 JWT token generation and validation
- 💾 Cloudflare Workers KV backed storage
- 🛡️ Secure password hashing with bcrypt
- 🚀 Simple, flat API design
- 📦 Built on Cloudflare API v6 SDK
- 🎯 Production-ready error handling
- 📚 Comprehensive documentation

## 📦 Installation

```bash
go get github.com/zolagz/cloudflare-auth-sdk
```

## 🚀 Quick Start

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
    // Create SDK client
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
    user, err := client.Register(ctx, "user@example.com", "SecurePassword123!")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ User registered: %s\n", user.Email)

    // Login
    loginResp, err := client.Login(ctx, "user@example.com", "SecurePassword123!")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Login successful! Token: %s\n", loginResp.Token)

    // Validate token
    validUser, err := client.ValidateToken(ctx, loginResp.Token)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Token valid for: %s\n", validUser.Email)
}
```

## 📖 Documentation

- **[Getting Started Guide](./docs/getting-started.md)** - Learn the basics
- **[API Reference](./docs/api-reference.md)** - Complete API documentation
- **[Advanced Usage](./docs/advanced-usage.md)** - Complex scenarios and best practices
- **[Examples](./examples/)** - Code examples for common use cases

## 🔑 Core Features

### Authentication

```go
// Register new user
user, err := client.Register(ctx, "user@example.com", "password")

// Login and get JWT token
loginResp, err := client.Login(ctx, "user@example.com", "password")

// Validate JWT token
userInfo, err := client.ValidateToken(ctx, token)

// Get user by ID
user, err := client.GetUserByID(ctx, userID)

// Delete user
err := client.DeleteUser(ctx, userID)
```

### Key-Value Operations

```go
// Store data
err := client.KVSet(ctx, "key", []byte("value"), nil)

// Store with expiration
err := client.KVSet(ctx, "session:123", data, &sdk.KVWriteOptions{
    ExpirationTTL: 3600, // 1 hour in seconds
})

// Retrieve data
value, err := client.KVGet(ctx, "key")

// List keys with prefix
keys, err := client.KVList(ctx, "users:", 100)

// Delete single key
err := client.KVDelete(ctx, "key")

// Bulk delete
err := client.KVDeleteBulk(ctx, []string{"key1", "key2", "key3"})
```

## ⚙️ Configuration

### ClientOptions

```go
type ClientOptions struct {
    APIToken           string // Cloudflare API Token (required)
    AccountID          string // Cloudflare Account ID (required)
    NamespaceID        string // Workers KV Namespace ID (required)
    JWTSecret          string // JWT signing secret (required)
    JWTExpirationHours int    // JWT expiration (optional, default: 24)
}
```

### Environment Variables

For production, use environment variables:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_NAMESPACE_ID="your-namespace-id"
export JWT_SECRET="your-jwt-secret"
```

## 🛡️ Error Handling

The SDK provides helper functions for common error types:

```go
user, err := client.GetUserByID(ctx, userID)
if err != nil {
    if sdk.IsUserNotFound(err) {
        // Handle user not found
    } else if sdk.IsUnauthorized(err) {
        // Handle unauthorized access
    } else if sdk.IsInternalError(err) {
        // Handle internal error
    }
}
```

Available error helpers:
- `IsUserNotFound(err)` - User doesn't exist
- `IsUserAlreadyExists(err)` - User already registered
- `IsInvalidCredentials(err)` - Wrong password
- `IsUnauthorized(err)` - Invalid or expired token
- `IsInternalError(err)` - Server-side error

## 📁 Project Structure

```
.
├── client.go                  # Main SDK client
├── types.go                   # Type definitions
├── errors.go                  # Error handling
├── options.go                 # Client options
├── kv.go                      # KV operations
├── docs/                      # Documentation
│   ├── getting-started.md
│   ├── api-reference.md
│   └── advanced-usage.md
├── examples/                  # Usage examples
│   ├── quickstart/
│   ├── custom-auth/
│   └── kv-operations/
├── pkg/
│   └── version/              # Version info
└── internal/                 # Internal packages
    └── testutil/             # Testing utilities
```

## 🧪 Examples

### Quickstart

```bash
cd examples/quickstart
export CLOUDFLARE_API_TOKEN="..."
export CLOUDFLARE_ACCOUNT_ID="..."
export CLOUDFLARE_NAMESPACE_ID="..."
export JWT_SECRET="..."
go run main.go
```

### KV Operations

```bash
cd examples/kv-operations
go run main.go
```

See [examples/](./examples/) for more examples.

## 🔒 Security

- **Password Storage**: Passwords are hashed using bcrypt (cost factor 10)
- **JWT Signing**: Use a strong secret (minimum 32 characters recommended)
- **API Token**: Keep your Cloudflare API token secure, never commit to version control
- **HTTPS Only**: Always use HTTPS in production environments
- **Token Validation**: Tokens are validated on every request
- **Error Messages**: Sensitive information is never exposed in error messages

## 🚀 Best Practices

1. **Reuse Client**: Create one client instance and reuse it
2. **Context Timeouts**: Always use context with appropriate timeouts
3. **Error Handling**: Use the provided error helper functions
4. **Strong Passwords**: Enforce password complexity in your application
5. **Secure Storage**: Store tokens in secure, HTTP-only cookies or secure storage
6. **Rate Limiting**: Implement rate limiting for authentication endpoints
7. **Logging**: Use structured logging, never log tokens or passwords

## 📊 Version

Current version: **1.0.0**

## 📝 License

[MIT License](LICENSE)

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## 📮 Support

If you encounter any issues or have questions:
- Open an issue on GitHub
- Check the [documentation](./docs/)
- Review the [examples](./examples/)

---

Built with ❤️ using Cloudflare Workers KV
