package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Group represents a Zendesk group
type Group struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	Name      string `json:"name"`
	Description string `json:"description"`
	Default   bool   `json:"default"`
	Deleted   bool   `json:"deleted"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GroupsResponse represents the response from listing groups
type GroupsResponse struct {
	Groups       []Group `json:"groups"`
	NextPage     string  `json:"next_page"`
	PreviousPage string  `json:"previous_page"`
	Count        int     `json:"count"`
}

// GroupResponse represents a single group response
type GroupResponse struct {
	Group Group `json:"group"`
}

// GroupMembership represents a user's membership in a group
type GroupMembership struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	GroupID   int64  `json:"group_id"`
	Default   bool   `json:"default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GroupMembershipsResponse represents the response from listing group memberships
type GroupMembershipsResponse struct {
	GroupMemberships []GroupMembership `json:"group_memberships"`
	NextPage         string            `json:"next_page"`
	PreviousPage     string            `json:"previous_page"`
	Count            int               `json:"count"`
}

// ListGroups retrieves a list of groups
func (c *Client) ListGroups(ctx context.Context, page int, perPage int) (*GroupsResponse, error) {
	cacheKey := fmt.Sprintf("%s:groups:list:%d:%d", c.subdomain, page, perPage)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp GroupsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/groups.json?page=%d&per_page=%d", page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list groups (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var groupsResp GroupsResponse
	if err := json.Unmarshal(body, &groupsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &groupsResp, nil
}

// GetGroup retrieves a specific group by ID
func (c *Client) GetGroup(ctx context.Context, groupID int64) (*Group, error) {
	cacheKey := fmt.Sprintf("%s:groups:%d", c.subdomain, groupID)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp GroupResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp.Group, nil
			}
		}
	}

	// Fetch from API
	path := fmt.Sprintf("/groups/%d.json", groupID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get group (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var groupResp GroupResponse
	if err := json.Unmarshal(body, &groupResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &groupResp.Group, nil
}

// GetGroupUsers retrieves users in a group
func (c *Client) GetGroupUsers(ctx context.Context, groupID int64, page int, perPage int) (*UsersResponse, error) {
	cacheKey := fmt.Sprintf("%s:groups:%d:users:%d:%d", c.subdomain, groupID, page, perPage)

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
	path := fmt.Sprintf("/groups/%d/users.json?page=%d&per_page=%d", groupID, page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get group users (status %d): %s", resp.StatusCode, string(body))
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

// GetGroupMemberships retrieves memberships for a group
func (c *Client) GetGroupMemberships(ctx context.Context, groupID int64, page int, perPage int) (*GroupMembershipsResponse, error) {
	cacheKey := fmt.Sprintf("%s:groups:%d:memberships:%d:%d", c.subdomain, groupID, page, perPage)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp GroupMembershipsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/groups/%d/memberships.json?page=%d&per_page=%d", groupID, page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get group memberships (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var membershipsResp GroupMembershipsResponse
	if err := json.Unmarshal(body, &membershipsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &membershipsResp, nil
}
