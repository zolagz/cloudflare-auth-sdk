// Package cloudflare_auth_sdk provides a Cloudflare Workers KV-based authentication SDK.
//
// This SDK offers a complete solution for user authentication with JWT tokens,
// backed by Cloudflare Workers KV storage.
//
// Basic usage:
//
//	client, err := cloudflare_auth_sdk.NewClient(&cloudflare_auth_sdk.ClientOptions{
//	    APIToken:    "your-cloudflare-api-token",
//	    AccountID:   "your-account-id",
//	    NamespaceID: "your-kv-namespace-id",
//	    JWTSecret:   "your-jwt-secret",
//	})
//
//	user, err := client.Register(ctx, "user@example.com", "password")
//	loginResp, err := client.Login(ctx, "user@example.com", "password")
package cloudflare_auth_sdk

import (
	"context"
	"fmt"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Client is the main SDK client that provides all authentication and KV operations.
type Client struct {
	cfClient    *cloudflare.Client
	accountID   string
	namespaceID string
	jwtSecret   []byte
	jwtExpiry   time.Duration
}

// NewClient creates a new SDK client with the provided options.
//
// Example:
//
//	client, err := cloudflare_auth_sdk.NewClient(&cloudflare_auth_sdk.ClientOptions{
//	    APIToken:           "your-api-token",
//	    AccountID:          "your-account-id",
//	    NamespaceID:        "your-namespace-id",
//	    JWTSecret:          "your-jwt-secret",
//	    JWTExpirationHours: 24,
//	})
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, ErrInvalidConfig
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Create Cloudflare client
	var cfClient *cloudflare.Client
	if opts.APIToken != "" {
		cfClient = cloudflare.NewClient(
			option.WithAPIToken(opts.APIToken),
		)
	} else {
		cfClient = cloudflare.NewClient(
			option.WithAPIKey(opts.APIKey),
			option.WithAPIEmail(opts.Email),
		)
	}

	// Set default JWT expiration
	jwtExpiry := time.Duration(opts.JWTExpirationHours) * time.Hour
	if jwtExpiry == 0 {
		jwtExpiry = 24 * time.Hour
	}

	return &Client{
		cfClient:    cfClient,
		accountID:   opts.AccountID,
		namespaceID: opts.NamespaceID,
		jwtSecret:   []byte(opts.JWTSecret),
		jwtExpiry:   jwtExpiry,
	}, nil
}

// Register creates a new user account.
//
// The password will be securely hashed using bcrypt before storage.
// Returns the created user information or an error if registration fails.
func (c *Client) Register(ctx context.Context, email, password string) (*User, error) {
	const op = "Client.Register"

	if email == "" || password == "" {
		return nil, NewAppError(op, ErrInvalidInput, "email and password are required", 400)
	}

	// Check if user already exists
	userKey := getUserKey(email)
	existingData, _ := c.kvGet(ctx, userKey)
	if existingData != nil {
		return nil, NewAppError(op, ErrUserAlreadyExists, "user already exists", 409)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, NewAppError(op, err, "failed to hash password", 500)
	}

	// Create user
	now := time.Now()
	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save user
	if err := c.saveUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token.
//
// Returns login response with token and user info, or an error if authentication fails.
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	const op = "Client.Login"

	if email == "" || password == "" {
		return nil, NewAppError(op, ErrInvalidInput, "email and password are required", 400)
	}

	// Get user
	user, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return nil, NewAppError(op, ErrUserNotFound, "user not found", 404)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, NewAppError(op, ErrInvalidCredentials, "invalid credentials", 401)
	}

	// Generate JWT token
	expiresAt := time.Now().Add(c.jwtExpiry)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(c.jwtSecret)
	if err != nil {
		return nil, NewAppError(op, err, "failed to generate token", 500)
	}

	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:    user.ID,
			Email: user.Email,
		},
	}, nil
}

// ValidateToken validates a JWT token and returns the user information.
//
// Returns user info if the token is valid, or an error if validation fails.
func (c *Client) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	const op = "Client.ValidateToken"

	claims, err := c.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	return c.GetUserByID(ctx, claims.UserID)
}

// GetUserByID retrieves user information by user ID.
func (c *Client) GetUserByID(ctx context.Context, userID string) (*User, error) {
	const op = "Client.GetUserByID"

	// Get email from ID mapping
	idKey := getUserIDKey(userID)
	emailData, err := c.kvGet(ctx, idKey)
	if err != nil {
		return nil, NewAppError(op, ErrUserNotFound, "user not found", 404)
	}

	return c.getUserByEmail(ctx, string(emailData))
}

// GetUserByEmail retrieves user information by email address.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return c.getUserByEmail(ctx, email)
}

// DeleteUser deletes a user account.
func (c *Client) DeleteUser(ctx context.Context, email string) error {
	const op = "Client.DeleteUser"

	user, err := c.getUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Delete user data
	userKey := getUserKey(email)
	if err := c.kvDelete(ctx, userKey); err != nil {
		return NewAppError(op, err, "failed to delete user", 500)
	}

	// Delete ID mapping
	idKey := getUserIDKey(user.ID)
	if err := c.kvDelete(ctx, idKey); err != nil {
		return NewAppError(op, err, "failed to delete user ID mapping", 500)
	}

	return nil
}

// parseToken parses and validates a JWT token
func (c *Client) parseToken(tokenString string) (*Claims, error) {
	const op = "Client.parseToken"

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.jwtSecret, nil
	})

	if err != nil {
		return nil, NewAppError(op, err, "invalid token", 401)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, NewAppError(op, ErrInvalidToken, "invalid token claims", 401)
	}

	return claims, nil
}

// getUserByEmail retrieves a user by email
func (c *Client) getUserByEmail(ctx context.Context, email string) (*User, error) {
	const op = "Client.getUserByEmail"

	userKey := getUserKey(email)
	userData, err := c.kvGet(ctx, userKey)
	if err != nil {
		return nil, NewAppError(op, ErrUserNotFound, "user not found", 404)
	}

	return userFromJSON(userData)
}

// saveUser saves a user to KV storage
func (c *Client) saveUser(ctx context.Context, user *User) error {
	const op = "Client.saveUser"

	userData, err := user.toJSON()
	if err != nil {
		return NewAppError(op, err, "failed to serialize user", 500)
	}

	userKey := getUserKey(user.Email)
	if err := c.kvSet(ctx, userKey, userData); err != nil {
		return NewAppError(op, err, "failed to save user", 500)
	}

	// Save ID mapping
	idKey := getUserIDKey(user.ID)
	if err := c.kvSet(ctx, idKey, []byte(user.Email)); err != nil {
		return NewAppError(op, err, "failed to save user ID mapping", 500)
	}

	return nil
}

// Helper functions for key generation
func getUserKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func getUserIDKey(userID string) string {
	return fmt.Sprintf("user:id:%s", userID)
}

// KV operation wrappers
func (c *Client) kvGet(ctx context.Context, key string) ([]byte, error) {
	resp, err := c.cfClient.KV.Namespaces.Values.Get(ctx, c.namespaceID, key,
		kv.NamespaceValueGetParams{
			AccountID: cloudflare.F(c.accountID),
		})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return readAll(resp.Body)
}

func (c *Client) kvSet(ctx context.Context, key string, value []byte) error {
	_, err := c.cfClient.KV.Namespaces.Values.Update(ctx, c.namespaceID, key,
		kv.NamespaceValueUpdateParams{
			AccountID: cloudflare.F(c.accountID),
			Value:     cloudflare.F(string(value)),
		})
	return err
}

func (c *Client) kvDelete(ctx context.Context, key string) error {
	_, err := c.cfClient.KV.Namespaces.Values.Delete(ctx, c.namespaceID, key,
		kv.NamespaceValueDeleteParams{
			AccountID: cloudflare.F(c.accountID),
		})
	return err
}
