package view

import "testing"

func TestTypedName_isQuotedWhenAShellWouldSplitIt(t *testing.T) {
	for name, want := range map[string]string{"sprint": "sprint", "my sprint": `"my sprint"`} {
		if got := TypedName(name); got != want {
			t.Errorf("TypedName(%q) = %q, want %q", name, got, want)
		}
	}
}
