package config

import (
	"errors"
	"log"
	"os"
	
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Cloudflare API credentials
	CloudflareAPIKey   string
	CloudflareEmail    string
	CloudflareAPIToken string
	
	// Cloudflare Account and KV Namespace
	AccountID     string
	NamespaceID   string
	
	// JWT configuration
	JWTSecret     string
	JWTExpiration int // in hours
	
	// Server configuration
	ServerPort    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// 尝试加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件或加载失败，将使用系统环境变量")
	}
	
	cfg := &Config{
		CloudflareAPIKey:   os.Getenv("CLOUDFLARE_API_KEY"),
		CloudflareEmail:    os.Getenv("CLOUDFLARE_EMAIL"),
		CloudflareAPIToken: os.Getenv("CLOUDFLARE_API_TOKEN"),
		AccountID:          os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		NamespaceID:        os.Getenv("CLOUDFLARE_NAMESPACE_ID"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		ServerPort:         getEnvOrDefault("SERVER_PORT", "8080"),
	}
	
	// JWT expiration defaults to 24 hours
	cfg.JWTExpiration = 24
	
	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	
	return cfg, nil
}

// Validate checks if all required configuration fields are set
func (c *Config) Validate() error {
	// At least one authentication method is required
	if c.CloudflareAPIToken == "" && (c.CloudflareAPIKey == "" || c.CloudflareEmail == "") {
		return errors.New("cloudflare authentication required: either API_TOKEN or (API_KEY + EMAIL)")
	}
	
	if c.AccountID == "" {
		return errors.New("CLOUDFLARE_ACCOUNT_ID is required")
	}
	
	if c.NamespaceID == "" {
		return errors.New("CLOUDFLARE_NAMESPACE_ID is required")
	}
	
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET is required")
	}
	
	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
