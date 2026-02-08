package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ZendeskError represents a Zendesk API error response
type ZendeskError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
	Details     interface{} `json:"details"`
}

// APIError represents a formatted API error
type APIError struct {
	StatusCode  int
	Message     string
	Description string
	Details     interface{}
}

func (e *APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Description)
	}
	return e.Message
}

// ParseAPIError parses a Zendesk API error response
func ParseAPIError(statusCode int, body []byte) error {
	// Try to parse as Zendesk error
	var zdError ZendeskError
	if err := json.Unmarshal(body, &zdError); err == nil && zdError.Error != "" {
		return &APIError{
			StatusCode:  statusCode,
			Message:     zdError.Error,
			Description: zdError.Description,
			Details:     zdError.Details,
		}
	}

	// Fallback to generic error
	return &APIError{
		StatusCode:  statusCode,
		Message:     getStatusMessage(statusCode),
		Description: string(body),
	}
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsAuthError checks if the error is an authentication error
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// getStatusMessage returns a user-friendly message for HTTP status codes
func getStatusMessage(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Authentication Failed"
	case http.StatusForbidden:
		return "Access Denied"
	case http.StatusNotFound:
		return "Resource Not Found"
	case http.StatusUnprocessableEntity:
		return "Invalid Input"
	case http.StatusTooManyRequests:
		return "Rate Limit Exceeded"
	case http.StatusInternalServerError:
		return "Zendesk Server Error"
	case http.StatusServiceUnavailable:
		return "Zendesk Service Unavailable"
	default:
		return fmt.Sprintf("HTTP %d Error", statusCode)
	}
}

// FormatUserFriendlyError formats an error with helpful suggestions
func FormatUserFriendlyError(err error) string {
	if apiErr, ok := err.(*APIError); ok {
		msg := fmt.Sprintf("Error: %s", apiErr.Message)

		if apiErr.Description != "" {
			msg += fmt.Sprintf("\n  %s", apiErr.Description)
		}

		// Add helpful suggestions
		switch apiErr.StatusCode {
		case http.StatusUnauthorized:
			msg += "\n\nSuggestion: Check your credentials with 'zd test'"
		case http.StatusForbidden:
			msg += "\n\nSuggestion: You may not have permission for this operation"
		case http.StatusNotFound:
			msg += "\n\nSuggestion: Verify the resource ID exists"
		case http.StatusTooManyRequests:
			msg += "\n\nSuggestion: You've hit the rate limit. Wait a minute and try again, or use --refresh less frequently"
		case http.StatusUnprocessableEntity:
			msg += "\n\nSuggestion: Check your input values and required fields"
		}

		return msg
	}

	return err.Error()
}
