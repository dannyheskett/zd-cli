package auth

import (
	"encoding/base64"
	"fmt"
)

// EncodeToken encodes email and API token in Zendesk format (email/token:token)
// Zendesk uses Basic authentication with email/token as username and token as password
func EncodeToken(email, apiToken string) string {
	credentials := fmt.Sprintf("%s/token:%s", email, apiToken)
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}

// ValidateTokenAuth validates that email and API token are provided
func ValidateTokenAuth(email, apiToken string) error {
	if email == "" {
		return fmt.Errorf("email is required for token authentication")
	}
	if apiToken == "" {
		return fmt.Errorf("API token is required for token authentication")
	}
	return nil
}
