package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-desk-cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestValidateTicketSort(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"created_at is allowed", "created_at", false},
		{"updated_at is allowed", "updated_at", false},
		{"unknown column is rejected", "priority", true},
		{"sql fragment is rejected", "created_at desc", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTicketSort(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateTicketSort(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateTicketSort(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestValidateTicketOrder(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"asc is allowed", "asc", false},
		{"desc is allowed", "desc", false},
		{"uppercase is rejected", "DESC", true},
		{"unknown direction is rejected", "descending", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTicketOrder(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateTicketOrder(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateTicketOrder(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestValidateTicketFieldFilter(t *testing.T) {
	cases := []struct {
		name    string
		field   string
		value   string
		wantErr string
	}{
		{"both empty is allowed", "", "", ""},
		{"both set is allowed", "Browser", "Chrome", ""},
		{"value without name is rejected", "", "Chrome", "--field-name is required"},
		{"name without value is rejected", "Browser", "", "--field-value is required"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTicketFieldFilter(tc.field, tc.value)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("validateTicketFieldFilter(%q, %q) = %v, want nil", tc.field, tc.value, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateTicketFieldFilter(%q, %q) = nil, want error", tc.field, tc.value)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %v, want it to mention %q", err, tc.wantErr)
			}
		})
	}
}

func TestTicketsListDefaultsToNewestFirst(t *testing.T) {
	sort, err := ticketsListCmd.Flags().GetString("sort")
	if err != nil {
		t.Fatalf("sort flag not registered: %v", err)
	}
	if sort != defaultTicketSort {
		t.Errorf("default sort = %q, want %q", sort, defaultTicketSort)
	}

	order, err := ticketsListCmd.Flags().GetString("order")
	if err != nil {
		t.Fatalf("order flag not registered: %v", err)
	}
	if order != defaultTicketOrder {
		t.Errorf("default order = %q, want %q", order, defaultTicketOrder)
	}
}

func TestTicketsListRejectsBadSortBeforeContactingTheAPI(t *testing.T) {
	cases := []struct {
		name string
		flag string
		bad  string
		want string
	}{
		{"invalid sort", "sort", "priority", "invalid sort"},
		{"invalid order", "order", "descending", "invalid order"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				_ = ticketsListCmd.Flags().Set("sort", defaultTicketSort)
				_ = ticketsListCmd.Flags().Set("order", defaultTicketOrder)
			})

			if err := ticketsListCmd.Flags().Set(tc.flag, tc.bad); err != nil {
				t.Fatalf("could not set --%s: %v", tc.flag, err)
			}

			err := ticketsListCmd.RunE(ticketsListCmd, nil)
			if err == nil {
				t.Fatalf("tickets list accepted --%s %q", tc.flag, tc.bad)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

type capturedRequest struct {
	method string
	path   string
	query  string
	body   map[string]interface{}
}

func startTestAPI(t *testing.T, response string) *capturedRequest {
	t.Helper()
	captured := &capturedRequest{}
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = strings.TrimPrefix(r.URL.Path, "/api/external/v2")
		captured.query = r.URL.RawQuery
		captured.body = nil
		_ = json.NewDecoder(r.Body).Decode(&captured.body)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(api.Close)

	home := t.TempDir()
	dir := filepath.Join(home, ".config", "neetodesk")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not create the config directory: %v", err)
	}
	creds := `{"credentials":[{"subdomain":"acme","email":"a@acme.com","session_token":"tok"}]}`
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(creds), 0o600); err != nil {
		t.Fatalf("could not write the credentials: %v", err)
	}
	t.Setenv("HOME", home)
	t.Setenv("NEETODESK_BASE_URL", api.URL)

	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		t.Fatalf("could not parse the product config: %v", err)
	}
	Register(cli.New(*cfg))
	return captured
}

func setFlags(t *testing.T, cmd *cobra.Command, flags map[string][]string) {
	t.Helper()
	t.Cleanup(func() {
		for name := range flags {
			flag := cmd.Flags().Lookup(name)
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				_ = slice.Replace([]string{})
			} else {
				_ = flag.Value.Set(flag.DefValue)
			}
			flag.Changed = false
		}
	})
	for name, values := range flags {
		for _, value := range values {
			if err := cmd.Flags().Set(name, value); err != nil {
				t.Fatalf("could not set --%s: %v", name, err)
			}
		}
	}
}

func TestTicketsListSendsNewFiltersAsAPIParameters(t *testing.T) {
	req := startTestAPI(t, `{"tickets":[]}`)
	setFlags(t, ticketsListCmd, map[string][]string{
		"priority": {"urgent,high"}, "unassigned": {"true"}, "customer-domain": {"acme.com"},
		"tag": {"vip", "billing"}, "query": {"refund"}, "external-id": {"ext-1"},
	})

	if err := ticketsListCmd.RunE(ticketsListCmd, nil); err != nil {
		t.Fatalf("tickets list returned %v", err)
	}

	got, _ := url.ParseQuery(req.query)
	want := map[string]string{
		"priority": "urgent,high", "unassigned": "true", "customer_domain": "acme.com",
		"tags": "vip,billing", "query": "refund", "external_id": "ext-1",
	}
	for key, value := range want {
		if got.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, got.Get(key), value)
		}
	}
}

func TestTicketsListOmitsUnassignedWhenNotSet(t *testing.T) {
	req := startTestAPI(t, `{"tickets":[]}`)
	if err := ticketsListCmd.RunE(ticketsListCmd, nil); err != nil {
		t.Fatalf("tickets list returned %v", err)
	}
	if strings.Contains(req.query, "unassigned") {
		t.Errorf("query = %q, want no unassigned parameter", req.query)
	}
}

func TestTicketsListRejectsUnassignedWithAssignee(t *testing.T) {
	for _, flag := range []string{"assignee-id", "assignee-email"} {
		t.Run(flag, func(t *testing.T) {
			setFlags(t, ticketsListCmd, map[string][]string{"unassigned": {"true"}, flag: {"x"}})
			err := ticketsListCmd.RunE(ticketsListCmd, nil)
			if err == nil || !strings.Contains(err.Error(), "--unassigned cannot be combined") {
				t.Errorf("error = %v, want the unassigned conflict", err)
			}
		})
	}
}

func TestTicketsCreateSendsTagsGroupAndFields(t *testing.T) {
	req := startTestAPI(t, `{"id":"1"}`)
	setFlags(t, ticketsCreateCmd, map[string][]string{
		"email": {"a@acme.com"}, "subject": {"s"}, "description": {"d"}, "group": {"Engineering"},
		"tag": {"bug", "vip"}, "sub-category-one": {"one"}, "sub-category-two": {"two"},
		"resolution-due-date": {"2026-12-01T10:00:00Z"}, "field": {"Browser=Chrome,Firefox", "Note=a=b"},
	})

	if err := ticketsCreateCmd.RunE(ticketsCreateCmd, nil); err != nil {
		t.Fatalf("tickets create returned %v", err)
	}

	if req.method != "POST" || req.path != "/tickets" {
		t.Errorf("request = %s %s, want POST /tickets", req.method, req.path)
	}
	wantBody := `{"description":"d","email":"a@acme.com","group":"Engineering","resolution_due_date":"2026-12-01T10:00:00Z",` +
		`"sub_category_one":"one","sub_category_two":"two","subject":"s","tags":["bug","vip"],` +
		`"ticket_fields":{"Browser":"Chrome,Firefox","Note":"a=b"}}`
	gotBody, _ := json.Marshal(req.body)
	if string(gotBody) != wantBody {
		t.Errorf("body = %s, want %s", gotBody, wantBody)
	}
}

func TestTicketsCreateRejectsMalformedField(t *testing.T) {
	setFlags(t, ticketsCreateCmd, map[string][]string{
		"email": {"a@acme.com"}, "subject": {"s"}, "description": {"d"}, "field": {"Browser"},
	})
	err := ticketsCreateCmd.RunE(ticketsCreateCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--field must be Name=Value") {
		t.Errorf("error = %v, want the field format error", err)
	}
}

func TestTicketsUpdateSendsTagAdjustments(t *testing.T) {
	req := startTestAPI(t, `{"id":"1"}`)
	setFlags(t, ticketsUpdateCmd, map[string][]string{
		"add-tag": {"billing"}, "remove-tag": {"old", "stale"}, "group": {"Billing"},
		"assignee-email": {"riley@example.com"}, "field": {"Browser=Chrome"},
	})

	if err := ticketsUpdateCmd.RunE(ticketsUpdateCmd, []string{"1042"}); err != nil {
		t.Fatalf("tickets update returned %v", err)
	}

	if req.method != "PATCH" || req.path != "/tickets/1042" {
		t.Errorf("request = %s %s, want PATCH /tickets/1042", req.method, req.path)
	}
	wantBody := `{"add_tags":["billing"],"assignee_email":"riley@example.com","group":"Billing",` +
		`"remove_tags":["old","stale"],"ticket_fields":{"Browser":"Chrome"}}`
	gotBody, _ := json.Marshal(req.body)
	if string(gotBody) != wantBody {
		t.Errorf("body = %s, want %s", gotBody, wantBody)
	}
}

func TestTicketsUpdateReplacesTagsWithTagFlag(t *testing.T) {
	req := startTestAPI(t, `{"id":"1"}`)
	setFlags(t, ticketsUpdateCmd, map[string][]string{"tag": {"a", "b"}})
	if err := ticketsUpdateCmd.RunE(ticketsUpdateCmd, []string{"1042"}); err != nil {
		t.Fatalf("tickets update returned %v", err)
	}
	gotBody, _ := json.Marshal(req.body)
	if string(gotBody) != `{"tags":["a","b"]}` {
		t.Errorf("body = %s, want only tags", gotBody)
	}
}

func TestTicketsUpdateRejectsTagWithAddOrRemove(t *testing.T) {
	for _, flag := range []string{"add-tag", "remove-tag"} {
		t.Run(flag, func(t *testing.T) {
			setFlags(t, ticketsUpdateCmd, map[string][]string{"tag": {"a"}, flag: {"b"}})
			err := ticketsUpdateCmd.RunE(ticketsUpdateCmd, []string{"1042"})
			if err == nil || !strings.Contains(err.Error(), "--tag cannot be combined") {
				t.Errorf("error = %v, want the tag conflict", err)
			}
		})
	}
}

func TestParseNameValues(t *testing.T) {
	got, err := parseNameValues([]string{"A=1", "B=x=y", "C="})
	if err != nil {
		t.Fatalf("parseNameValues returned %v", err)
	}
	if got["A"] != "1" || got["B"] != "x=y" || got["C"] != "" {
		t.Errorf("parsed = %v", got)
	}
	for _, bad := range []string{"nope", "=v"} {
		if _, err := parseNameValues([]string{bad}); err == nil {
			t.Errorf("parseNameValues accepted %q", bad)
		}
	}
}

func TestTicketsListSendsFiltersAsAPIParameters(t *testing.T) {
	req := startTestAPI(t, `{"tickets":[]}`)
	query := &req.query

	flags := map[string]string{
		"status": "open,pending", "customer-id": "cus-1", "customer-email": "a@acme.com",
		"assignee-id": "agent-1", "assignee-email": "b@acme.com",
		"field-name": "single one", "field-value": "two",
		"range-type": "custom", "start-date": "2026-01-01", "end-date": "2026-01-31",
		"sort": "updated_at", "order": "asc", "page": "2", "page-size": "25",
	}
	t.Cleanup(func() {
		for flag := range flags {
			_ = ticketsListCmd.Flags().Set(flag, ticketsListCmd.Flags().Lookup(flag).DefValue)
		}
	})
	for flag, value := range flags {
		if err := ticketsListCmd.Flags().Set(flag, value); err != nil {
			t.Fatalf("could not set --%s: %v", flag, err)
		}
	}

	if err := ticketsListCmd.RunE(ticketsListCmd, nil); err != nil {
		t.Fatalf("tickets list returned %v", err)
	}

	want := "assignee_email=b%40acme.com&assignee_id=agent-1&customer_email=a%40acme.com" +
		"&customer_id=cus-1&end_date=2026-01-31&field_name=single+one&field_value=two" +
		"&order=asc&page=2&page_number=2&page_size=25&range_type=custom&sort=updated_at" +
		"&start_date=2026-01-01&status=open%2Cpending"
	if *query != want {
		t.Errorf("query = %q, want %q", *query, want)
	}
}
