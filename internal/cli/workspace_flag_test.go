package cli

import "testing"

func TestValueOf_readsAFlagBeforeTheCommandLineIsParsed(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"task", "list", "--workspace", "home"}, "home"},
		{[]string{"task", "list", "--workspace=home"}, "home"},
		{[]string{"task", "list", "--workspace"}, ""},
		{[]string{"task", "new", "--", "--workspace", "home"}, ""},
		{[]string{"task", "list"}, ""},
	} {
		if got := valueOf(workspaceFlag, c.args); got != c.want {
			t.Errorf("valueOf(%v) = %q, want %q", c.args, got, c.want)
		}
	}
}
