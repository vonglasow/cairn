package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func homeArchivingSince(t *testing.T, daysAgo int) string {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now().AddDate(0, 0, -daysAgo).Format(time.DateOnly)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), fmt.Appendf(nil, "[archive]\nstart = %s\n", start), 0o600); err != nil {
		t.Fatal(err)
	}
	return home
}

func doneLongAgo(t *testing.T, e *Engine, number int) {
	t.Helper()
	if _, err := e.Ask(TaskEdit, given("number", fmt.Sprint(number), "state", "done")); err != nil {
		t.Fatal(err)
	}
	ago := time.Now().AddDate(0, 0, -10)
	if err := e.history.Record(trace.Made(trace.Carried{Task: number, Did: trace.Changed, Field: "column", Was: "todo", Became: "done", At: ago})); err != nil {
		t.Fatal(err)
	}
}

func TestAsk_sweepsTheArchiveOwedBeforeWhatNeedsTheStore(t *testing.T) {
	e := openEngine(t)
	e.acting = Acting{Through: "cli", By: "sam"}
	add(t, e, "Migration Jenkins")
	doneLongAgo(t, e, 1)
	e.home = Home{Dir: homeArchivingSince(t, 20)}
	e.sweepOwed = true
	if _, err := e.Ask(TaskList, Input{}); err != nil {
		t.Fatal(err)
	}
	if !archivedOnItsOwn(t, e, 1) {
		t.Error("a listing left the archive it owed unswept")
	}
}

func TestTheSweepArchivesEveryTaskDoneBeforeThePeriodPastOneDoneInIt(t *testing.T) {
	e := openEngine(t)
	e.acting = Acting{Through: "cli", By: "sam"}
	for _, title := range []string{"Spike Argo", "Migration Jenkins", "Ship it"} {
		add(t, e, title)
	}
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	doneLongAgo(t, e, 2)
	doneLongAgo(t, e, 3)
	today := midnight(time.Now())
	if err := e.sweptByPeriod(today.AddDate(0, 0, -20), 14, today); err != nil {
		t.Fatal(err)
	}
	if archivedOnItsOwn(t, e, 1) {
		t.Error("a task done in the period under way was archived")
	}
	for _, n := range []int{2, 3} {
		if !archivedOnItsOwn(t, e, n) {
			t.Errorf("task %d, done before the period under way, is not archived", n)
		}
	}
}

func TestADeadlineIsWrittenOnlyWhenItIsLateOrWithinAWeek(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 0, 0, 0, time.Local)
	on := func(days int) time.Time { return startOfDay(now).AddDate(0, 0, days) }
	for _, tc := range []struct {
		due        time.Time
		text       string
		show, late bool
	}{
		{time.Time{}, "", false, false},
		{on(-2), "−2d", true, true},
		{on(0), "0d", true, false},
		{on(3), "3d", true, false},
		{on(soon), "7d", true, false},
		{on(soon + 1), "", false, false},
	} {
		text, show, late := Deadline(tc.due, now)
		if text != tc.text || show != tc.show || late != tc.late {
			t.Errorf("Deadline(%v) = %q, %v, %v, want %q, %v, %v", tc.due, text, show, late, tc.text, tc.show, tc.late)
		}
	}
}

func TestACardDoneSaysNothingOfItsDay(t *testing.T) {
	now := time.Date(2026, 6, 10, 15, 0, 0, 0, time.Local)
	due := startOfDay(now).AddDate(0, 0, 2)
	if text, show, _ := CardDeadline(Card{State: "done", Deadline: due}, now); show || text != "" {
		t.Errorf("a card done says %q of its day", text)
	}
	if text, show, _ := CardDeadline(Card{State: "todo", Deadline: due}, now); !show || text != "2d" {
		t.Errorf("a card to do says %q of its day, shown %v", text, show)
	}
}

func TestStepsAreTalliedLeftAndDone(t *testing.T) {
	parts := []Part{{Title: "a", Done: true}, {Title: "b"}, {Title: "c"}}
	if left, done := Tally(parts); left != 2 || done != 1 {
		t.Errorf("Tally = %d left, %d done, want 2 and 1", left, done)
	}
	if got := Tallied(parts); got != "2 to do · 1 done" {
		t.Errorf("Tallied = %q", got)
	}
	if got := Tallied(nil); got != "" {
		t.Errorf("Tallied with no step = %q, want nothing", got)
	}
}

func TestAFilterSaysTheValuesItTakesAsTheyRead(t *testing.T) {
	f := Filter{Values: []string{"api", "doc"}, Shown: []string{"API", "Doc"}, Ticked: []string{"api", "doc", "gone"}}
	if got := f.TakenSays(); got != "API + Doc + gone" {
		t.Errorf("TakenSays = %q", got)
	}
}

func TestFilters_keepOneValueHeldAsTheOneTheyAreOn(t *testing.T) {
	fs := Filters{several(offered("label", "label", []string{"api", "doc"}, nil))}
	one := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"doc", "gone"}}})
	if one[0].At != 2 || one[0].Ticked != nil {
		t.Errorf("one value held: at %d, ticked %v", one[0].At, one[0].Ticked)
	}
	two := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"doc", "api"}}})
	if two[0].At != 0 || !slices.Equal(two[0].Ticked, []string{"doc", "api"}) {
		t.Errorf("two values held: at %d, ticked %v", two[0].At, two[0].Ticked)
	}
	none := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"gone"}}})
	if none[0].At != 0 || none[0].Ticked != nil {
		t.Errorf("nothing held still offered: at %d, ticked %v", none[0].At, none[0].Ticked)
	}
}

func keptNumbers(t *testing.T, fs Filters, cards []Card) []int {
	t.Helper()
	kept, err := fs.Kept(cards)
	if err != nil {
		t.Fatal(err)
	}
	var numbers []int
	for _, c := range kept {
		numbers = append(numbers, c.Number)
	}
	return numbers
}

func TestFilters_keepByWhatFollowsALabel(t *testing.T) {
	label := several(offered("label", "label", []string{"api"}, nil))
	label.At = 1
	priority := offered("priority", "priority", []string{"high"}, nil)
	priority.At = 1
	cards := []Card{{Number: 1, Labels: []string{"api"}, Priority: "high"}, {Number: 2, Labels: []string{"api"}, Priority: "low"}}
	if got := keptNumbers(t, Filters{label, priority}, cards); !slices.Equal(got, []int{1}) {
		t.Errorf("kept %v, want [1]", got)
	}
}

func TestFilters_keepOnlyWhatIsDueByTheDay(t *testing.T) {
	due := offered("due by", "due-before", []string{"2026-06-14"}, nil)
	due.At = 1
	cards := []Card{{Number: 1, Deadline: dayNamed("2026-06-10")}, {Number: 2, Deadline: dayNamed("2026-06-20")}, {Number: 3}}
	if got := keptNumbers(t, Filters{due}, cards); !slices.Equal(got, []int{1}) {
		t.Errorf("kept %v, want [1]", got)
	}
}

func TestPartsIn_readsAMarkOnlyWhereItIsWhole(t *testing.T) {
	for _, tc := range []struct {
		said string
		want []Part
	}{
		{"#3", []Part{{Title: "#3"}}},
		{"a]b", []Part{{Title: "a]b"}}},
		{"[ ]\nNext", []Part{{Title: "Next"}}},
	} {
		if got := PartsIn(tc.said); !slices.Equal(got, tc.want) {
			t.Errorf("PartsIn(%q) = %+v, want %+v", tc.said, got, tc.want)
		}
	}
}

func TestStepsIn_readsEveryLine(t *testing.T) {
	steps, err := StepsIn("[x] A\n[ ] B")
	if err != nil {
		t.Fatal(err)
	}
	if want := []task.Step{{Title: "A", Done: true}, {Title: "B"}}; !slices.Equal(steps, want) {
		t.Errorf("StepsIn = %v, want %v", steps, want)
	}
}

func TestAsked_passesOverAFieldThatIsNoArgument(t *testing.T) {
	in := Asked([]Field{{Name: "note"}, {Name: "title", Arg: "title", Value: "Ship it"}}, TaskEdit)
	if in.Given["title"] != "Ship it" {
		t.Errorf("a field behind one that is no argument was lost: %+v", in.Given)
	}
}

func TestAnEstimateIsOverrunOnlyByATimeSpentThatReads(t *testing.T) {
	if overrun("1h", "soon") {
		t.Error("a time spent that does not read overran its estimate")
	}
}

func TestPages_areNoneOnlyWhenEveryOneIsOff(t *testing.T) {
	if PagesOff(PageTasks, PageTime).None() {
		t.Error("with the objectives shown, every page was said to be off")
	}
}

func TestOfFocus_saysEverySessionAndNoneAsAnEmptyList(t *testing.T) {
	if got := ofFocus(nil, domain.Progress{}, nil, nil).Sessions; got == nil || len(got) != 0 {
		t.Errorf("with no session, sessions = %#v", got)
	}
	held := []FocusSession{{Task: 4}}
	if got := ofFocus(nil, domain.Progress{}, nil, held).Sessions; !slices.Equal(got, held) {
		t.Errorf("sessions = %v, want %v", got, held)
	}
}

type sheetWork struct{ sheet trace.Sheet }

func (w sheetWork) Ask(Operation, Input) (Answer, error) { return Answer{Sheet: &w.sheet}, nil }

func (sheetWork) Version() (int64, error) { return 0, nil }

func TestReadPeriod_answersTheSheetRead(t *testing.T) {
	w := sheetWork{trace.Sheet{Days: []trace.Day{{Day: dayNamed("2026-10-13"), Total: 90 * time.Minute}}}}
	sheet, err := ReadPeriod(w, dayNamed("2026-10-12"), dayNamed("2026-10-18"), dayNamed("2026-10-14"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sheet.Days) != 1 || sheet.Days[0].Minutes != 90 {
		t.Errorf("read %+v, want the one day of 90 minutes", sheet)
	}
}

func TestATimesheetTaskSaysItWasDeleted(t *testing.T) {
	if got := (TimesheetTask{Title: "Ship it", Deleted: true}).Shown(); got != "Ship it (deleted)" {
		t.Errorf("Shown = %q", got)
	}
}

func TestLook_readsEachCardWithItsStepsAndTheFocus(t *testing.T) {
	told := task.From(aTask("Ship it"))
	told.Numbered(1)
	told.Described("why it matters")
	told.Step("Write it", false)
	told.Step("Test it", true)
	f := newFake()
	f.tasks = []task.Task{told.Task(), aTask("Bare")}
	f.focus = []int{1}
	s, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(s.Focus, []int{1}) {
		t.Errorf("focus = %v, want [1]", s.Focus)
	}
	c := s.Cards[0]
	if c.Number != 1 || !c.InFocus || !c.Described() {
		t.Errorf("card = %+v", c)
	}
	if want := []Part{{Number: 1, Title: "Write it"}, {Number: 2, Title: "Test it", Done: true}}; !slices.Equal(c.Steps, want) {
		t.Errorf("steps = %+v, want %+v", c.Steps, want)
	}
	if bare := s.Cards[1]; bare.Steps != nil || bare.Described() {
		t.Errorf("a bare card = %+v", bare)
	}
}

func TestLabelRename_renamesItOnEveryObjectiveAndListsIt(t *testing.T) {
	path := storeAt(t)
	for _, title := range []string{"one", "two"} {
		if _, err := Do(ObjectiveNew, path, Input{Given: map[string]string{"title": title}, Listed: map[string][]string{"label": {"sp26-20"}}}); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "sp26-20", "to": "sprint-20"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Labels == nil || !slices.ContainsFunc(*answer.Labels, func(l LabelCarried) bool { return l.Label == "sprint-20" }) {
		t.Errorf("the rename answered %+v", answer.Labels)
	}
	for _, n := range []string{"1", "2"} {
		goal, _ := Do(ObjectiveShow, path, given("number", n))
		if labels := factsOfObjective(*goal.Objective).Labels; !slices.Equal(labels, []string{"sprint-20"}) {
			t.Errorf("objective %s carries %v", n, labels)
		}
	}
}

func TestStepNew_answersTheTaskWithItsSteps(t *testing.T) {
	path := storeAt(t, "Ship it")
	answer, err := Do(StepNew, path, given("number", "1", "title", "Write it"))
	if err != nil {
		t.Fatal(err)
	}
	if answer.Task == nil || len(answer.Task.Carried().Steps) != 1 {
		t.Errorf("step new answered %+v", answer.Task)
	}
}

func TestStepList_saysATaskHasNoSteps(t *testing.T) {
	path := storeAt(t, "Ship it")
	answer, err := Do(StepList, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := RenderText(StepList, answer); got != "task 1 has no steps\n" {
		t.Errorf("step list says %q", got)
	}
}

func TestChangedNumbers_refusesNoughtAndWhatDoesNotRead(t *testing.T) {
	notANumber, bothWays, atBirth := errors.New("nan"), errors.New("both"), errors.New("birth")
	for _, said := range []string{"0", "99999999999999999999"} {
		if _, err := changedNumbers([]string{said}, nil, false, notANumber, bothWays, atBirth); !errors.Is(err, notANumber) {
			t.Errorf("%q gave %v", said, err)
		}
	}
}

func TestAPageMayStartAtTheFirstTask(t *testing.T) {
	p, err := askedPage(given("offset", "0"))
	if err != nil || !p.asked || p.offset != 0 {
		t.Errorf("offset 0 gave %+v, %v", p, err)
	}
}
