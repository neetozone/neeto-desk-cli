package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ticketDraftsCmd = &cobra.Command{
	Use:   "drafts",
	Short: "Manage ticket drafts",
}

var ticketDraftsCreateCmd = &cobra.Command{
	Use:   "create <ticket-id>",
	Short: "Create or upsert a draft on a ticket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("content"); v != "" {
			body["content"] = v
		}
		if v, _ := cmd.Flags().GetString("comment-type"); v != "" {
			body["comment_type"] = v
		}
		if v, _ := cmd.Flags().GetString("author-email"); v != "" {
			body["author_email"] = v
		}

		path := fmt.Sprintf("/tickets/%s/drafts", args[0])
		data, err := c.Post(path, body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	ticketDraftsCreateCmd.Flags().String("content", "", "Draft content (HTML)")
	ticketDraftsCreateCmd.Flags().String("comment-type", "reply", "Draft type: reply or note")
	ticketDraftsCreateCmd.Flags().String("author-email", "", "Author agent email")
	_ = ticketDraftsCreateCmd.MarkFlagRequired("content")

	ticketDraftsCmd.AddCommand(ticketDraftsCreateCmd)
	ticketsCmd.AddCommand(ticketDraftsCmd)
}
