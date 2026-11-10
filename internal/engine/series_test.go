package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func onDay(day string) Home {
	return Home{Lookup: func(name string) (string, bool) {
		if name == TodayVariable {
			return day, true
		}
		return "", false
	}}
}

func doneOn(t *testing.T, path, day string, op Operation, a request) Answer {
	t.Helper()
	answer, err := DoAt(op, Places{Store: path, Home: onDay(day)}, a.input(), Acting{Through: "cli", By: "me"})
	if err != nil {
		t.Fatalf("%s %v on %s: %v", op.Line(), a, day, err)
	}
	return answer
}

func refusedOn(t *testing.T, path, day string, op Operation, a request, reason string) string {
	t.Helper()
	_, err := DoAt(op, Places{Store: path, Home: onDay(day)}, a.input(), Acting{})
	if err == nil {
		t.Fatalf("%s %v: it went ahead, want it refused as %s", op.Line(), a, reason)
	}
	if got := RefusalOf(err).Reason; got != reason {
		t.Errorf("%s %v: refused as %q (%v), want %q", op.Line(), a, got, err, reason)
	}
	return err.Error()
}

func seriesStore(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cairn.db")
}

func shownSeries(t *testing.T, path, day string, n string) (Series, string) {
	t.Helper()
	answer := doneOn(t, path, day, SeriesShow, ask("number", n))
	return Of(SeriesShow, answer).(Series), RenderText(SeriesShow, answer)
}

func occurrencesListed(t *testing.T, path, day string, a request) []int {
	t.Helper()
	answer := doneOn(t, path, day, TaskList, a)
	return rowsListed(Of(TaskList, answer).(Tasks))
}

func TestSeries_aTaskMadeWithARuleIsTheFirstOccurrence(t *testing.T) {
	path := seriesStore(t)
	answer := doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday 10:30", "estimate", "15m", "steps", "[ ] Yesterday").with("label", "work"))
	said := Of(TaskNew, answer).(Task)
	if said.Series == nil || said.Series.Series != 1 || said.Series.Day != "2026-11-09" || said.Deadline != "2026-11-09" || said.State != "todo" {
		t.Fatalf("the task made reads %+v, series %+v", said.Facts, said.Series)
	}
	if !slices.Equal(said.Series.Then, []string{"2026-11-10", "2026-11-11", "2026-11-12"}) {
		t.Errorf("then reads %v", said.Series.Then)
	}
	text := RenderText(TaskNew, answer)
	for _, want := range []string{"repeats  ↻1 every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending", "day      Mon 9 Nov", "then     Tue 10 Nov · Wed 11 Nov · Thu 12 Nov"} {
		if !strings.Contains(text, want) {
			t.Errorf("the answer does not say %q:\n%s", want, text)
		}
	}
	s, _ := shownSeries(t, path, "2026-11-09", "↻1")
	if s.Template.Title != "Stand-up" || !slices.Equal(s.Template.Labels, []string{"work"}) || s.Template.Estimate != "15m" || !slices.Equal(s.Template.Steps, []string{"Yesterday"}) {
		t.Errorf("the template reads %+v", s.Template)
	}
	if s.State != "active" || !slices.Equal(s.Open, []int{1}) || len(s.Acts) != 1 || s.Acts[0].Did != "made" || s.Acts[0].By != "me" {
		t.Errorf("the series reads %+v", s)
	}
}

func TestSeries_doneMakesTheNextAndWritesTheDaysMissed(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	answer := doneOn(t, path, "2026-11-12", TaskEdit, ask("number", "1", "state", "done"))
	said := Of(TaskEdit, answer).(Task)
	if len(said.Made) != 1 || said.Made[0].Task != 2 || said.Made[0].Day != "2026-11-12" || !slices.Equal(said.Made[0].Missed, []string{"2026-11-10", "2026-11-11"}) {
		t.Fatalf("made reads %+v", said.Made)
	}
	text := RenderText(TaskEdit, answer)
	if !strings.Contains(text, "↻1 missed Tue 10 Nov, Wed 11 Nov") || !strings.Contains(text, "↻1 made #2 for Thu 12 Nov") {
		t.Errorf("the answer reads:\n%s", text)
	}
	if got := occurrencesListed(t, path, "2026-11-12", ask("series", "1", "with-archived", "true")); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("the occurrences listed are %v", got)
	}
	s, text := shownSeries(t, path, "2026-11-12", "1")
	if s.Done != 1 || s.Missed != 2 || s.Skipped != 0 || !strings.Contains(text, "missed Tue 10 Nov") {
		t.Errorf("the series reads %+v:\n%s", s, text)
	}
	doneOn(t, path, "2026-11-12", TaskEdit, ask("number", "1", "title", "again"))
	if got := occurrencesListed(t, path, "2026-11-12", ask("repeats", "true")); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("an act on a task done made another: %v", got)
	}
}

func TestSeries_skippingArchivesAndMakesTheNext(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Water", "every", "every 3 days after done"))
	answer := doneOn(t, path, "2026-11-10", TaskSkip, ask("number", "1"))
	if text := RenderText(TaskSkip, answer); !strings.HasPrefix(text, "#1 skipped\n") || !strings.Contains(text, "↻1 made #2 for Fri 13 Nov") {
		t.Errorf("skipping answers:\n%s", text)
	}
	shown := doneOn(t, path, "2026-11-10", TaskShow, ask("number", "1"))
	if !factsOf(*shown.Task).Archived {
		t.Error("the occurrence skipped is still on the board")
	}
	s, text := shownSeries(t, path, "2026-11-10", "1")
	if s.Skipped != 1 || !strings.Contains(text, "skipped #1") || !strings.Contains(text, "3 days after it is done") {
		t.Errorf("the series reads %+v:\n%s", s, text)
	}
	refusedOn(t, path, "2026-11-10", TaskSkip, ask("number", "1"), "occurrence-not-open")
	doneOn(t, path, "2026-11-10", TaskNew, ask("title", "Spec"))
	refusedOn(t, path, "2026-11-10", TaskSkip, ask("number", "3"), "no-occurrence")
}

func TestSeries_stoppedMakesNothingAndResumedCatchesNothingUp(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	text := RenderText(SeriesStop, doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "1")))
	if text != "↻1 stopped; #1 stays open, and nothing more is made\n" {
		t.Errorf("stopping says %q", text)
	}
	refusedOn(t, path, "2026-11-09", SeriesStop, ask("number", "1"), "series-stopped")
	refusedOn(t, path, "2026-11-09", TaskSkip, ask("number", "1"), "series-stopped")
	doneOn(t, path, "2026-11-10", TaskEdit, ask("number", "1", "state", "done"))
	if got := occurrencesListed(t, path, "2026-11-12", ask("repeats", "true", "with-archived", "true")); !slices.Equal(got, []int{1}) {
		t.Errorf("a stopped series made %v", got)
	}
	answer := doneOn(t, path, "2026-11-16", SeriesResume, ask("number", "1"))
	if text := RenderText(SeriesResume, answer); text != "↻1 resumed · next Mon 16 Nov, made as #2\n" {
		t.Errorf("resuming says %q", text)
	}
	s, _ := shownSeries(t, path, "2026-11-16", "1")
	if s.Missed != 0 || s.State != "active" {
		t.Errorf("resumed, the series reads %+v", s)
	}
	refusedOn(t, path, "2026-11-16", SeriesResume, ask("number", "1"), "series-active")
}

func recurring(t *testing.T, home string, on bool) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, ".config", "cairn"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "cairn", "config.toml"), fmt.Appendf(nil, "[recurrence]\non = %t\n", on), 0o600); err != nil {
		t.Fatal(err)
	}
}

func atHome(home, path, day string) Places {
	return Places{Store: path, Home: Home{Dir: home, Lookup: func(name string) (string, bool) {
		switch name {
		case TodayVariable:
			return day, true
		case "CAIRN_STORE":
			return path, true
		}
		return "", false
	}}}
}

func TestSeries_withNothingOpenTheNextIsMadeOnItsDayAndNotBefore(t *testing.T) {
	path, home := seriesStore(t), t.TempDir()
	recurring(t, home, true)
	at := func(day string, op Operation, a request) Answer {
		t.Helper()
		answer, err := DoAt(op, atHome(home, path, day), a.input(), Acting{})
		if err != nil {
			t.Fatalf("%s on %s: %v", op.Line(), day, err)
		}
		return answer
	}
	at("2026-11-09", TaskNew, ask("title", "Review", "every", "every friday"))
	recurring(t, home, false)
	at("2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
	recurring(t, home, true)
	listed := func(day string) []int {
		return rowsListed(Of(TaskList, at(day, TaskList, ask("repeats", "true", "with-archived", "true"))).(Tasks))
	}
	if got := listed("2026-11-19"); !slices.Equal(got, []int{1}) {
		t.Errorf("made ahead of its day: %v", got)
	}
	if got := listed("2026-11-20"); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("on its day, the occurrences are %v", got)
	}
	s := Of(SeriesShow, at("2026-11-20", SeriesShow, ask("number", "1"))).(Series)
	if s.Missed != 0 || s.Acts[len(s.Acts)-1].Through != "series" {
		t.Errorf("made on its day, the series reads %+v", s)
	}
}

func TestSeries_anActThatLeavesNothingOpenMakesTheNext(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Review", "every", "every friday"))
	doneOn(t, path, "2026-11-09", TaskDelete, ask("number", "1"))
	if got := occurrencesListed(t, path, "2026-11-09", ask("repeats", "true")); !slices.Equal(got, []int{2}) {
		t.Fatalf("the act that left nothing open made %v", got)
	}
	if f := factsOf(*doneOn(t, path, "2026-11-09", TaskShow, ask("number", "2")).Task); f.Deadline.Format("2006-01-02") != "2026-11-20" {
		t.Errorf("the next is for %v", f.Deadline)
	}
}

func TestSeries_endsAfterSoManyTimes(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Course", "every", "every day 2 times"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
	doneOn(t, path, "2026-11-10", TaskEdit, ask("number", "2", "state", "done"))
	if got := occurrencesListed(t, path, "2026-11-12", ask("repeats", "true", "with-archived", "true")); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("an ended series made %v", got)
	}
	s, text := shownSeries(t, path, "2026-11-12", "1")
	if !s.Ended || !strings.Contains(text, "it has ended") {
		t.Errorf("the series reads %+v:\n%s", s, text)
	}
}

func TestSeries_editingChangesWhatComesNext(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	doneOn(t, path, "2026-11-09", SeriesEdit, ask("number", "1", "title", "Daily", "every", "every tuesday", "priority", "high").with("label", "team"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
	next := doneOn(t, path, "2026-11-09", TaskShow, ask("number", "2"))
	f := factsOf(*next.Task)
	if f.Title != "Daily" || string(f.Priority) != "high" || !slices.Equal(f.Labels, []string{"team"}) || f.Deadline.Format("2006-01-02") != "2026-11-10" {
		t.Errorf("the next reads %+v", f)
	}
	if f := factsOf(*doneOn(t, path, "2026-11-09", TaskShow, ask("number", "1")).Task); f.Title != "Stand-up" {
		t.Errorf("the one made before was changed: %q", f.Title)
	}
	_, text := shownSeries(t, path, "2026-11-09", "1")
	if !strings.Contains(text, `changed title from "Stand-up" to "Daily"`) || !strings.Contains(text, "changed rule") {
		t.Errorf("the history reads:\n%s", text)
	}
	refusedOn(t, path, "2026-11-09", SeriesEdit, ask("number", "1"), "nothing-to-change")
	refusedOn(t, path, "2026-11-09", SeriesShow, ask("number", "9"), "no-such-series")
	refusedOn(t, path, "2026-11-09", SeriesShow, ask("number", "nine"), "no-such-series")
}

func TestSeries_anOccurrenceChangedWithSeriesChangesTheTemplate(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "title", "Daily", "estimate", "20m", "series", "true", "steps", "[ ] One").with("objective"))
	s, _ := shownSeries(t, path, "2026-11-09", "1")
	if s.Template.Title != "Daily" || s.Template.Estimate != "20m" || !slices.Equal(s.Template.Steps, []string{"One"}) {
		t.Errorf("the template reads %+v", s.Template)
	}
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Spec"))
	refusedOn(t, path, "2026-11-09", TaskEdit, ask("number", "2", "title", "x", "series", "true"), "no-occurrence")
	refusedOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "every", "every day"), "already-an-occurrence")
}

func TestSeries_aRuleGivenToATaskMakesItTheFirst(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Water", "deadline", "2026-11-20"))
	answer := doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "every", "every 3 days", "after-done", "true", "from", "2026-11-11"))
	said := Of(TaskEdit, answer).(Task)
	if said.Series == nil || said.Series.Day != "2026-11-11" || said.Deadline != "2026-11-20" || !strings.Contains(said.Series.Says, "after the last is done") {
		t.Errorf("the task reads %+v, series %+v", said.Facts, said.Series)
	}
}

func TestSeries_listedNarrowedAndSearched(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Expense report", "every", "every 4 weeks on monday"))
	doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "2"))
	listed := func(a request) []int {
		var numbers []int
		for _, s := range Of(SeriesList, doneOn(t, path, "2026-11-09", SeriesList, a)).(SeriesListed).Series {
			numbers = append(numbers, s.Number)
		}
		return numbers
	}
	if got := listed(ask()); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("every series: %v", got)
	}
	if got := listed(ask("state", "stopped")); !slices.Equal(got, []int{2}) {
		t.Errorf("the stopped: %v", got)
	}
	if got := listed(ask("search", "STAND")); !slices.Equal(got, []int{1}) {
		t.Errorf("searched: %v", got)
	}
	text := RenderText(SeriesList, doneOn(t, path, "2026-11-09", SeriesList, ask()))
	if !strings.Contains(text, "↻1  Stand-up  · active  · next Mon 9 Nov") || !strings.Contains(text, "↻2  Expense report  · stopped since Mon 9 Nov") {
		t.Errorf("the listing reads:\n%s", text)
	}
	if text := RenderText(SeriesList, doneOn(t, path, "2026-11-09", SeriesList, ask("search", "nothing"))); text != "there is no series\n" {
		t.Errorf("nothing listed reads %q", text)
	}
	refusedOn(t, path, "2026-11-09", SeriesList, ask("state", "asleep"), "not-one-of")
}

func TestSeries_deletingAsksWhatBecomesOfItsTasks(t *testing.T) {
	for _, c := range []struct {
		choice string
		left   []int
	}{{"keep-tasks", []int{2, 1}}, {"delete-open", []int{1}}, {"delete-tasks", nil}} {
		path := seriesStore(t)
		doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
		doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
		said := refusedOn(t, path, "2026-11-09", SeriesDelete, ask("number", "1"), "series-delete-asks")
		if !strings.Contains(said, "↻1 made 2 tasks") || !strings.Contains(said, "--delete-open keeps the done and deletes #2") {
			t.Errorf("the refusal reads %q", said)
		}
		refusedOn(t, path, "2026-11-09", SeriesDelete, ask("number", "1", "keep-tasks", "true", "delete-tasks", "true"), "series-delete-asks")
		answer := doneOn(t, path, "2026-11-09", SeriesDelete, ask("number", "1", c.choice, "true"))
		if text := RenderText(SeriesDelete, answer); !strings.HasPrefix(text, "↻1 was deleted") {
			t.Errorf("deleting says %q", text)
		}
		if got := occurrencesListed(t, path, "2026-11-09", ask("with-archived", "true")); !slices.Equal(got, c.left) {
			t.Errorf("%s left %v, want %v", c.choice, got, c.left)
		}
		if got := occurrencesListed(t, path, "2026-11-09", ask("repeats", "true", "with-archived", "true")); len(got) != 0 {
			t.Errorf("%s: occurrences of a deleted series: %v", c.choice, got)
		}
	}
}

func TestSeries_anOccurrenceServesWhatTheSeriesServes(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", ObjectiveNew, ask("title", "Team"))
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday").with("objective", "1"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
	if f := factsOf(*doneOn(t, path, "2026-11-09", TaskShow, ask("number", "2")).Task); !slices.Equal(f.Objectives, []int{1}) {
		t.Errorf("the next serves %v", f.Objectives)
	}
}

func TestSeries_aListingMarksTheOccurrences(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Spec"))
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Review", "every", "every friday"))
	answer := doneOn(t, path, "2026-11-09", TaskList, ask())
	if text := RenderText(TaskList, answer); !strings.Contains(text, "· ↻1 every week on Friday") {
		t.Errorf("the listing reads:\n%s", text)
	}
	rows := Of(TaskList, answer).(Tasks).Tasks
	if rows[0].Series != 0 && rows[1].Series != 1 {
		t.Errorf("the rows read %+v", rows)
	}
	refusedOn(t, path, "2026-11-09", TaskList, ask("series", "7"), "no-such-series")
}

func TestSeries_offMakesNothingAndRefusesARule(t *testing.T) {
	path := seriesStore(t)
	home := t.TempDir()
	off := Home{Dir: home, Lookup: func(name string) (string, bool) {
		if name == "CAIRN_STORE" {
			return path, true
		}
		return "", false
	}}
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	if _, err := DoAt(TaskNew, Places{Store: path, Home: off}, ask("title", "x", "every", "every day").input(), Acting{}); RefusalOf(err).Reason != "recurrence-off" {
		t.Errorf("a rule given with recurrence off: %v", err)
	}
	if _, err := DoAt(SeriesList, Places{Store: path, Home: off}, ask().input(), Acting{}); RefusalOf(err).Reason != "recurrence-off" {
		t.Errorf("series asked for with recurrence off: %v", err)
	}
	answer, err := DoAt(TaskEdit, Places{Store: path, Home: off}, ask("number", "1", "state", "done").input(), Acting{})
	if err != nil || len(answer.Followed) != 0 || answer.Occurring != nil {
		t.Errorf("done with recurrence off: %v, %+v", err, answer.Followed)
	}
	if strings.Contains(RenderText(TaskEdit, answer), "↻") {
		t.Errorf("with recurrence off, the card says it repeats:\n%s", RenderText(TaskEdit, answer))
	}
	if !PagesFrom(off).Offers(TaskList) || PagesFrom(off).Offers(SeriesList) || PagesFrom(off).Offers(TaskSkip) {
		t.Error("with recurrence off, the series are offered")
	}
	if slices.ContainsFunc(PagesFrom(off).Shaped(TaskNew).Takes, func(a Argument) bool { return a.Name == "every" }) {
		t.Error("with recurrence off, task new takes a rule")
	}
	if err := os.MkdirAll(filepath.Join(home, ".config", "cairn"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "cairn", "config.toml"), []byte("[recurrence]\non = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !PagesFrom(off).Offers(SeriesList) || !slices.ContainsFunc(PagesFrom(off).Shaped(TaskNew).Takes, func(a Argument) bool { return a.Name == "every" }) {
		t.Error("turned on, the series are not offered")
	}
}

func TestSeries_countedAfterDoneWithNothingOpenComesOnItsDayOrToday(t *testing.T) {
	for _, c := range []struct{ before, on, day string }{{"2026-11-11", "2026-11-12", "2026-11-12"}, {"2026-11-11", "2026-11-20", "2026-11-20"}} {
		path, home := seriesStore(t), t.TempDir()
		recurring(t, home, true)
		at := func(day string, op Operation, a request) Answer {
			t.Helper()
			answer, err := DoAt(op, atHome(home, path, day), a.input(), Acting{})
			if err != nil {
				t.Fatalf("%s on %s: %v", op.Line(), day, err)
			}
			return answer
		}
		at("2026-11-09", TaskNew, ask("title", "Water", "every", "every 3 days after done"))
		recurring(t, home, false)
		at("2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
		recurring(t, home, true)
		listed := func(day string) []int {
			return rowsListed(Of(TaskList, at(day, TaskList, ask("repeats", "true"))).(Tasks))
		}
		if got := listed(c.before); !slices.Equal(got, []int{1}) {
			t.Errorf("made ahead of its day: %v", got)
		}
		if got := listed(c.on); !slices.Equal(got, []int{2, 1}) {
			t.Fatalf("on %s the occurrences are %v", c.on, got)
		}
		if f := factsOf(*at(c.on, TaskShow, ask("number", "2")).Task); f.Deadline.Format("2006-01-02") != c.day {
			t.Errorf("made for %v, want %s", f.Deadline, c.day)
		}
	}
}

func TestSeries_resumedWithOneOpenOrEndedMakesNothing(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "1"))
	_, text := shownSeries(t, path, "2026-11-12", "1")
	if !strings.Contains(text, "stopped since Mon 9 Nov") || !strings.Contains(text, "next      Mon 9 Nov") {
		t.Errorf("a stopped series reads:\n%s", text)
	}
	if text := RenderText(SeriesResume, doneOn(t, path, "2026-11-12", SeriesResume, ask("number", "1"))); text != "↻1 resumed; #1 is still open, and the next comes when it is done or skipped\n" {
		t.Errorf("resumed with one open: %q", text)
	}
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Once", "every", "every day 1 times"))
	doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "2"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "2", "state", "done"))
	_, text = shownSeries(t, path, "2026-11-12", "2")
	if !strings.Contains(text, "next      none, it has ended") {
		t.Errorf("an ended series reads:\n%s", text)
	}
	if text := RenderText(SeriesResume, doneOn(t, path, "2026-11-12", SeriesResume, ask("number", "2"))); text != "↻2 resumed; it has ended, and nothing more is made\n" {
		t.Errorf("resumed once ended: %q", text)
	}
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Later", "every", "every friday"))
	doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "3"))
	doneOn(t, path, "2026-11-09", TaskDelete, ask("number", "3"))
	_, text = shownSeries(t, path, "2026-11-16", "3")
	if !strings.Contains(text, "next      if resumed today: Fri 20 Nov · Fri 27 Nov · Fri 4 Dec") || !strings.Contains(text, "deleted") {
		t.Errorf("a stopped series with nothing open reads:\n%s", text)
	}
}

func TestSeries_aRuleIsRefusedWhatItCannotDo(t *testing.T) {
	path := seriesStore(t)
	refusedOn(t, path, "2026-11-09", TaskNew, ask("title", "x", "after-done", "true"), "every-required")
	refusedOn(t, path, "2026-11-09", TaskNew, ask("title", "x", "every", "every day until 2026-11-01"), "series-ends-before")
	refusedOn(t, path, "2026-11-09", TaskNew, ask("title", "x", "every", "every blue moon"), "not-a-rule")
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Spec"))
	refusedOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "from", "2026-11-12"), "every-required")
	refusedOn(t, path, "2026-11-09", TaskEdit, ask("number", "9", "every", "every day"), "no-such-task")
	refusedOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "every", "every day until 2026-11-01"), "series-ends-before")
	if got := occurrencesListed(t, path, "2026-11-09", ask("repeats", "true", "with-archived", "true")); len(got) != 0 {
		t.Errorf("a refused rule made %v", got)
	}
}

func TestSeries_anEmptyDayAndAfterDoneFalseNameNoRule(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Spec", "from", "", "after-done", "false"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "title", "Spec it", "from", "", "after-done", "false"))
	if got := occurrencesListed(t, path, "2026-11-09", ask("repeats", "true", "with-archived", "true")); len(got) != 0 {
		t.Errorf("an empty day made %v", got)
	}
	refusedOn(t, path, "2026-11-09", TaskNew, ask("title", "x", "from", "2026-11-12"), "every-required")
}

func TestSeries_editingItsStartAndItsRuleIsWrittenDown(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Review", "every", "every week"))
	answer := doneOn(t, path, "2026-11-09", SeriesEdit, ask("number", "↻1", "from", "2026-11-11", "every", "every 2 weeks", "after-done", "true"))
	if text := RenderText(SeriesEdit, answer); !strings.HasPrefix(text, "changed ↻1 Review\nrule      every 2 weeks, after the last is done, from Wed 11 Nov 2026") {
		t.Errorf("editing says:\n%s", text)
	}
	s, _ := shownSeries(t, path, "2026-11-09", "1")
	var fields []string
	for _, a := range s.Acts {
		if a.Did == "changed" {
			fields = append(fields, a.Field)
		}
	}
	if !slices.Equal(fields, []string{"rule", "from"}) {
		t.Errorf("the changes written are %v", fields)
	}
	refusedOn(t, path, "2026-11-09", SeriesEdit, ask("number", "1", "after-done", "true"), "every-required")
	refusedOn(t, path, "2026-11-09", SeriesEdit, ask("number", "1", "every", "every moon"), "not-a-rule")
	refusedOn(t, path, "2026-11-09", SeriesEdit, ask("number", "1").with("objective", "7"), "no-such-objective")
}

func TestSeries_anOccurrenceSaysWhatComesAfterIt(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Last", "every", "every day until 2026-11-09"))
	if text := RenderText(TaskShow, doneOn(t, path, "2026-11-09", TaskShow, ask("number", "1"))); !strings.Contains(text, "then     nothing: the series ends") {
		t.Errorf("the last occurrence reads:\n%s", text)
	}
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Water", "every", "every 2 days after done"))
	doneOn(t, path, "2026-11-09", SeriesStop, ask("number", "2"))
	if text := RenderText(TaskShow, doneOn(t, path, "2026-11-09", TaskShow, ask("number", "2"))); !strings.Contains(text, "then     nothing: the series is stopped") {
		t.Errorf("an occurrence of a stopped series reads:\n%s", text)
	}
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "2", "state", "done"))
	answer := doneOn(t, path, "2026-11-14", SeriesResume, ask("number", "2"))
	if len(answer.Followed) != 1 || answer.Followed[0].Day != "2026-11-14" {
		t.Errorf("resumed, counted after done, it made %+v", answer.Followed)
	}
}

func TestSeries_anOccurrenceMadeSaysItsSeriesInItsTrace(t *testing.T) {
	path := seriesStore(t)
	doneOn(t, path, "2026-11-09", TaskNew, ask("title", "Stand-up", "every", "every weekday"))
	doneOn(t, path, "2026-11-09", TaskEdit, ask("number", "1", "state", "done"))
	made := RenderText(TraceShow, doneOn(t, path, "2026-11-09", TraceShow, ask("number", "2")))
	if !strings.Contains(made, "made from series") || !strings.Contains(made, "↻1") {
		t.Errorf("the trace of the occurrence made reads:\n%s", made)
	}
	if first := RenderText(TraceShow, doneOn(t, path, "2026-11-09", TraceShow, ask("number", "1"))); strings.Contains(first, "made from series") {
		t.Errorf("the task the series began from says it was made from it:\n%s", first)
	}
}
