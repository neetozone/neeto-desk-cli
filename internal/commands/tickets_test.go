package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-desk-cli"
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

func TestTicketsListSendsFiltersAsAPIParameters(t *testing.T) {
	var query string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"tickets":[]}`))
	}))
	defer api.Close()

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
	if query != want {
		t.Errorf("query = %q, want %q", query, want)
	}
}
