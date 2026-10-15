package commands

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var ticketsCmd = &cobra.Command{
	Use:   "tickets",
	Short: "Manage tickets",
}

const (
	defaultTicketSort  = "created_at"
	defaultTicketOrder = "desc"
)

var (
	validTicketSorts  = []string{"created_at", "updated_at"}
	validTicketOrders = []string{"asc", "desc"}
)

func validateTicketSort(v string) error {
	if v == "" || slices.Contains(validTicketSorts, v) {
		return nil
	}
	return fmt.Errorf("invalid sort %q: must be one of %s", v, strings.Join(validTicketSorts, ", "))
}

func validateTicketOrder(v string) error {
	if v == "" || slices.Contains(validTicketOrders, v) {
		return nil
	}
	return fmt.Errorf("invalid order %q: must be one of %s", v, strings.Join(validTicketOrders, ", "))
}

func validateTicketFieldFilter(name, value string) error {
	if (name == "") == (value == "") {
		return nil
	}
	if name == "" {
		return fmt.Errorf("--field-name is required when --field-value is set")
	}
	return fmt.Errorf("--field-value is required when --field-name is set")
}

func parseNameValues(pairs []string) (map[string]interface{}, error) {
	values := map[string]interface{}{}
	for _, pair := range pairs {
		name, value, found := strings.Cut(pair, "=")
		if !found || name == "" {
			return nil, fmt.Errorf("--field must be Name=Value")
		}
		values[name] = value
	}
	return values, nil
}

var ticketExtraFlags = map[string]string{
	"group": "group", "sub-category-one": "sub_category_one",
	"sub-category-two": "sub_category_two", "resolution-due-date": "resolution_due_date",
}

func flagValues(cmd *cobra.Command, name string) []string {
	values, _ := cmd.Flags().GetStringArray(name)
	return values
}

func setStringFlags(cmd *cobra.Command, body map[string]interface{}, flagToKey map[string]string) {
	for flag, key := range flagToKey {
		if v, _ := cmd.Flags().GetString(flag); v != "" {
			body[key] = v
		}
	}
}

func fieldValues(cmd *cobra.Command) (map[string]interface{}, error) {
	return parseNameValues(flagValues(cmd, "field"))
}

func setFieldValues(body, values map[string]interface{}) {
	if len(values) > 0 {
		body["ticket_fields"] = values
	}
}

var ticketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickets",
	RunE: func(cmd *cobra.Command, args []string) error {
		sort, _ := cmd.Flags().GetString("sort")
		if err := validateTicketSort(sort); err != nil {
			return err
		}
		order, _ := cmd.Flags().GetString("order")
		if err := validateTicketOrder(order); err != nil {
			return err
		}
		fieldName, _ := cmd.Flags().GetString("field-name")
		fieldValue, _ := cmd.Flags().GetString("field-value")
		if err := validateTicketFieldFilter(fieldName, fieldValue); err != nil {
			return err
		}

		unassigned, _ := cmd.Flags().GetBool("unassigned")
		assigneeID, _ := cmd.Flags().GetString("assignee-id")
		assigneeEmail, _ := cmd.Flags().GetString("assignee-email")
		if unassigned && (assigneeID != "" || assigneeEmail != "") {
			return fmt.Errorf("--unassigned cannot be combined with --assignee-id or --assignee-email")
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if unassigned {
			params.Set("unassigned", "true")
		}
		for flag, key := range map[string]string{
			"priority": "priority", "customer-domain": "customer_domain",
			"query": "query", "external-id": "external_id",
		} {
			if v, _ := cmd.Flags().GetString(flag); v != "" {
				params.Set(key, v)
			}
		}
		if tags, _ := cmd.Flags().GetStringArray("tag"); len(tags) > 0 {
			params.Set("tags", strings.Join(tags, ","))
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			params.Set("status", status)
		}
		if v, _ := cmd.Flags().GetString("customer-id"); v != "" {
			params.Set("customer_id", v)
		}
		if v, _ := cmd.Flags().GetString("customer-email"); v != "" {
			params.Set("customer_email", v)
		}
		if v, _ := cmd.Flags().GetString("assignee-id"); v != "" {
			params.Set("assignee_id", v)
		}
		if v, _ := cmd.Flags().GetString("assignee-email"); v != "" {
			params.Set("assignee_email", v)
		}
		if fieldName != "" {
			params.Set("field_name", fieldName)
			params.Set("field_value", fieldValue)
		}
		params = mergeParams(params, dateRangeParams(cmd))
		if sort != "" {
			params.Set("sort", sort)
		}
		if order != "" {
			params.Set("order", order)
		}

		data, err := c.Get("/tickets", params)
		if err != nil {
			return err
		}

		printList(data, "tickets", []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk tickets show <ticket-number>"},
			{Label: "Update", Command: "neetodesk tickets update <ticket-number>"},
		})
		return nil
	},
}

var ticketsShowCmd = &cobra.Command{
	Use:   "show <ticket-number>",
	Short: "Show a ticket by number or id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/tickets/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Comments", Command: fmt.Sprintf("neetodesk tickets comments list %s", args[0])},
			{Label: "Update", Command: fmt.Sprintf("neetodesk tickets update %s", args[0])},
		})
		return nil
	},
}

var ticketsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a ticket",
	RunE: func(cmd *cobra.Command, args []string) error {
		ticketFields, err := fieldValues(cmd)
		if err != nil {
			return err
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("email"); v != "" {
			body["email"] = v
		}
		if v, _ := cmd.Flags().GetString("subject"); v != "" {
			body["subject"] = v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body["description"] = v
		}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			body["name"] = v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			body["status"] = v
		}
		if v, _ := cmd.Flags().GetString("priority"); v != "" {
			body["priority"] = v
		}
		if v, _ := cmd.Flags().GetString("category"); v != "" {
			body["category"] = v
		}
		if v, _ := cmd.Flags().GetString("agent-id"); v != "" {
			body["agent_id"] = v
		}
		if v, _ := cmd.Flags().GetString("group-id"); v != "" {
			body["group_id"] = v
		}
		if v, _ := cmd.Flags().GetString("assignee-email"); v != "" {
			body["assignee_email"] = v
		}
		setStringFlags(cmd, body, ticketExtraFlags)
		if tags, _ := cmd.Flags().GetStringArray("tag"); len(tags) > 0 {
			body["tags"] = tags
		}
		setFieldValues(body, ticketFields)

		attachments, err := attachmentSignedIDs(cmd, c)
		if err != nil {
			return err
		}
		if len(attachments) > 0 {
			body["attachments"] = attachments
		}

		data, err := c.Post("/tickets", body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk tickets show <ticket-number>"},
		})
		return nil
	},
}

var ticketsUpdateCmd = &cobra.Command{
	Use:   "update <ticket-number>",
	Short: "Update a ticket by number or id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		replacing := len(flagValues(cmd, "tag")) > 0
		adjusting := len(flagValues(cmd, "add-tag")) > 0 || len(flagValues(cmd, "remove-tag")) > 0
		if replacing && adjusting {
			return fmt.Errorf("--tag cannot be combined with --add-tag or --remove-tag")
		}
		ticketFields, err := fieldValues(cmd)
		if err != nil {
			return err
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("subject"); v != "" {
			body["subject"] = v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body["description"] = v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			body["status"] = v
		}
		if v, _ := cmd.Flags().GetString("priority"); v != "" {
			body["priority"] = v
		}
		if v, _ := cmd.Flags().GetString("category"); v != "" {
			body["category"] = v
		}
		if v, _ := cmd.Flags().GetString("agent-id"); v != "" {
			body["agent_id"] = v
		}
		if v, _ := cmd.Flags().GetString("group-id"); v != "" {
			body["group_id"] = v
		}
		if v, _ := cmd.Flags().GetString("assignee-email"); v != "" {
			body["assignee_email"] = v
		}
		setStringFlags(cmd, body, ticketExtraFlags)
		if tags, _ := cmd.Flags().GetStringArray("tag"); len(tags) > 0 {
			body["tags"] = tags
		}
		if tags, _ := cmd.Flags().GetStringArray("add-tag"); len(tags) > 0 {
			body["add_tags"] = tags
		}
		if tags, _ := cmd.Flags().GetStringArray("remove-tag"); len(tags) > 0 {
			body["remove_tags"] = tags
		}
		setFieldValues(body, ticketFields)

		attachments, err := attachmentSignedIDs(cmd, c)
		if err != nil {
			return err
		}
		if len(attachments) > 0 {
			body["attachments"] = attachments
		}

		data, err := c.Patch(fmt.Sprintf("/tickets/%s", args[0]), body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(ticketsListCmd)
	ticketsListCmd.Flags().String("status", "", `Filter by status (comma-separated, e.g. new,open). "open" includes new tickets`)
	ticketsListCmd.Flags().String("customer-id", "", "Filter by customer ID")
	ticketsListCmd.Flags().String("customer-email", "", "Filter by customer email")
	ticketsListCmd.Flags().String("assignee-id", "", "Filter by assignee ID")
	ticketsListCmd.Flags().String("assignee-email", "", "Filter by assignee email")
	ticketsListCmd.Flags().String("field-name", "", "Custom ticket field to filter by, used with --field-value")
	ticketsListCmd.Flags().String("field-value", "", "Value the custom ticket field must have, used with --field-name")
	ticketsListCmd.Flags().String("priority", "", "Filter by priority (comma-separated, e.g. urgent,high)")
	ticketsListCmd.Flags().Bool("unassigned", false, "Only tickets with no assigned agent")
	ticketsListCmd.Flags().String("customer-domain", "", "Filter by customer email domain (e.g. acme.com)")
	ticketsListCmd.Flags().StringArray("tag", []string{}, "Filter by tag, matches any (repeat flag for multiple)")
	ticketsListCmd.Flags().String("query", "", "Text to find in the subject")
	ticketsListCmd.Flags().String("external-id", "", "Filter by external ID")
	addDateRangeFlags(ticketsListCmd)
	ticketsListCmd.Flags().String("sort", defaultTicketSort, "Sort field: created_at or updated_at")
	ticketsListCmd.Flags().String("order", defaultTicketOrder, "Sort direction: asc or desc")

	ticketsCreateCmd.Flags().String("email", "", "Customer email")
	ticketsCreateCmd.Flags().String("subject", "", "Ticket subject")
	ticketsCreateCmd.Flags().String("description", "", "Ticket description")
	ticketsCreateCmd.Flags().String("name", "", "Customer name")
	ticketsCreateCmd.Flags().String("status", "", "Ticket status")
	ticketsCreateCmd.Flags().String("priority", "", "Ticket priority (low, medium, high, urgent)")
	ticketsCreateCmd.Flags().String("category", "", "Ticket category")
	ticketsCreateCmd.Flags().String("agent-id", "", "Agent ID to assign")
	ticketsCreateCmd.Flags().String("group-id", "", "Group ID to assign")
	ticketsCreateCmd.Flags().String("assignee-email", "", "Agent email to assign")
	addTicketExtraFlags(ticketsCreateCmd)
	ticketsCreateCmd.Flags().StringArray("tag", []string{}, "Tag to add (repeat flag for multiple)")
	addAttachFlag(ticketsCreateCmd)
	_ = ticketsCreateCmd.MarkFlagRequired("email")
	_ = ticketsCreateCmd.MarkFlagRequired("subject")
	_ = ticketsCreateCmd.MarkFlagRequired("description")

	ticketsUpdateCmd.Flags().String("subject", "", "Ticket subject")
	ticketsUpdateCmd.Flags().String("description", "", "Ticket description")
	ticketsUpdateCmd.Flags().String("status", "", "Ticket status, e.g. open, waiting_on_customer, closed")
	ticketsUpdateCmd.Flags().String("priority", "", "Ticket priority")
	ticketsUpdateCmd.Flags().String("category", "", "Ticket category")
	ticketsUpdateCmd.Flags().String("agent-id", "", "Agent ID to assign")
	ticketsUpdateCmd.Flags().String("group-id", "", "Group ID to assign")
	ticketsUpdateCmd.Flags().String("assignee-email", "", "Agent email to assign")
	addTicketExtraFlags(ticketsUpdateCmd)
	ticketsUpdateCmd.Flags().StringArray("tag", []string{}, "Replace every tag with these (repeat flag for multiple)")
	ticketsUpdateCmd.Flags().StringArray("add-tag", []string{}, "Tag to add, keeping existing tags (repeat flag for multiple)")
	ticketsUpdateCmd.Flags().StringArray("remove-tag", []string{}, "Tag to remove (repeat flag for multiple)")
	addAttachFlag(ticketsUpdateCmd)

	ticketsCmd.AddCommand(ticketsListCmd)
	ticketsCmd.AddCommand(ticketsShowCmd)
	ticketsCmd.AddCommand(ticketsCreateCmd)
	ticketsCmd.AddCommand(ticketsUpdateCmd)

	register(func(root *cobra.Command) { root.AddCommand(ticketsCmd) })
}

func addTicketExtraFlags(cmd *cobra.Command) {
	cmd.Flags().String("group", "", "Group name to assign")
	cmd.Flags().String("sub-category-one", "", "Ticket sub category one")
	cmd.Flags().String("sub-category-two", "", "Ticket sub category two")
	cmd.Flags().String("resolution-due-date", "", "Resolution due date (ISO 8601)")
	cmd.Flags().StringArray("field", []string{}, "Custom ticket field as Name=Value (repeat flag for multiple)")
}

func addDateRangeFlags(cmd *cobra.Command) {
	cmd.Flags().String("range-type", "", "Date range: last_7_days, last_30_days, last_24_hours, this_week, last_week, this_month, last_month, this_quarter, last_quarter, this_year, last_year, all_time, or custom")
	cmd.Flags().String("start-date", "", "Start date for custom range (YYYY-MM-DD)")
	cmd.Flags().String("end-date", "", "End date for custom range (YYYY-MM-DD)")
}

func dateRangeParams(cmd *cobra.Command) url.Values {
	params := url.Values{}
	if v, _ := cmd.Flags().GetString("range-type"); v != "" {
		params.Set("range_type", v)
	}
	if v, _ := cmd.Flags().GetString("start-date"); v != "" {
		params.Set("start_date", v)
	}
	if v, _ := cmd.Flags().GetString("end-date"); v != "" {
		params.Set("end_date", v)
	}
	return params
}
