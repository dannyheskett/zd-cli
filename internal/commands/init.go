package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"zd-cli/internal/auth"
	"zd-cli/internal/config"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Zendesk CLI configuration",
		Long:  "Initialize the Zendesk CLI by creating the configuration directory and setting up your first instance.",
		RunE:  runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	cfg, err := config.Load()
	if err == nil && len(cfg.Instances) > 0 {
		color.Yellow("Configuration already exists.")
		prompt := promptui.Prompt{
			Label:     "Do you want to add another instance",
			IsConfirm: true,
		}
		result, err := prompt.Run()
		if err != nil || strings.ToLower(result) != "y" {
			return nil
		}
		return runAddInstance(cmd, args)
	}

	// Create new config
	cfg = config.NewConfig()

	color.Cyan("Welcome to Zendesk CLI!")
	color.White("Let's set up your first Zendesk instance.\n")

	// Prompt for instance details
	instance, err := promptForInstance("")
	if err != nil {
		return err
	}

	// Add instance to config
	if err := cfg.AddInstance(instance); err != nil {
		return fmt.Errorf("failed to add instance: %w", err)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("\n✓ Configuration initialized successfully!")
	color.White("Instance '%s' is now active.\n", instance.Name)
	color.White("Run 'zd test' to verify your connection.\n")

	return nil
}

func promptForInstance(defaultName string) (*config.Instance, error) {
	instance := &config.Instance{}

	// Instance name
	namePrompt := promptui.Prompt{
		Label:   "Instance name",
		Default: defaultName,
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("instance name cannot be empty")
			}
			return nil
		},
	}
	name, err := namePrompt.Run()
	if err != nil {
		return nil, err
	}
	instance.Name = strings.TrimSpace(name)

	// Subdomain
	subdomainPrompt := promptui.Prompt{
		Label: "Zendesk subdomain (e.g., 'mycompany' for mycompany.zendesk.com)",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("subdomain cannot be empty")
			}
			return nil
		},
	}
	subdomain, err := subdomainPrompt.Run()
	if err != nil {
		return nil, err
	}
	instance.Subdomain = strings.TrimSpace(subdomain)

	// Auth type
	authTypePrompt := promptui.Select{
		Label: "Authentication method",
		Items: []string{"API Token", "OAuth"},
	}
	authTypeIdx, _, err := authTypePrompt.Run()
	if err != nil {
		return nil, err
	}

	if authTypeIdx == 0 {
		// API Token Authentication
		instance.AuthType = config.AuthTypeToken

		// Email
		emailPrompt := promptui.Prompt{
			Label: "Email address",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("email cannot be empty")
				}
				if !strings.Contains(input, "@") {
					return fmt.Errorf("invalid email format")
				}
				return nil
			},
		}
		email, err := emailPrompt.Run()
		if err != nil {
			return nil, err
		}
		instance.Email = strings.TrimSpace(email)

		// API Token
		tokenPrompt := promptui.Prompt{
			Label: "API Token",
			Mask:  '*',
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("API token cannot be empty")
				}
				return nil
			},
		}
		token, err := tokenPrompt.Run()
		if err != nil {
			return nil, err
		}
		instance.APIToken = strings.TrimSpace(token)

	} else {
		// OAuth Authentication
		if err := setupOAuth(instance); err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func setupOAuth(instance *config.Instance) error {
	instance.AuthType = config.AuthTypeOAuth

	color.Cyan("\nOAuth Setup\n")
	color.White("You need to create an OAuth client in your Zendesk instance first.\n")
	color.White("Go to: Admin Center → Apps and integrations → APIs → Zendesk API → OAuth Clients\n")
	color.White("Use redirect URL: %s\n\n", "http://localhost:8080/callback")

	// OAuth Client ID
	clientIDPrompt := promptui.Prompt{
		Label: "OAuth Client ID",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("client ID cannot be empty")
			}
			return nil
		},
	}
	clientID, err := clientIDPrompt.Run()
	if err != nil {
		return err
	}
	instance.OAuthClientID = strings.TrimSpace(clientID)

	// OAuth Client Secret
	secretPrompt := promptui.Prompt{
		Label: "OAuth Client Secret",
		Mask:  '*',
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("client secret cannot be empty")
			}
			return nil
		},
	}
	secret, err := secretPrompt.Run()
	if err != nil {
		return err
	}
	instance.OAuthSecret = strings.TrimSpace(secret)

	// Perform OAuth flow
	color.Cyan("\nStarting OAuth authorization flow...\n")

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

	// Store tokens
	instance.OAuthToken = token.AccessToken
	instance.OAuthRefresh = token.RefreshToken
	instance.SetOAuthExpiry(token.Expiry)

	color.Green("\n✓ OAuth authorization successful!\n")

	return nil
}
