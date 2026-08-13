package commands

import (
	"fmt"
	"strings"

	"github.com/neetozone/neeto-desk-cli/internal/output"
	"github.com/spf13/cobra"
)

var fieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage custom ticket fields",
}

var fieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom ticket fields",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if state, _ := cmd.Flags().GetString("state"); state != "" {
			params.Set("state", state)
		}

		data, err := c.Get("/fields", params)
		if err != nil {
			return err
		}

		printList(data, "fields", []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk fields show <id>"},
			{Label: "Update", Command: "neetodesk fields update <id>"},
		})
		return nil
	},
}

var fieldsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a custom ticket field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/fields/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var fieldsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom ticket field",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		kind, _ := cmd.Flags().GetString("kind")
		options, _ := cmd.Flags().GetStringArray("option")

		if isDropdownKind(kind) && len(options) == 0 {
			return fmt.Errorf("--option is required at least once for %s fields", kind)
		}

		field := map[string]interface{}{
			"name": name,
			"kind": kind,
		}
		if len(options) > 0 {
			field["data"] = options
		}
		if v, _ := cmd.Flags().GetString("state"); v != "" {
			field["state"] = v
		}
		if v, _ := cmd.Flags().GetString("placeholder-slug"); v != "" {
			field["placeholder_slug"] = v
		}
		if cmd.Flags().Changed("display-order") {
			v, _ := cmd.Flags().GetInt("display-order")
			field["display_order"] = v
		}
		addRequiredFlags(cmd, field)

		data, err := c.Post("/fields", map[string]interface{}{"field": field})
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk fields show <id>"},
		})
		return nil
	},
}

var fieldsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a custom ticket field",
	Long: "Updates a custom ticket field.\n\n" +
		"The field type cannot be changed after creation. Dropdown options are updated\n" +
		"incrementally: options that are not referenced are left unchanged.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		field := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			field["name"] = v
		}
		if v, _ := cmd.Flags().GetString("state"); v != "" {
			field["state"] = v
		}
		if v, _ := cmd.Flags().GetString("placeholder-slug"); v != "" {
			field["placeholder_slug"] = v
		}
		if cmd.Flags().Changed("display-order") {
			v, _ := cmd.Flags().GetInt("display-order")
			field["display_order"] = v
		}
		addRequiredFlags(cmd, field)

		addOptions, _ := cmd.Flags().GetStringArray("add-option")
		renameOptions, _ := cmd.Flags().GetStringArray("rename-option")
		removeOptions, _ := cmd.Flags().GetStringArray("remove-option")

		data, err := buildOptionChanges(addOptions, renameOptions)
		if err != nil {
			return err
		}
		if len(data) > 0 {
			field["data"] = data
		}
		if len(removeOptions) > 0 {
			field["remove_option_ids"] = removeOptions
		}

		if len(field) == 0 {
			return fmt.Errorf("Nothing to update. Pass at least one flag, such as --name or --state")
		}

		result, err := c.Patch(fmt.Sprintf("/fields/%s", args[0]), map[string]interface{}{"field": field})
		if err != nil {
			return err
		}

		printResource(result, nil)
		return nil
	},
}

var fieldsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a custom ticket field",
	Long: "Deletes a custom ticket field along with the values stored against it in tickets.\n\n" +
		"Only an inactive field can be deleted. Deactivate it first with\n" +
		"`neetodesk fields update <id> --state inactive`.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete(fmt.Sprintf("/fields/%s", args[0])); err != nil {
			return err
		}

		output.PrintMessage("Ticket field deleted.")
		return nil
	},
}

// buildOptionChanges turns --add-option and --rename-option values into the
// option payload the API expects. An entry with only a label adds an option;
// an entry with an id renames the existing option.
func buildOptionChanges(addOptions, renameOptions []string) ([]map[string]interface{}, error) {
	data := []map[string]interface{}{}

	for _, label := range addOptions {
		data = append(data, map[string]interface{}{"label": label})
	}

	for _, pair := range renameOptions {
		id, label, found := strings.Cut(pair, "=")
		if !found || strings.TrimSpace(id) == "" || strings.TrimSpace(label) == "" {
			return nil, fmt.Errorf("--rename-option expects <id>=<new label>, got %q", pair)
		}
		data = append(data, map[string]interface{}{"id": id, "label": label})
	}

	return data, nil
}

// addRequiredFlags copies the three "is required" toggles into the payload,
// but only when the caller actually passed them, so that an omitted flag
// leaves the stored value alone instead of resetting it to false.
func addRequiredFlags(cmd *cobra.Command, field map[string]interface{}) {
	toggles := map[string]string{
		"required":           "is_required",
		"required-on-submit": "is_required_for_agent_when_submitting_form",
		"required-on-close":  "is_required_for_agent_when_closing_ticket",
	}

	for flag, attribute := range toggles {
		if cmd.Flags().Changed(flag) {
			v, _ := cmd.Flags().GetBool(flag)
			field[attribute] = v
		}
	}
}

func isDropdownKind(kind string) bool {
	return kind == "single_option" || kind == "multi_option"
}

func init() {
	addPaginationFlags(fieldsListCmd)
	fieldsListCmd.Flags().String("state", "", "Filter by state (active or inactive)")

	fieldsCreateCmd.Flags().String("name", "", "Field name")
	fieldsCreateCmd.Flags().String("kind", "",
		"Field type: text, textarea, integer, decimal, date, checkbox, single_option or multi_option")
	fieldsCreateCmd.Flags().StringArray("option", []string{},
		"Option label for single_option and multi_option fields (repeat flag for multiple)")
	fieldsCreateCmd.Flags().String("state", "", "State of the field (active or inactive), defaults to active")
	fieldsCreateCmd.Flags().String("placeholder-slug", "", "Slug used in placeholders and automations")
	fieldsCreateCmd.Flags().Int("display-order", 0, "Position in the list of ticket fields")
	fieldsCreateCmd.Flags().Bool("required", false, "Require customers to fill this field")
	fieldsCreateCmd.Flags().Bool("required-on-submit", false, "Require agents to fill this field when creating a ticket")
	fieldsCreateCmd.Flags().Bool("required-on-close", false, "Require agents to fill this field before closing a ticket")
	_ = fieldsCreateCmd.MarkFlagRequired("name")
	_ = fieldsCreateCmd.MarkFlagRequired("kind")

	fieldsUpdateCmd.Flags().String("name", "", "New field name")
	fieldsUpdateCmd.Flags().String("state", "", "New state (active or inactive)")
	fieldsUpdateCmd.Flags().String("placeholder-slug", "", "New slug used in placeholders and automations")
	fieldsUpdateCmd.Flags().Int("display-order", 0, "New position in the list of ticket fields")
	fieldsUpdateCmd.Flags().Bool("required", false, "Require customers to fill this field")
	fieldsUpdateCmd.Flags().Bool("required-on-submit", false, "Require agents to fill this field when creating a ticket")
	fieldsUpdateCmd.Flags().Bool("required-on-close", false, "Require agents to fill this field before closing a ticket")
	fieldsUpdateCmd.Flags().StringArray("add-option", []string{},
		"Label of an option to add (repeat flag for multiple)")
	fieldsUpdateCmd.Flags().StringArray("rename-option", []string{},
		"Rename an existing option, as <id>=<new label> (repeat flag for multiple)")
	fieldsUpdateCmd.Flags().StringArray("remove-option", []string{},
		"ID of an option to remove (repeat flag for multiple)")

	fieldsCmd.AddCommand(fieldsListCmd)
	fieldsCmd.AddCommand(fieldsShowCmd)
	fieldsCmd.AddCommand(fieldsCreateCmd)
	fieldsCmd.AddCommand(fieldsUpdateCmd)
	fieldsCmd.AddCommand(fieldsDeleteCmd)

	rootCmd.AddCommand(fieldsCmd)
}
