package main

import (
	"context"
	"log"
	
	"github.com/cloudflare/cloudflare-go"
	
	"github.com/zolagz/cloudflare-auth-sdk/internal/auth"
	"github.com/zolagz/cloudflare-auth-sdk/internal/config"
	"github.com/zolagz/cloudflare-auth-sdk/internal/kv"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Initialize Cloudflare API client
	var api *cloudflare.API
	if cfg.CloudflareAPIToken != "" {
		api, err = cloudflare.NewWithAPIToken(cfg.CloudflareAPIToken)
	} else {
		api, err = cloudflare.New(cfg.CloudflareAPIKey, cfg.CloudflareEmail)
	}
	if err != nil {
		log.Fatalf("Failed to create Cloudflare API client: %v", err)
	}
	
	// Initialize KV client
	kvClient := kv.NewClient(api, cfg.AccountID, cfg.NamespaceID)
	
	// Initialize auth service
	authService := auth.NewService(kvClient, cfg.JWTSecret, cfg.JWTExpiration)
	
	ctx := context.Background()
	
	// Example: Register a new user
	log.Println("Registering a new user...")
	registerReq := &auth.RegisterRequest{
		Email:    "user@example.com",
		Password: "SecurePassword123!",
	}
	
	user, err := authService.Register(ctx, registerReq)
	if err != nil {
		log.Printf("Registration failed: %v", err)
	} else {
		log.Printf("User registered successfully: %s (ID: %s)", user.Email, user.ID)
	}
	
	// Example: Login
	log.Println("Logging in...")
	loginReq := &auth.LoginRequest{
		Email:    "user@example.com",
		Password: "SecurePassword123!",
	}
	
	loginResp, err := authService.Login(ctx, loginReq)
	if err != nil {
		log.Printf("Login failed: %v", err)
	} else {
		log.Printf("Login successful!")
		log.Printf("Token: %s", loginResp.Token)
		log.Printf("Expires at: %s", loginResp.ExpiresAt)
	}
	
	// Example: Validate token
	if loginResp != nil {
		log.Println("Validating token...")
		claims, err := authService.ValidateToken(loginResp.Token)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
		} else {
			log.Printf("Token valid for user: %s (ID: %s)", claims.Email, claims.UserID)
		}
	}
	
	// Example: KV operations
	log.Println("\nKV Operations:")
	
	// Set a value
	log.Println("Setting key-value pair...")
	err = kvClient.Set(ctx, "test:key1", []byte("Hello, Cloudflare KV!"), nil)
	if err != nil {
		log.Printf("Failed to set key: %v", err)
	} else {
		log.Println("Key set successfully")
	}
	
	// Get a value
	log.Println("Getting value...")
	value, err := kvClient.Get(ctx, "test:key1")
	if err != nil {
		log.Printf("Failed to get key: %v", err)
	} else {
		log.Printf("Value: %s", string(value))
	}
	
	// Set with expiration
	log.Println("Setting key with 3600s TTL...")
	err = kvClient.Set(ctx, "test:key2", []byte("This expires in 1 hour"), &kv.WriteOptions{
		ExpirationTTL: 3600,
	})
	if err != nil {
		log.Printf("Failed to set key with TTL: %v", err)
	} else {
		log.Println("Key with TTL set successfully")
	}
	
	// List keys
	log.Println("Listing keys with prefix 'test:'...")
	keys, err := kvClient.List(ctx, "test:", 10)
	if err != nil {
		log.Printf("Failed to list keys: %v", err)
	} else {
		log.Printf("Found %d keys:", len(keys))
		for _, key := range keys {
			log.Printf("  - %s", key.Name)
		}
	}
	
	log.Println("\nExample completed!")
}
