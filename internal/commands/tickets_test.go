package commands

import (
	"strings"
	"testing"
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
