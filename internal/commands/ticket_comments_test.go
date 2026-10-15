package commands

import "testing"

func TestValidateCommentType(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"reply is allowed", "reply", false},
		{"note is allowed", "note", false},
		{"unknown value is rejected", "internal", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCommentType(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateCommentType(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateCommentType(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestTicketCommentsCreateSendsAuthorEmail(t *testing.T) {
	req := startTestAPI(t, `{"id":"1"}`)
	setFlags(t, ticketCommentsCreateCmd, map[string][]string{
		"content": {"<p>hi</p>"}, "author-email": {"riley@example.com"},
	})

	if err := ticketCommentsCreateCmd.RunE(ticketCommentsCreateCmd, []string{"1042"}); err != nil {
		t.Fatalf("tickets comments create returned %v", err)
	}

	if req.path != "/tickets/1042/comments" {
		t.Errorf("path = %q", req.path)
	}
	if req.body["author_email"] != "riley@example.com" {
		t.Errorf("body = %v, want author_email", req.body)
	}
}
