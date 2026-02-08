package auth

import (
	"context"
	"fmt"
	"time"
	
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	
	apperrors "github.com/zolagz/cloudflare-auth-sdk/internal/errors"
	"github.com/zolagz/cloudflare-auth-sdk/internal/kv"
)

// Service handles authentication operations
type Service struct {
	kvClient      *kv.Client
	jwtSecret     []byte
	jwtExpiration time.Duration
}

// NewService creates a new auth service
func NewService(kvClient *kv.Client, jwtSecret string, expirationHours int) *Service {
	return &Service{
		kvClient:      kvClient,
		jwtSecret:     []byte(jwtSecret),
		jwtExpiration: time.Duration(expirationHours) * time.Hour,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register creates a new user
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	const op = "auth.Register"
	
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.NewAppError(op, apperrors.ErrInvalidInput,
			"email and password are required", 400)
	}
	
	// Check if user already exists
	userKey := s.getUserKey(req.Email)
	existingData, _ := s.kvClient.Get(ctx, userKey)
	if existingData != nil {
		return nil, apperrors.NewAppError(op, apperrors.ErrUserAlreadyExists,
			"user already exists", 409)
	}
	
	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to hash password", 500)
	}
	
	// Create user
	now := time.Now()
	user := &User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	// Save to KV
	userData, err := user.ToJSON()
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to serialize user", 500)
	}
	
	if err := s.kvClient.Set(ctx, userKey, userData, nil); err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to save user", 500)
	}
	
	// Also save user ID mapping for easier lookups
	idKey := s.getUserIDKey(user.ID)
	if err := s.kvClient.Set(ctx, idKey, []byte(req.Email), nil); err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to save user ID mapping", 500)
	}
	
	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	const op = "auth.Login"
	
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, apperrors.NewAppError(op, apperrors.ErrInvalidInput,
			"email and password are required", 400)
	}
	
	// Get user from KV
	userKey := s.getUserKey(req.Email)
	userData, err := s.kvClient.Get(ctx, userKey)
	if err != nil {
		return nil, apperrors.NewAppError(op, apperrors.ErrUserNotFound,
			"user not found", 404)
	}
	
	user, err := FromJSON(userData)
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to parse user data", 500)
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.NewAppError(op, apperrors.ErrInvalidCredentials,
			"invalid credentials", 401)
	}
	
	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiration)
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
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to generate token", 500)
	}
	
	return &LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User:      user.ToUserInfo(),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	const op = "auth.ValidateToken"
	
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"invalid token", 401)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperrors.NewAppError(op, apperrors.ErrInvalidToken,
			"invalid token claims", 401)
	}
	
	return claims, nil
}

// GetUser retrieves a user by email
func (s *Service) GetUser(ctx context.Context, email string) (*User, error) {
	const op = "auth.GetUser"
	
	userKey := s.getUserKey(email)
	userData, err := s.kvClient.Get(ctx, userKey)
	if err != nil {
		return nil, apperrors.NewAppError(op, apperrors.ErrUserNotFound,
			"user not found", 404)
	}
	
	user, err := FromJSON(userData)
	if err != nil {
		return nil, apperrors.NewAppError(op, err,
			"failed to parse user data", 500)
	}
	
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID string) (*User, error) {
	const op = "auth.GetUserByID"
	
	// First get email from ID mapping
	idKey := s.getUserIDKey(userID)
	emailData, err := s.kvClient.Get(ctx, idKey)
	if err != nil {
		return nil, apperrors.NewAppError(op, apperrors.ErrUserNotFound,
			"user not found", 404)
	}
	
	return s.GetUser(ctx, string(emailData))
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, email string) error {
	const op = "auth.DeleteUser"
	
	// Get user first to get ID
	user, err := s.GetUser(ctx, email)
	if err != nil {
		return err
	}
	
	// Delete user data
	userKey := s.getUserKey(email)
	if err := s.kvClient.Delete(ctx, userKey); err != nil {
		return apperrors.NewAppError(op, err,
			"failed to delete user", 500)
	}
	
	// Delete ID mapping
	idKey := s.getUserIDKey(user.ID)
	if err := s.kvClient.Delete(ctx, idKey); err != nil {
		return apperrors.NewAppError(op, err,
			"failed to delete user ID mapping", 500)
	}
	
	return nil
}

// Helper functions
func (s *Service) getUserKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func (s *Service) getUserIDKey(userID string) string {
	return fmt.Sprintf("user:id:%s", userID)
}
