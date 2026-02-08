package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"zd-cli/internal/auth"
	"zd-cli/internal/cache"
	"zd-cli/internal/config"
)

// Client wraps the Zendesk API client
type Client struct {
	subdomain  string
	httpClient *http.Client
	authHeader string
	cache      *cache.Cache
	useCache   bool
}

// NewClient creates a new Zendesk API client from an instance configuration
func NewClient(instance *config.Instance) (*Client, error) {
	return NewClientWithCache(instance, true)
}

// NewClientWithCache creates a new Zendesk API client with optional caching
func NewClientWithCache(instance *config.Instance, useCache bool) (*Client, error) {
	client := &Client{
		subdomain:  instance.Subdomain,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		useCache:   useCache,
	}

	switch instance.AuthType {
	case config.AuthTypeToken:
		if err := auth.ValidateTokenAuth(instance.Email, instance.APIToken); err != nil {
			return nil, err
		}
		encodedToken := auth.EncodeToken(instance.Email, instance.APIToken)
		client.authHeader = fmt.Sprintf("Basic %s", encodedToken)

	case config.AuthTypeOAuth:
		if err := auth.ValidateOAuthToken(instance.OAuthToken, instance.OAuthRefresh, instance.OAuthExpiry); err != nil {
			return nil, err
		}
		if err := auth.ValidateOAuthConfig(instance.OAuthClientID, instance.OAuthSecret); err != nil {
			return nil, err
		}
		client.authHeader = fmt.Sprintf("Bearer %s", instance.OAuthToken)

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", instance.AuthType)
	}

	// Initialize cache with default TTL
	if useCache {
		c, err := cache.New(cache.DefaultTTL)
		if err != nil {
			// Cache initialization failed, continue without cache
			client.useCache = false
		} else {
			client.cache = c
		}
	}

	return client, nil
}

// GetBaseURL returns the base API URL for the instance
func (c *Client) GetBaseURL() string {
	return fmt.Sprintf("https://%s.zendesk.com/api/v2", c.subdomain)
}

// makeRequest makes an HTTP request to the Zendesk API
func (c *Client) makeRequest(ctx context.Context, method, path string) (*http.Response, error) {
	url := c.GetBaseURL() + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// TestConnection tests the connection to the Zendesk instance
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users/me.json")
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return ParseAPIError(resp.StatusCode, body)
	}

	return nil
}

// GetCurrentUser retrieves information about the authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (map[string]interface{}, error) {
	cacheKey := fmt.Sprintf("%s:users:me", c.subdomain)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var result map[string]interface{}
			if err := json.Unmarshal(cached, &result); err == nil {
				return result, nil
			}
		}
	}

	// Cache miss or disabled, fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users/me.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get current user (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return result, nil
}

// ClearCache clears all cached data for this client
func (c *Client) ClearCache() error {
	if c.cache != nil {
		return c.cache.Clear()
	}
	return nil
}
