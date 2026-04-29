package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with namespace-aware operations.
type Client struct {
	api       *vaultapi.Client
	namespace string
}

// NewClient creates a new Vault client from the provided address, token, and optional namespace.
func NewClient(address, token, namespace string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	if namespace != "" {
		api.SetNamespace(namespace)
	}

	return &Client{api: api, namespace: namespace}, nil
}

// GetSecrets reads all key-value pairs from the given KV v2 path.
func (c *Client) GetSecrets(ctx context.Context, mountPath, secretPath string) (map[string]string, error) {
	fullPath := strings.TrimSuffix(mountPath, "/") + "/data/" + strings.TrimPrefix(secretPath, "/")

	secret, err := c.api.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", fullPath, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no secret found at path %q", fullPath)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected secret data format at path %q", fullPath)
	}

	result := make(map[string]string, len(data))
	for k, v := range data {
		result[k] = fmt.Sprintf("%v", v)
	}

	return result, nil
}

// ListPaths lists all secret paths under the given KV v2 mount and prefix.
// Paths ending with "/" are subdirectories and can be listed recursively.
func (c *Client) ListPaths(ctx context.Context, mountPath, prefix string) ([]string, error) {
	fullPath := strings.TrimSuffix(mountPath, "/") + "/metadata/" + strings.TrimPrefix(prefix, "/")

	secret, err := c.api.Logical().ListWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("listing paths at %q: %w", fullPath, err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	paths := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			paths = append(paths, s)
		}
	}

	return paths, nil
}

// ListPathsRecursive lists all leaf secret paths (non-directory) under the given
// KV v2 mount and prefix, traversing subdirectories automatically.
func (c *Client) ListPathsRecursive(ctx context.Context, mountPath, prefix string) ([]string, error) {
	entries, err := c.ListPaths(ctx, mountPath, prefix)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, entry := range entries {
		if strings.HasSuffix(entry, "/") {
			subPrefix := strings.TrimSuffix(prefix, "/") + "/" + entry
			children, err := c.ListPathsRecursive(ctx, mountPath, subPrefix)
			if err != nil {
				return nil, fmt.Errorf("listing paths under %q: %w", subPrefix, err)
			}
			result = append(result, children...)
		} else {
			// Prepend the prefix so callers receive full relative paths.
			leafPath := strings.TrimSuffix(prefix, "/") + "/" + entry
			result = append(result, strings.TrimPrefix(leafPath, "/"))
		}
	}

	return result, nil
}
