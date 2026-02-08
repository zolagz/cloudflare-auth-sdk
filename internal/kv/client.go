package kv

import (
	"context"
	"fmt"
	"io"
	
	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
	apperrors "github.com/zolagz/cloudflare-auth-sdk/internal/errors"
)

// Client wraps Cloudflare API client for KV operations
type Client struct {
	client      *cloudflare.Client
	accountID   string
	namespaceID string
}

// NewClient creates a new KV client
func NewClient(client *cloudflare.Client, accountID, namespaceID string) *Client {
	return &Client{
		client:      client,
		accountID:   accountID,
		namespaceID: namespaceID,
	}
}

// WriteOptions contains options for writing KV pairs
type WriteOptions struct {
	ExpirationTTL int    // Time to live in seconds
	Metadata      string // Optional metadata
}

// Get retrieves a value from KV store
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	const op = "kv.Get"
	
	resp, err := c.client.KV.Namespaces.Values.Get(ctx, c.namespaceID, key, 
		kv.NamespaceValueGetParams{
			AccountID: cloudflare.F(c.accountID),
		})
	if err != nil {
		return nil, apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to get key: %s", key), 500)
	}
	defer resp.Body.Close()
	
	value, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to read response for key: %s", key), 500)
	}
	
	return value, nil
}

// Set stores a key-value pair
func (c *Client) Set(ctx context.Context, key string, value []byte, opts *WriteOptions) error {
	const op = "kv.Set"
	
	params := kv.NamespaceValueUpdateParams{
		AccountID: cloudflare.F(c.accountID),
		Value:     cloudflare.F(string(value)),
	}
	
	if opts != nil {
		if opts.ExpirationTTL > 0 {
			params.ExpirationTTL = cloudflare.F(float64(opts.ExpirationTTL))
		}
		if opts.Metadata != "" {
			params.Metadata = cloudflare.F[any](opts.Metadata)
		}
	}
	
	_, err := c.client.KV.Namespaces.Values.Update(ctx, c.namespaceID, key, params)
	if err != nil {
		return apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to set key: %s", key), 500)
	}
	
	return nil
}

// Delete removes a key from KV store
func (c *Client) Delete(ctx context.Context, key string) error {
	const op = "kv.Delete"
	
	_, err := c.client.KV.Namespaces.Values.Delete(ctx, c.namespaceID, key,
		kv.NamespaceValueDeleteParams{
			AccountID: cloudflare.F(c.accountID),
		})
	if err != nil {
		return apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to delete key: %s", key), 500)
	}
	
	return nil
}

// Key represents a KV key with metadata
type Key struct {
	Name       string      `json:"name"`
	Expiration float64     `json:"expiration,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// List lists keys in the KV namespace
func (c *Client) List(ctx context.Context, prefix string, limit int) ([]Key, error) {
	const op = "kv.List"
	
	params := kv.NamespaceKeyListParams{
		AccountID: cloudflare.F(c.accountID),
	}
	
	if prefix != "" {
		params.Prefix = cloudflare.F(prefix)
	}
	
	if limit > 0 {
		params.Limit = cloudflare.F(float64(limit))
	}
	
	resp, err := c.client.KV.Namespaces.Keys.List(ctx, c.namespaceID, params)
	if err != nil {
		return nil, apperrors.NewAppError(op, err, 
			"failed to list keys", 500)
	}
	
	// Convert response to our Key type
	var keys []Key
	for _, item := range resp.Result {
		keys = append(keys, Key{
			Name:       item.Name,
			Expiration: item.Expiration,
			Metadata:   item.Metadata,
		})
	}
	
	return keys, nil
}

// DeleteBulk deletes multiple keys
func (c *Client) DeleteBulk(ctx context.Context, keys []string) error {
	const op = "kv.DeleteBulk"
	
	_, err := c.client.KV.Namespaces.Keys.BulkDelete(ctx, c.namespaceID,
		kv.NamespaceKeyBulkDeleteParams{
			AccountID: cloudflare.F(c.accountID),
			Body:      keys,
		})
	if err != nil {
		return apperrors.NewAppError(op, err, 
			"failed to delete keys in bulk", 500)
	}
	
	return nil
}
