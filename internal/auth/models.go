package auth

import (
	"encoding/json"
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents public user information
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// ToJSON converts User to JSON bytes
func (u *User) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// FromJSON parses User from JSON bytes
func FromJSON(data []byte) (*User, error) {
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
