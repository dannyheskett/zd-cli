package commands

import (
	"context"
	"fmt"
	"time"

	"zd-cli/internal/client"
	"zd-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewTestCommand creates the test command
func NewTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test connection to the current Zendesk instance",
		Long:  "Verify that the current instance configuration is valid and the API is accessible.",
		RunE:  runTest,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data from API")

	return cmd
}

func runTest(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err == config.ErrConfigNotFound {
		return fmt.Errorf("no configuration found. Run 'zd init' to get started")
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get current instance
	instance, err := cfg.GetCurrentInstance()
	if err != nil {
		return fmt.Errorf("no current instance set. Run 'zd instance switch <name>' to select an instance")
	}

	// Check if refresh flag is set
	refresh, _ := cmd.Flags().GetBool("refresh")
	useCache := !refresh

	color.Cyan("Testing connection to '%s' (%s.zendesk.com)...\n", instance.Name, instance.Subdomain)
	if !useCache {
		color.Yellow("(bypassing cache)\n")
	}

	// Create client with cache option
	zdClient, err := client.NewClientWithCache(instance, useCache)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := zdClient.TestConnection(ctx); err != nil {
		color.Red("✗ Connection test failed: %v\n", err)
		return err
	}

	// Get current user info
	user, err := zdClient.GetCurrentUser(ctx)
	if err != nil {
		color.Yellow("⚠ Connection succeeded but failed to get user info: %v\n", err)
		return nil
	}

	color.Green("✓ Connection successful!\n")

	// Display user info if available
	if userObj, ok := user["user"].(map[string]interface{}); ok {
		if name, ok := userObj["name"].(string); ok {
			color.White("  Authenticated as: %s\n", name)
		}
		if email, ok := userObj["email"].(string); ok {
			color.White("  Email: %s\n", email)
		}
		if role, ok := userObj["role"].(string); ok {
			color.White("  Role: %s\n", role)
		}
	}

	return nil
}
