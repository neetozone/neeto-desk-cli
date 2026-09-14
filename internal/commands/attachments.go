package commands

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neetozone/neeto-cli-commons/client"
	"github.com/spf13/cobra"
)

const uploadTimeout = 10 * time.Minute

type attachmentUpload struct {
	SignedID      string            `json:"signed_id"`
	UploadURL     string            `json:"upload_url"`
	UploadHeaders map[string]string `json:"upload_headers"`
}

var attachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Upload files to attach to tickets and comments",
}

var attachmentsUploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file and print the signed id that attaches it",
	Long: "Upload a file and print the signed id that attaches it.\n\n" +
		"Pass the signed id as attachments when creating a ticket or a comment through the API, " +
		"or use --attach to upload and attach in one step.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		_, data, err := uploadFile(c, args[0])
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	attachmentsCmd.AddCommand(attachmentsUploadCmd)

	register(func(root *cobra.Command) { root.AddCommand(attachmentsCmd) })
}

func addAttachFlag(cmd *cobra.Command) {
	cmd.Flags().StringArray("attach", nil, "Path to a file to attach (repeat the flag for more than one)")
}

func attachmentSignedIDs(cmd *cobra.Command, c *client.Client) ([]string, error) {
	paths, _ := cmd.Flags().GetStringArray("attach")
	if len(paths) == 0 {
		return nil, nil
	}

	signedIDs := make([]string, 0, len(paths))
	for _, path := range paths {
		upload, _, err := uploadFile(c, path)
		if err != nil {
			return nil, err
		}
		signedIDs = append(signedIDs, upload.SignedID)
	}

	return signedIDs, nil
}

func uploadFile(c *client.Client, path string) (attachmentUpload, json.RawMessage, error) {
	var upload attachmentUpload

	info, err := os.Stat(path)
	if err != nil {
		return upload, nil, fmt.Errorf("Could not read %s: %w", path, err)
	}
	if info.IsDir() {
		return upload, nil, fmt.Errorf("%s is a directory, not a file.", path)
	}

	checksum, err := fileChecksum(path)
	if err != nil {
		return upload, nil, err
	}

	data, err := c.Post("/attachment_uploads", map[string]interface{}{
		"filename":     filepath.Base(path),
		"byte_size":    info.Size(),
		"checksum":     checksum,
		"content_type": contentTypeFor(path),
	})
	if err != nil {
		return upload, nil, err
	}

	if err := json.Unmarshal(data, &upload); err != nil {
		return upload, nil, fmt.Errorf("Could not read the upload details for %s: %w", path, err)
	}
	if upload.UploadURL == "" || upload.SignedID == "" {
		return upload, nil, fmt.Errorf("Could not reserve an upload for %s.", path)
	}

	if err := sendFile(upload, path, info.Size()); err != nil {
		return upload, nil, err
	}

	return upload, data, nil
}

// The upload URL points at the storage service rather than at the product, so this
// request is built by hand: the API client would attach the session token to it.
func sendFile(upload attachmentUpload, path string, size int64) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Could not read %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	req, err := http.NewRequest(http.MethodPut, upload.UploadURL, file)
	if err != nil {
		return err
	}
	req.ContentLength = size
	for name, value := range upload.UploadHeaders {
		req.Header.Set(name, value)
	}

	resp, err := (&http.Client{Timeout: uploadTimeout}).Do(req)
	if err != nil {
		return fmt.Errorf("Could not upload %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf(
			"Could not upload %s: the storage service answered %d. %s",
			path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

// Base64 of the raw digest, which is the form Content-MD5 takes. Not the hex digest.
func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Could not read %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	digest := md5.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", fmt.Errorf("Could not read %s: %w", path, err)
	}

	return base64.StdEncoding.EncodeToString(digest.Sum(nil)), nil
}

func contentTypeFor(path string) string {
	declared := mime.TypeByExtension(filepath.Ext(path))
	if declared == "" {
		return "application/octet-stream"
	}

	return strings.TrimSpace(strings.Split(declared, ";")[0])
}
