package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Organization represents a Zendesk organization
type Organization struct {
	ID                 int64    `json:"id"`
	URL                string   `json:"url"`
	ExternalID         *string  `json:"external_id"`
	Name               string   `json:"name"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	DomainNames        []string `json:"domain_names"`
	Details            string   `json:"details"`
	Notes              string   `json:"notes"`
	GroupID            *int64   `json:"group_id"`
	SharedTickets      bool     `json:"shared_tickets"`
	SharedComments     bool     `json:"shared_comments"`
	Tags               []string `json:"tags"`
	OrganizationFields map[string]interface{} `json:"organization_fields"`
}

// OrganizationsResponse represents the response from listing organizations
type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
	NextPage      string         `json:"next_page"`
	PreviousPage  string         `json:"previous_page"`
	Count         int            `json:"count"`
}

// OrganizationResponse represents a single organization response
type OrganizationResponse struct {
	Organization Organization `json:"organization"`
}

// ListOrganizations retrieves a list of organizations
func (c *Client) ListOrganizations(ctx context.Context, page int, perPage int) (*OrganizationsResponse, error) {
	cacheKey := fmt.Sprintf("%s:organizations:list:%d:%d", c.subdomain, page, perPage)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp OrganizationsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/organizations.json?page=%d&per_page=%d", page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list organizations (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var orgsResp OrganizationsResponse
	if err := json.Unmarshal(body, &orgsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &orgsResp, nil
}

// GetOrganization retrieves a specific organization by ID
func (c *Client) GetOrganization(ctx context.Context, orgID int64) (*Organization, error) {
	cacheKey := fmt.Sprintf("%s:organizations:%d", c.subdomain, orgID)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp OrganizationResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp.Organization, nil
			}
		}
	}

	// Fetch from API
	path := fmt.Sprintf("/organizations/%d.json", orgID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get organization (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var orgResp OrganizationResponse
	if err := json.Unmarshal(body, &orgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &orgResp.Organization, nil
}

// SearchOrganizations searches for organizations by query
func (c *Client) SearchOrganizations(ctx context.Context, query string) ([]Organization, error) {
	cacheKey := fmt.Sprintf("%s:organizations:search:%s", c.subdomain, query)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp OrganizationsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return resp.Organizations, nil
			}
		}
	}

	// Build query parameters - Zendesk requires 'name' parameter for org search
	path := fmt.Sprintf("/organizations/search.json?name=%s", url.QueryEscape(query))

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search organizations (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var orgsResp OrganizationsResponse
	if err := json.Unmarshal(body, &orgsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return orgsResp.Organizations, nil
}

// GetOrganizationUsers retrieves users in an organization
func (c *Client) GetOrganizationUsers(ctx context.Context, orgID int64, page int, perPage int) (*UsersResponse, error) {
	cacheKey := fmt.Sprintf("%s:organizations:%d:users:%d:%d", c.subdomain, orgID, page, perPage)

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
	path := fmt.Sprintf("/organizations/%d/users.json?page=%d&per_page=%d", orgID, page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get organization users (status %d): %s", resp.StatusCode, string(body))
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

// GetOrganizationTickets retrieves tickets for an organization
func (c *Client) GetOrganizationTickets(ctx context.Context, orgID int64, page int, perPage int) (*TicketsResponse, error) {
	cacheKey := fmt.Sprintf("%s:organizations:%d:tickets:%d:%d", c.subdomain, orgID, page, perPage)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp TicketsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp, nil
			}
		}
	}

	// Build query parameters
	path := fmt.Sprintf("/organizations/%d/tickets.json?page=%d&per_page=%d", orgID, page, perPage)

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get organization tickets (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ticketsResp TicketsResponse
	if err := json.Unmarshal(body, &ticketsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &ticketsResp, nil
}
