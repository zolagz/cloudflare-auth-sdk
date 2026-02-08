// Package main demonstrates the quickstart example for the Cloudflare Auth SDK.
//
// This example shows the basic usage of the SDK for user registration,
// login, and token validation.
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
		APIToken:           getEnv("CLOUDFLARE_API_TOKEN", ""),
		AccountID:          getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
		NamespaceID:        getEnv("CLOUDFLARE_NAMESPACE_ID", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpirationHours: 24,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Register a new user
	fmt.Println("=== Registering User ===")
	user, err := client.Register(ctx, "user@example.com", "SecurePassword123!")
	if err != nil {
		if sdk.IsUserAlreadyExists(err) {
			fmt.Println("User already exists, skipping registration")
		} else {
			log.Fatalf("Registration failed: %v", err)
		}
	} else {
		fmt.Printf("✓ User registered: %s (ID: %s)\n", user.Email, user.ID)
	}

	// Login
	fmt.Println("\n=== Logging In ===")
	loginResp, err := client.Login(ctx, "user@example.com", "SecurePassword123!")
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("✓ Login successful!\n")
	fmt.Printf("  Token: %s...\n", loginResp.Token[:20])
	fmt.Printf("  Expires: %s\n", loginResp.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  User: %s (%s)\n", loginResp.User.Email, loginResp.User.ID)

	// Validate token
	fmt.Println("\n=== Validating Token ===")
	validatedUser, err := client.ValidateToken(ctx, loginResp.Token)
	if err != nil {
		log.Fatalf("Token validation failed: %v", err)
	}
	fmt.Printf("✓ Token is valid for user: %s (ID: %s)\n", validatedUser.Email, validatedUser.ID)

	// Get user by ID
	fmt.Println("\n=== Getting User by ID ===")
	fetchedUser, err := client.GetUserByID(ctx, validatedUser.ID)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("✓ User found: %s\n", fetchedUser.Email)

	fmt.Println("\n✅ All operations completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
