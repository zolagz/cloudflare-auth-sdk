# SDK Directory Structure

```
cloudflare-auth-sdk/
│
├── 📄 README.md                    # Main documentation
├── 📄 go.mod                       # Go module definition
├── 📄 go.sum                       # Dependency checksums
├── 📄 Makefile                     # Build automation
│
├── 🔧 Core SDK Files (Root Level - Public API)
│   ├── client.go                  # Main SDK client
│   ├── types.go                   # Type definitions
│   ├── errors.go                  # Error handling
│   ├── options.go                 # Client configuration
│   └── kv.go                      # KV operations
│
├── 📚 docs/                        # Documentation
│   ├── getting-started.md         # Beginner guide
│   ├── api-reference.md           # Complete API docs
│   └── advanced-usage.md          # Advanced patterns
│
├── 💡 examples/                    # Usage examples
│   ├── quickstart/                # Basic usage
│   │   └── main.go
│   ├── kv-operations/             # KV-specific examples
│   │   └── main.go
│   └── custom-auth/               # Custom authentication
│       └── main.go
│
├── 📦 pkg/                         # Public packages
│   └── version/                   # Version information
│       └── version.go
│
└── 🔒 internal/                    # Internal packages
    └── testutil/                  # Testing utilities
        └── (test helpers)

Import Path: github.com/zolagz/cloudflare-auth-sdk
```

## Quick Reference

### Public API (Importable)

```go
import sdk "github.com/zolagz/cloudflare-auth-sdk"
```

All public types and functions are accessible from the root package:
- `sdk.Client` - Main SDK client
- `sdk.ClientOptions` - Configuration options
- `sdk.User`, `sdk.UserInfo`, `sdk.LoginResponse` - Data types
- `sdk.NewClient()` - Client constructor
- `sdk.IsUserNotFound()`, `sdk.IsUnauthorized()` - Error helpers

### File Responsibilities

| File | Purpose | Key Exports |
|------|---------|-------------|
| `client.go` | Main SDK client | `Client`, `NewClient()`, `Register()`, `Login()`, `ValidateToken()` |
| `types.go` | Type definitions | `User`, `UserInfo`, `LoginResponse`, `Claims`, `KVKey` |
| `errors.go` | Error handling | `AppError`, `ErrUserNotFound`, error helpers |
| `options.go` | Configuration | `ClientOptions`, builder methods |
| `kv.go` | KV operations | `KVGet()`, `KVSet()`, `KVDelete()`, `KVList()` |

### Directory Purposes

| Directory | Access | Purpose |
|-----------|--------|---------|
| `docs/` | Public | User documentation |
| `examples/` | Public | Code examples |
| `pkg/` | Public | Reusable packages |
| `internal/` | Private | Internal utilities (not importable) |

## Design Principles

1. **Flat API** - All functionality accessible from single import
2. **Single Client** - One client instance for all operations
3. **Clear Separation** - Public API at root, internals in `internal/`
4. **Standard Layout** - Follows Go community conventions
5. **Documentation First** - Comprehensive docs before code

## Benefits

✅ **Simple imports** - One line: `import sdk "..."`  
✅ **Type safety** - All types in one place  
✅ **Better IDE support** - All methods auto-complete  
✅ **Easy discovery** - Examples organized by use case  
✅ **Professional** - Industry-standard structure
