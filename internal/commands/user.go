package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zd-cli/internal/client"
	"zd-cli/internal/config"
	"zd-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewUserCommand creates the user management command
func NewUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Zendesk users",
		Long:  "View and search Zendesk users.",
	}

	cmd.AddCommand(newUserMeCommand())
	cmd.AddCommand(newUserListCommand())
	cmd.AddCommand(newUserSearchCommand())
	cmd.AddCommand(newUserShowCommand())
	cmd.AddCommand(newUserCreateCommand())
	cmd.AddCommand(newUserUpdateCommand())
	cmd.AddCommand(newUserSuspendCommand())
	cmd.AddCommand(newUserUnsuspendCommand())
	cmd.AddCommand(newUserDeleteCommand())

	// Add global output format flag to all subcommands
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, csv")

	return cmd
}

func newUserMeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Show current authenticated user",
		RunE:  runUserMe,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newUserListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE:  runUserList,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newUserSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search users by name or email",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runUserSearch,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newUserShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <user-id>",
		Short: "Show detailed information for a specific user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserShow,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func runUserMe(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	return outputUser(cmd, user, true)
}

func runUserList(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.ListUsers(ctx, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(resp.Users) == 0 {
		color.Yellow("No users found.\n")
		return nil
	}

	return outputUsers(cmd, resp.Users, page, resp.Count, resp.NextPage)
}

func runUserSearch(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	query := strings.Join(args, " ")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := zdClient.SearchUsers(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search users: %w", err)
	}

	if len(users) == 0 {
		color.Yellow("No users found matching '%s'.\n", query)
		return nil
	}

	return outputUsers(cmd, users, 0, len(users), "")
}

func runUserShow(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.GetUser(ctx, userID)
	if err != nil {
		// Use user-friendly error formatting
		return fmt.Errorf("%s", client.FormatUserFriendlyError(err))
	}

	return outputUser(cmd, user, true)
}

// Helper function to get client with cache option from flags
func getClientFromFlags(cmd *cobra.Command) (*client.Client, error) {
	cfg, err := config.Load()
	if err == config.ErrConfigNotFound {
		return nil, fmt.Errorf("no configuration found. Run 'zd init' to get started")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	instance, err := cfg.GetCurrentInstance()
	if err != nil {
		return nil, fmt.Errorf("no current instance set. Run 'zd instance switch <name>' to select an instance")
	}

	refresh, _ := cmd.Flags().GetBool("refresh")
	useCache := !refresh

	return client.NewClientWithCache(instance, useCache)
}

// Display a user summary (compact format)
func displayUserSummary(user *client.User, index int) {
	email := user.Email
	if email == "" {
		email = "(no email)"
	}

	// Build status badges
	badges := []string{}
	if user.Verified {
		badges = append(badges, color.GreenString("✓"))
	}
	if user.Suspended {
		badges = append(badges, color.RedString("suspended"))
	}
	if !user.Active {
		badges = append(badges, color.YellowString("inactive"))
	}

	statusStr := ""
	if len(badges) > 0 {
		statusStr = " | " + strings.Join(badges, " ")
	}

	// Single line format
	fmt.Printf("#%-3d %s | %s | %s | ID: %d%s\n",
		index,
		color.CyanString(user.Name),
		email,
		user.Role,
		user.ID,
		statusStr)
}

// User modification commands

func newUserCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		RunE:  runUserCreate,
	}

	cmd.Flags().String("name", "", "User name")
	cmd.Flags().String("email", "", "User email")
	cmd.Flags().String("role", "end-user", "User role: end-user, agent, admin")
	cmd.Flags().String("phone", "", "Phone number")

	return cmd
}

func newUserUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <user-id>",
		Short: "Update a user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserUpdate,
	}

	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("email", "", "New email")
	cmd.Flags().String("phone", "", "New phone")
	cmd.Flags().String("role", "", "New role: end-user, agent, admin")
	cmd.Flags().Bool("verified", false, "Mark as verified")

	return cmd
}

func newUserSuspendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend <user-id>",
		Short: "Suspend a user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserSuspend,
	}

	return cmd
}

func newUserUnsuspendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsuspend <user-id>",
		Short: "Unsuspend a user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserUnsuspend,
	}

	return cmd
}

func newUserDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <user-id>",
		Short: "Delete a user",
		Args:  cobra.ExactArgs(1),
		RunE:  runUserDelete,
	}

	cmd.Flags().Bool("force", false, "Skip confirmation prompt")

	return cmd
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// Get flags
	name, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	role, _ := cmd.Flags().GetString("role")
	phone, _ := cmd.Flags().GetString("phone")

	// Interactive prompts if not provided
	if name == "" {
		name, err = promptString("Name", true)
		if err != nil {
			return err
		}
	}

	if email == "" {
		email, err = promptString("Email", true)
		if err != nil {
			return err
		}
	}

	// Build request
	req := client.CreateUserRequest{
		Name:  name,
		Email: email,
		Role:  role,
		Phone: phone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.CreateUser(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	color.Green("✓ User created successfully!\n")
	color.White("User ID: %d\n", user.ID)
	color.White("Name: %s\n", user.Name)
	color.White("Email: %s\n", user.Email)
	color.White("Role: %s\n", user.Role)

	return nil
}

func runUserUpdate(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	// Build update request from flags
	req := client.UpdateUserRequest{}
	updated := false

	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		req.Name = &name
		updated = true
	}

	if cmd.Flags().Changed("email") {
		email, _ := cmd.Flags().GetString("email")
		req.Email = &email
		updated = true
	}

	if cmd.Flags().Changed("phone") {
		phone, _ := cmd.Flags().GetString("phone")
		req.Phone = &phone
		updated = true
	}

	if cmd.Flags().Changed("role") {
		role, _ := cmd.Flags().GetString("role")
		req.Role = &role
		updated = true
	}

	if cmd.Flags().Changed("verified") {
		verified, _ := cmd.Flags().GetBool("verified")
		req.Verified = &verified
		updated = true
	}

	if !updated {
		return fmt.Errorf("no updates specified. Use flags like --name, --email, --role, etc.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.UpdateUser(ctx, userID, req)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	color.Green("✓ User #%d updated successfully!\n", userID)
	displayUser(user, false)

	return nil
}

func runUserSuspend(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.SuspendUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	color.Green("✓ User #%d suspended\n", user.ID)
	color.White("Name: %s\n", user.Name)

	return nil
}

func runUserUnsuspend(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	user, err := zdClient.UnsuspendUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to unsuspend user: %w", err)
	}

	color.Green("✓ User #%d unsuspended\n", user.ID)
	color.White("Name: %s\n", user.Name)

	return nil
}

func runUserDelete(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[0])
	}

	// Confirmation unless --force
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		color.Yellow("WARNING: This will permanently delete user %d\n", userID)
		confirm, err := promptString("Type 'yes' to confirm", true)
		if err != nil {
			return err
		}
		if strings.ToLower(confirm) != "yes" {
			color.Yellow("Deletion cancelled.\n")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := zdClient.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	color.Green("✓ User #%d deleted\n", userID)

	return nil
}

// Display full user details
func displayUser(user *client.User, detailed bool) {
	color.Cyan("User: %s\n", user.Name)
	color.White(strings.Repeat("─", 80) + "\n")

	color.White("ID:           %d\n", user.ID)
	color.White("Email:        %s\n", user.Email)
	color.White("Role:         %s\n", user.Role)

	if user.Phone != "" {
		color.White("Phone:        %s\n", user.Phone)
	}

	if user.OrganizationID != nil {
		color.White("Org ID:       %d\n", *user.OrganizationID)
	}

	color.White("Time Zone:    %s\n", user.TimeZone)
	color.White("Locale:       %s\n", user.Locale)

	// Status
	color.White("\nStatus:\n")
	if user.Active {
		color.Green("  ✓ Active\n")
	} else {
		color.Red("  ✗ Inactive\n")
	}

	if user.Verified {
		color.Green("  ✓ Verified\n")
	} else {
		color.Yellow("  ○ Not verified\n")
	}

	if user.Suspended {
		color.Red("  ✗ Suspended\n")
	}

	if user.TwoFactorAuthEnabled {
		color.Green("  ✓ 2FA Enabled\n")
	}

	// Dates
	color.White("\nDates:\n")
	color.White("  Created:      %s\n", formatDate(user.CreatedAt))
	color.White("  Last Updated: %s\n", formatDate(user.UpdatedAt))
	if user.LastLoginAt != nil && *user.LastLoginAt != "" {
		color.White("  Last Login:   %s\n", formatDate(*user.LastLoginAt))
	}

	// Additional details
	if detailed {
		if user.Alias != "" {
			color.White("\nAlias:        %s\n", user.Alias)
		}

		if len(user.Tags) > 0 {
			color.White("\nTags:         %s\n", strings.Join(user.Tags, ", "))
		}

		if user.Notes != "" {
			color.White("\nNotes:\n%s\n", user.Notes)
		}

		if user.Details != "" {
			color.White("\nDetails:\n%s\n", user.Details)
		}
	}

	color.White("\nURL:          %s\n", user.URL)
}

// Format ISO 8601 date to readable format
func formatDate(dateStr string) string {
	if dateStr == "" {
		return "N/A"
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("2006-01-02 15:04:05 MST")
}

// outputUser outputs a single user in the requested format
func outputUser(cmd *cobra.Command, user *client.User, detailed bool) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(user)

	case output.FormatCSV:
		headers := []string{"id", "name", "email", "role", "active", "verified", "suspended", "created_at", "updated_at"}
		return writer.WriteCSV(user, headers)

	default:
		// Table format (default)
		displayUser(user, detailed)
		return nil
	}
}

// outputUsers outputs multiple users in the requested format
func outputUsers(cmd *cobra.Command, users []client.User, page, total int, nextPage string) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(users)

	case output.FormatCSV:
		headers := []string{"id", "name", "email", "role", "active", "verified", "suspended", "organization_id", "phone", "time_zone", "created_at", "updated_at"}
		return writer.WriteCSV(users, headers)

	default:
		// Table format (default)
		if page > 0 {
			color.Cyan("Users (Page %d, showing %d of %d total)\n", page, len(users), total)
		} else {
			color.Cyan("Found %d user(s)\n", len(users))
		}
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, user := range users {
			displayUserSummary(&user, i+1)
		}

		// Show pagination info
		if nextPage != "" {
			fmt.Println()
			color.White("More results available. Use --page %d to see next page.\n", page+1)
		}

		return nil
	}
}
