package commands

import (
	"encoding/json"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/output"
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

		if items, headers, rows, ok := ticketStatusReport(data); ok {
			app.Printer.PrintTable(items, headers, rows, nil)
		} else {
			printList(data, "ticket_statuses", nil)
		}
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

		if headers, rows, ok := ticketTimeSeriesReport(data); ok {
			app.Printer.PrintTable(data, headers, rows, nil)
		} else {
			printResource(data, nil)
		}
		return nil
	},
}

func ticketStatusReport(data json.RawMessage) (json.RawMessage, []string, [][]interface{}, bool) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, nil, nil, false
	}

	items, ok := parsed["ticket_statuses"]
	if !ok {
		return nil, nil, nil, false
	}

	var statuses []map[string]interface{}
	if err := json.Unmarshal(items, &statuses); err != nil || len(statuses) == 0 {
		return nil, nil, nil, false
	}

	headers := []string{"NAME", "PRESENT", "PREVIOUS", "CHANGE %"}
	rows := make([][]interface{}, len(statuses))
	for i, status := range statuses {
		value, _ := status["value"].(map[string]interface{})
		rows[i] = []interface{}{
			status["name"],
			value["present"],
			value["previous"],
			value["change_percentage"],
		}
	}

	return items, headers, rows, true
}

func ticketTimeSeriesReport(data json.RawMessage) ([]string, [][]interface{}, bool) {
	var parsed struct {
		Dates []interface{}            `json:"dates"`
		Data  []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, nil, false
	}
	if len(parsed.Dates) == 0 || len(parsed.Data) == 0 {
		return nil, nil, false
	}

	headers := []string{"DATE"}
	for _, series := range parsed.Data {
		name, _ := series["name"].(string)
		headers = append(headers, output.FormatHeader(name))
	}

	rows := make([][]interface{}, len(parsed.Dates))
	for i, date := range parsed.Dates {
		row := make([]interface{}, 0, len(parsed.Data)+1)
		row = append(row, date)
		for _, series := range parsed.Data {
			values, _ := series["values"].([]interface{})
			if i < len(values) {
				row = append(row, values[i])
			} else {
				row = append(row, nil)
			}
		}
		rows[i] = row
	}

	return headers, rows, true
}

func reportParams(cmd *cobra.Command) url.Values {
	return mergeParams(dateRangeParams(cmd), paginationParams(cmd))
}

func mergeParams(dst, src url.Values) url.Values {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
	return dst
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

	addPaginationFlags(reportsAgentsCmd)
	addPaginationFlags(reportsGroupsCmd)

	reportsCmd.AddCommand(reportsAgentsCmd)
	reportsCmd.AddCommand(reportsGroupsCmd)
	reportsCmd.AddCommand(reportsSurveysCmd)
	reportsCmd.AddCommand(reportsTicketsCmd)
	reportsCmd.AddCommand(reportsTicketTimeSeriesCmd)

	register(func(root *cobra.Command) { root.AddCommand(reportsCmd) })
}
