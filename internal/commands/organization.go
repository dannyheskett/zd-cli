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

// NewOrganizationCommand creates the organization management command
func NewOrganizationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "org",
		Aliases: []string{"organization"},
		Short:   "Manage Zendesk organizations",
		Long:    "View and search Zendesk organizations.",
	}

	cmd.AddCommand(newOrgListCommand())
	cmd.AddCommand(newOrgShowCommand())
	cmd.AddCommand(newOrgSearchCommand())
	cmd.AddCommand(newOrgUsersCommand())
	cmd.AddCommand(newOrgTicketsCommand())

	// Add global output format flag to all subcommands
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, csv")

	return cmd
}

func newOrgListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all organizations",
		RunE:  runOrgList,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newOrgShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <org-id>",
		Short: "Show detailed information for a specific organization",
		Args:  cobra.ExactArgs(1),
		RunE:  runOrgShow,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newOrgSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search organizations by name",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runOrgSearch,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newOrgUsersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users <org-id>",
		Short: "List users in an organization",
		Args:  cobra.ExactArgs(1),
		RunE:  runOrgUsers,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newOrgTicketsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tickets <org-id>",
		Short: "List tickets for an organization",
		Args:  cobra.ExactArgs(1),
		RunE:  runOrgTickets,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 100, "Results per page (max 100)")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func runOrgList(cmd *cobra.Command, args []string) error {
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

	resp, err := zdClient.ListOrganizations(ctx, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to list organizations: %w", err)
	}

	if len(resp.Organizations) == 0 {
		color.Yellow("No organizations found.\n")
		return nil
	}

	return outputOrganizations(cmd, resp.Organizations, page, resp.Count, resp.NextPage)
}

func runOrgShow(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	orgID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	org, err := zdClient.GetOrganization(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	return outputOrganization(cmd, org, true)
}

func runOrgSearch(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	query := strings.Join(args, " ")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgs, err := zdClient.SearchOrganizations(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search organizations: %w", err)
	}

	if len(orgs) == 0 {
		color.Yellow("No organizations found matching '%s'.\n", query)
		return nil
	}

	return outputOrganizations(cmd, orgs, 0, len(orgs), "")
}

func runOrgUsers(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	orgID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %s", args[0])
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.GetOrganizationUsers(ctx, orgID, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to get organization users: %w", err)
	}

	if len(resp.Users) == 0 {
		color.Yellow("No users found in organization %d.\n", orgID)
		return nil
	}

	return outputUsers(cmd, resp.Users, page, resp.Count, resp.NextPage)
}

func runOrgTickets(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	orgID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid organization ID: %s", args[0])
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.GetOrganizationTickets(ctx, orgID, page, perPage)
	if err != nil {
		return fmt.Errorf("failed to get organization tickets: %w", err)
	}

	if len(resp.Tickets) == 0 {
		color.Yellow("No tickets found for organization %d.\n", orgID)
		return nil
	}

	return outputTickets(cmd, resp.Tickets, page, resp.Count, resp.NextPage)
}

// outputOrganization outputs a single organization in the requested format
func outputOrganization(cmd *cobra.Command, org *client.Organization, detailed bool) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(org)

	case output.FormatCSV:
		headers := []string{"id", "name", "created_at", "updated_at", "shared_tickets", "shared_comments"}
		return writer.WriteCSV(org, headers)

	default:
		// Table format (default)
		displayOrganization(org, detailed)
		return nil
	}
}

// outputOrganizations outputs multiple organizations in the requested format
func outputOrganizations(cmd *cobra.Command, orgs []client.Organization, page, total int, nextPage string) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(orgs)

	case output.FormatCSV:
		headers := []string{"id", "name", "created_at", "updated_at", "shared_tickets", "shared_comments", "group_id"}
		return writer.WriteCSV(orgs, headers)

	default:
		// Table format (default)
		if page > 0 {
			color.Cyan("Organizations (Page %d, showing %d of %d total)\n", page, len(orgs), total)
		} else {
			color.Cyan("Found %d organization(s)\n", len(orgs))
		}
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, org := range orgs {
			displayOrganizationSummary(&org, i+1)
		}

		// Show pagination info
		if nextPage != "" {
			fmt.Println()
			color.White("More results available. Use --page %d to see next page.\n", page+1)
		}

		return nil
	}
}

// Display an organization summary (compact format)
func displayOrganizationSummary(org *client.Organization, index int) {
	sharedInfo := ""
	if org.SharedTickets {
		sharedInfo += " | shared tickets"
	}
	if org.SharedComments {
		sharedInfo += " | shared comments"
	}

	fmt.Printf("#%-3d %s | ID: %d%s\n",
		index,
		color.CyanString(org.Name),
		org.ID,
		sharedInfo)
}

// Display full organization details
func displayOrganization(org *client.Organization, detailed bool) {
	color.Cyan("Organization: %s\n", org.Name)
	color.White(strings.Repeat("─", 80) + "\n")

	color.White("ID:           %d\n", org.ID)

	if len(org.DomainNames) > 0 {
		color.White("Domains:      %s\n", strings.Join(org.DomainNames, ", "))
	}

	if org.GroupID != nil {
		color.White("Group ID:     %d\n", *org.GroupID)
	}

	color.White("\nSharing:\n")
	if org.SharedTickets {
		color.Green("  ✓ Shared Tickets\n")
	} else {
		color.White("  ○ Private Tickets\n")
	}

	if org.SharedComments {
		color.Green("  ✓ Shared Comments\n")
	} else {
		color.White("  ○ Private Comments\n")
	}

	// Dates
	color.White("\nDates:\n")
	color.White("  Created:      %s\n", formatDate(org.CreatedAt))
	color.White("  Last Updated: %s\n", formatDate(org.UpdatedAt))

	// Tags
	if len(org.Tags) > 0 {
		color.White("\nTags: %s\n", strings.Join(org.Tags, ", "))
	}

	// Additional details
	if detailed {
		if org.Details != "" {
			color.White("\nDetails:\n%s\n", org.Details)
		}

		if org.Notes != "" {
			color.White("\nNotes:\n%s\n", org.Notes)
		}
	}

	color.White("\nURL: %s\n", org.URL)
}
