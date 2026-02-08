// Package main demonstrates advanced KV operations with the Cloudflare Auth SDK.
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

	// Store a simple key-value pair
	fmt.Println("=== Storing Data ===")
	err = client.KVSet(ctx, "app:config", []byte("my-config-value"), nil)
	if err != nil {
		log.Fatalf("Failed to set value: %v", err)
	}
	fmt.Println("✓ Data stored successfully")

	// Store with expiration
	fmt.Println("\n=== Storing Data with Expiration ===")
	err = client.KVSet(ctx, "session:12345", []byte("session-data"), &sdk.KVWriteOptions{
		ExpirationTTL: 3600, // 1 hour
		Metadata:      "temporary session data",
	})
	if err != nil {
		log.Fatalf("Failed to set value with expiration: %v", err)
	}
	fmt.Println("✓ Session data stored with 1-hour expiration")

	// Retrieve data
	fmt.Println("\n=== Retrieving Data ===")
	value, err := client.KVGet(ctx, "app:config")
	if err != nil {
		log.Fatalf("Failed to get value: %v", err)
	}
	fmt.Printf("✓ Retrieved value: %s\n", string(value))

	// List keys with prefix
	fmt.Println("\n=== Listing Keys ===")
	keys, err := client.KVList(ctx, "app:", 10)
	if err != nil {
		log.Fatalf("Failed to list keys: %v", err)
	}
	fmt.Printf("✓ Found %d keys:\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key.Name)
	}

	// Delete a key
	fmt.Println("\n=== Deleting Key ===")
	err = client.KVDelete(ctx, "session:12345")
	if err != nil {
		log.Fatalf("Failed to delete key: %v", err)
	}
	fmt.Println("✓ Key deleted successfully")

	// Bulk delete
	fmt.Println("\n=== Bulk Delete ===")
	keysToDelete := []string{"temp:key1", "temp:key2", "temp:key3"}
	err = client.KVDeleteBulk(ctx, keysToDelete)
	if err != nil {
		log.Fatalf("Failed to bulk delete: %v", err)
	}
	fmt.Println("✓ Bulk delete completed")

	fmt.Println("\n✅ All KV operations completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
