package commands

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildOptionChanges_AddsAndRenames(t *testing.T) {
	data, err := buildOptionChanges(
		[]string{"Safari"},
		[]string{"aaaabbbb-cccc-dddd-eeee-ffff00004444=Chromium"})
	if err != nil {
		t.Fatalf("buildOptionChanges() error = %v, want nil", err)
	}

	want := []map[string]interface{}{
		{"label": "Safari"},
		{"id": "aaaabbbb-cccc-dddd-eeee-ffff00004444", "label": "Chromium"},
	}
	if !reflect.DeepEqual(data, want) {
		t.Errorf("data = %#v, want %#v", data, want)
	}
}

func TestBuildOptionChanges_LabelWithEqualsSign(t *testing.T) {
	data, err := buildOptionChanges(nil, []string{"opt-1=A=B"})
	if err != nil {
		t.Fatalf("buildOptionChanges() error = %v, want nil", err)
	}

	want := []map[string]interface{}{{"id": "opt-1", "label": "A=B"}}
	if !reflect.DeepEqual(data, want) {
		t.Errorf("data = %#v, want %#v (only the first = separates id from label)", data, want)
	}
}

func TestBuildOptionChanges_RejectsMalformedRename(t *testing.T) {
	for _, pair := range []string{"no-separator", "=missing-id", "missing-label="} {
		if _, err := buildOptionChanges(nil, []string{pair}); err == nil {
			t.Errorf("buildOptionChanges(%q) error = nil, want an error", pair)
		}
	}
}

func TestBuildOptionChanges_EmptyWhenNothingPassed(t *testing.T) {
	data, err := buildOptionChanges(nil, nil)
	if err != nil {
		t.Fatalf("buildOptionChanges() error = %v, want nil", err)
	}
	if len(data) != 0 {
		t.Errorf("data = %#v, want empty so that `data` is left out of the payload", data)
	}
}

func TestAddRequiredFlags_OnlyIncludesFlagsThatWerePassed(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("required", false, "")
	cmd.Flags().Bool("required-on-submit", false, "")
	cmd.Flags().Bool("required-on-close", false, "")

	if err := cmd.Flags().Set("required-on-close", "true"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	field := map[string]interface{}{}
	addRequiredFlags(cmd, field)

	want := map[string]interface{}{"is_required_for_agent_when_closing_ticket": true}
	if !reflect.DeepEqual(field, want) {
		t.Errorf("field = %#v, want %#v (untouched toggles must not be sent)", field, want)
	}
}

func TestAddRequiredFlags_SendsExplicitFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("required", true, "")
	cmd.Flags().Bool("required-on-submit", false, "")
	cmd.Flags().Bool("required-on-close", false, "")

	if err := cmd.Flags().Set("required", "false"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	field := map[string]interface{}{}
	addRequiredFlags(cmd, field)

	want := map[string]interface{}{"is_required": false}
	if !reflect.DeepEqual(field, want) {
		t.Errorf("field = %#v, want %#v (an explicit --required=false must reach the API)", field, want)
	}
}

func TestIsDropdownKind(t *testing.T) {
	for _, kind := range []string{"single_option", "multi_option"} {
		if !isDropdownKind(kind) {
			t.Errorf("isDropdownKind(%q) = false, want true", kind)
		}
	}
	for _, kind := range []string{"text", "textarea", "integer", "decimal", "date", "checkbox", ""} {
		if isDropdownKind(kind) {
			t.Errorf("isDropdownKind(%q) = true, want false", kind)
		}
	}
}
