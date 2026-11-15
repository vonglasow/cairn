package engine

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestAuthor_isTheOneNamedOrTheOneLoggedIn(t *testing.T) {
	named := Home{Lookup: func(k string) (string, bool) { return "claude session 7f3a", k == AuthorVariable }}
	if got := author(named); got != "claude session 7f3a" {
		t.Errorf("named, the author is %q", got)
	}
	me, err := user.Current()
	if err != nil {
		t.Skip("no user to compare with:", err)
	}
	if got := author(Home{}); got != me.Username {
		t.Errorf("unnamed, the author is %q, want %q", got, me.Username)
	}
}

func TestActs_sayWhoDidThemAndThroughWhat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := doThere(TaskNew, Places{Store: path}, given("title", "Write the spec"), Acting{Through: "gui", By: "sam"}); err != nil {
		t.Fatal(err)
	}
	if _, err := doThere(TaskNew, Places{Store: path}, given("title", "Ship it"), Acting{Through: "mcp", By: "claude-code"}); err != nil {
		t.Fatal(err)
	}
	shown, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, act := range shown.Acts {
		if c := act.Carried(); c.By != "sam" || c.Through != "gui" {
			t.Errorf("an act of task 1 says %q through %q", c.By, c.Through)
		}
	}
	listed, err := Do(traceList, path, given("by", "claude-code"))
	if err != nil {
		t.Fatal(err)
	}
	for _, act := range listed.Since {
		if c := act.Carried(); c.By != "claude-code" {
			t.Errorf("listing what claude-code did, an act of %q came", c.By)
		}
	}
	if len(listed.Since) == 0 {
		t.Error("nothing claude-code did is listed")
	}
}
