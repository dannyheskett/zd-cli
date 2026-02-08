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

// Ticket represents a Zendesk ticket
type Ticket struct {
	ID              int64    `json:"id"`
	URL             string   `json:"url"`
	ExternalID      *string  `json:"external_id"`
	Type            string   `json:"type"`
	Subject         string   `json:"subject"`
	RawSubject      string   `json:"raw_subject"`
	Description     string   `json:"description"`
	Priority        string   `json:"priority"`
	Status          string   `json:"status"`
	Recipient       *string  `json:"recipient"`
	RequesterID     int64    `json:"requester_id"`
	SubmitterID     int64    `json:"submitter_id"`
	AssigneeID      *int64   `json:"assignee_id"`
	OrganizationID  *int64   `json:"organization_id"`
	GroupID         *int64   `json:"group_id"`
	CollaboratorIDs []int64  `json:"collaborator_ids"`
	FollowerIDs     []int64  `json:"follower_ids"`
	EmailCCIDs      []int64  `json:"email_cc_ids"`
	ForumTopicID    *int64   `json:"forum_topic_id"`
	ProblemID       *int64   `json:"problem_id"`
	HasIncidents    bool     `json:"has_incidents"`
	DueAt           *string  `json:"due_at"`
	Tags            []string `json:"tags"`
	Via             struct {
		Channel string `json:"channel"`
		Source  struct {
			From interface{} `json:"from"`
			To   interface{} `json:"to"`
			Rel  *string     `json:"rel"`
		} `json:"source"`
	} `json:"via"`
	CustomFields    []interface{} `json:"custom_fields"`
	SatisfactionRating *struct {
		Score   string `json:"score"`
		Comment string `json:"comment"`
	} `json:"satisfaction_rating"`
	SharingAgreementIDs []int64 `json:"sharing_agreement_ids"`
	Fields              []interface{} `json:"fields"`
	FollowupIDs         []int64 `json:"followup_ids"`
	TicketFormID        *int64  `json:"ticket_form_id"`
	BrandID             int64   `json:"brand_id"`
	AllowChannelback    bool    `json:"allow_channelback"`
	AllowAttachments    bool    `json:"allow_attachments"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// TicketsResponse represents the response from listing tickets
type TicketsResponse struct {
	Tickets      []Ticket `json:"tickets"`
	NextPage     string   `json:"next_page"`
	PreviousPage string   `json:"previous_page"`
	Count        int      `json:"count"`
}

// TicketResponse represents a single ticket response
type TicketResponse struct {
	Ticket Ticket `json:"ticket"`
}

// Comment represents a ticket comment
type Comment struct {
	ID          int64    `json:"id"`
	Type        string   `json:"type"`
	AuthorID    int64    `json:"author_id"`
	Body        string   `json:"body"`
	HTMLBody    string   `json:"html_body"`
	PlainBody   string   `json:"plain_body"`
	Public      bool     `json:"public"`
	Attachments []interface{} `json:"attachments"`
	AuditID     int64    `json:"audit_id"`
	Via         struct {
		Channel string `json:"channel"`
		Source  struct {
			From interface{} `json:"from"`
			To   interface{} `json:"to"`
			Rel  *string     `json:"rel"`
		} `json:"source"`
	} `json:"via"`
	CreatedAt   string `json:"created_at"`
	Metadata    interface{} `json:"metadata"`
}

// CommentsResponse represents the response from listing comments
type CommentsResponse struct {
	Comments     []Comment `json:"comments"`
	NextPage     string    `json:"next_page"`
	PreviousPage string    `json:"previous_page"`
	Count        int       `json:"count"`
}

// ListTickets retrieves a list of tickets
func (c *Client) ListTickets(ctx context.Context, page int, perPage int, status string) (*TicketsResponse, error) {
	cacheKey := fmt.Sprintf("%s:tickets:list:%d:%d:%s", c.subdomain, page, perPage, status)

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
	path := fmt.Sprintf("/tickets.json?page=%d&per_page=%d", page, perPage)
	if status != "" {
		path += fmt.Sprintf("&status=%s", url.QueryEscape(status))
	}

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list tickets (status %d): %s", resp.StatusCode, string(body))
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

// GetTicket retrieves a specific ticket by ID
func (c *Client) GetTicket(ctx context.Context, ticketID int64) (*Ticket, error) {
	cacheKey := fmt.Sprintf("%s:tickets:%d", c.subdomain, ticketID)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp TicketResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return &resp.Ticket, nil
			}
		}
	}

	// Fetch from API
	path := fmt.Sprintf("/tickets/%d.json", ticketID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get ticket (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ticketResp TicketResponse
	if err := json.Unmarshal(body, &ticketResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return &ticketResp.Ticket, nil
}

// GetTicketComments retrieves comments for a ticket
func (c *Client) GetTicketComments(ctx context.Context, ticketID int64) ([]Comment, error) {
	cacheKey := fmt.Sprintf("%s:tickets:%d:comments", c.subdomain, ticketID)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp CommentsResponse
			if err := json.Unmarshal(cached, &resp); err == nil {
				return resp.Comments, nil
			}
		}
	}

	// Fetch from API
	path := fmt.Sprintf("/tickets/%d/comments.json", ticketID)
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get ticket comments (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var commentsResp CommentsResponse
	if err := json.Unmarshal(body, &commentsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return commentsResp.Comments, nil
}

// CreateTicketRequest represents a ticket creation request
type CreateTicketRequest struct {
	Subject     string   `json:"subject"`
	Description string   `json:"comment,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Type        string   `json:"type,omitempty"`
	Status      string   `json:"status,omitempty"`
	AssigneeID  *int64   `json:"assignee_id,omitempty"`
	GroupID     *int64   `json:"group_id,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateTicketRequest represents a ticket update request
type UpdateTicketRequest struct {
	Subject    *string  `json:"subject,omitempty"`
	Priority   *string  `json:"priority,omitempty"`
	Status     *string  `json:"status,omitempty"`
	AssigneeID *int64   `json:"assignee_id,omitempty"`
	GroupID    *int64   `json:"group_id,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Comment    *struct {
		Body   string `json:"body"`
		Public bool   `json:"public"`
	} `json:"comment,omitempty"`
}

// CreateTicket creates a new ticket
func (c *Client) CreateTicket(ctx context.Context, req CreateTicketRequest) (*Ticket, error) {
	requestBody := map[string]interface{}{
		"ticket": map[string]interface{}{
			"subject": req.Subject,
			"comment": map[string]interface{}{
				"body": req.Description,
			},
		},
	}

	// Add optional fields
	ticket := requestBody["ticket"].(map[string]interface{})
	if req.Priority != "" {
		ticket["priority"] = req.Priority
	}
	if req.Type != "" {
		ticket["type"] = req.Type
	}
	if req.Status != "" {
		ticket["status"] = req.Status
	}
	if req.AssigneeID != nil {
		ticket["assignee_id"] = *req.AssigneeID
	}
	if req.GroupID != nil {
		ticket["group_id"] = *req.GroupID
	}
	if len(req.Tags) > 0 {
		ticket["tags"] = req.Tags
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.makeTicketRequest(ctx, http.MethodPost, "/tickets.json", body)
}

// UpdateTicket updates an existing ticket
func (c *Client) UpdateTicket(ctx context.Context, ticketID int64, req UpdateTicketRequest) (*Ticket, error) {
	requestBody := map[string]interface{}{
		"ticket": req,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/tickets/%d.json", ticketID)
	ticket, err := c.makeTicketRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	// Invalidate cache for this ticket
	if c.cache != nil {
		cacheKey := fmt.Sprintf("%s:tickets:%d", c.subdomain, ticketID)
		c.cache.Delete(cacheKey)
	}

	return ticket, nil
}

// makeTicketRequest makes a request that returns a ticket
func (c *Client) makeTicketRequest(ctx context.Context, method, path string, body []byte) (*Ticket, error) {
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
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var ticketResp TicketResponse
	if err := json.Unmarshal(respBody, &ticketResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ticketResp.Ticket, nil
}

// SearchTickets searches for tickets by query
func (c *Client) SearchTickets(ctx context.Context, query string) ([]Ticket, error) {
	cacheKey := fmt.Sprintf("%s:tickets:search:%s", c.subdomain, query)

	// Try cache first
	if c.useCache && c.cache != nil {
		if cached, found := c.cache.Get(cacheKey); found {
			var resp struct {
				Results []Ticket `json:"results"`
			}
			if err := json.Unmarshal(cached, &resp); err == nil {
				return resp.Results, nil
			}
		}
	}

	// Build search query - type:ticket is required for ticket search
	searchQuery := fmt.Sprintf("type:ticket %s", query)
	path := fmt.Sprintf("/search.json?query=%s", url.QueryEscape(searchQuery))

	// Fetch from API
	resp, err := c.makeRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to search tickets (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp struct {
		Results []Ticket `json:"results"`
		Count   int      `json:"count"`
	}
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the result
	if c.useCache && c.cache != nil {
		c.cache.Set(cacheKey, body)
	}

	return searchResp.Results, nil
}
