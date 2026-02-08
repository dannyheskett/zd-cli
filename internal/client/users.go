package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// User represents a Zendesk user
type User struct {
	ID                  int64       `json:"id"`
	URL                 string      `json:"url"`
	Name                string      `json:"name"`
	Email               string      `json:"email"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"updated_at"`
	TimeZone            string      `json:"time_zone"`
	Phone               string      `json:"phone"`
	Photo               interface{} `json:"photo"`
	LocaleID            int         `json:"locale_id"`
	Locale              string      `json:"locale"`
	OrganizationID      *int64      `json:"organization_id"`
	Role                string      `json:"role"`
	Verified            bool        `json:"verified"`
	ExternalID          *string     `json:"external_id"`
	Tags                []string    `json:"tags"`
	Alias               string      `json:"alias"`
	Active              bool        `json:"active"`
	Shared              bool        `json:"shared"`
	SharedAgent         bool        `json:"shared_agent"`
	LastLoginAt         *string     `json:"last_login_at"`
	TwoFactorAuthEnabled bool       `json:"two_factor_auth_enabled"`
	Signature           string      `json:"signature"`
	Details             string      `json:"details"`
	Notes               string      `json:"notes"`
	CustomRoleID        *int64      `json:"custom_role_id"`
	Moderator           bool        `json:"moderator"`
	TicketRestriction   *string     `json:"ticket_restriction"`
	OnlyPrivateComments bool        `json:"only_private_comments"`
	RestrictedAgent     bool        `json:"restricted_agent"`
	Suspended           bool        `json:"suspended"`
}

// UsersResponse represents the response from listing users
type UsersResponse struct {
	Users      []User `json:"users"`
	NextPage   string `json:"next_page"`
	PreviousPage string `json:"previous_page"`
	Count      int    `json:"count"`
}

// UserResponse represents a single user response
type UserResponse struct {
	User User `json:"user"`
}

// GetMe retrieves information about the authenticated user
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	cacheKey := fmt.Sprintf("%s:users:me", c.subdomain)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp UserResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp.User, nil
			}
		}
	}

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, "/users/me.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userResp UserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &userResp.User, nil
}

// ListUsers retrieves a list of users
func (c *Client) ListUsers(ctx context.Context, page int, perPage int) (*UsersResponse, error) {
	cacheKey := fmt.Sprintf("%s:users:list:%d:%d", c.subdomain, page, perPage)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp UsersResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/users.json?page=%d&per_page=%d", page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var usersResp UsersResponse
	if err := json.Unmarshal(body, &usersResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &usersResp, nil
}

// SearchUsers searches for users by query
func (c *Client) SearchUsers(ctx context.Context, query string) ([]User, error) {
	cacheKey := fmt.Sprintf("%s:users:search:%s", c.subdomain, query)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp UsersResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return resp.Users, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/users/search.json?query=%s", url.QueryEscape(query))

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, ParseAPIError(resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var usersResp UsersResponse
	if err := json.Unmarshal(body, &usersResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return usersResp.Users, nil
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	Role   *string `json:"role,omitempty"`
	Verified *bool `json:"verified,omitempty"`
}

// CreateUser creates a new user
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	requestBody := map[string]interface{}{
		"user": req,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.makeUserRequest(ctx, http.MethodPost, "/users.json", body)
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(ctx context.Context, userID int64, req UpdateUserRequest) (*User, error) {
	requestBody := map[string]interface{}{
		"user": req,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/users/%d.json", userID)
	user, err := c.makeUserRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for this user
	if c.cache != nil {
		cacheKey := fmt.Sprintf("%s:users:%d", c.subdomain, userID)
		c.cache.Delete(cacheKey)
	}

	return user, nil
}

// SuspendUser suspends a user
func (c *Client) SuspendUser(ctx context.Context, userID int64) (*User, error) {
	path := fmt.Sprintf("/users/%d.json", userID)

	requestBody := map[string]interface{}{
		"user": map[string]interface{}{
			"suspended": true,
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	user, err := c.makeUserRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if c.cache != nil {
		cacheKey := fmt.Sprintf("%s:users:%d", c.subdomain, userID)
		c.cache.Delete(cacheKey)
	}

	return user, nil
}

// UnsuspendUser unsuspends a user
func (c *Client) UnsuspendUser(ctx context.Context, userID int64) (*User, error) {
	path := fmt.Sprintf("/users/%d.json", userID)

	requestBody := map[string]interface{}{
		"user": map[string]interface{}{
			"suspended": false,
		},
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	user, err := c.makeUserRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	if c.cache != nil {
		cacheKey := fmt.Sprintf("%s:users:%d", c.subdomain, userID)
		c.cache.Delete(cacheKey)
	}

	return user, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, userID int64) error {
	path := fmt.Sprintf("/users/%d.json", userID)

	url := c.GetBaseURL() + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return ParseAPIError(resp.StatusCode, body)
	}

	// Invalidate cache
	if c.cache != nil {
		cacheKey := fmt.Sprintf("%s:users:%d", c.subdomain, userID)
		c.cache.Delete(cacheKey)
	}

	return nil
}

// makeUserRequest makes a request that returns a user
func (c *Client) makeUserRequest(ctx context.Context, method, path string, body []byte) (*User, error) {
	url := c.GetBaseURL() + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Body = io.NopCloser(strings.NewReader(string(body)))
		req.ContentLength = int64(len(body))
	}

	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ParseAPIError(resp.StatusCode, respBody)
	}

	var userResp UserResponse
	if err := json.Unmarshal(respBody, &userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userResp.User, nil
}

// GetUser retrieves a specific user by ID
func (c *Client) GetUser(ctx context.Context, userID int64) (*User, error) {
	cacheKey := fmt.Sprintf("%s:users:%d", c.subdomain, userID)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp UserResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp.User, nil
			}
		}
	}

	// Fetch from API
	path := fmt.Sprintf("/users/%d.json", userID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userResp UserResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &userResp.User, nil
}
