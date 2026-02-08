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
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// NewTicketCommand creates the ticket management command
func NewTicketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "Manage Zendesk tickets",
		Long:  "View, search, and manage Zendesk tickets.",
	}

	cmd.AddCommand(newTicketListCommand())
	cmd.AddCommand(newTicketShowCommand())
	cmd.AddCommand(newTicketCommentsCommand())
	cmd.AddCommand(newTicketSearchCommand())
	cmd.AddCommand(newTicketCreateCommand())
	cmd.AddCommand(newTicketUpdateCommand())
	cmd.AddCommand(newTicketCommentCommand())
	cmd.AddCommand(newTicketAssignCommand())
	cmd.AddCommand(newTicketCloseCommand())

	// Add global output format flag to all subcommands
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, csv")

	return cmd
}

func newTicketListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets",
		RunE:  runTicketList,
	}

	cmd.Flags().Int("page", 1, "Page number")
	cmd.Flags().Int("per-page", 30, "Results per page (max 100)")
	cmd.Flags().String("status", "", "Filter by status: new, open, pending, hold, solved, closed")
	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newTicketShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <ticket-id>",
		Short: "Show detailed information for a specific ticket",
		Args:  cobra.ExactArgs(1),
		RunE:  runTicketShow,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newTicketCommentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comments <ticket-id>",
		Short: "Show comments/conversation for a ticket",
		Args:  cobra.ExactArgs(1),
		RunE:  runTicketComments,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func newTicketSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search tickets by keyword",
		Long: `Search tickets by keyword. Examples:
  zd ticket search "login issue"
  zd ticket search "status:open priority:urgent"
  zd ticket search "assignee:me"`,
		Args: cobra.MinimumNArgs(1),
		RunE: runTicketSearch,
	}

	cmd.Flags().Bool("refresh", false, "Bypass cache and fetch fresh data")

	return cmd
}

func runTicketList(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	status, _ := cmd.Flags().GetString("status")

	if perPage > 100 {
		perPage = 100
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := zdClient.ListTickets(ctx, page, perPage, status)
	if err != nil {
		return fmt.Errorf("failed to list tickets: %w", err)
	}

	if len(resp.Tickets) == 0 {
		color.Yellow("No tickets found.\n")
		return nil
	}

	return outputTickets(cmd, resp.Tickets, page, resp.Count, resp.NextPage)
}

func runTicketShow(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.GetTicket(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	return outputTicket(cmd, ticket, true)
}

func runTicketComments(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	comments, err := zdClient.GetTicketComments(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to get ticket comments: %w", err)
	}

	if len(comments) == 0 {
		color.Yellow("No comments found for ticket %d.\n", ticketID)
		return nil
	}

	return outputComments(cmd, comments, ticketID)
}

func runTicketSearch(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	query := strings.Join(args, " ")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tickets, err := zdClient.SearchTickets(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search tickets: %w", err)
	}

	if len(tickets) == 0 {
		color.Yellow("No tickets found matching '%s'.\n", query)
		return nil
	}

	return outputTickets(cmd, tickets, 0, len(tickets), "")
}

// outputTicket outputs a single ticket in the requested format
func outputTicket(cmd *cobra.Command, ticket *client.Ticket, detailed bool) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(ticket)

	case output.FormatCSV:
		headers := []string{"id", "subject", "status", "priority", "requester_id", "assignee_id", "created_at", "updated_at"}
		return writer.WriteCSV(ticket, headers)

	default:
		// Table format (default)
		displayTicket(ticket, detailed)
		return nil
	}
}

// outputTickets outputs multiple tickets in the requested format
func outputTickets(cmd *cobra.Command, tickets []client.Ticket, page, total int, nextPage string) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(tickets)

	case output.FormatCSV:
		headers := []string{"id", "subject", "status", "priority", "type", "requester_id", "assignee_id", "group_id", "organization_id", "created_at", "updated_at"}
		return writer.WriteCSV(tickets, headers)

	default:
		// Table format (default)
		if page > 0 {
			color.Cyan("Tickets (Page %d, showing %d of %d total)\n", page, len(tickets), total)
		} else {
			color.Cyan("Found %d ticket(s)\n", len(tickets))
		}
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, ticket := range tickets {
			displayTicketSummary(&ticket, i+1)
		}

		// Show pagination info
		if nextPage != "" {
			fmt.Println()
			color.White("More results available. Use --page %d to see next page.\n", page+1)
		}

		return nil
	}
}

// outputComments outputs comments in the requested format
func outputComments(cmd *cobra.Command, comments []client.Comment, ticketID int64) error {
	format, _ := cmd.Flags().GetString("output")
	writer := output.NewWriter(output.Format(format))

	switch output.Format(format) {
	case output.FormatJSON:
		return writer.WriteJSON(comments)

	case output.FormatCSV:
		headers := []string{"id", "author_id", "body", "public", "created_at"}
		return writer.WriteCSV(comments, headers)

	default:
		// Table format (default)
		color.Cyan("Comments for Ticket #%d (%d total)\n", ticketID, len(comments))
		color.White(strings.Repeat("─", 80) + "\n\n")

		for i, comment := range comments {
			displayComment(&comment, i+1)
		}

		return nil
	}
}

// Display a ticket summary (compact format)
func displayTicketSummary(ticket *client.Ticket, index int) {
	// Status color
	statusColor := color.WhiteString
	switch ticket.Status {
	case "new":
		statusColor = color.CyanString
	case "open":
		statusColor = color.BlueString
	case "pending":
		statusColor = color.YellowString
	case "solved":
		statusColor = color.GreenString
	case "closed":
		statusColor = color.HiBlackString
	}

	// Priority indicator
	priorityIndicator := ""
	switch ticket.Priority {
	case "urgent":
		priorityIndicator = color.RedString("!")
	case "high":
		priorityIndicator = color.YellowString("↑")
	}

	fmt.Printf("#%-4d %s%-8s %s| %s | ID: %d\n",
		index,
		priorityIndicator,
		statusColor(ticket.Status),
		color.WhiteString("| "),
		ticket.Subject,
		ticket.ID)
}

// Display full ticket details
func displayTicket(ticket *client.Ticket, detailed bool) {
	color.Cyan("Ticket #%d: %s\n", ticket.ID, ticket.Subject)
	color.White(strings.Repeat("─", 80) + "\n")

	// Status and Priority
	fmt.Printf("Status:       %s\n", getColoredStatus(ticket.Status))
	fmt.Printf("Priority:     %s\n", getColoredPriority(ticket.Priority))
	fmt.Printf("Type:         %s\n", ticket.Type)

	// People
	color.White("\nPeople:\n")
	color.White("  Requester ID: %d\n", ticket.RequesterID)
	color.White("  Submitter ID: %d\n", ticket.SubmitterID)
	if ticket.AssigneeID != nil {
		color.White("  Assignee ID:  %d\n", *ticket.AssigneeID)
	} else {
		color.White("  Assignee ID:  (unassigned)\n")
	}

	// Organization and Group
	if ticket.OrganizationID != nil {
		color.White("  Organization: %d\n", *ticket.OrganizationID)
	}
	if ticket.GroupID != nil {
		color.White("  Group:        %d\n", *ticket.GroupID)
	}

	// Dates
	color.White("\nDates:\n")
	color.White("  Created:      %s\n", formatDate(ticket.CreatedAt))
	color.White("  Updated:      %s\n", formatDate(ticket.UpdatedAt))
	if ticket.DueAt != nil && *ticket.DueAt != "" {
		color.White("  Due:          %s\n", formatDate(*ticket.DueAt))
	}

	// Tags
	if len(ticket.Tags) > 0 {
		color.White("\nTags: %s\n", strings.Join(ticket.Tags, ", "))
	}

	// Description
	if detailed && ticket.Description != "" {
		color.White("\nDescription:\n")
		color.White("%s\n", ticket.Description)
	}

	color.White("\nURL: %s\n", ticket.URL)
}

// Display a comment
func displayComment(comment *client.Comment, index int) {
	visibility := "Public"
	if !comment.Public {
		visibility = color.YellowString("Private")
	}

	color.White("#%-3d [%s] Author ID: %d | %s\n", index, visibility, comment.AuthorID, formatDate(comment.CreatedAt))

	// Use plain body if available, otherwise HTML body, otherwise regular body
	body := comment.PlainBody
	if body == "" {
		body = comment.Body
	}

	// Truncate long comments for list view
	if len(body) > 200 {
		body = body[:200] + "..."
	}

	color.White("%s\n\n", body)
}

func getColoredStatus(status string) string {
	switch status {
	case "new":
		return color.CyanString(status)
	case "open":
		return color.BlueString(status)
	case "pending":
		return color.YellowString(status)
	case "solved":
		return color.GreenString(status)
	case "closed":
		return color.HiBlackString(status)
	default:
		return status
	}
}

func getColoredPriority(priority string) string {
	switch priority {
	case "urgent":
		return color.RedString(priority)
	case "high":
		return color.YellowString(priority)
	case "normal":
		return color.WhiteString(priority)
	case "low":
		return color.HiBlackString(priority)
	default:
		return priority
	}
}

// Ticket modification commands

func newTicketCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ticket",
		RunE:  runTicketCreate,
	}

	cmd.Flags().String("subject", "", "Ticket subject")
	cmd.Flags().String("description", "", "Ticket description")
	cmd.Flags().String("priority", "", "Priority: low, normal, high, urgent")
	cmd.Flags().String("type", "incident", "Type: problem, incident, question, task")
	cmd.Flags().String("status", "new", "Status: new, open, pending, hold, solved, closed")
	cmd.Flags().Int64("assignee", 0, "Assignee user ID")
	cmd.Flags().Int64("group", 0, "Group ID")
	cmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")

	return cmd
}

func newTicketUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <ticket-id>",
		Short: "Update a ticket",
		Args:  cobra.ExactArgs(1),
		RunE:  runTicketUpdate,
	}

	cmd.Flags().String("subject", "", "New subject")
	cmd.Flags().String("priority", "", "New priority: low, normal, high, urgent")
	cmd.Flags().String("status", "", "New status: new, open, pending, hold, solved, closed")
	cmd.Flags().Int64("assignee", 0, "New assignee user ID")
	cmd.Flags().Int64("group", 0, "New group ID")
	cmd.Flags().StringSlice("tags", []string{}, "Tags to set")

	return cmd
}

func newTicketCommentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment <ticket-id>",
		Short: "Add a comment to a ticket",
		Args:  cobra.ExactArgs(1),
		RunE:  runTicketComment,
	}

	cmd.Flags().String("message", "", "Comment message")
	cmd.Flags().Bool("public", true, "Make comment public")
	cmd.Flags().Bool("private", false, "Make comment private")

	return cmd
}

func newTicketAssignCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign <ticket-id> <user-id>",
		Short: "Assign a ticket to a user",
		Args:  cobra.ExactArgs(2),
		RunE:  runTicketAssign,
	}

	return cmd
}

func newTicketCloseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close <ticket-id>",
		Short: "Close a ticket",
		Args:  cobra.ExactArgs(1),
		RunE:  runTicketClose,
	}

	cmd.Flags().String("comment", "", "Optional closing comment")

	return cmd
}

func runTicketCreate(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	// Get flags
	subject, _ := cmd.Flags().GetString("subject")
	description, _ := cmd.Flags().GetString("description")
	priority, _ := cmd.Flags().GetString("priority")
	ticketType, _ := cmd.Flags().GetString("type")
	status, _ := cmd.Flags().GetString("status")
	assigneeID, _ := cmd.Flags().GetInt64("assignee")
	groupID, _ := cmd.Flags().GetInt64("group")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	// Interactive prompts if not provided
	if subject == "" {
		subject, err = promptString("Subject", true)
		if err != nil {
			return err
		}
	}

	if description == "" {
		description, err = promptString("Description", true)
		if err != nil {
			return err
		}
	}

	// Build request
	req := client.CreateTicketRequest{
		Subject:     subject,
		Description: description,
		Priority:    priority,
		Type:        ticketType,
		Status:      status,
		Tags:        tags,
	}

	if assigneeID > 0 {
		req.AssigneeID = &assigneeID
	}
	if groupID > 0 {
		req.GroupID = &groupID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.CreateTicket(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	color.Green("✓ Ticket created successfully!\n")
	color.White("Ticket ID: %d\n", ticket.ID)
	color.White("Status: %s\n", ticket.Status)
	color.White("URL: %s\n", ticket.URL)

	return nil
}

func runTicketUpdate(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	// Build update request from flags
	req := client.UpdateTicketRequest{}
	updated := false

	if cmd.Flags().Changed("subject") {
		subject, _ := cmd.Flags().GetString("subject")
		req.Subject = &subject
		updated = true
	}

	if cmd.Flags().Changed("priority") {
		priority, _ := cmd.Flags().GetString("priority")
		req.Priority = &priority
		updated = true
	}

	if cmd.Flags().Changed("status") {
		status, _ := cmd.Flags().GetString("status")
		req.Status = &status
		updated = true
	}

	if cmd.Flags().Changed("assignee") {
		assigneeID, _ := cmd.Flags().GetInt64("assignee")
		req.AssigneeID = &assigneeID
		updated = true
	}

	if cmd.Flags().Changed("group") {
		groupID, _ := cmd.Flags().GetInt64("group")
		req.GroupID = &groupID
		updated = true
	}

	if cmd.Flags().Changed("tags") {
		tags, _ := cmd.Flags().GetStringSlice("tags")
		req.Tags = tags
		updated = true
	}

	if !updated {
		return fmt.Errorf("no updates specified. Use flags like --status, --priority, --assignee, etc.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.UpdateTicket(ctx, ticketID, req)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	color.Green("✓ Ticket #%d updated successfully!\n", ticketID)
	displayTicket(ticket, false)

	return nil
}

func runTicketComment(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	message, _ := cmd.Flags().GetString("message")
	if message == "" {
		message, err = promptString("Comment", true)
		if err != nil {
			return err
		}
	}

	isPublic := true
	if cmd.Flags().Changed("private") {
		private, _ := cmd.Flags().GetBool("private")
		isPublic = !private
	}

	// Create update request with just a comment
	req := client.UpdateTicketRequest{}
	req.Comment = &struct {
		Body   string `json:"body"`
		Public bool   `json:"public"`
	}{
		Body:   message,
		Public: isPublic,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.UpdateTicket(ctx, ticketID, req)
	if err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	visibility := "public"
	if !isPublic {
		visibility = "private"
	}

	color.Green("✓ Added %s comment to ticket #%d\n", visibility, ticket.ID)

	return nil
}

func runTicketAssign(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	assigneeID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %s", args[1])
	}

	req := client.UpdateTicketRequest{
		AssigneeID: &assigneeID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.UpdateTicket(ctx, ticketID, req)
	if err != nil {
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	color.Green("✓ Ticket #%d assigned to user %d\n", ticket.ID, assigneeID)

	return nil
}

func runTicketClose(cmd *cobra.Command, args []string) error {
	zdClient, err := getClientFromFlags(cmd)
	if err != nil {
		return err
	}

	ticketID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", args[0])
	}

	closedStatus := "closed"
	req := client.UpdateTicketRequest{
		Status: &closedStatus,
	}

	// Add closing comment if provided
	if cmd.Flags().Changed("comment") {
		message, _ := cmd.Flags().GetString("comment")
		req.Comment = &struct {
			Body   string `json:"body"`
			Public bool   `json:"public"`
		}{
			Body:   message,
			Public: true,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticket, err := zdClient.UpdateTicket(ctx, ticketID, req)
	if err != nil {
		return fmt.Errorf("failed to close ticket: %w", err)
	}

	color.Green("✓ Ticket #%d closed\n", ticket.ID)

	return nil
}

// Helper function for interactive string prompts
func promptString(label string, required bool) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}

	if required {
		prompt.Validate = func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("%s cannot be empty", label)
			}
			return nil
		}
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}
