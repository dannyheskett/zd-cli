package commands

import (
	"context"
	"fmt"
	"time"

	"zd-cli/internal/auth"
	"zd-cli/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewReauthCommand creates the reauth command
func NewReauthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reauth [instance-name]",
		Short: "Re-authorize an OAuth instance",
		Long:  "Re-run the OAuth authorization flow for an existing OAuth-configured instance.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runReauth,
	}

	return cmd
}

func runReauth(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine which instance to reauth
	var instanceName string
	if len(args) > 0 {
		instanceName = args[0]
	} else {
		// Use current instance
		if cfg.Current == "" {
			return fmt.Errorf("no current instance set. Specify instance name: zd reauth <instance-name>")
		}
		instanceName = cfg.Current
	}

	// Get the instance
	instance, ok := cfg.Instances[instanceName]
	if !ok {
		return fmt.Errorf("instance '%s' not found", instanceName)
	}

	// Verify it's an OAuth instance
	if instance.AuthType != config.AuthTypeOAuth {
		return fmt.Errorf("instance '%s' uses %s authentication, not OAuth", instanceName, instance.AuthType)
	}

	// Verify OAuth config exists
	if err := auth.ValidateOAuthConfig(instance.OAuthClientID, instance.OAuthSecret); err != nil {
		return fmt.Errorf("OAuth client credentials missing: %w", err)
	}

	color.Cyan("Re-authorizing instance '%s' (%s.zendesk.com)...\n", instanceName, instance.Subdomain)

	// Perform OAuth flow
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	oauthCfg := auth.OAuthConfig{
		ClientID:     instance.OAuthClientID,
		ClientSecret: instance.OAuthSecret,
		Subdomain:    instance.Subdomain,
		RedirectURL:  auth.DefaultRedirectURL,
	}

	token, err := auth.PerformOAuthFlow(ctx, oauthCfg)
	if err != nil {
		return fmt.Errorf("OAuth authorization failed: %w", err)
	}

	// Update instance with new tokens
	instance.OAuthToken = token.AccessToken
	instance.OAuthRefresh = token.RefreshToken
	instance.SetOAuthExpiry(token.Expiry)

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("\n✓ Re-authorization successful!\n")
	color.White("Instance '%s' has been updated with new OAuth tokens.\n", instanceName)

	return nil
}
