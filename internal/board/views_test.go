package board

import (
	"cairn/internal/engine"
	"slices"
	"testing"
	"time"
)

func keptAView(t *testing.T, e *engine.Engine, name string, a request) {
	t.Helper()
	a.given["name"] = name
	if _, err := e.Ask(engine.ViewNew, a.input()); err != nil {
		t.Fatalf("view new %s: %v", name, err)
	}
}

func TestLook_readsEveryViewWithTheTasksItKeeps(t *testing.T) {
	e := homeTree(t)
	keptAView(t, e, "Home", ask().with("objective", "1"))
	keptAView(t, e, "Relay", ask("search", "relay"))
	s := lookAt(t, e)
	if len(s.Views) != 2 {
		t.Fatalf("the snapshot holds %d views, want 2", len(s.Views))
	}
	home, relay := s.Views[0], s.Views[1]
	if home.Name != "Home" || home.Place != 1 || !slices.Equal(home.Kept, []int{1, 2}) {
		t.Errorf("the first view reads %+v", home)
	}
	if relay.Name != "Relay" || relay.Place != 2 || !slices.Equal(relay.Kept, []int{2}) {
		t.Errorf("the second view reads %+v", relay)
	}
	if v, found := s.View("relay"); !found || v.Name != "Relay" {
		t.Errorf("a view looked for by its name in another case reads %+v, %v", v, found)
	}
	if _, found := s.View("nowhere"); found {
		t.Error("a view nobody kept was found")
	}
	if got := filteredNumbers(s.Viewed("Relay", s.Cards)); got != "2" {
		t.Errorf("the cards the view Relay keeps are %q", got)
	}
	if got := filteredNumbers(s.Viewed("", s.Cards)); got != "1 2" {
		t.Errorf("no view keeps %q, want every card", got)
	}
}

func TestLook_saysAViewKeepingNothingWhenEveryConditionIsGone(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(engine.TaskNew, engine.Input{Given: map[string]string{"title": "Buy seeds"}, Listed: map[string][]string{"label": {"garden"}}}); err != nil {
		t.Fatal(err)
	}
	keptAView(t, e, "Garden", ask().with("label", "garden"))
	if _, err := e.Ask(engine.LabelDelete, engine.Input{Given: map[string]string{"label": "garden"}, Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	s := lookAt(t, e)
	if v := s.Views[0]; len(v.Kept) != 0 || !slices.Equal(v.Gone, []string{"label garden"}) {
		t.Errorf("a view whose every condition is gone reads %+v", v)
	}
	if got := GoneSays(s.Views[0].View); got != "label garden is gone and narrows nothing" {
		t.Errorf("what is gone reads %q", got)
	}
	if got := GoneSays(engine.View{}); got != "" {
		t.Errorf("nothing gone reads %q", got)
	}
	if got := GoneSays(engine.View{Gone: []string{"label a", "objective 2"}}); got != "label a and objective 2 are gone and narrow nothing" {
		t.Errorf("two gone read %q", got)
	}
}

func TestLook_failsWhenTheViewsCannotBeRead(t *testing.T) {
	f := newFake()
	f.refuse = map[string]error{"view list": errNope}
	if _, err := Look(f); err == nil {
		t.Error("the views could not be read, and Look went on")
	}
}

func TestTouching_namesEveryWayATaskLeftOutTouchesOneShown(t *testing.T) {
	all := []Card{
		{Number: 1, Title: "Order the relay"},
		{Number: 2, Blockers: []int{1}},
		{Number: 3, Blockers: []int{2}, Linked: []int{2}},
		{Number: 4, Linked: []int{9}},
		{Number: 5},
	}
	shown := []Card{all[1]}
	refs := Touching(shown, all)
	if len(refs) != 2 || refs[0].Card.Number != 1 || refs[1].Card.Number != 3 {
		t.Fatalf("the references are %+v", refs)
	}
	if got := refs[0].Says(); got != "outside this view · blocks #2" {
		t.Errorf("a blocker says %q", got)
	}
	if got := refs[1].Says(); got != "outside this view · waits on #2 · linked to #2" {
		t.Errorf("a task waiting on and linked to one shown says %q", got)
	}
	if got := refs[1].Touches(""); got != "waits on 2 · linked to 2" {
		t.Errorf("what it touches reads %q", got)
	}
	twice := Touching([]Card{all[1], {Number: 6, Blockers: []int{1}}}, append(all, Card{Number: 6, Blockers: []int{1}}))
	if got := twice[0].Says(); got != "outside this view · blocks #2, #6" {
		t.Errorf("a blocker of two says %q", got)
	}
	if got := Touching([]Card{{Number: 7, Linked: []int{5}}}, all); len(got) != 1 || got[0].Says() != "outside this view · linked to #7" {
		t.Errorf("a task linked from one shown reads %+v", got)
	}
}

func TestFilters_holdWhatAViewKeeps(t *testing.T) {
	s := lookAt(t, homeTree(t))
	s.Cards = append(s.Cards, Card{Number: 9, Labels: []string{"backend"}})
	fs := FiltersOf(s, time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)).On("priority", "high").On("objective", "1 itself").On("deadline-before", "2026-11-16")
	fs = fs.Toggled(2, "backend")
	in := fs.Holding(engine.View{}, "relay  board")
	if in.Given["priority"] != "high" || in.Given["deadline-before"] != "2026-11-16" || in.Given["itself"] != "true" || in.Given["search"] != "relay board" {
		t.Errorf("the filters hold %+v", in.Given)
	}
	if !slices.Equal(in.Listed["objective"], []string{"1"}) || !slices.Equal(in.Listed["label"], []string{"backend"}) {
		t.Errorf("the filters list %+v", in.Listed)
	}
	if got := fs.Holding(engine.View{}, "").Given["search"]; got != "" {
		t.Errorf("no search holds %q", got)
	}
	if got := fs.Proposed(s, ""); got != "high · backend · Home automation · 2026-11-16" {
		t.Errorf("the name proposed is %q", got)
	}
	if got := FiltersOf(s, time.Now()).Proposed(s, "relay"); got != "relay" {
		t.Errorf("a search alone proposes %q", got)
	}
}

func TestFilters_holdAViewAndWhatIsOnTopAsANewView(t *testing.T) {
	s := lookAt(t, homeTree(t))
	over := engine.View{Name: "Home", Holds: engine.Holds{State: "todo", Priority: "high", DueBefore: "+7d", Search: "pin", Label: []string{"area"}, Objective: []int{1}, Itself: true}}
	in := FiltersOf(s, time.Now()).Holding(over, "")
	if in.Named("view") || in.Given["state"] != "todo" || in.Given["priority"] != "high" || in.Given["deadline-before"] != "+7d" || in.Given["search"] != "pin" || in.Given["itself"] != "true" {
		t.Errorf("a view held as a new one gives %+v", in.Given)
	}
	if !slices.Equal(in.Listed["label"], []string{"area"}) || !slices.Equal(in.Listed["objective"], []string{"1"}) {
		t.Errorf("a view held as a new one lists %+v", in.Listed)
	}
	on := FiltersOf(s, time.Now()).On("state", "current").Holding(over, "map")
	if on.Given["state"] != "current" || on.Given["search"] != "pin map" {
		t.Errorf("a filter on top held as a new view gives %+v", on.Given)
	}
}

func TestFilters_saveWhatIsOnTopIntoAView(t *testing.T) {
	s := lookAt(t, homeTree(t))
	fs := FiltersOf(s, time.Now()).On("objective", "2").On("state", "todo")
	v := engine.View{Name: "Home", Holds: engine.Holds{Label: []string{"area"}, Objective: []int{1}, State: "backlog", Search: "pin"}}
	in := fs.Saving(v, "map")
	if in.Given["view"] != "Home" || in.Given["state"] != "todo" || in.Given["search"] != "pin map" {
		t.Errorf("saving gives %+v", in.Given)
	}
	if !slices.Equal(in.Listed["objective"], []string{"1", "2"}) || in.Named("label") {
		t.Errorf("saving lists %+v", in.Listed)
	}
	if got := fs.Saving(engine.View{Name: "Home"}, "").Given["search"]; got != "" {
		t.Errorf("no search on top saves %q", got)
	}
	s.Cards = append(s.Cards, Card{Number: 9, Labels: []string{"backend"}})
	labelled := FiltersOf(s, time.Now()).Toggled(2, "backend")
	if got := labelled.Saving(v, "").Listed["label"]; !slices.Equal(got, []string{"area", "backend"}) {
		t.Errorf("a label on top saves %v", got)
	}
}

func TestHolds_sayWhichFilterAViewHolds(t *testing.T) {
	h := engine.Holds{State: "todo", Label: []string{"x"}}
	for arg, want := range map[string]bool{"state": true, "label": true, "priority": false, "objective": false, "deadline-before": false} {
		if got := Narrows(h, arg); got != want {
			t.Errorf("Narrows(%s) = %v", arg, got)
		}
	}
	if !Narrows(engine.Holds{Priority: "high", Objective: []int{1}, DueBefore: "+7d"}, "deadline-before") {
		t.Error("a day due by is not held")
	}
}

func TestHolds_narrowsByTheFocusOnlyWhenOneIsHeld(t *testing.T) {
	if Narrows((engine.Holds{}), "focus") || !Narrows((engine.Holds{Focus: "in"}), "focus") {
		t.Error("a view narrows by the focus whether it holds one or not")
	}
}

func heldFilters(t *testing.T) Filters {
	t.Helper()
	return FiltersOf(Snapshot{Cards: []Card{{Number: 1, State: "todo", Labels: []string{"x"}}}}, time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC))
}

func TestHolding_theFocusFilteredWinsOverTheOneHeld(t *testing.T) {
	in := on(heldFilters(t), "focus", "in").Holding(engine.View{Name: "Home", Holds: engine.Holds{Focus: "out"}}, "")
	if in.Given["focus"] != "true" || in.Given["no-focus"] != "" {
		t.Errorf("the focus kept is %v", in.Given)
	}
}
