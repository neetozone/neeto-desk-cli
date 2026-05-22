package commands

import (
	"github.com/neetozone/neeto-desk-cli/internal/output"
	"github.com/spf13/cobra"
)

var formsCmd = &cobra.Command{
	Use:   "forms",
	Short: "List contact forms",
}

var formsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List enabled contact forms",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/forms", nil)
		if err != nil {
			return err
		}

		printList(data, "forms", []output.Breadcrumb{
			{Label: "View", Command: "neetodesk forms list"},
		})
		return nil
	},
}

func init() {
	formsCmd.AddCommand(formsListCmd)
	rootCmd.AddCommand(formsCmd)
}
