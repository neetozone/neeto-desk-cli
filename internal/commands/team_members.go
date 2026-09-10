package commands

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var teamMembersCmd = &cobra.Command{
	Use:   "team-members",
	Short: "Manage team members (agents)",
}

const (
	defaultTeamMemberSort  = "created_at"
	defaultTeamMemberOrder = "desc"
)

var (
	validTeamMemberSorts  = []string{"created_at", "updated_at", "email", "first_name", "last_name"}
	validTeamMemberOrders = []string{"asc", "desc"}
)

func validateTeamMemberSort(v string) error {
	if v == "" || slices.Contains(validTeamMemberSorts, v) {
		return nil
	}
	return fmt.Errorf("invalid sort %q: must be one of %s", v, strings.Join(validTeamMemberSorts, ", "))
}

func validateTeamMemberOrder(v string) error {
	if v == "" || slices.Contains(validTeamMemberOrders, v) {
		return nil
	}
	return fmt.Errorf("invalid order %q: must be one of %s", v, strings.Join(validTeamMemberOrders, ", "))
}

var teamMembersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members",
	RunE: func(cmd *cobra.Command, args []string) error {
		sort, _ := cmd.Flags().GetString("sort")
		if err := validateTeamMemberSort(sort); err != nil {
			return err
		}
		order, _ := cmd.Flags().GetString("order")
		if err := validateTeamMemberOrder(order); err != nil {
			return err
		}

		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			params.Set("email", email)
		}
		if sort != "" {
			params.Set("sort", sort)
		}
		if order != "" {
			params.Set("order", order)
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

		printMessage("Team member removed.")
		return nil
	},
}

func init() {
	addPaginationFlags(teamMembersListCmd)
	teamMembersListCmd.Flags().String("email", "", "Filter by email")
	teamMembersListCmd.Flags().String("sort", defaultTeamMemberSort, "Sort field: created_at, updated_at, email, first_name or last_name")
	teamMembersListCmd.Flags().String("order", defaultTeamMemberOrder, "Sort direction: asc or desc")

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

	register(func(root *cobra.Command) { root.AddCommand(teamMembersCmd) })
}
