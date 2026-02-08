# API Reference

Complete API reference for the Cloudflare Auth SDK.

## Package Import

```go
import sdk "github.com/zolagz/cloudflare-auth-sdk"
```

## Types

### ClientOptions

Configuration options for creating a new SDK client.

```go
type ClientOptions struct {
    APIToken           string // Cloudflare API Token (required)
    AccountID          string // Cloudflare Account ID (required)
    NamespaceID        string // Workers KV Namespace ID (required)
    JWTSecret          string // JWT signing secret (required)
    JWTExpirationHours int    // JWT expiration time in hours (optional, default: 24)
}
```

**Methods:**

- `Validate() error` - Validates the configuration options

**Builder Methods:**

- `WithAPIToken(token string) *ClientOptions`
- `WithAccountID(id string) *ClientOptions`
- `WithNamespaceID(id string) *ClientOptions`
- `WithJWTSecret(secret string) *ClientOptions`
- `WithJWTExpiration(hours int) *ClientOptions`

**Example:**

```go
opts := &sdk.ClientOptions{}.
    WithAPIToken("your-token").
    WithAccountID("account-id").
    WithNamespaceID("namespace-id").
    WithJWTSecret("secret").
    WithJWTExpiration(48)
```

### Client

Main SDK client for authentication and KV operations.

```go
type Client struct {
    // Contains unexported fields
}
```

### User

Represents a user in the system.

```go
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

### UserInfo

Public user information (without sensitive data).

```go
type UserInfo struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### LoginResponse

Response from login operation.

```go
type LoginResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    User      UserInfo  `json:"user"`
}
```

### Claims

JWT token claims.

```go
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}
```

### KVKey

Represents a key in Workers KV.

```go
type KVKey struct {
    Name       string                 `json:"name"`
    Expiration int64                  `json:"expiration,omitempty"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

### KVWriteOptions

Options for writing to Workers KV.

```go
type KVWriteOptions struct {
    Expiration    int64  // Unix timestamp when key expires
    ExpirationTTL int    // TTL in seconds
    Metadata      string // Arbitrary metadata
}
```

### AppError

Application error with code and context.

```go
type AppError struct {
    Code    string // Error code (e.g., "user_not_found")
    Message string // Human-readable error message
    Err     error  // Underlying error
}
```

**Methods:**

- `Error() string` - Returns error message
- `Unwrap() error` - Returns underlying error

## Client Functions

### NewClient

Creates a new SDK client.

```go
func NewClient(opts *ClientOptions) (*Client, error)
```

**Parameters:**

- `opts` - Client configuration options

**Returns:**

- `*Client` - New client instance
- `error` - Configuration validation error

**Example:**

```go
client, err := sdk.NewClient(&sdk.ClientOptions{
    APIToken:           os.Getenv("CLOUDFLARE_API_TOKEN"),
    AccountID:          os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
    NamespaceID:        os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
    JWTSecret:          os.Getenv("JWT_SECRET"),
    JWTExpirationHours: 24,
})
```

## Client Methods

### Authentication Methods

#### Register

Registers a new user.

```go
func (c *Client) Register(ctx context.Context, email, password string) (*User, error)
```

**Parameters:**

- `ctx` - Context for cancellation and timeouts
- `email` - User email address
- `password` - User password (will be hashed with bcrypt)

**Returns:**

- `*User` - Created user
- `error` - `ErrUserAlreadyExists` if user exists, other errors

**Example:**

```go
user, err := client.Register(ctx, "user@example.com", "SecurePassword123!")
if err != nil {
    if sdk.IsUserAlreadyExists(err) {
        // Handle existing user
    }
    return err
}
```

#### Login

Authenticates a user and returns a JWT token.

```go
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error)
```

**Parameters:**

- `ctx` - Context
- `email` - User email
- `password` - User password

**Returns:**

- `*LoginResponse` - Token and user information
- `error` - `ErrInvalidCredentials` if credentials are wrong

**Example:**

```go
resp, err := client.Login(ctx, "user@example.com", "password123")
if err != nil {
    if sdk.IsInvalidCredentials(err) {
        // Handle bad credentials
    }
    return err
}

token := resp.Token
```

#### ValidateToken

Validates a JWT token and returns user information.

```go
func (c *Client) ValidateToken(ctx context.Context, tokenString string) (*UserInfo, error)
```

**Parameters:**

- `ctx` - Context
- `tokenString` - JWT token to validate

**Returns:**

- `*UserInfo` - User information from token
- `error` - `ErrUnauthorized` if token is invalid

**Example:**

```go
user, err := client.ValidateToken(ctx, token)
if err != nil {
    if sdk.IsUnauthorized(err) {
        // Handle invalid token
    }
    return err
}
```

#### GetUserByID

Retrieves user information by user ID.

```go
func (c *Client) GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
```

**Parameters:**

- `ctx` - Context
- `userID` - User ID

**Returns:**

- `*UserInfo` - User information
- `error` - `ErrUserNotFound` if user doesn't exist

**Example:**

```go
user, err := client.GetUserByID(ctx, "user-id")
if err != nil {
    if sdk.IsUserNotFound(err) {
        // Handle not found
    }
    return err
}
```

#### DeleteUser

Deletes a user from the system.

```go
func (c *Client) DeleteUser(ctx context.Context, userID string) error
```

**Parameters:**

- `ctx` - Context
- `userID` - User ID to delete

**Returns:**

- `error` - Error if deletion fails

**Example:**

```go
err := client.DeleteUser(ctx, "user-id")
```

### Key-Value Methods

#### KVGet

Retrieves a value from Workers KV.

```go
func (c *Client) KVGet(ctx context.Context, key string) ([]byte, error)
```

**Parameters:**

- `ctx` - Context
- `key` - Key to retrieve

**Returns:**

- `[]byte` - Value data
- `error` - Error if key doesn't exist or retrieval fails

**Example:**

```go
value, err := client.KVGet(ctx, "app:config")
if err != nil {
    return err
}
fmt.Printf("Value: %s\n", string(value))
```

#### KVSet

Stores a value in Workers KV.

```go
func (c *Client) KVSet(ctx context.Context, key string, value []byte, opts *KVWriteOptions) error
```

**Parameters:**

- `ctx` - Context
- `key` - Key to store
- `value` - Value data
- `opts` - Write options (can be nil)

**Returns:**

- `error` - Error if storage fails

**Example:**

```go
// Simple set
err := client.KVSet(ctx, "app:config", []byte("value"), nil)

// Set with expiration
err := client.KVSet(ctx, "session:123", data, &sdk.KVWriteOptions{
    ExpirationTTL: 3600, // 1 hour
    Metadata:      "session data",
})
```

#### KVDelete

Deletes a key from Workers KV.

```go
func (c *Client) KVDelete(ctx context.Context, key string) error
```

**Parameters:**

- `ctx` - Context
- `key` - Key to delete

**Returns:**

- `error` - Error if deletion fails

**Example:**

```go
err := client.KVDelete(ctx, "app:config")
```

#### KVList

Lists keys in Workers KV with optional prefix.

```go
func (c *Client) KVList(ctx context.Context, prefix string, limit int) ([]KVKey, error)
```

**Parameters:**

- `ctx` - Context
- `prefix` - Key prefix filter (empty string for all keys)
- `limit` - Maximum number of keys to return

**Returns:**

- `[]KVKey` - List of keys
- `error` - Error if listing fails

**Example:**

```go
keys, err := client.KVList(ctx, "users:", 100)
if err != nil {
    return err
}

for _, key := range keys {
    fmt.Printf("Key: %s\n", key.Name)
}
```

#### KVDeleteBulk

Deletes multiple keys from Workers KV.

```go
func (c *Client) KVDeleteBulk(ctx context.Context, keys []string) error
```

**Parameters:**

- `ctx` - Context
- `keys` - List of keys to delete

**Returns:**

- `error` - Error if deletion fails

**Example:**

```go
keysToDelete := []string{"temp:key1", "temp:key2", "temp:key3"}
err := client.KVDeleteBulk(ctx, keysToDelete)
```

## Error Handling

### Error Constants

```go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrInvalidToken      = errors.New("invalid token")
    ErrInternalError     = errors.New("internal error")
)
```

### Error Helper Functions

#### IsUserNotFound

```go
func IsUserNotFound(err error) bool
```

#### IsUserAlreadyExists

```go
func IsUserAlreadyExists(err error) bool
```

#### IsInvalidCredentials

```go
func IsInvalidCredentials(err error) bool
```

#### IsUnauthorized

```go
func IsUnauthorized(err error) bool
```

#### IsInternalError

```go
func IsInternalError(err error) bool
```

**Example:**

```go
user, err := client.GetUserByID(ctx, userID)
if err != nil {
    if sdk.IsUserNotFound(err) {
        return nil, fmt.Errorf("user not found")
    }
    if sdk.IsInternalError(err) {
        log.Printf("Internal error: %v", err)
    }
    return nil, err
}
```

## Version

```go
const Version = "1.0.0"
```

## See Also

- [Getting Started Guide](./getting-started.md)
- [Advanced Usage](./advanced-usage.md)
- [Examples](../examples/)
