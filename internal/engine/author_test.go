package engine

import (
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/trace"
)

func TestAuthor_isTheOneNamedOrTheOneLoggedIn(t *testing.T) {
	named := Home{Lookup: func(k string) (string, bool) { return "claude session 7f3a", k == AuthorVariable }}
	if got := Author(named); got != "claude session 7f3a" {
		t.Errorf("named, the author is %q", got)
	}
	me, err := user.Current()
	if err != nil {
		t.Skip("no user to compare with:", err)
	}
	if got := Author(Home{}); got != me.Username {
		t.Errorf("unnamed, the author is %q, want %q", got, me.Username)
	}
}

func TestActs_sayWhoDidThemAndThroughWhat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := DoAt(TaskNew, Places{Store: path}, given("title", "Write the spec"), Acting{Through: "gui", By: "sam"}); err != nil {
		t.Fatal(err)
	}
	if _, err := DoAt(TaskNew, Places{Store: path}, given("title", "Ship it"), Acting{Through: "mcp", By: "claude-code"}); err != nil {
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
	listed, err := Do(TraceList, path, given("by", "claude-code"))
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

func TestHold_writesDownWhatItIsAskedAsDoneThroughItsSurface(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	h := Home{Dir: t.TempDir(), Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}
	held, err := Hold(h, Acting{Through: "tui", By: "sam"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := held.Ask(TaskNew, given("title", "Write the spec")); err != nil {
		t.Fatal(err)
	}
	_ = held.Close()
	shown, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if c := shown.Acts[0].Carried(); c.Through != "tui" || c.By != "sam" {
		t.Errorf("an act from the board says %q through %q", c.By, c.Through)
	}
}

func TestSaidOf_aBoardLineSaysWhoDidTheAct(t *testing.T) {
	var a trace.AsTold
	a.On(1)
	a.Did(trace.FocusTaken)
	a.Through("gui")
	a.By("sam")
	if got := SaidOf(a.Act()).Line(); !strings.HasSuffix(got, "(by sam, through gui)") {
		t.Errorf("a board reads %q", got)
	}
}

func TestTraceList_showsReadingOneAuthorWithAnExample(t *testing.T) {
	if !strings.Contains(TraceList.Example, "cairn trace list --by sam") {
		t.Fatalf("the example does not show reading one author: %q", TraceList.Example)
	}
}
