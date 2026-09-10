package commands

import (
	"strings"
	"testing"
)

func TestValidateTeamMemberSort(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"created_at is allowed", "created_at", false},
		{"updated_at is allowed", "updated_at", false},
		{"email is allowed", "email", false},
		{"first_name is allowed", "first_name", false},
		{"last_name is allowed", "last_name", false},
		{"unknown column is rejected", "role", true},
		{"sql fragment is rejected", "email desc", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTeamMemberSort(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateTeamMemberSort(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateTeamMemberSort(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestValidateTeamMemberOrder(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"asc is allowed", "asc", false},
		{"desc is allowed", "desc", false},
		{"uppercase is rejected", "ASC", true},
		{"unknown direction is rejected", "ascending", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTeamMemberOrder(tc.value)
			if tc.wantErr && err == nil {
				t.Errorf("validateTeamMemberOrder(%q) = nil, want error", tc.value)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("validateTeamMemberOrder(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestTeamMembersListDefaultsToNewestFirst(t *testing.T) {
	sort, err := teamMembersListCmd.Flags().GetString("sort")
	if err != nil {
		t.Fatalf("sort flag not registered: %v", err)
	}
	if sort != defaultTeamMemberSort {
		t.Errorf("default sort = %q, want %q", sort, defaultTeamMemberSort)
	}

	order, err := teamMembersListCmd.Flags().GetString("order")
	if err != nil {
		t.Fatalf("order flag not registered: %v", err)
	}
	if order != defaultTeamMemberOrder {
		t.Errorf("default order = %q, want %q", order, defaultTeamMemberOrder)
	}
}

func TestTeamMembersListRejectsBadSortBeforeContactingTheAPI(t *testing.T) {
	cases := []struct {
		name string
		flag string
		bad  string
		want string
	}{
		{"invalid sort", "sort", "role", "invalid sort"},
		{"invalid order", "order", "ascending", "invalid order"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				_ = teamMembersListCmd.Flags().Set("sort", defaultTeamMemberSort)
				_ = teamMembersListCmd.Flags().Set("order", defaultTeamMemberOrder)
			})

			if err := teamMembersListCmd.Flags().Set(tc.flag, tc.bad); err != nil {
				t.Fatalf("could not set --%s: %v", tc.flag, err)
			}

			err := teamMembersListCmd.RunE(teamMembersListCmd, nil)
			if err == nil {
				t.Fatalf("team-members list accepted --%s %q", tc.flag, tc.bad)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}
