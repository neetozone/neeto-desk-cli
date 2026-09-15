package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var ticketCommentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage ticket comments",
}

var validCommentTypes = map[string]bool{"reply": true, "note": true}

func validateCommentType(v string) error {
	if v == "" || validCommentTypes[v] {
		return nil
	}
	return fmt.Errorf("invalid comment-type %q: must be one of reply, note", v)
}

var ticketCommentsListCmd = &cobra.Command{
	Use:   "list <ticket-number>",
	Short: "List comments on a ticket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		path := fmt.Sprintf("/tickets/%s/comments", args[0])
		data, err := c.Get(path, params)
		if err != nil {
			return err
		}

		printList(data, "comments", []output.Breadcrumb{
			{Label: "Show", Command: fmt.Sprintf("neetodesk tickets comments show %s <comment-id>", args[0])},
		})
		return nil
	},
}

var ticketCommentsShowCmd = &cobra.Command{
	Use:   "show <ticket-number> <comment-id>",
	Short: "Show a comment on a ticket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/tickets/%s/comments/%s", args[0], args[1])
		data, err := c.Get(path, nil)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var ticketCommentsCreateCmd = &cobra.Command{
	Use:   "create <ticket-number>",
	Short: "Add a comment to a ticket",
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
			if err := validateCommentType(v); err != nil {
				return err
			}
			body["comment_type"] = v
		}

		attachments, err := attachmentSignedIDs(cmd, c)
		if err != nil {
			return err
		}
		if len(attachments) > 0 {
			body["attachments"] = attachments
		}

		path := fmt.Sprintf("/tickets/%s/comments", args[0])
		data, err := c.Post(path, body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	addPaginationFlags(ticketCommentsListCmd)

	ticketCommentsCreateCmd.Flags().String("content", "", "Comment content (HTML)")
	ticketCommentsCreateCmd.Flags().String("comment-type", "reply", "Comment type: reply or note")
	addAttachFlag(ticketCommentsCreateCmd)
	_ = ticketCommentsCreateCmd.MarkFlagRequired("content")

	ticketCommentsCmd.AddCommand(ticketCommentsListCmd)
	ticketCommentsCmd.AddCommand(ticketCommentsShowCmd)
	ticketCommentsCmd.AddCommand(ticketCommentsCreateCmd)

	ticketsCmd.AddCommand(ticketCommentsCmd)
}
