package view

import "testing"

func TestCmdKey_isACommandKeyToNineThenAPlace(t *testing.T) {
	for place, want := range map[int]string{1: "⌘1", 9: "⌘9", 10: "place 10"} {
		if got := CmdKey(place); got != want {
			t.Errorf("CmdKey(%d) = %q, want %q", place, got, want)
		}
	}
}

func TestTypedName_isQuotedWhenAShellWouldSplitIt(t *testing.T) {
	for name, want := range map[string]string{"sprint": "sprint", "my sprint": `"my sprint"`} {
		if got := TypedName(name); got != want {
			t.Errorf("TypedName(%q) = %q, want %q", name, got, want)
		}
	}
}
