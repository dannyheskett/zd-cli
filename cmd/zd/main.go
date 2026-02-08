package main

import (
	"fmt"
	"os"

	"zd-cli/internal/commands"
	"github.com/spf13/cobra"
)

var (
	version = "0.3.0"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "zd",
	Short: "Zendesk CLI - Manage your Zendesk instances from the command line",
	Long: `zd is a command-line interface for managing Zendesk instances.

It supports multiple instances with easy switching, API token and OAuth authentication,
and provides commands for managing tickets, users, and more.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Register commands
	rootCmd.AddCommand(commands.NewInitCommand())
	rootCmd.AddCommand(commands.NewInstanceCommand())
	rootCmd.AddCommand(commands.NewTestCommand())
	rootCmd.AddCommand(commands.NewCompletionCommand(rootCmd))
	rootCmd.AddCommand(commands.NewInstallCommand())
	rootCmd.AddCommand(commands.NewCacheCommand())
	rootCmd.AddCommand(commands.NewUserCommand())
	rootCmd.AddCommand(commands.NewTicketCommand())
	rootCmd.AddCommand(commands.NewOrganizationCommand())
	rootCmd.AddCommand(commands.NewGroupCommand())
	rootCmd.AddCommand(commands.NewReauthCommand())

	// Global flags
	rootCmd.PersistentFlags().String("instance", "", "Override the current instance")
	rootCmd.PersistentFlags().String("config", "", "Config file path (default: ~/.zd/config)")

	// Disable the default completion command since we have our own
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
