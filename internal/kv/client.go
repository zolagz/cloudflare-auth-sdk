package kv

import (
	"context"
	"fmt"
	
	"github.com/cloudflare/cloudflare-go"
	apperrors "github.com/zolagz/cloudflare-auth-sdk/internal/errors"
)

// Client wraps Cloudflare API client for KV operations
type Client struct {
	api         *cloudflare.API
	accountID   string
	namespaceID string
}

// NewClient creates a new KV client
func NewClient(api *cloudflare.API, accountID, namespaceID string) *Client {
	return &Client{
		api:         api,
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
	
	value, err := c.api.GetWorkersKV(ctx, cloudflare.AccountIdentifier(c.accountID), 
		cloudflare.GetWorkersKVParams{
			NamespaceID: c.namespaceID,
			Key:         key,
		})
	if err != nil {
		return nil, apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to get key: %s", key), 500)
	}
	
	return value, nil
}

// Set stores a key-value pair
func (c *Client) Set(ctx context.Context, key string, value []byte, opts *WriteOptions) error {
	const op = "kv.Set"
	
	params := cloudflare.SetWorkersKVParams{
		NamespaceID: c.namespaceID,
		Key:         key,
		Value:       value,
	}
	
	if opts != nil {
		if opts.ExpirationTTL > 0 {
			params.ExpirationTTL = opts.ExpirationTTL
		}
		if opts.Metadata != "" {
			params.Metadata = opts.Metadata
		}
	}
	
	_, err := c.api.SetWorkersKV(ctx, cloudflare.AccountIdentifier(c.accountID), params)
	if err != nil {
		return apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to set key: %s", key), 500)
	}
	
	return nil
}

// Delete removes a key from KV store
func (c *Client) Delete(ctx context.Context, key string) error {
	const op = "kv.Delete"
	
	_, err := c.api.DeleteWorkersKV(ctx, cloudflare.AccountIdentifier(c.accountID),
		cloudflare.DeleteWorkersKVParams{
			NamespaceID: c.namespaceID,
			Key:         key,
		})
	if err != nil {
		return apperrors.NewAppError(op, err, 
			fmt.Sprintf("failed to delete key: %s", key), 500)
	}
	
	return nil
}

// List lists keys in the KV namespace
func (c *Client) List(ctx context.Context, prefix string, limit int) ([]cloudflare.StorageKey, error) {
	const op = "kv.List"
	
	params := cloudflare.ListWorkersKVsOptions{
		Limit: limit,
	}
	
	if prefix != "" {
		params.Prefix = prefix
	}
	
	keys, _, err := c.api.ListWorkersKVs(ctx, cloudflare.AccountIdentifier(c.accountID),
		cloudflare.ListWorkersKVsParams{
			NamespaceID: c.namespaceID,
		})
	if err != nil {
		return nil, apperrors.NewAppError(op, err, 
			"failed to list keys", 500)
	}
	
	return keys, nil
}

// DeleteBulk deletes multiple keys
func (c *Client) DeleteBulk(ctx context.Context, keys []string) error {
	const op = "kv.DeleteBulk"
	
	_, err := c.api.DeleteWorkersKVBulk(ctx, cloudflare.AccountIdentifier(c.accountID),
		cloudflare.DeleteWorkersKVBulkParams{
			NamespaceID: c.namespaceID,
			Keys:        keys,
		})
	if err != nil {
		return apperrors.NewAppError(op, err, 
			"failed to delete keys in bulk", 500)
	}
	
	return nil
}
