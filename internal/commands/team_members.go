package commands

import (
	"fmt"

	"github.com/neetozone/neeto-desk-cli/internal/output"
	"github.com/spf13/cobra"
)

var teamMembersCmd = &cobra.Command{
	Use:   "team-members",
	Short: "Manage team members (agents)",
}

var teamMembersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			params.Set("email", email)
		}

		data, err := c.Get("/team-members", params)
		if err != nil {
			return err
		}

		printList(data, "team_members", []output.Breadcrumb{
			{Label: "Show", Command: "neetodesk team-members show <id>"},
			{Label: "Update", Command: "neetodesk team-members update <id>"},
		})
		return nil
	},
}

var teamMembersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/team-members/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var teamMembersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Invite team members by email",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		emails, _ := cmd.Flags().GetStringArray("email")
		role, _ := cmd.Flags().GetString("role")
		sendInvite, _ := cmd.Flags().GetBool("send-invitation-email")

		body := map[string]interface{}{
			"emails":                emails,
			"organization_role":     role,
			"send_invitation_email": sendInvite,
		}

		data, err := c.Post("/team-members", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

var teamMembersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("email"); v != "" {
			body["email"] = v
		}
		if v, _ := cmd.Flags().GetString("first-name"); v != "" {
			body["first_name"] = v
		}
		if v, _ := cmd.Flags().GetString("last-name"); v != "" {
			body["last_name"] = v
		}
		if v, _ := cmd.Flags().GetString("time-zone"); v != "" {
			body["time_zone"] = v
		}
		if v, _ := cmd.Flags().GetString("role"); v != "" {
			body["organization_role"] = v
		}

		data, err := c.Patch(fmt.Sprintf("/team-members/%s", args[0]), body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var teamMembersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete(fmt.Sprintf("/team-members/%s", args[0])); err != nil {
			return err
		}

		output.PrintMessage("Team member removed.")
		return nil
	},
}

func init() {
	addPaginationFlags(teamMembersListCmd)
	teamMembersListCmd.Flags().String("email", "", "Filter by email")

	teamMembersCreateCmd.Flags().StringArray("email", []string{}, "Email addresses to invite (repeat flag for multiple)")
	teamMembersCreateCmd.Flags().String("role", "", "Organization role (e.g. agent, admin)")
	teamMembersCreateCmd.Flags().Bool("send-invitation-email", true, "Send invitation email to new members")
	_ = teamMembersCreateCmd.MarkFlagRequired("email")
	_ = teamMembersCreateCmd.MarkFlagRequired("role")

	teamMembersUpdateCmd.Flags().String("email", "", "New email")
	teamMembersUpdateCmd.Flags().String("first-name", "", "First name")
	teamMembersUpdateCmd.Flags().String("last-name", "", "Last name")
	teamMembersUpdateCmd.Flags().String("time-zone", "", "Time zone")
	teamMembersUpdateCmd.Flags().String("role", "", "Organization role")

	teamMembersCmd.AddCommand(teamMembersListCmd)
	teamMembersCmd.AddCommand(teamMembersShowCmd)
	teamMembersCmd.AddCommand(teamMembersCreateCmd)
	teamMembersCmd.AddCommand(teamMembersUpdateCmd)
	teamMembersCmd.AddCommand(teamMembersDeleteCmd)

	rootCmd.AddCommand(teamMembersCmd)
}
