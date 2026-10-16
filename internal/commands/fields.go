package commands

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var fieldsCmd = &cobra.Command{
	Use:   "fields",
	Short: "Manage custom ticket fields",
}

var validFieldKinds = []string{"text", "textarea", "integer", "decimal", "date", "checkbox", "single_option", "multi_option"}

var fieldBoolFlags = map[string]string{
	"required":                      "is_required",
	"required-when-submitting-form": "is_required_for_agent_when_submitting_form",
	"required-when-closing":         "is_required_for_agent_when_closing_ticket",
}

func validateFieldKind(v string) error {
	if slices.Contains(validFieldKinds, v) {
		return nil
	}
	return fmt.Errorf("invalid kind %q: must be one of %s", v, strings.Join(validFieldKinds, ", "))
}

func parseOptionRenames(pairs []string) ([]map[string]interface{}, error) {
	renames := make([]map[string]interface{}, 0, len(pairs))
	for _, pair := range pairs {
		id, label, found := strings.Cut(pair, "=")
		if !found || id == "" || label == "" {
			return nil, fmt.Errorf("--rename-option must be id=Label")
		}
		renames = append(renames, map[string]interface{}{"id": id, "label": label})
	}
	return renames, nil
}

func fieldCommonBody(cmd *cobra.Command) (map[string]interface{}, error) {
	field := map[string]interface{}{}
	for flag, key := range map[string]string{"name": "name", "state": "state", "placeholder-slug": "placeholder_slug"} {
		if cmd.Flags().Changed(flag) {
			v, _ := cmd.Flags().GetString(flag)
			field[key] = v
		}
	}
	for flag, key := range fieldBoolFlags {
		if cmd.Flags().Changed(flag) {
			v, _ := cmd.Flags().GetBool(flag)
			field[key] = v
		}
	}
	if cmd.Flags().Changed("display-order") {
		v, _ := cmd.Flags().GetInt("display-order")
		field["display_order"] = v
	}
	if state, ok := field["state"].(string); ok && state != "active" && state != "inactive" {
		return nil, fmt.Errorf("invalid state %q: must be active or inactive", state)
	}
	return field, nil
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
		if v, _ := cmd.Flags().GetString("state"); v != "" {
			params.Set("state", v)
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

		data, err := c.Get(fmt.Sprintf("/fields/%s", url.PathEscape(args[0])), nil)
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
		kind, _ := cmd.Flags().GetString("kind")
		if err := validateFieldKind(kind); err != nil {
			return err
		}
		field, err := fieldCommonBody(cmd)
		if err != nil {
			return err
		}
		field["kind"] = kind
		if options, _ := cmd.Flags().GetStringArray("option"); len(options) > 0 {
			field["data"] = options
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post("/fields", map[string]interface{}{"field": field})
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var fieldsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a custom ticket field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		field, err := fieldCommonBody(cmd)
		if err != nil {
			return err
		}
		renamePairs, _ := cmd.Flags().GetStringArray("rename-option")
		options, err := parseOptionRenames(renamePairs)
		if err != nil {
			return err
		}
		added, _ := cmd.Flags().GetStringArray("add-option")
		for _, label := range added {
			options = append(options, map[string]interface{}{"label": label})
		}
		if len(options) > 0 {
			field["data"] = options
		}
		if removed, _ := cmd.Flags().GetStringArray("remove-option"); len(removed) > 0 {
			field["remove_option_ids"] = removed
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Patch(fmt.Sprintf("/fields/%s", url.PathEscape(args[0])), map[string]interface{}{"field": field})
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var fieldsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a custom ticket field",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete(fmt.Sprintf("/fields/%s", url.PathEscape(args[0]))); err != nil {
			return err
		}

		printMessage("Field deleted.")
		return nil
	},
}

func addFieldCommonFlags(cmd *cobra.Command) {
	cmd.Flags().String("state", "", "Field state: active or inactive")
	cmd.Flags().Bool("required", false, "Make the field required")
	cmd.Flags().Bool("required-when-submitting-form", false, "Require agents to fill the field when submitting the form")
	cmd.Flags().Bool("required-when-closing", false, "Require agents to fill the field when closing a ticket")
	cmd.Flags().Int("display-order", 0, "Position of the field in the ticket form")
	cmd.Flags().String("placeholder-slug", "", "Placeholder slug for the field")
}

func init() {
	addPaginationFlags(fieldsListCmd)
	fieldsListCmd.Flags().String("state", "", "Filter by state: active or inactive")

	fieldsCreateCmd.Flags().String("name", "", "Field name")
	fieldsCreateCmd.Flags().String("kind", "", "Field kind: "+strings.Join(validFieldKinds, ", "))
	fieldsCreateCmd.Flags().StringArray("option", []string{}, "Option label for single_option and multi_option fields (repeat flag for multiple)")
	addFieldCommonFlags(fieldsCreateCmd)
	_ = fieldsCreateCmd.MarkFlagRequired("name")
	_ = fieldsCreateCmd.MarkFlagRequired("kind")

	fieldsUpdateCmd.Flags().String("name", "", "Field name")
	fieldsUpdateCmd.Flags().StringArray("add-option", []string{}, "Option label to add (repeat flag for multiple)")
	fieldsUpdateCmd.Flags().StringArray("rename-option", []string{}, "Rename an option as id=Label (repeat flag for multiple)")
	fieldsUpdateCmd.Flags().StringArray("remove-option", []string{}, "Option id to remove (repeat flag for multiple)")
	addFieldCommonFlags(fieldsUpdateCmd)

	fieldsCmd.AddCommand(fieldsListCmd)
	fieldsCmd.AddCommand(fieldsShowCmd)
	fieldsCmd.AddCommand(fieldsCreateCmd)
	fieldsCmd.AddCommand(fieldsUpdateCmd)
	fieldsCmd.AddCommand(fieldsDeleteCmd)

	register(func(root *cobra.Command) { root.AddCommand(fieldsCmd) })
}
