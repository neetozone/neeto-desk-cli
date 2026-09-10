package commands

import "testing"

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
