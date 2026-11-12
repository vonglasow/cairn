package engine

import (
	"cairn/internal/store"
	"errors"
	"slices"
	"strings"
	"testing"
)

type request struct {
	given  map[string]string
	listed map[string][]string
}

func ask(pairs ...string) request {
	a := request{given: map[string]string{}, listed: map[string][]string{}}
	for i := 0; i+1 < len(pairs); i += 2 {
		a.given[pairs[i]] = pairs[i+1]
	}
	return a
}

func (a request) with(name string, values ...string) request {
	a.listed[name] = values
	return a
}

func (a request) input() Input { return Input{Given: a.given, Listed: a.listed, Confirmed: true} }

func did(t *testing.T, path string, op Operation, a request) Answer {
	t.Helper()
	answer, err := Do(op, path, a.input())
	if err != nil {
		t.Fatalf("%s: %v", op.Line(), err)
	}
	return answer
}

func refusedAs(t *testing.T, path string, op Operation, a request, reason string) string {
	t.Helper()
	_, err := Do(op, path, a.input())
	if err == nil {
		t.Fatalf("%s %v: it went ahead, want it refused as %s", op.Line(), a, reason)
	}
	if got := RefusalOf(err).Reason; got != reason {
		t.Errorf("%s %v: refused as %q (%v), want %q", op.Line(), a, got, err, reason)
	}
	return err.Error()
}

func listedIn(t *testing.T, path string, a request) (Tasks, string) {
	t.Helper()
	answer := did(t, path, TaskList, a)
	return Of(TaskList, answer).(Tasks), RenderText(TaskList, answer)
}

func rowsListed(listed Tasks) []int {
	var out []int
	for _, row := range listed.Tasks {
		out = append(out, row.Number)
	}
	return out
}

func viewed(t *testing.T, path string, op Operation, a request) (View, string) {
	t.Helper()
	answer := did(t, path, op, a)
	return Of(op, answer).(View), RenderText(op, answer)
}

func homeStore(t *testing.T) string {
	t.Helper()
	path := storeAt(t)
	did(t, path, ObjectiveNew, ask("title", "Home automation"))
	did(t, path, ObjectiveNew, ask("title", "Firmware", "parent", "1"))
	did(t, path, TaskNew, ask("title", "Wiring diagram").with("objective", "1").with("label", "doc"))
	did(t, path, TaskNew, ask("title", "Read the sensors").with("objective", "2"))
	did(t, path, TaskNew, ask("title", "Buy milk", "state", "todo").with("label", "shopping"))
	return path
}

func TestViewNew_keepsTheFiltersAndTaskListOpensThemAgain(t *testing.T) {
	path := homeStore(t)
	kept, text := viewed(t, path, ViewNew, ask("name", " Home automation ").with("objective", "1"))
	if kept.Name != "Home automation" || kept.Place != 1 || kept.Keeps != 2 || kept.Of != 3 || !slices.Equal(kept.Holds.Objective, []int{1}) {
		t.Errorf("keeping the view answers %+v", kept)
	}
	if text != "kept view Home automation · ⌘1 · keeps 2 of 3 tasks\n" {
		t.Errorf("keeping the view reads %q", text)
	}
	listed, _ := listedIn(t, path, ask("view", "HOME AUTOMATION"))
	if !slices.Equal(rowsListed(listed), []int{1, 2}) || listed.View != "Home automation" {
		t.Errorf("the view lists %v under %q", rowsListed(listed), listed.View)
	}
}

func TestViewNew_holdsEveryFilterTaskListTakesAndASearch(t *testing.T) {
	path := homeStore(t)
	did(t, path, TaskEdit, ask("number", "1", "deadline", "2026-11-20", "priority", "high"))
	did(t, path, TaskNew, ask("title", "Old wiring", "state", "done").with("objective", "1"))
	did(t, path, TaskArchive, ask("number", "4"))
	cases := []struct {
		name  string
		asked request
		want  []int
	}{
		{"state", ask("name", "a", "state", "todo"), []int{3}},
		{"priority", ask("name", "b", "priority", "high"), []int{1}},
		{"due", ask("name", "c", "deadline-before", "2026-11-30"), []int{1}},
		{"itself", ask("name", "d", "itself", "true").with("objective", "1"), []int{1}},
		{"archived", ask("name", "e", "archived", "true"), []int{4}},
		{"with archived", ask("name", "f", "with-archived", "true").with("objective", "1"), []int{1, 2, 4}},
		{"search", ask("name", "g", "search", "  sensors  "), []int{2}},
		{"labels", ask("name", "h").with("label", "Doc"), []int{1}},
	}
	for _, c := range cases {
		kept, _ := viewed(t, path, ViewNew, c.asked)
		listed, _ := listedIn(t, path, ask("view", c.asked.given["name"]))
		if !slices.Equal(rowsListed(listed), c.want) {
			t.Errorf("%s: the view lists %v, want %v", c.name, rowsListed(listed), c.want)
		}
		if kept.Keeps != len(c.want) {
			t.Errorf("%s: the view keeps %d, want %d", c.name, kept.Keeps, len(c.want))
		}
	}
	views := Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if len(views.Views) != len(cases) || views.Views[7].Place != 8 || views.Views[6].Holds.Search != "sensors" || !slices.Equal(views.Views[7].Holds.Label, []string{"doc"}) {
		t.Errorf("the views read %+v", views)
	}
	text := RenderText(ViewList, did(t, path, ViewList, ask()))
	for _, said := range []string{"state todo", "priority high", "deadline before 2026-11-30", "objective 1 itself", "archived", "with archived", `search "sensors"`, "label doc"} {
		if !strings.Contains(text, said) {
			t.Errorf("the views read %q, want %q", text, said)
		}
	}
}

func TestViewNew_refusesWhatCannotBeAView(t *testing.T) {
	path := homeStore(t)
	did(t, path, ViewNew, ask("name", "Groceries", "state", "todo"))
	if said := refusedAs(t, path, ViewNew, ask("name", "grocéries", "priority", "high"), "view-name-taken"); said != "a view is already named Groceries" {
		t.Errorf("the name taken reads %q", said)
	}
	if said := refusedAs(t, path, ViewNew, ask("name", "Everything", "with-archived", "true"), "view-holds-nothing"); said != "a view holds at least one filter or a search" {
		t.Errorf("a view holding nothing reads %q", said)
	}
	refusedAs(t, path, ViewNew, ask("name", "Nothing", "search", "  "), "view-holds-nothing")
	refusedAs(t, path, ViewNew, ask("name", "Gone").with("objective", "9"), "no-such-objective")
	refusedAs(t, path, ViewNew, ask("name", "Wrong").with("objective", "x"), "objective-required")
	refusedAs(t, path, ViewNew, ask("name", "  ", "state", "todo"), "title-required")
	refusedAs(t, path, ViewNew, ask("name", "Odd", "state", "sideways"), "not-a-state")
	refusedAs(t, path, ViewNew, ask("name", "Odd").with("label", "a:b"), "label-has-no-family")
	if views := Of(ViewList, did(t, path, ViewList, ask())).(Views); len(views.Views) != 1 {
		t.Errorf("a refused view was kept: %+v", views)
	}
}

func TestViews_aViewNobodyKeptIsRefusedByName(t *testing.T) {
	path := homeStore(t)
	for _, op := range []Operation{ViewShow, ViewDelete} {
		if said := refusedAs(t, path, op, ask("view", "Gardn"), "no-such-view"); said != "no view is named Gardn" {
			t.Errorf("%s reads %q", op.Line(), said)
		}
	}
	refusedAs(t, path, ViewEdit, ask("view", "Gardn", "state", "todo"), "no-such-view")
	refusedAs(t, path, TaskList, ask("view", "Gardn"), "no-such-view")
}

func TestTaskList_flagsBesideAViewNarrowFurther(t *testing.T) {
	path := homeStore(t)
	did(t, path, ViewNew, ask("name", "Home").with("objective", "1"))
	listed, _ := listedIn(t, path, ask("view", "Home").with("label", "doc"))
	if !slices.Equal(rowsListed(listed), []int{1}) {
		t.Errorf("a label beside the view lists %v", rowsListed(listed))
	}
	listed, _ = listedIn(t, path, ask("view", "Home", "state", "todo"))
	if len(listed.Tasks) != 0 {
		t.Errorf("a column beside the view lists %v", rowsListed(listed))
	}
	answer := did(t, path, TaskList, ask("view", "Home", "count", "state"))
	if said := Of(TaskList, answer).(Tasks); *said.Total != 2 || len(said.Outside) != 0 {
		t.Errorf("counted, the view says %+v", said)
	}
}

func TestTaskList_aViewNamesWhatLiesOutsideItOnceAndCountsNoneOfIt(t *testing.T) {
	path := homeStore(t)
	did(t, path, TaskNew, ask("title", "Upgrade the printer"))
	did(t, path, TaskNew, ask("title", "Publish the diagram"))
	did(t, path, TaskNew, ask("title", "Archived elsewhere"))
	did(t, path, TaskEdit, ask("number", "1").with("blocker", "4").with("link", "4"))
	did(t, path, TaskEdit, ask("number", "5").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "6").with("link", "1"))
	did(t, path, TaskArchive, ask("number", "6"))
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, ViewNew, ask("name", "Docs").with("label", "doc"))
	listed, text := listedIn(t, path, ask("view", "Docs"))
	if !slices.Equal(rowsListed(listed), []int{1}) {
		t.Fatalf("the view lists %v", rowsListed(listed))
	}
	want := []Outside{
		{Number: 2, Title: "Read the sensors", State: "backlog", Why: []Touch{{Is: "waits on", Of: 1}}},
		{Number: 4, Title: "Upgrade the printer", State: "backlog", Why: []Touch{{Is: "blocks", Of: 1}, {Is: "linked to", Of: 1}}},
		{Number: 5, Title: "Publish the diagram", State: "backlog", Why: []Touch{{Is: "waits on", Of: 1}}},
	}
	if !slices.EqualFunc(listed.Outside, want, func(a, b Outside) bool {
		return a.Number == b.Number && a.Title == b.Title && a.State == b.State && slices.Equal(a.Why, b.Why)
	}) {
		t.Errorf("outside reads %+v, want %+v", listed.Outside, want)
	}
	for _, said := range []string{"#1  Wiring diagram", "outside this view, never counted:\n", "#4  Upgrade the printer  · backlog · blocks #1 · linked to #1\n", "#2  Read the sensors  · backlog · waits on #1\n"} {
		if !strings.Contains(text, said) {
			t.Errorf("the listing reads %q, want %q", text, said)
		}
	}
	plain, plainText := listedIn(t, path, ask().with("label", "doc"))
	if len(plain.Outside) != 0 || strings.Contains(plainText, "outside") {
		t.Errorf("without a view, outside reads %+v", plain.Outside)
	}
	shown, shownText := viewed(t, path, ViewShow, ask("view", "docs"))
	if len(shown.Outside) != 3 || shown.Keeps != 1 || shown.Of != 5 {
		t.Errorf("showing the view answers %+v", shown)
	}
	for _, said := range []string{"view     Docs  (⌘1)\n", "holds    label doc\n", "keeps    1 of 5 tasks · 0 done\n", "outside  #2 waits on #1 · #4 blocks #1, linked to #1 · #5 waits on #1\n"} {
		if !strings.Contains(shownText, said) {
			t.Errorf("showing the view reads %q, want %q", shownText, said)
		}
	}
}

func TestViewEdit_renamesChangesAndMovesAView(t *testing.T) {
	path := homeStore(t)
	did(t, path, ViewNew, ask("name", "First", "state", "todo"))
	did(t, path, ViewNew, ask("name", "Second", "state", "backlog").with("objective", "1"))
	did(t, path, ViewNew, ask("name", "Third", "state", "done"))
	changed, text := viewed(t, path, ViewEdit, ask("view", "second", "name", "Docs", "state", "").with("label", "doc").with("objective"))
	if changed.Name != "Docs" || changed.Was != "Second" || changed.Holds.State != "" || len(changed.Holds.Objective) != 0 || changed.Keeps != 1 {
		t.Errorf("the view changed answers %+v", changed)
	}
	if text != "changed view Docs (was Second) · ⌘2 · keeps 1 of 3 tasks\n" {
		t.Errorf("the view changed reads %q", text)
	}
	moved, text := viewed(t, path, ViewEdit, ask("view", "Third", "place", "1"))
	if moved.Place != 1 || moved.Was != "" || text != "changed view Third · ⌘1 · keeps 0 of 3 tasks\n" {
		t.Errorf("the view moved answers %+v, %q", moved, text)
	}
	views := Of(ViewList, did(t, path, ViewList, ask())).(Views)
	var order []string
	for _, v := range views.Views {
		order = append(order, v.Name)
	}
	if !slices.Equal(order, []string{"Third", "First", "Docs"}) {
		t.Errorf("the views stand %v", order)
	}
	did(t, path, ViewEdit, ask("view", "Third", "place", "3"))
	did(t, path, ViewEdit, ask("view", "Third", "name", "third"))
	refusedAs(t, path, ViewEdit, ask("view", "Third"), "nothing-to-change")
	refusedAs(t, path, ViewEdit, ask("view", "Third", "place", "4"), "no-such-place")
	refusedAs(t, path, ViewEdit, ask("view", "Third", "place", "0"), "no-such-place")
	refusedAs(t, path, ViewEdit, ask("view", "Third", "name", "First"), "view-name-taken")
	refusedAs(t, path, ViewEdit, ask("view", "Third", "state", ""), "view-holds-nothing")
	refusedAs(t, path, ViewEdit, ask("view", "Third", "state", "sideways"), "not-a-state")
	views = Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if views.Views[2].Name != "third" || views.Views[2].Holds.State != "done" {
		t.Errorf("after the refusals the views read %+v", views.Views)
	}
}

func TestViewDelete_closesThePlacesAndTouchesNoTask(t *testing.T) {
	path := homeStore(t)
	before := len(Of(TraceList, did(t, path, TraceList, ask())).(Acts).Acts)
	did(t, path, ViewNew, ask("name", "First", "state", "todo"))
	did(t, path, ViewNew, ask("name", "Second", "state", "backlog"))
	did(t, path, ViewEdit, ask("view", "Second", "name", "Next"))
	did(t, path, ViewEdit, ask("view", "Next", "name", "Second"))
	answer := did(t, path, ViewDelete, ask("view", "first"))
	if said := Of(ViewDelete, answer).(ViewDeleted); said.Deleted != "First" {
		t.Errorf("deleting answers %+v", said)
	}
	if text := RenderText(ViewDelete, answer); text != "deleted view First · no task changed\n" {
		t.Errorf("deleting reads %q", text)
	}
	views := Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if len(views.Views) != 1 || views.Views[0].Name != "Second" || views.Views[0].Place != 1 {
		t.Errorf("the views left read %+v", views.Views)
	}
	acts := Of(TraceList, did(t, path, TraceList, ask())).(Acts)
	if len(acts.Acts) != before {
		t.Errorf("the trace holds %d acts, want the %d it held before any view", len(acts.Acts), before)
	}
	if text := RenderText(ViewDelete, did(t, path, ViewDelete, ask("view", "Second"))); text != "deleted view Second · no task changed\n" {
		t.Errorf("deleting the last reads %q", text)
	}
	if text := RenderText(ViewList, did(t, path, ViewList, ask())); text != "there is no view\n" {
		t.Errorf("no view reads %q", text)
	}
}

func TestViews_followALabelRenamedAndSayALabelDeletedIsGone(t *testing.T) {
	path := homeStore(t)
	did(t, path, ViewNew, ask("name", "Docs", "state", "backlog").with("label", "doc", "release"))
	did(t, path, ViewNew, ask("name", "Only").with("label", "doc"))
	labels := Of(LabelList, did(t, path, LabelList, ask())).(Labels)
	if !slices.ContainsFunc(labels.Labels, func(l LabelCarried) bool { return l.Label == "release" }) {
		t.Errorf("a label a view holds is not kept: %+v", labels)
	}
	did(t, path, LabelRename, ask("label", "doc", "to", "docs"))
	views := Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if !slices.Equal(views.Views[0].Holds.Label, []string{"docs", "release"}) || views.Views[0].Keeps != 0 {
		t.Errorf("after the rename the view holds %+v", views.Views[0])
	}
	deleted := did(t, path, LabelDelete, ask("label", "docs"))
	if said := RenderText(LabelDelete, deleted); !strings.Contains(said, "warning: 2 views hold the label docs; it narrows nothing there now\n") {
		t.Errorf("deleting the label reads %q", said)
	}
	if said := Of(LabelDelete, deleted).(Labels); len(said.Warnings) != 1 {
		t.Errorf("deleting the label answers %+v", said)
	}
	did(t, path, LabelDelete, ask("label", "release"))
	listed, text := listedIn(t, path, ask("view", "Docs"))
	if !slices.Equal(rowsListed(listed), []int{1, 2}) || !slices.Equal(listed.Gone, []string{"label docs", "label release"}) {
		t.Errorf("a view with labels gone lists %v, gone %v", rowsListed(listed), listed.Gone)
	}
	if !strings.HasPrefix(text, "view Docs: label docs and label release are gone and narrow nothing\n  cairn view edit Docs, or cairn view delete Docs\n") {
		t.Errorf("a view with labels gone reads %q", text)
	}
	listed, text = listedIn(t, path, ask("view", "Only"))
	if len(listed.Tasks) != 0 || !slices.Equal(listed.Gone, []string{"label docs"}) {
		t.Errorf("a view with every condition gone lists %v, gone %v", rowsListed(listed), listed.Gone)
	}
	if !strings.HasPrefix(text, "every condition view Only holds is gone, so it keeps nothing: label docs\n  cairn view edit Only, or cairn view delete Only\n") {
		t.Errorf("a view with every condition gone reads %q", text)
	}
	views = Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if views.Views[1].Keeps != 0 || !slices.Equal(views.Views[1].Gone, []string{"label docs"}) {
		t.Errorf("the view gone reads %+v", views.Views[1])
	}
	if text := RenderText(ViewList, did(t, path, ViewList, ask())); !strings.Contains(text, "label docs, gone") {
		t.Errorf("the views read %q", text)
	}
	if text := RenderText(ViewShow, did(t, path, ViewShow, ask("view", "Only"))); !strings.Contains(text, "gone     label docs\n") {
		t.Errorf("the view shown reads %q", text)
	}
	if said := RenderText(LabelDelete, did(t, path, LabelDelete, ask("label", "shopping"))); strings.Contains(said, "warning") {
		t.Errorf("a label no view holds reads %q", said)
	}
}

func TestObjectiveDelete_asksWhatBecomesOfTheViewsNamingIt(t *testing.T) {
	path := homeStore(t)
	did(t, path, ObjectiveNew, ask("title", "Garden"))
	did(t, path, ViewNew, ask("name", "Home", "state", "backlog").with("objective", "1"))
	did(t, path, ViewNew, ask("name", "Firmware").with("objective", "2"))
	did(t, path, ViewNew, ask("name", "Garden").with("objective", "3"))
	did(t, path, ViewNew, ask("name", "Next", "state", "todo"))
	said := refusedAs(t, path, ObjectiveDelete, ask("number", "3"), "objective-named-by-a-view")
	if said != "objective 3: a view names it, and deleting it asks what goes with them:\n  view Garden\nsay it with --views off|delete" {
		t.Errorf("the refusal reads %q", said)
	}
	said = refusedAs(t, path, ObjectiveDelete, ask("number", "1", "sub-objectives", "delete"), "objective-carries-tasks")
	if !strings.Contains(said, "carries tasks and a view names it,") || !strings.Contains(said, "  view Home\n  view Firmware\nsay it with --tasks detach|delete; and --views off|delete") {
		t.Errorf("the refusal reads %q", said)
	}
	said = refusedAs(t, path, ObjectiveDelete, ask("number", "1", "tasks", "detach", "views", "off"), "objective-carries-tasks")
	if strings.Contains(said, "view") {
		t.Errorf("the views decided, the refusal reads %q", said)
	}
	said = refusedAs(t, path, ObjectiveDelete, ask("number", "1", "sub-objectives", "keep", "tasks", "detach"), "objective-named-by-a-view")
	if !strings.HasPrefix(said, "objective 1: a view names it,") {
		t.Errorf("the tasks decided, the refusal reads %q", said)
	}
	did(t, path, ObjectiveDelete, ask("number", "1", "sub-objectives", "delete", "tasks", "detach", "views", "off"))
	views := Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if len(views.Views) != 4 || len(views.Views[0].Holds.Objective) != 0 || views.Views[0].Holds.State != "backlog" {
		t.Errorf("taken off, the views read %+v", views.Views)
	}
	if !slices.Equal(views.Views[1].Holds.Objective, []int{2}) || !slices.Equal(views.Views[1].Gone, []string{"objective 2"}) {
		t.Errorf("a view holding nothing else keeps the objective gone: %+v", views.Views[1])
	}
	if text := RenderText(ViewList, did(t, path, ViewList, ask())); !strings.Contains(text, "objective 2, gone") {
		t.Errorf("the views read %q", text)
	}
	did(t, path, ObjectiveDelete, ask("number", "3", "views", "delete"))
	views = Of(ViewList, did(t, path, ViewList, ask())).(Views)
	if len(views.Views) != 3 || views.Views[2].Name != "Next" || views.Views[2].Place != 3 {
		t.Errorf("deleted with it, the views read %+v", views.Views)
	}
	refusedAs(t, path, ObjectiveDelete, ask("number", "9", "views", "sideways"), "not-one-of")
}

func TestViewShow_givesEachObjectiveItNamesWithTheStoresFigures(t *testing.T) {
	path := homeStore(t)
	did(t, path, TaskEdit, ask("number", "1", "state", "done"))
	did(t, path, ObjectiveNew, ask("title", "Garden"))
	did(t, path, ViewNew, ask("name", "Home", "state", "backlog").with("objective", "1", "3"))
	shown, text := viewed(t, path, ViewShow, ask("view", "home"))
	if len(shown.Objectives) != 2 {
		t.Fatalf("the view shown gives the objectives %+v", shown.Objectives)
	}
	home := shown.Objectives[0]
	if home.Number != 1 || home.Title != "Home automation" || home.Progress == nil || home.Progress.Tasks.Done != 1 || home.Progress.Tasks.Total != 2 || home.Progress.Percent != 50 {
		t.Errorf("the view gives its objective as %+v", home)
	}
	if garden := shown.Objectives[1]; garden.Number != 3 || garden.Progress != nil {
		t.Errorf("the view gives an objective nothing serves as %+v", garden)
	}
	for _, said := range []string{"objective #1 Home automation  1/2 · 50% (the store's)\n", "objective #3 Garden  nothing serves it yet\n"} {
		if !strings.Contains(text, said) {
			t.Errorf("the view shown reads %q, want %q", text, said)
		}
	}
	kept, _ := viewed(t, path, ViewNew, ask("name", "Other").with("objective", "1"))
	if len(kept.Objectives) != 0 {
		t.Errorf("keeping a view gives the objectives %+v, which only showing it gives", kept.Objectives)
	}
}

func TestCmdKey_isAPlacePastNine(t *testing.T) {
	if got := cmdKey(9); got != "⌘9" {
		t.Errorf("place 9 is %q", got)
	}
	if got := cmdKey(10); got != "place 10" {
		t.Errorf("place 10 is %q", got)
	}
}

func TestObjectivesNamed_passesOverABlankOne(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	got, err := e.objectivesNamed([]string{" ", "1"})
	if err != nil || len(got) != 1 || got[0] != 1 {
		t.Errorf("named %v, %v", got, err)
	}
	if _, err := e.objectivesNamed([]string{"x"}); !errors.Is(err, ErrObjectiveRequired) {
		t.Errorf("a word for an objective: %v", err)
	}
}

func TestViewObjectives_passOverAnObjectiveGone(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	got, err := e.viewObjectives([]int{9, 1})
	if err != nil || len(got) != 1 {
		t.Errorf("objectives of a view %+v, %v", got, err)
	}
}

func TestView_findsTheFirstViewAndOnlyAViewThere(t *testing.T) {
	s := Snapshot{Views: []Viewing{{View: View{Name: "Home"}}, {View: View{Name: "Work"}}}}
	if _, found := s.View("home"); !found {
		t.Error("the first view is not found")
	}
	cards := []Card{{Number: 1}, {Number: 2}}
	if got := s.Viewed("nowhere", cards); len(got) != 2 {
		t.Errorf("a view not there keeps %d cards", len(got))
	}
}

func TestViewNew_refusesTheNameOfTheSecondView(t *testing.T) {
	path := storeAt(t, "one")
	did(t, path, ViewNew, ask("name", "Home", "state", "todo"))
	did(t, path, ViewNew, ask("name", "Work", "state", "done"))
	refusedAs(t, path, ViewNew, ask("name", "work", "state", "todo"), "view-name-taken")
}

func TestFocusHeld_saysWhatEachFocusAskedKeeps(t *testing.T) {
	for _, c := range []struct {
		was  string
		in   Input
		want string
	}{
		{"", given("focus", "true"), "in"},
		{"", given("no-focus", "true"), "out"},
		{"in", given("focus", "false"), ""},
		{"out", given("no-focus", "false"), ""},
		{"out", given("focus", "false"), "out"},
		{"in", given("no-focus", "false"), "in"},
		{"in", Input{}, "in"},
	} {
		got, err := focusHeld(c.was, c.in)
		if err != nil || got != c.want {
			t.Errorf("was %q, given %v: %q, %v; want %q", c.was, c.in.Given, got, err, c.want)
		}
	}
}

func TestViewShow_countsTheTasksKeptThatAreDone(t *testing.T) {
	path := storeAt(t, "one", "two")
	did(t, path, TaskEdit, ask("number", "1", "state", "done"))
	did(t, path, ViewNew, ask("name", "All", "search", "o"))
	if v := did(t, path, ViewShow, ask("view", "All")).Viewed; v == nil || v.Done != 1 {
		t.Errorf("the view shown reads %+v", v)
	}
}

func viewShown(t *testing.T, e *Engine, name string) *View {
	t.Helper()
	a, err := e.Ask(ViewShow, given("view", name))
	if err != nil {
		t.Fatal(err)
	}
	return a.Viewed
}

func TestViewShow_aViewWithOneLabelGoneStillKeepsWhatTheRestKeeps(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(TaskNew, Input{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"x"}}}); err != nil {
		t.Fatal(err)
	}
	if err := e.store.KeepViews([]store.View{{Name: "Mine", State: "backlog", Labels: []string{"x", "y"}}}); err != nil {
		t.Fatal(err)
	}
	if v := viewShown(t, e, "Mine"); v == nil || v.Keeps != 1 {
		t.Errorf("the view shown reads %+v", v)
	}
}

func TestViewShow_anObjectiveGoneLeavesTheNextNarrowing(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	serving(t, e, "serving", 1)
	add(t, e, "loose")
	if err := e.store.KeepViews([]store.View{{Name: "Mine", Objectives: []int{9, 1}}}); err != nil {
		t.Fatal(err)
	}
	if v := viewShown(t, e, "Mine"); v == nil || v.Keeps != 1 {
		t.Errorf("the view shown reads %+v", v)
	}
}

func TestViewShow_saysTheTasksOutsideInTheirNumbersOrder(t *testing.T) {
	path := storeAt(t, "one", "two", "three")
	did(t, path, TaskEdit, ask("number", "3", "state", "todo").with("link", "1", "2"))
	did(t, path, TaskMove, ask("number", "2", "first", "true"))
	did(t, path, ViewNew, ask("name", "Todo", "state", "todo"))
	v := did(t, path, ViewShow, ask("view", "Todo")).Viewed
	if v == nil || len(v.Outside) != 2 || v.Outside[0].Number != 1 || v.Outside[1].Number != 2 {
		t.Errorf("outside reads %+v", v)
	}
}

func TestViewShownText_saysNoGoneNorOutsideWhenThereIsNone(t *testing.T) {
	said := viewShownText(View{Name: "Home", Place: 1, Holds: Holds{State: "todo"}, Keeps: 1, Of: 2})
	want := "view     Home  (⌘1)\nholds    " + HoldsSays(View{Holds: Holds{State: "todo"}}) + "\nkeeps    1 of 2 tasks · 0 done\n"
	if said != want {
		t.Errorf("said %q, want %q", said, want)
	}
}

func TestLostText_saysNothingOutsideAViewAndOneGoneInTheSingular(t *testing.T) {
	if got := lostText(Answer{Lost: []string{"label x"}}); got != "" {
		t.Errorf("outside a view: %q", got)
	}
	if got := lostText(Answer{InView: "Home", Lost: []string{"label x"}}); got != "view Home: label x is gone and narrows nothing\n  cairn view edit Home, or cairn view delete Home\n" {
		t.Errorf("one gone: %q", got)
	}
}

func TestViewShow_anObjectiveGoneBeforeAnotherLeavesTheOtherNarrowing(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Gone")
	addObjective(t, e, "Ship")
	serving(t, e, "serving", 2)
	add(t, e, "loose")
	if err := objectiveDeleted(e, map[string]string{"number": "1"}); err != nil {
		t.Fatal(err)
	}
	if err := e.store.KeepViews([]store.View{{Name: "Mine", Objectives: []int{1, 2}}}); err != nil {
		t.Fatal(err)
	}
	if v := viewShown(t, e, "Mine"); v == nil || v.Keeps != 1 {
		t.Errorf("the view shown reads %+v", v)
	}
}
