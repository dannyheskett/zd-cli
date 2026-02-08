package commands

import (
	"fmt"

	"zd-cli/internal/client"

	"github.com/fatih/color"
)

// HandleError handles errors with user-friendly formatting
func HandleError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check if it's an API error
	if apiErr, ok := err.(*client.APIError); ok {
		color.Red("✗ %s\n", apiErr.Message)

		if apiErr.Description != "" {
			color.Yellow("  %s\n", apiErr.Description)
		}

		// Add context-specific suggestions
		switch apiErr.StatusCode {
		case 401:
			color.White("\nSuggestion: Run 'zd test' to verify your credentials\n")
		case 403:
			color.White("\nSuggestion: You may not have permission for this operation\n")
		case 404:
			color.White("\nSuggestion: Verify the resource ID exists\n")
		case 422:
			color.White("\nSuggestion: Check your input values and required fields\n")
		case 429:
			color.White("\nSuggestion: Rate limit exceeded. Wait a moment and try again\n")
			color.White("  Or reduce the frequency of --refresh flag usage\n")
		case 500, 502, 503:
			color.White("\nSuggestion: Zendesk is experiencing issues. Try again later\n")
		}

		return fmt.Errorf("operation failed")
	}

	// Generic error
	return fmt.Errorf("%s: %w", operation, err)
}
