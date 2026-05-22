package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "View NeetoDesk reports",
}

var reportsAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Agent performance report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/reports/agents", reportParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "members", nil)
		return nil
	},
}

var reportsGroupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Group performance report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/reports/groups", reportParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "teams", nil)
		return nil
	},
}

var reportsSurveysCmd = &cobra.Command{
	Use:   "surveys",
	Short: "Customer satisfaction surveys report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/reports/surveys", dateRangeParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "surveys", nil)
		return nil
	},
}

var reportsTicketsCmd = &cobra.Command{
	Use:   "tickets",
	Short: "Ticket status change report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/reports/tickets", dateRangeParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "ticket_statuses", nil)
		return nil
	},
}

var reportsTicketTimeSeriesCmd = &cobra.Command{
	Use:   "ticket-time-series",
	Short: "Ticket creation and closure time series report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/reports/ticket-time-series", dateRangeParams(cmd))
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func reportParams(cmd *cobra.Command) url.Values {
	params := dateRangeParams(cmd)
	page, _ := cmd.Flags().GetInt("page")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		params.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	return params
}

func init() {
	for _, cmd := range []*cobra.Command{
		reportsAgentsCmd,
		reportsGroupsCmd,
		reportsSurveysCmd,
		reportsTicketsCmd,
		reportsTicketTimeSeriesCmd,
	} {
		addDateRangeFlags(cmd)
	}

	reportsAgentsCmd.Flags().Int("page", 0, "Page number")
	reportsAgentsCmd.Flags().Int("page-size", 0, "Items per page")
	reportsGroupsCmd.Flags().Int("page", 0, "Page number")
	reportsGroupsCmd.Flags().Int("page-size", 0, "Items per page")

	reportsCmd.AddCommand(reportsAgentsCmd)
	reportsCmd.AddCommand(reportsGroupsCmd)
	reportsCmd.AddCommand(reportsSurveysCmd)
	reportsCmd.AddCommand(reportsTicketsCmd)
	reportsCmd.AddCommand(reportsTicketTimeSeriesCmd)

	rootCmd.AddCommand(reportsCmd)
}
