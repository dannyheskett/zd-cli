package commands

import (
	"fmt"
	"strings"

	"zd-cli/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewInstanceCommand creates the instance management command
func NewInstanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage Zendesk instances",
		Long:  "Add, remove, list, and switch between Zendesk instances.",
	}

	cmd.AddCommand(newInstanceAddCommand())
	cmd.AddCommand(newInstanceListCommand())
	cmd.AddCommand(newInstanceSwitchCommand())
	cmd.AddCommand(newInstanceRemoveCommand())
	cmd.AddCommand(newInstanceCurrentCommand())

	return cmd
}

func newInstanceAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new Zendesk instance",
		RunE:  runAddInstance,
	}
}

func runAddInstance(cmd *cobra.Command, args []string) error {
	// Load existing config
	cfg, err := config.LoadOrCreate()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Prompt for instance details
	instance, err := promptForInstance("")
	if err != nil {
		return err
	}

	// Check if instance already exists
	if _, exists := cfg.Instances[instance.Name]; exists {
		return fmt.Errorf("instance '%s' already exists", instance.Name)
	}

	// Add instance and make it current
	if err := cfg.AddInstanceAndSwitch(instance); err != nil {
		return fmt.Errorf("failed to add instance: %w", err)
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("✓ Instance '%s' added successfully!\n", instance.Name)
	color.White("Instance '%s' is now active.\n", instance.Name)

	return nil
}

func newInstanceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured instances",
		RunE:  runInstanceList,
	}
}

func runInstanceList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err == config.ErrConfigNotFound {
		color.Yellow("No instances configured. Run 'zd init' to get started.\n")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.Instances) == 0 {
		color.Yellow("No instances configured. Run 'zd init' to get started.\n")
		return nil
	}

	// Print header
	fmt.Printf("%-3s %-20s %-30s %-10s %s\n", "", "NAME", "SUBDOMAIN", "AUTH TYPE", "EMAIL")
	fmt.Println(strings.Repeat("-", 80))

	// Print instances
	for name, instance := range cfg.Instances {
		current := " "
		if name == cfg.Current {
			current = "*"
		}

		email := instance.Email
		if instance.AuthType == config.AuthTypeOAuth {
			email = "(OAuth)"
		}

		fmt.Printf("%-3s %-20s %-30s %-10s %s\n",
			current,
			name,
			instance.Subdomain+".zendesk.com",
			string(instance.AuthType),
			email,
		)
	}

	return nil
}

func newInstanceSwitchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a different instance",
		Args:  cobra.ExactArgs(1),
		RunE:  runInstanceSwitch,
	}
}

func runInstanceSwitch(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.SwitchInstance(name); err != nil {
		return fmt.Errorf("failed to switch instance: %w", err)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("✓ Switched to instance '%s'\n", name)

	return nil
}

func newInstanceRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an instance",
		Args:  cobra.ExactArgs(1),
		RunE:  runInstanceRemove,
	}
}

func runInstanceRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.RemoveInstance(name); err != nil {
		return fmt.Errorf("failed to remove instance: %w", err)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	color.Green("✓ Instance '%s' removed\n", name)

	if cfg.Current != "" {
		color.White("Current instance is now '%s'\n", cfg.Current)
	}

	return nil
}

func newInstanceCurrentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show the current active instance",
		RunE:  runInstanceCurrent,
	}
}

func runInstanceCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.Current == "" {
		color.Yellow("No current instance set\n")
		return nil
	}

	instance, ok := cfg.Instances[cfg.Current]
	if !ok {
		return fmt.Errorf("current instance '%s' not found in configuration", cfg.Current)
	}

	color.Cyan("Current instance: %s\n", cfg.Current)
	color.White("  Subdomain: %s.zendesk.com\n", instance.Subdomain)
	color.White("  Auth Type: %s\n", instance.AuthType)
	if instance.Email != "" {
		color.White("  Email: %s\n", instance.Email)
	}

	return nil
}
