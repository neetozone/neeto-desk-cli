package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neetozone/neeto-cli-commons/client"
)

func TestFileChecksumIsTheBase64MD5Digest(t *testing.T) {
	path := writeTempFile(t, "sample.txt", "sample")

	checksum, err := fileChecksum(path)
	if err != nil {
		t.Fatalf("fileChecksum() returned %v", err)
	}

	// base64 of the raw MD5 digest of "sample", the form Content-MD5 takes.
	if want := "Xo/5v1W6NQgZnSLphBKb5g=="; checksum != want {
		t.Errorf("fileChecksum() = %q, want %q", checksum, want)
	}
}

func TestFileChecksumReportsAMissingFile(t *testing.T) {
	if _, err := fileChecksum(filepath.Join(t.TempDir(), "absent.txt")); err == nil {
		t.Error("fileChecksum() = nil, want error for a missing file")
	}
}

func TestContentTypeFor(t *testing.T) {
	cases := map[string]string{
		"screenshot.png": "image/png",
		"SCREENSHOT.PNG": "image/png",
		"report.PDF":     "application/pdf",
		"report.pdf":     "application/pdf",
		"notes":          "application/octet-stream",
		"archive.bin":    "application/octet-stream",
	}

	for name, want := range cases {
		if got := contentTypeFor(name); got != want {
			t.Errorf("contentTypeFor(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestUploadFileReservesTheUploadThenSendsTheBytes(t *testing.T) {
	path := writeTempFile(t, "report.pdf", "sample")
	var reserved map[string]any
	var sentBody, sentContentType, sentSessionToken string
	var sentMethod string

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sentMethod = r.Method
		sentBody = string(body)
		sentContentType = r.Header.Get("Content-Type")
		sentSessionToken = r.Header.Get("Session-Token")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer storage.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/attachment_uploads" {
			t.Errorf("reserved at %q, want /attachment_uploads", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&reserved)
		_, _ = fmt.Fprintf(w, `{"signed_id":"signed-abc","upload_url":%q,
			"upload_headers":{"Content-Type":"application/pdf"}}`, storage.URL+"/put")
	}))
	defer api.Close()

	upload, _, err := uploadFile(testClient(api), path)
	if err != nil {
		t.Fatalf("uploadFile() returned %v", err)
	}

	if upload.SignedID != "signed-abc" {
		t.Errorf("signed id = %q, want signed-abc", upload.SignedID)
	}
	if reserved["filename"] != "report.pdf" {
		t.Errorf("reserved filename = %v, want report.pdf", reserved["filename"])
	}
	if reserved["byte_size"] != float64(6) {
		t.Errorf("reserved byte_size = %v, want 6", reserved["byte_size"])
	}
	if reserved["checksum"] != "Xo/5v1W6NQgZnSLphBKb5g==" {
		t.Errorf("reserved checksum = %v, want the base64 MD5 digest", reserved["checksum"])
	}
	if reserved["content_type"] != "application/pdf" {
		t.Errorf("reserved content_type = %v, want application/pdf", reserved["content_type"])
	}
	if sentMethod != http.MethodPut {
		t.Errorf("sent with %s, want PUT", sentMethod)
	}
	if sentBody != "sample" {
		t.Errorf("sent body = %q, want sample", sentBody)
	}
	if sentContentType != "application/pdf" {
		t.Errorf("sent Content-Type = %q, want the header the API returned", sentContentType)
	}
	if sentSessionToken != "" {
		t.Errorf("sent Session-Token %q to the storage service, want none", sentSessionToken)
	}
}

func TestUploadFileReportsAStorageFailureAndDoesNotReturnASignedID(t *testing.T) {
	path := writeTempFile(t, "report.pdf", "sample")

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("SignatureDoesNotMatch"))
	}))
	defer storage.Close()

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"signed_id":"signed-abc","upload_url":%q,"upload_headers":{}}`, storage.URL)
	}))
	defer api.Close()

	_, _, err := uploadFile(testClient(api), path)
	if err == nil {
		t.Fatal("uploadFile() = nil, want an error when the storage service refuses the file")
	}
	if !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Errorf("error = %q, want it to carry the storage service's answer", err.Error())
	}
}

func TestUploadFileReportsAMissingFileWithoutReservingAnUpload(t *testing.T) {
	reserved := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reserved = true
	}))
	defer api.Close()

	_, _, err := uploadFile(testClient(api), filepath.Join(t.TempDir(), "absent.pdf"))
	if err == nil {
		t.Fatal("uploadFile() = nil, want an error for a missing file")
	}
	if reserved {
		t.Error("reserved an upload for a file that could not be read")
	}
}

func TestUploadFileRejectsADirectory(t *testing.T) {
	_, _, err := uploadFile(testClient(nil), t.TempDir())
	if err == nil {
		t.Fatal("uploadFile() = nil, want an error for a directory")
	}
}

func writeTempFile(t *testing.T, name, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("could not write the fixture file: %v", err)
	}

	return path
}

func testClient(server *httptest.Server) *client.Client {
	c := &client.Client{SessionToken: "session-token", HTTPClient: http.DefaultClient}
	if server != nil {
		c.BaseURL = server.URL
		c.HTTPClient = server.Client()
	}

	return c
}
