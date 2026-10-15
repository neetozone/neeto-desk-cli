package commands

import (
	"encoding/json"
	"testing"
)

func TestCustomersListFlagsAreRegistered(t *testing.T) {
	for _, name := range []string{"email", "page", "page-size"} {
		if customersListCmd.Flags().Lookup(name) == nil {
			t.Errorf("customers list is missing the --%s flag", name)
		}
	}
}

func TestCustomersShowTakesExactlyOneArgument(t *testing.T) {
	if err := customersShowCmd.Args(customersShowCmd, nil); err == nil {
		t.Error("customers show accepted no arguments")
	}
	if err := customersShowCmd.Args(customersShowCmd, []string{"a", "b"}); err == nil {
		t.Error("customers show accepted two arguments")
	}
	if err := customersShowCmd.Args(customersShowCmd, []string{"a"}); err != nil {
		t.Errorf("customers show rejected one argument: %v", err)
	}
}

func TestCustomersSubcommandsAreRegistered(t *testing.T) {
	want := map[string]bool{"list": false, "show": false, "create": false}
	for _, sub := range customersCmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("customers %s is not registered", name)
		}
	}
}

func TestCustomersCreateSendsRepeatedEmailsPhonesAndLinks(t *testing.T) {
	req := startTestAPI(t, `{"id":"1"}`)
	setFlags(t, customersCreateCmd, map[string][]string{
		"email": {"a@acme.com", "b@acme.com"}, "phone": {"+1555", "+1666"}, "link": {"https://acme.com"},
	})

	if err := customersCreateCmd.RunE(customersCreateCmd, nil); err != nil {
		t.Fatalf("customers create returned %v", err)
	}

	want := `{"customer":{"emails":[{"email":"a@acme.com","primary":true},{"email":"b@acme.com","primary":false}],` +
		`"links":["https://acme.com"],"phones":[{"label":"","phone":"+1555"},{"label":"","phone":"+1666"}]}}`
	got, _ := json.Marshal(req.body)
	if string(got) != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
