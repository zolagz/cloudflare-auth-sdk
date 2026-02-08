package cloudflare_auth_sdk

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserInfo represents public user information (without sensitive data).
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// LoginResponse represents the response from a successful login.
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// KVKey represents a key in the KV namespace with metadata.
type KVKey struct {
	Name       string      `json:"name"`
	Expiration float64     `json:"expiration,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// KVWriteOptions contains options for writing KV pairs.
type KVWriteOptions struct {
	ExpirationTTL int    // Time to live in seconds
	Metadata      string // Optional metadata
}

// toJSON converts User to JSON bytes
func (u *User) toJSON() ([]byte, error) {
	return json.Marshal(u)
}

// userFromJSON parses User from JSON bytes
func userFromJSON(data []byte) (*User, error) {
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ToUserInfo converts User to UserInfo
func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:    u.ID,
		Email: u.Email,
	}
}
