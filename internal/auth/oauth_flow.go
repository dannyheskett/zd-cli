package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	// DefaultRedirectURL is the local callback URL for OAuth
	DefaultRedirectURL = "http://localhost:8080/callback"
	// CallbackPort is the port for the local callback server
	CallbackPort = 8080
)

// OAuthConfig holds OAuth2 configuration for Zendesk
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Subdomain    string
}

// GetOAuthConfig creates an OAuth2 config for Zendesk
func GetOAuthConfig(cfg OAuthConfig) *oauth2.Config {
	authURL := fmt.Sprintf("https://%s.zendesk.com/oauth/authorizations/new", cfg.Subdomain)
	tokenURL := fmt.Sprintf("https://%s.zendesk.com/oauth/tokens", cfg.Subdomain)

	if cfg.RedirectURL == "" {
		cfg.RedirectURL = DefaultRedirectURL
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		// Zendesk requires scope parameter as space-separated string
		Scopes: []string{"read", "write"},
	}
}

// PerformOAuthFlow performs the OAuth authorization flow with browser
func PerformOAuthFlow(ctx context.Context, cfg OAuthConfig) (*oauth2.Token, error) {
	oauthCfg := GetOAuthConfig(cfg)

	// Generate random state for CSRF protection
	state := fmt.Sprintf("state-%d", time.Now().Unix())

	// Create channel to receive authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server to receive callback
	server := &http.Server{Addr: fmt.Sprintf(":%d", CallbackPort)}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Log the callback for debugging
		fmt.Printf("\n📥 Received callback: %s\n", r.URL.String())

		// Verify state
		if r.URL.Query().Get("state") != state {
			errChan <- fmt.Errorf("state mismatch - possible CSRF attack")
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		// Check for errors
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errDesc := r.URL.Query().Get("error_description")
			errChan <- fmt.Errorf("authorization failed: %s - %s", errMsg, errDesc)
			http.Error(w, "Authorization failed", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		// Send success page
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Authorization Successful</title>
				<style>
					body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
					h1 { color: #28a745; }
				</style>
			</head>
			<body>
				<h1>Authorization Successful!</h1>
				<p>You can close this window and return to the terminal.</p>
				<p>The zd-cli tool is now authorized to access your Zendesk instance.</p>
			</body>
			</html>
		`)

		codeChan <- code
	})

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start callback server: %w", err)
		}
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Generate authorization URL (don't use AccessTypeOffline - Zendesk doesn't support it)
	authURL := oauthCfg.AuthCodeURL(state)

	fmt.Printf("\nOpening browser for authorization...\n")
	fmt.Printf("\nIf browser doesn't open automatically, visit:\n")
	fmt.Printf("%s\n\n", authURL)

	// Try to open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("⚠ Could not open browser automatically: %v\n", err)
		fmt.Printf("Please open the URL manually.\n\n")
	}

	fmt.Println("Waiting for authorization (timeout: 5 minutes)...")

	// Wait for code or error
	var code string
	select {
	case code = <-codeChan:
		// Success - continue
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-ctx.Done():
		server.Shutdown(context.Background())
		return nil, ctx.Err()
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authorization timeout - no response received within 5 minutes")
	}

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)

	// Exchange code for token
	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for token: %w", err)
	}

	return token, nil
}

// RefreshToken refreshes an OAuth token if it's expired
func RefreshToken(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*oauth2.Token, error) {
	if token.Valid() {
		return token, nil
	}

	tokenSource := cfg.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// ValidateOAuthToken validates that OAuth token fields are populated
func ValidateOAuthToken(accessToken, refreshToken, expiry string) error {
	if accessToken == "" {
		return fmt.Errorf("OAuth access token is required")
	}
	if refreshToken == "" {
		return fmt.Errorf("OAuth refresh token is required")
	}
	if expiry == "" {
		return fmt.Errorf("OAuth token expiry is required")
	}
	return nil
}

// ValidateOAuthConfig validates OAuth client configuration
func ValidateOAuthConfig(clientID, clientSecret string) error {
	if clientID == "" {
		return fmt.Errorf("OAuth client ID is required")
	}
	if clientSecret == "" {
		return fmt.Errorf("OAuth client secret is required")
	}
	return nil
}
