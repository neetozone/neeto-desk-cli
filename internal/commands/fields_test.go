package commands

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateFieldKindAcceptsEachKind(t *testing.T) {
	for _, kind := range []string{"text", "textarea", "integer", "decimal", "date", "checkbox", "single_option", "multi_option"} {
		t.Run(kind, func(t *testing.T) {
			if err := validateFieldKind(kind); err != nil {
				t.Errorf("validateFieldKind(%q) = %v, want nil", kind, err)
			}
		})
	}
}

func TestValidateFieldKindRejectsUnknownKinds(t *testing.T) {
	for _, kind := range []string{"", "number", "SINGLE_OPTION"} {
		t.Run(kind, func(t *testing.T) {
			if err := validateFieldKind(kind); err == nil {
				t.Errorf("validateFieldKind(%q) = nil, want error", kind)
			}
		})
	}
}

func TestParseOptionRenamesSplitsOnTheFirstEquals(t *testing.T) {
	got, err := parseOptionRenames([]string{"abc=New Label", "def=a=b"})
	if err != nil {
		t.Fatalf("parseOptionRenames returned %v", err)
	}
	encoded, _ := json.Marshal(got)
	if string(encoded) != `[{"id":"abc","label":"New Label"},{"id":"def","label":"a=b"}]` {
		t.Errorf("parsed = %s", encoded)
	}
}

func TestParseOptionRenamesRejectsMalformedPairs(t *testing.T) {
	for _, bad := range []string{"abc", "=x", "abc="} {
		t.Run(bad, func(t *testing.T) {
			if _, err := parseOptionRenames([]string{bad}); err == nil {
				t.Errorf("parseOptionRenames accepted %q", bad)
			}
		})
	}
}

func TestFieldsListSendsStateFilter(t *testing.T) {
	req := startTestAPI(t, `{"fields":[]}`)
	setFlags(t, fieldsListCmd, map[string][]string{"state": {"active"}})

	if err := fieldsListCmd.RunE(fieldsListCmd, nil); err != nil {
		t.Fatalf("fields list returned %v", err)
	}
	if req.method != "GET" || req.path != "/fields" || req.query != "state=active" {
		t.Errorf("request = %s %s?%s", req.method, req.path, req.query)
	}
}

func TestFieldsShowRequestsTheField(t *testing.T) {
	req := startTestAPI(t, `{"id":"f1"}`)
	if err := fieldsShowCmd.RunE(fieldsShowCmd, []string{"f1"}); err != nil {
		t.Fatalf("fields show returned %v", err)
	}
	if req.method != "GET" || req.path != "/fields/f1" {
		t.Errorf("request = %s %s", req.method, req.path)
	}
}

func TestFieldsCreateSendsOnlyChangedFlags(t *testing.T) {
	req := startTestAPI(t, `{"id":"f1"}`)
	setFlags(t, fieldsCreateCmd, map[string][]string{
		"name": {"Browser"}, "kind": {"single_option"}, "option": {"Chrome", "Firefox"},
		"required": {"false"}, "display-order": {"3"},
	})

	if err := fieldsCreateCmd.RunE(fieldsCreateCmd, nil); err != nil {
		t.Fatalf("fields create returned %v", err)
	}

	if req.method != "POST" || req.path != "/fields" {
		t.Errorf("request = %s %s", req.method, req.path)
	}
	want := `{"field":{"data":["Chrome","Firefox"],"display_order":3,"is_required":false,"kind":"single_option","name":"Browser"}}`
	got, _ := json.Marshal(req.body)
	if string(got) != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestFieldsUpdateSendsOptionChanges(t *testing.T) {
	req := startTestAPI(t, `{"id":"f1"}`)
	setFlags(t, fieldsUpdateCmd, map[string][]string{
		"add-option": {"Safari"}, "rename-option": {"o1=Edge"}, "remove-option": {"o2"},
		"state": {"inactive"}, "required-when-closing": {"true"},
	})

	if err := fieldsUpdateCmd.RunE(fieldsUpdateCmd, []string{"f1"}); err != nil {
		t.Fatalf("fields update returned %v", err)
	}

	if req.method != "PATCH" || req.path != "/fields/f1" {
		t.Errorf("request = %s %s", req.method, req.path)
	}
	want := `{"field":{"data":[{"id":"o1","label":"Edge"},{"label":"Safari"}],` +
		`"is_required_for_agent_when_closing_ticket":true,"remove_option_ids":["o2"],"state":"inactive"}}`
	got, _ := json.Marshal(req.body)
	if string(got) != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestFieldsUpdateRejectsMalformedRename(t *testing.T) {
	setFlags(t, fieldsUpdateCmd, map[string][]string{"rename-option": {"nolabel"}})
	err := fieldsUpdateCmd.RunE(fieldsUpdateCmd, []string{"f1"})
	if err == nil || !strings.Contains(err.Error(), "--rename-option must be id=Label") {
		t.Errorf("error = %v, want the rename format error", err)
	}
}

func TestFieldsUpdateRejectsInvalidState(t *testing.T) {
	setFlags(t, fieldsUpdateCmd, map[string][]string{"state": {"paused"}})
	err := fieldsUpdateCmd.RunE(fieldsUpdateCmd, []string{"f1"})
	if err == nil || !strings.Contains(err.Error(), "invalid state") {
		t.Errorf("error = %v, want the state error", err)
	}
}

func TestFieldsDeleteSendsDelete(t *testing.T) {
	req := startTestAPI(t, `{}`)
	if err := fieldsDeleteCmd.RunE(fieldsDeleteCmd, []string{"f1"}); err != nil {
		t.Fatalf("fields delete returned %v", err)
	}
	if req.method != "DELETE" || req.path != "/fields/f1" {
		t.Errorf("request = %s %s", req.method, req.path)
	}
}
