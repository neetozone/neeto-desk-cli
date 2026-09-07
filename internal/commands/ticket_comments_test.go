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
