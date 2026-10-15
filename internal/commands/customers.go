package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage customers",
}

var customersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			params.Set("email", email)
		}

		data, err := c.Get("/customers", params)
		if err != nil {
			return err
		}

		printList(data, "customers", []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk customers show <id>"},
		})
		return nil
	},
}

var customersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/customers/%s", url.PathEscape(args[0])), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "List", Command: "neetodesk customers list"},
		})
		return nil
	},
}

var customersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a customer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		customer := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("first-name"); v != "" {
			customer["first_name"] = v
		}
		if v, _ := cmd.Flags().GetString("last-name"); v != "" {
			customer["last_name"] = v
		}
		if emails, _ := cmd.Flags().GetStringArray("email"); len(emails) > 0 {
			entries := make([]map[string]interface{}, len(emails))
			for i, email := range emails {
				entries[i] = map[string]interface{}{"email": email, "primary": i == 0}
			}
			customer["emails"] = entries
		}
		if phones, _ := cmd.Flags().GetStringArray("phone"); len(phones) > 0 {
			entries := make([]map[string]interface{}, len(phones))
			for i, phone := range phones {
				entries[i] = map[string]interface{}{"phone": phone, "label": ""}
			}
			customer["phones"] = entries
		}
		if links, _ := cmd.Flags().GetStringArray("link"); len(links) > 0 {
			customer["links"] = links
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			customer["language"] = v
		}
		if v, _ := cmd.Flags().GetString("time-zone"); v != "" {
			customer["time_zone"] = v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			customer["description"] = v
		}
		if v, _ := cmd.Flags().GetString("company-id"); v != "" {
			customer["company_id"] = v
		}

		body := map[string]interface{}{"customer": customer}
		data, err := c.Post("/customers", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(customersListCmd)
	customersListCmd.Flags().String("email", "", "Filter by exact email")

	customersCreateCmd.Flags().StringArray("email", []string{}, "Customer email, first is primary (repeat flag for multiple)")
	customersCreateCmd.Flags().String("first-name", "", "Customer first name")
	customersCreateCmd.Flags().String("last-name", "", "Customer last name")
	customersCreateCmd.Flags().StringArray("phone", []string{}, "Customer phone number (repeat flag for multiple)")
	customersCreateCmd.Flags().StringArray("link", []string{}, "Customer link (repeat flag for multiple)")
	customersCreateCmd.Flags().String("language", "", "Preferred language")
	customersCreateCmd.Flags().String("time-zone", "", "Time zone (e.g. America/New_York)")
	customersCreateCmd.Flags().String("description", "", "Customer description")
	customersCreateCmd.Flags().String("company-id", "", "Company ID to associate")
	_ = customersCreateCmd.MarkFlagRequired("email")

	customersCmd.AddCommand(customersListCmd)
	customersCmd.AddCommand(customersShowCmd)
	customersCmd.AddCommand(customersCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(customersCmd) })
}
