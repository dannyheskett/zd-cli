package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"zd-cli/internal/client"
	"zd-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewGroupCommand creates the group management command
func NewGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage Zendesk groups",
		Long:  "View Zendesk groups and their members.",
	}

	cmd.AddCommand(newGroupListCommand())
	cmd.AddCommand(newGroupShowCommand())
	cmd.AddCommand(newGroupUsersCommand())
	cmd.AddCommand(newGroupMembershipsCommand())

	// Add global output format flag to all subcommands
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, csv")

	return cmd
}

func newGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all groups",
		RunE:  runGroupList,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newGroupShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <group-id>",
		Short: "Show detailed information for a specific group",
		Args:  cobra.ExactArgs(1),
		RunE:  runGroupShow,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newGroupUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users <group-id>",
		Short: "List users in a group",
		Args:  cobra.ExactArgs(1),
		RunE:  runGroupUsers,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newGroupMembershipsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memberships <group-id>",
		Short: "List memberships for a group",
		Args:  cobra.ExactArgs(1),
		RunE:  runGroupMemberships,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func runGroupList(cmd *cobra.Command, args []string) error {
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

	resp, err := zdClient.ListGroups(ctx, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to list groups: %w", err)
	}

	if len(resp.Groups) == 0 {
		color.Yellow("No groups found.\n")
		return nil
	}

	return outputGroups(cmd, resp.Groups, page, resp.Count, resp.NextPage)
}

func runGroupShow(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	groupID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid group ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	group, err := zdClient.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to get group: %w", err)
	}

	return outputGroup(cmd, group, true)
}

func runGroupUsers(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	groupID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid group ID: %s", args[0])
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.GetGroupUsers(ctx, groupID, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to get group users: %w", err)
	}

	if len(resp.Users) == 0 {
		color.Yellow("No users found in group %d.\n", groupID)
		return nil
	}

	return outputUsers(cmd, resp.Users, page, resp.Count, resp.NextPage)
}

func runGroupMemberships(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	groupID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid group ID: %s", args[0])
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.GetGroupMemberships(ctx, groupID, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to get group memberships: %w", err)
	}

	if len(resp.GroupMemberships) == 0 {
		color.Yellow("No memberships found for group %d.\n", groupID)
		return nil
	}

	return outputMemberships(cmd, resp.GroupMemberships, page, resp.Count, resp.NextPage)
}

// outputGroup outputs a single group in the requested format
func outputGroup(cmd *cobra.Command, group *client.Group, detailed bool) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(group)

	case output.FormatCSV:
		headers := []string{"id", "name", "description", "default", "deleted", "created_at", "updated_at"}
		return writer.WriteCSV(group, headers)

	default:
		// Table format (default)
		displayGroup(group, detailed)
		return nil
	}
}

// outputGroups outputs multiple groups in the requested format
func outputGroups(cmd *cobra.Command, groups []client.Group, page, total int, nextPage string) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(groups)

	case output.FormatCSV:
		headers := []string{"id", "name", "description", "default", "deleted", "created_at", "updated_at"}
		return writer.WriteCSV(groups, headers)

	default:
		// Table format (default)
		if page > 0 {
			color.Cyan("Groups (Page %d, showing %d of %d total)\n", page, len(groups), total)
		} else {
			color.Cyan("Found %d group(s)\n", len(groups))
		}
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, group := range groups {
			displayGroupSummary(&group, i+1)
		}

		// Show pagination info
		if nextPage != "" {
			fmt.Println()
			color.White("More results available. Use --page %d to see next page.\n", page+1)
		}

		return nil
	}
}

// outputMemberships outputs memberships in the requested format
func outputMemberships(cmd *cobra.Command, memberships []client.GroupMembership, page, total int, nextPage string) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(memberships)

	case output.FormatCSV:
		headers := []string{"id", "user_id", "group_id", "default", "created_at", "updated_at"}
		return writer.WriteCSV(memberships, headers)

	default:
		// Table format (default)
		if page > 0 {
			color.Cyan("Group Memberships (Page %d, showing %d of %d total)\n", page, len(memberships), total)
		} else {
			color.Cyan("Found %d membership(s)\n", len(memberships))
		}
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, membership := range memberships {
			displayMembershipSummary(&membership, i+1)
		}

		// Show pagination info
		if nextPage != "" {
			fmt.Println()
			color.White("More results available. Use --page %d to see next page.\n", page+1)
		}

		return nil
	}
}

// Display a group summary (compact format)
func displayGroupSummary(group *client.Group, index int) {
	defaultBadge := ""
	if group.Default {
		defaultBadge = " | " + color.GreenString("default")
	}
	deletedBadge := ""
	if group.Deleted {
		deletedBadge = " | " + color.RedString("deleted")
	}

	fmt.Printf("#%-3d %s | ID: %d%s%s\n",
		index,
		color.CyanString(group.Name),
		group.ID,
		defaultBadge,
		deletedBadge)
}

// Display full group details
func displayGroup(group *client.Group, detailed bool) {
	color.Cyan("Group: %s\n", group.Name)
	color.White(strings.Repeat("─", 80) + "\n")

	color.White("ID:           %d\n", group.ID)

	if group.Description != "" {
		color.White("Description:  %s\n", group.Description)
	}

	// Status
	color.White("\nStatus:\n")
	if group.Default {
		color.Green("  ✓ Default Group\n")
	}
	if group.Deleted {
		color.Red("  ✗ Deleted\n")
	} else {
		color.Green("  ✓ Active\n")
	}

	// Dates
	color.White("\nDates:\n")
	color.White("  Created:      %s\n", formatDate(group.CreatedAt))
	color.White("  Last Updated: %s\n", formatDate(group.UpdatedAt))

	color.White("\nURL: %s\n", group.URL)
}

// Display a membership summary
func displayMembershipSummary(membership *client.GroupMembership, index int) {
	defaultBadge := ""
	if membership.Default {
		defaultBadge = " | " + color.GreenString("default")
	}

	fmt.Printf("#%-3d User ID: %d%s\n",
		index,
		membership.UserID,
		defaultBadge)
}
