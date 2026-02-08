package cloudflare_auth_sdk

import "errors"

// ClientOptions contains the configuration for creating a new SDK client.
type ClientOptions struct {
	// Cloudflare API authentication (use either APIToken or APIKey+Email)
	APIToken string // Recommended: Cloudflare API Token
	APIKey   string // Legacy: Cloudflare API Key
	Email    string // Legacy: Email (required if using APIKey)

	// Cloudflare Account and KV Namespace
	AccountID   string // Cloudflare Account ID
	NamespaceID string // Workers KV Namespace ID

	// JWT configuration
	JWTSecret          string // Secret key for signing JWT tokens
	JWTExpirationHours int    // Token expiration in hours (default: 24)
}

// Validate checks if all required options are set and valid.
func (o *ClientOptions) Validate() error {
	if o.AccountID == "" {
		return errors.New("AccountID is required")
	}

	if o.NamespaceID == "" {
		return errors.New("NamespaceID is required")
	}

	if o.JWTSecret == "" {
		return errors.New("JWTSecret is required")
	}

	// Check if either API Token or API Key+Email is provided
	if o.APIToken == "" && (o.APIKey == "" || o.Email == "") {
		return errors.New("either APIToken or both APIKey and Email are required")
	}

	return nil
}

// WithAPIToken sets the Cloudflare API Token.
func (o *ClientOptions) WithAPIToken(token string) *ClientOptions {
	o.APIToken = token
	return o
}

// WithAPIKey sets the Cloudflare API Key and Email.
func (o *ClientOptions) WithAPIKey(key, email string) *ClientOptions {
	o.APIKey = key
	o.Email = email
	return o
}

// WithAccountID sets the Cloudflare Account ID.
func (o *ClientOptions) WithAccountID(id string) *ClientOptions {
	o.AccountID = id
	return o
}

// WithNamespaceID sets the Workers KV Namespace ID.
func (o *ClientOptions) WithNamespaceID(id string) *ClientOptions {
	o.NamespaceID = id
	return o
}

// WithJWTSecret sets the JWT signing secret.
func (o *ClientOptions) WithJWTSecret(secret string) *ClientOptions {
	o.JWTSecret = secret
	return o
}

// WithJWTExpirationHours sets the JWT token expiration in hours.
func (o *ClientOptions) WithJWTExpirationHours(hours int) *ClientOptions {
	o.JWTExpirationHours = hours
	return o
}
