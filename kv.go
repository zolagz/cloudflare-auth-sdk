package cloudflare_auth_sdk

import (
	"context"
	"fmt"
	"io"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/kv"
)

// KVGet retrieves a value from the KV store.
func (c *Client) KVGet(ctx context.Context, key string) ([]byte, error) {
	const op = "Client.KVGet"

	resp, err := c.cfClient.KV.Namespaces.Values.Get(ctx, c.namespaceID, key,
		kv.NamespaceValueGetParams{
			AccountID: cloudflare.F(c.accountID),
		})
	if err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("failed to get key: %s", key), 500)
	}
	defer resp.Body.Close()

	value, err := readAll(resp.Body)
	if err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("failed to read response for key: %s", key), 500)
	}

	return value, nil
}

// KVSet stores a key-value pair in the KV store.
func (c *Client) KVSet(ctx context.Context, key string, value []byte, opts *KVWriteOptions) error {
	const op = "Client.KVSet"

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

	_, err := c.cfClient.KV.Namespaces.Values.Update(ctx, c.namespaceID, key, params)
	if err != nil {
		return NewAppError(op, err, fmt.Sprintf("failed to set key: %s", key), 500)
	}

	return nil
}

// KVDelete removes a key from the KV store.
func (c *Client) KVDelete(ctx context.Context, key string) error {
	const op = "Client.KVDelete"

	_, err := c.cfClient.KV.Namespaces.Values.Delete(ctx, c.namespaceID, key,
		kv.NamespaceValueDeleteParams{
			AccountID: cloudflare.F(c.accountID),
		})
	if err != nil {
		return NewAppError(op, err, fmt.Sprintf("failed to delete key: %s", key), 500)
	}

	return nil
}

// KVList lists keys in the KV namespace.
func (c *Client) KVList(ctx context.Context, prefix string, limit int) ([]KVKey, error) {
	const op = "Client.KVList"

	params := kv.NamespaceKeyListParams{
		AccountID: cloudflare.F(c.accountID),
	}

	if prefix != "" {
		params.Prefix = cloudflare.F(prefix)
	}

	if limit > 0 {
		params.Limit = cloudflare.F(float64(limit))
	}

	resp, err := c.cfClient.KV.Namespaces.Keys.List(ctx, c.namespaceID, params)
	if err != nil {
		return nil, NewAppError(op, err, "failed to list keys", 500)
	}

	var keys []KVKey
	for _, item := range resp.Result {
		keys = append(keys, KVKey{
			Name:       item.Name,
			Expiration: item.Expiration,
			Metadata:   item.Metadata,
		})
	}

	return keys, nil
}

// KVDeleteBulk deletes multiple keys from the KV store.
func (c *Client) KVDeleteBulk(ctx context.Context, keys []string) error {
	const op = "Client.KVDeleteBulk"

	_, err := c.cfClient.KV.Namespaces.Keys.BulkDelete(ctx, c.namespaceID,
		kv.NamespaceKeyBulkDeleteParams{
			AccountID: cloudflare.F(c.accountID),
			Body:      keys,
		})
	if err != nil {
		return NewAppError(op, err, "failed to delete keys in bulk", 500)
	}

	return nil
}

// readAll is a helper to read all data from an io.Reader
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
