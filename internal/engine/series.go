package engine

import (
	"cairn/internal/engine/act"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine/task"
	"cairn/internal/store/sqlite"
)

// Series is how a series crosses to a machine: its rule, what its next occurrence is made from, what it did so far, and, shown, every occurrence and its trace.
type Series struct {
	Number      int                 `json:"number"`
	Title       string              `json:"title"`
	State       string              `json:"state" jsonschema:"active, or stopped"`
	Stopped     string              `json:"stopped,omitempty" jsonschema:"the day it was stopped; left out while it is active"`
	Rule        string              `json:"rule" jsonschema:"how often it comes back, in the words cairn reads back"`
	Says        string              `json:"says" jsonschema:"the whole rule the way a person reads it: how often, what it counts from, from which day, and when it ends"`
	From        string              `json:"from" jsonschema:"the day it counts from"`
	Next        []string            `json:"next" jsonschema:"the days of its next occurrences, the open one first, or the ones it would make if resumed today; empty once it has ended"`
	Ended       bool                `json:"ended" jsonschema:"true once its end is reached: it makes nothing more"`
	Template    series.Template     `json:"template"`
	Open        []int               `json:"open" jsonschema:"the numbers of its occurrences neither done nor archived"`
	Done        int                 `json:"done"`
	Skipped     int                 `json:"skipped"`
	Missed      int                 `json:"missed"`
	TimeSpent   string              `json:"time_spent" jsonschema:"how long was spent over its occurrences, as 1h30m; empty when nothing was"`
	Occurrences []series.Occurrence `json:"occurrences,omitempty" jsonschema:"every day it made a task for or missed, the latest first, archived ones included; left out unless it is shown"`
	Acts        []series.Act        `json:"acts,omitempty" jsonschema:"its trace: everything that happened to it, oldest first; left out unless it is shown"`
}

// seriesListed is how a listing of series crosses.
type seriesListed struct {
	Series []Series `json:"series"`
}

// ErrRecurrenceOff is what anything about a series is refused with while recurrence is off.
var ErrRecurrenceOff = Refusing("series-off", "series are off: [series] on = true, in the configuration or in the store's settings, turns them on")

var everyArguments = []Argument{
	{Name: "every", Kind: Text, Repeats: true, Means: "a rule it comes back by, in plain words: every weekday 10:30, every other thursday 14:00, every 3 days after done, every monday 10 times, every day until 2026-12-31; the task is its first occurrence"},
	{Name: "after-done", Kind: Bool, Repeats: true, Means: "count the rule from the day the last one is done, rather than on the calendar"},
	{Name: "from", Kind: day, Repeats: true, Means: "the day the rule counts from; today when it is not said"},
}

func seriesNumber(means string) Argument {
	return Argument{Name: "number", Kind: Text, Positional: true, Required: true, Of: "a series", Means: "the number of the series, as 3 or ↻3, " + means}
}

// SeriesList lists the series, active and stopped, or only some.
var SeriesList = Operation{
	Thing:   "series",
	Verb:    "list",
	Short:   "List the series",
	Does:    "List every series, active and stopped, with its rule, its next day and what it did so far, or only those in one state, or carrying every word of a search in their title or their rule.",
	Example: "cairn series list\ncairn series list --state stopped --search expense",
	Level:   Read,
	Needs:   aStore,
	Repeats: true,
	Answers: seriesListed{},
	Takes: []Argument{
		{Name: "state", Kind: Text, OneOf: []string{"active", "stopped"}, Means: "keep only the series in this state"},
		{Name: "search", Kind: Text, Means: "keep only the series carrying every word of it in their title or their rule"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		if err := on.Engine.mayRecur(); err != nil {
			return Answer{}, err
		}
		if err := accepted(in, Argument{Name: "state", OneOf: []string{"active", "stopped"}}); err != nil {
			return Answer{}, err
		}
		every, err := on.Engine.store.EverySeries()
		if err != nil {
			return Answer{}, err
		}
		tasks, err := on.Engine.everyTask()
		if err != nil {
			return Answer{}, err
		}
		listed := seriesListed{Series: []Series{}}
		for _, kept := range every {
			s, err := on.Engine.aSeries(kept)
			if err != nil {
				return Answer{}, err
			}
			shown := s.shape(tasks, on.Engine.today())
			if state := in.Given["state"]; state != "" && shown.State != state {
				continue
			}
			if search := in.Given["search"]; search != "" && !Reaches([]string{shown.Title, shown.Rule, shown.Says, "↻" + strconv.Itoa(shown.Number)}, search) {
				continue
			}
			listed.Series = append(listed.Series, shown)
		}
		return Answer{SeriesKept: &listed}, nil
	},
}

// SeriesShow says everything about a series: its rule, what its next occurrence is made from, every occurrence it made and its trace.
var SeriesShow = Operation{
	Thing:   "series",
	Verb:    "show",
	Short:   "Show a series, its occurrences and what happened to it",
	Does:    "Show a series: its rule in words and its next days, what its next occurrence is made from, what it did so far, every day it made a task for or missed, archived tasks included, and everything that happened to it, its trace.",
	Example: "cairn series show 3",
	Level:   Read,
	Needs:   aStore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to show")},
	Run: func(on workedOn, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		return on.Engine.shownSeries(s, true)
	},
}

// SeriesEdit changes a series: what its next occurrence is made from, its rule, or the day it counts from; the ones made never change.
var SeriesEdit = Operation{
	Thing: "series",
	Verb:  "edit",
	Short: "Change a series, for the ones to come",
	Does:  "Change what the next occurrence of a series is made from — its title, priority, labels, steps, estimate, the objectives it serves — its rule, or the day it counts from. The occurrences already made never change; each change is written down in its trace.",
	Example: `cairn series edit 3 --title "Stand-up" --estimate 15m
cairn series edit 3 --every "every tuesday 10:00"`,
	Level:      write,
	Needs:      aStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Repeats:    true,
	Answers:    Series{},
	Takes: append([]Argument{
		seriesNumber("to change"),
		{Name: "title", Kind: Text, Means: "what its occurrences are called"},
		{Name: "priority", Kind: Text, OneOf: oneOf(item.Priorities), Outside: item.ErrNotAPriority, Means: "the priority its occurrences carry"},
		{Name: "label", Kind: List, Means: "a label its occurrences carry; x or +x adds one, -x takes one off, =x replaces them all"},
		{Name: "estimate", Kind: duration, Means: "how long each occurrence should take"},
		{Name: "objective", Kind: List, Of: "an objective", Means: "the number of an objective its occurrences serve; 4 or +4 adds one, -1 takes one away, =4 replaces them all"},
		{Name: "state", Kind: Text, OneOf: series.Arrivals, Outside: series.ErrNotAnArrival, Means: "the column its occurrences arrive in: backlog, todo or current"},
		stepsArgument,
	}, everyArguments...),
	Run: func(on workedOn, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.editSeries(s, in); err != nil {
			return Answer{}, err
		}
		return on.Engine.shownNumber(s.kept.Number)
	},
}

// SeriesStop stops a series: its open occurrence stays an ordinary open task, and nothing more is made until it is resumed.
var SeriesStop = Operation{
	Thing:   "series",
	Verb:    "stop",
	Short:   "Stop a series",
	Does:    "Stop a series: nothing more is made until it is resumed. Its open occurrence stays an ordinary open task, and every task it made stays where it is.",
	Example: "cairn series stop 3",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to stop")},
	Run: func(on workedOn, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		if !s.active() {
			return Answer{}, Refusing("series-stopped", fmt.Sprintf("↻%d is already stopped, since %s", s.kept.Number, series.ShortDay(s.stopped())))
		}
		s.kept.Stopped = calendar.FormatDay(on.Engine.today())
		if err := on.Engine.store.UpdateSeries(s.kept); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "stopped"}); err != nil {
			return Answer{}, err
		}
		return on.Engine.shownNumber(s.kept.Number)
	},
}

// SeriesResume resumes a stopped series from today, catching nothing up.
var SeriesResume = Operation{
	Thing:   "series",
	Verb:    "resume",
	Short:   "Resume a stopped series",
	Does:    "Resume a stopped series from today: with no occurrence open, the next is made at once, on the first day of its rule from today. Nothing missed while it was stopped is made or written as missed.",
	Example: "cairn series resume 3",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to resume")},
	Run: func(on workedOn, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		if s.active() {
			return Answer{}, Refusing("series-active", fmt.Sprintf("↻%d is not stopped", s.kept.Number))
		}
		s.kept.Stopped = ""
		if err := on.Engine.store.UpdateSeries(s.kept); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "resumed"}); err != nil {
			return Answer{}, err
		}
		made, err := on.Engine.resumed(s.kept.Number)
		if err != nil {
			return Answer{}, err
		}
		answer, err := on.Engine.shownNumber(s.kept.Number)
		answer.Followed = made
		return answer, err
	},
}

// SeriesDelete deletes a series, and what becomes of the tasks it made is said, never assumed.
var SeriesDelete = Operation{
	Thing:   "series",
	Verb:    "delete",
	Short:   "Delete a series, saying what becomes of its tasks",
	Does:    "Delete a series: its rule and what its next occurrence is made from; its trace stays. Refused until what becomes of the tasks it made is said: keep-tasks keeps every one as an ordinary task, delete-open deletes the one open and keeps the rest, delete-tasks deletes them all.",
	Example: "cairn series delete 3 --keep-tasks --yes",
	Level:   Full,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: series.Deleted{},
	Takes: []Argument{
		seriesNumber("to delete"),
		{Name: "keep-tasks", Kind: Bool, Means: "keep every task it made, as an ordinary task"},
		{Name: "delete-open", Kind: Bool, Means: "delete the occurrence open, and keep the others"},
		{Name: "delete-tasks", Kind: Bool, Means: "delete every task it made as well"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		gone, err := on.Engine.deleteSeries(s, in)
		if err != nil {
			return Answer{}, err
		}
		return Answer{SeriesGone: &gone}, nil
	},
}

// TaskSkip skips an occurrence: it is taken off the board, and its series makes the next.
var TaskSkip = Operation{
	Thing:   "task",
	Verb:    "skip",
	Short:   "Skip an occurrence of a series",
	Does:    "Skip an occurrence of a series: it is archived as it stands, written down as skipped by its series, and the series makes the next.",
	Example: "cairn task skip 12",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the occurrence"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		if err := on.Engine.mayRecur(); err != nil {
			return Answer{}, err
		}
		n, _ := in.whole("number")
		s, err := on.Engine.occurrenceOf(n)
		if err != nil {
			return Answer{}, err
		}
		if !s.active() {
			return Answer{}, Refusing("series-stopped", fmt.Sprintf("↻%d is stopped: nothing would come after #%d; cairn task archive %d takes it off the board", s.kept.Number, n, n))
		}
		found, err := on.Engine.show(n)
		if err != nil {
			return Answer{}, err
		}
		if !series.IsOpen(found) {
			return Answer{}, Refusing("occurrence-not-open", fmt.Sprintf("task %d is done or archived already: only an open occurrence is skipped", n))
		}
		if err := on.Engine.archive(n); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "skipped", Task: n, Day: s.dayOf(n)}); err != nil {
			return Answer{}, err
		}
		return shownTask(on, n)
	},
}

type aSeries struct {
	kept  sqlite.Series
	rule  series.Rule
	start time.Time
	acts  []sqlite.SeriesAct
}

func (e *Engine) aSeries(kept sqlite.Series) (aSeries, error) {
	rule, err := series.ParseRule(kept.Rule)
	if err != nil {
		return aSeries{}, err
	}
	start, err := calendar.ParseDeadline(kept.Start)
	if err != nil {
		return aSeries{}, err
	}
	acts, err := e.store.SeriesActs(kept.Number)
	if err != nil {
		return aSeries{}, err
	}
	return aSeries{kept: kept, rule: rule, start: start, acts: acts}, nil
}

func (e *Engine) seriesAt(n int) (aSeries, error) {
	kept, err := e.store.FindSeries(n)
	if errors.Is(err, sqlite.ErrNoSuchSeries) {
		return aSeries{}, Refusing("no-such-series", fmt.Sprintf("there is no series ↻%d", n))
	}
	if err != nil {
		return aSeries{}, err
	}
	return e.aSeries(kept)
}

func (e *Engine) recurringSeries(in Input) (aSeries, error) {
	if err := e.mayRecur(); err != nil {
		return aSeries{}, err
	}
	given := in.Given["number"]
	n, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(given), "↻"))
	if err != nil {
		return aSeries{}, Refusing("no-such-series", fmt.Sprintf("%q names no series: a series is named by its number, as 3 or ↻3", given))
	}
	return e.seriesAt(n)
}

func (e *Engine) occurrenceOf(n int) (aSeries, error) {
	of, err := e.store.Occurrences()
	if err != nil {
		return aSeries{}, err
	}
	numbered, is := of[n]
	if !is {
		return aSeries{}, Refusing("no-occurrence", fmt.Sprintf("task %d is no occurrence of a series", n))
	}
	return e.seriesAt(numbered)
}

func (s aSeries) active() bool { return s.kept.Stopped == "" }

func (s aSeries) stopped() time.Time {
	day, _ := calendar.ParseDeadline(s.kept.Stopped)
	return day
}

func (s aSeries) count(did string) int {
	n := 0
	for _, a := range s.acts {
		if a.Did == did {
			n++
		}
	}
	return n
}

func (s aSeries) made() []int {
	var tasks []int
	for _, a := range s.acts {
		if a.Did == "made" {
			tasks = append(tasks, a.Task)
		}
	}
	return tasks
}

func (s aSeries) dayOf(task int) string {
	for _, a := range s.acts {
		if a.Did == "made" && a.Task == task {
			return a.Day
		}
	}
	return ""
}

func (s aSeries) last() time.Time {
	var last time.Time
	for _, a := range s.acts {
		if a.Did != "made" && a.Did != "missed" {
			continue
		}
		if day, err := calendar.ParseDeadline(a.Day); err == nil {
			last = day
		}
	}
	return last
}

func (s aSeries) room(more int) bool {
	return s.rule.Times() == 0 || s.count("made")+s.count("missed")+more < s.rule.Times()
}

func (s aSeries) ended() bool { return !s.room(0) }

func (s aSeries) opened(tasks map[int]task.Task) []int {
	var open []int
	for _, n := range s.made() {
		if t, there := tasks[n]; there && series.IsOpen(t) {
			open = append(open, n)
		}
	}
	return open
}

func (e *Engine) everyTask() (map[int]task.Task, error) {
	every := filter.Narrowing{}
	every.WithArchived()
	all, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	out := make(map[int]task.Task, len(all))
	for _, t := range all {
		out[t.Carried().Number] = t
	}
	return out, nil
}

func (e *Engine) today() time.Time { return series.DayOf(today(e.home)) }

func (e *Engine) recurring() bool {
	if e.home.Dir == "" {
		return true
	}
	read, err := e.configured()
	return err != nil || read.Recurrence()
}

func (e *Engine) mayRecur() error {
	if e.recurring() {
		return nil
	}
	return ErrRecurrenceOff
}

func (e *Engine) recordSeries(a sqlite.SeriesAct) error {
	a.Through, a.By, a.At = e.acting.Through, e.acting.By, e.home.now()
	return e.store.RecordSeries(a)
}

func ruleOf(in Input) (series.Rule, error) {
	said := in.Given["every"]
	if in.yes("after-done") {
		said += " after done"
	}
	return series.ParseRule(said)
}

func (e *Engine) fromOf(in Input) (time.Time, error) {
	if given := in.Given["from"]; given != "" {
		return calendar.ParseDeadline(given)
	}
	return e.today(), nil
}

func everyNamed(in Input) (bool, error) {
	named := in.Named("every")
	if !named && (in.yes("after-done") || in.Given["from"] != "") {
		return false, Refusing("every-required", "--after-done and --from go with --every, the rule they belong to")
	}
	return named, nil
}

func templateOf(t task.Task) sqlite.Series {
	f := t.Carried()
	steps := []string{}
	for _, s := range f.Steps {
		steps = append(steps, s.Title)
	}
	return sqlite.Series{Title: f.Title, Priority: string(f.Priority), Estimate: f.Estimate, Steps: steps,
		Labels: listed(f.Labels), Objectives: listed(f.Objectives), Arrives: series.ArrivalOf(f.State)}
}

func (e *Engine) beginSeries(t task.Task, rule series.Rule, from, day time.Time) error {
	kept := templateOf(t)
	kept.Rule, kept.Start = rule.Words(), calendar.FormatDay(from)
	made, err := e.store.KeepSeries(kept)
	if err != nil {
		return err
	}
	for _, name := range made.Labels {
		if err := e.store.KeepLabel(name); err != nil {
			return err
		}
	}
	return e.recordSeries(sqlite.SeriesAct{Series: made.Number, Did: "made", Task: t.Carried().Number, Day: calendar.FormatDay(day)})
}

func firstDay(rule series.Rule, from time.Time) (time.Time, error) {
	day, ok := rule.First(from)
	if !ok {
		return time.Time{}, Refusing("series-ends-before", "the rule ends before its first day")
	}
	return day, nil
}

func (e *Engine) occur(s aSeries, day time.Time) (int, error) {
	scopes, err := e.scopes()
	if err != nil {
		return 0, err
	}
	labels := slices.DeleteFunc(slices.Clone(s.kept.Labels), func(name string) bool { return scopes.Of(name).Refuses(name, label.OnTasks) != nil })
	state, deadline, estimate := series.ArrivalOf(item.State(s.kept.Arrives)), calendar.FormatDay(day), calendar.FormatDuration(s.kept.Estimate)
	objectives := make([]string, 0, len(s.kept.Objectives))
	for _, n := range s.kept.Objectives {
		objectives = append(objectives, strconv.Itoa(n))
	}
	steps := make([]task.Step, 0, len(s.kept.Steps))
	for _, title := range s.kept.Steps {
		steps = append(steps, task.Step{Title: title})
	}
	priority := s.kept.Priority
	made, err := e.add(s.kept.Title, Changes{State: &state, Priority: &priority, Labels: &labels, Deadline: &deadline,
		Estimate: &estimate, Objective: &objectives, Steps: &steps})
	if err != nil {
		return 0, err
	}
	n := made.Carried().Number
	if err := e.record(n, series.MadeFromSeries, "", "", "↻"+strconv.Itoa(s.kept.Number)); err != nil {
		return 0, err
	}
	return n, e.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "made", Task: n, Day: deadline})
}

func (e *Engine) openSeries() (map[int]bool, error) {
	if !e.recurring() {
		return nil, nil
	}
	every, err := e.store.EverySeries()
	if err != nil {
		return nil, err
	}
	if len(every) == 0 {
		return nil, nil
	}
	tasks, err := e.everyTask()
	if err != nil {
		return nil, err
	}
	open := map[int]bool{}
	for _, kept := range every {
		s, err := e.aSeries(kept)
		if err != nil {
			return nil, err
		}
		open[kept.Number] = s.active() && len(s.opened(tasks)) > 0
	}
	return open, nil
}

func (e *Engine) madeAfter(before map[int]bool) ([]series.MadeNext, error) {
	var made []series.MadeNext
	after, err := e.openSeries()
	if err != nil {
		return nil, err
	}
	for _, n := range slices.Sorted(maps.Keys(before)) {
		if !before[n] || after[n] {
			continue
		}
		if _, there := after[n]; !there {
			continue
		}
		s, err := e.seriesAt(n)
		if err != nil {
			return nil, err
		}
		if !s.active() {
			continue
		}
		next, did, err := e.follow(s)
		if err != nil {
			return nil, err
		}
		if did {
			made = append(made, next)
		}
	}
	return made, nil
}

func (e *Engine) follow(s aSeries) (series.MadeNext, bool, error) {
	today := e.today()
	var missed []time.Time
	day, ok := s.rule.Next(today, s.start)
	if !s.rule.AfterDone() {
		day, ok = s.rule.Next(s.last(), s.start)
		for ok && day.Before(today) && s.room(len(missed)) {
			missed = append(missed, day)
			day, ok = s.rule.Next(day, s.start)
		}
	}
	next := series.MadeNext{Series: s.kept.Number}
	for _, d := range missed {
		next.Missed = append(next.Missed, calendar.FormatDay(d))
		if err := e.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "missed", Day: calendar.FormatDay(d)}); err != nil {
			return series.MadeNext{}, false, err
		}
	}
	if !ok || !s.room(len(missed)) {
		return next, len(missed) > 0, nil
	}
	n, err := e.occur(s, day)
	if err != nil {
		return series.MadeNext{}, false, err
	}
	next.Task, next.Day = n, calendar.FormatDay(day)
	return next, true, nil
}

func (s aSeries) fromToday(today time.Time) (time.Time, bool) {
	if s.rule.AfterDone() {
		return today, true
	}
	after := today.AddDate(0, 0, -1)
	if last := s.last(); last.After(after) {
		after = last
	}
	return s.rule.Next(after, s.start)
}

func (e *Engine) resumed(n int) ([]series.MadeNext, error) {
	s, err := e.seriesAt(n)
	if err != nil {
		return nil, err
	}
	tasks, err := e.everyTask()
	if err != nil {
		return nil, err
	}
	if len(s.opened(tasks)) > 0 || s.ended() {
		return nil, nil
	}
	day, ok := s.fromToday(e.today())
	if !ok {
		return nil, nil
	}
	made, err := e.occur(s, day)
	if err != nil {
		return nil, err
	}
	return []series.MadeNext{{Series: n, Task: made, Day: calendar.FormatDay(day)}}, nil
}

func (e *Engine) seriesDue() error {
	if !e.recurring() {
		return nil
	}
	every, err := e.store.EverySeries()
	if err != nil {
		return err
	}
	if len(every) == 0 {
		return nil
	}
	tasks, err := e.everyTask()
	if err != nil {
		return err
	}
	today := e.today()
	var due []aSeries
	for _, kept := range every {
		s, err := e.aSeries(kept)
		if err != nil {
			return err
		}
		if !s.active() || s.ended() || len(s.opened(tasks)) > 0 {
			continue
		}
		day, ok := s.dueFrom(today)
		if ok && day.Equal(today) {
			due = append(due, s)
		}
	}
	if len(due) == 0 {
		return nil
	}
	return e.store.Within(func(kept *sqlite.Store) error {
		in := e.keptBy(kept)
		in.acting.Through = "series"
		for _, s := range due {
			if _, err := in.occur(s, today); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s aSeries) dueFrom(today time.Time) (time.Time, bool) {
	if !s.rule.AfterDone() {
		return s.fromToday(today)
	}
	day, ok := s.rule.Next(s.last(), s.start)
	if day.Before(today) {
		day = today
	}
	return day, ok
}

func (s aSeries) nextDays(tasks map[int]task.Task, today time.Time) []string {
	var next []time.Time
	if open := s.opened(tasks); len(open) > 0 {
		if day, err := calendar.ParseDeadline(s.dayOf(open[0])); err == nil {
			next = append(next, day)
		}
	} else if !s.ended() {
		if day, ok := s.fromToday(today); ok {
			next = append(next, day)
		}
	}
	if len(next) > 0 && !s.rule.AfterDone() {
		next = append(next, s.rule.Dates(next[0], s.start, 2)...)
	}
	days := []string{}
	for _, d := range next {
		days = append(days, calendar.FormatDay(d))
	}
	return days
}

func (s aSeries) shape(tasks map[int]task.Task, today time.Time) Series {
	said := Series{
		Number: s.kept.Number, Title: s.kept.Title, State: "active", Stopped: s.kept.Stopped,
		Rule: s.rule.Words(), Says: s.rule.Sentence(s.start), From: s.kept.Start, Ended: s.ended(),
		Next: s.nextDays(tasks, today), Open: listed(s.opened(tasks)),
		Skipped: s.count("skipped"), Missed: s.count("missed"),
		Template: series.Template{Title: s.kept.Title, Priority: s.kept.Priority, Labels: listed(s.kept.Labels), Steps: listed(s.kept.Steps),
			Estimate: calendar.FormatDuration(s.kept.Estimate), Objectives: listed(s.kept.Objectives), Arrives: s.kept.Arrives},
	}
	if !s.active() {
		said.State = "stopped"
	}
	var spent time.Duration
	for _, n := range s.made() {
		if t, there := tasks[n]; there {
			f := t.Carried()
			spent += f.Spent
			if f.State == item.Done {
				said.Done++
			}
		}
	}
	said.TimeSpent = calendar.FormatDuration(spent)
	return said
}

func (e *Engine) shownNumber(n int) (Answer, error) {
	s, err := e.seriesAt(n)
	if err != nil {
		return Answer{}, err
	}
	return e.shownSeries(s, false)
}

func (e *Engine) shownSeries(s aSeries, whole bool) (Answer, error) {
	tasks, err := e.everyTask()
	if err != nil {
		return Answer{}, err
	}
	shown := s.shape(tasks, e.today())
	if whole {
		shown.Occurrences, shown.Acts = s.occurrences(tasks), s.trace()
	}
	return Answer{SeriesShown: &shown}, nil
}

func (s aSeries) occurrences(tasks map[int]task.Task) []series.Occurrence {
	skipped := map[int]bool{}
	for _, a := range s.acts {
		if a.Did == "skipped" {
			skipped[a.Task] = true
		}
	}
	out := []series.Occurrence{}
	for _, a := range s.acts {
		switch a.Did {
		case "missed":
			out = append(out, series.Occurrence{Day: a.Day, Is: "missed"})
		case "made":
			one := series.Occurrence{Day: a.Day, Task: a.Task, Is: "deleted"}
			if t, there := tasks[a.Task]; there {
				f := t.Carried()
				one.Is, one.Archived, one.TimeSpent = string(f.State), f.Archived, calendar.FormatDuration(f.Spent)
				if skipped[a.Task] {
					one.Is = "skipped"
				}
			}
			out = append(out, one)
		}
	}
	slices.Reverse(out)
	slices.SortStableFunc(out, func(a, b series.Occurrence) int { return strings.Compare(b.Day, a.Day) })
	return out
}

func (s aSeries) trace() []series.Act {
	out := make([]series.Act, 0, len(s.acts))
	for _, a := range s.acts {
		out = append(out, series.Act{Did: a.Did, Task: a.Task, Day: a.Day, Field: a.Field, Was: a.Was, Became: a.Became,
			Through: a.Through, By: a.By, At: a.At.UTC().Format(time.RFC3339)})
	}
	return out
}

func (e *Engine) occurring(numbers []int) (map[int]series.Occurs, error) {
	if !e.recurring() || len(numbers) == 0 {
		return nil, nil
	}
	of, err := e.store.Occurrences()
	if err != nil {
		return nil, err
	}
	known := map[int]aSeries{}
	out := map[int]series.Occurs{}
	for _, n := range numbers {
		number, is := of[n]
		if !is {
			continue
		}
		s, seen := known[number]
		if !seen {
			if s, err = e.seriesAt(number); err != nil {
				return nil, err
			}
			known[number] = s
		}
		out[n] = s.occurs(n)
	}
	return out, nil
}

func (s aSeries) occurs(n int) series.Occurs {
	o := series.Occurs{Series: s.kept.Number, Rule: s.rule.Words(), Often: s.rule.Short(), Says: s.rule.Sentence(s.start),
		Day: s.dayOf(n), At: s.rule.At(), Then: []string{}, Stopped: !s.active()}
	if s.rule.AfterDone() {
		o.After = s.rule.Then()
		return o
	}
	if day, err := calendar.ParseDeadline(o.Day); err == nil {
		for _, d := range s.rule.Dates(day, s.start, 3) {
			o.Then = append(o.Then, calendar.FormatDay(d))
		}
	}
	return o
}

func (e *Engine) editSeries(s aSeries, in Input) error {
	changes, err := taskChanges(in)
	if err != nil {
		return err
	}
	changes.Title = in.change("title")
	arrives := changes.State
	changes.State = nil
	everyGiven, err := everyNamed(in)
	if err != nil {
		return err
	}
	if !changes.named() && !everyGiven && arrives == nil {
		return errNothingToChange
	}
	was := s.kept
	if changes.named() {
		steps := []task.Step{}
		for _, title := range was.Steps {
			steps = append(steps, task.Step{Title: title})
		}
		t := task.Made(task.Carried{Title: was.Title, State: item.Backlog, Priority: item.Priority(was.Priority),
			Labels: was.Labels, Estimate: was.Estimate, Steps: steps, Objectives: was.Objectives})
		if t, err = apply(t, changes, changing); err != nil {
			return err
		}
		if err := e.inScope(was.Labels, t.Carried().Labels, label.OnTasks); err != nil {
			return err
		}
		if err := e.canServe(t); err != nil {
			return err
		}
		template := templateOf(t)
		template.Number, template.Rule, template.Start, template.Stopped = was.Number, was.Rule, was.Start, was.Stopped
		template.Arrives = was.Arrives
		s.kept = template
	}
	if arrives != nil {
		if !slices.Contains(series.Arrivals, *arrives) {
			return fmt.Errorf("%q: %w", *arrives, series.ErrNotAnArrival)
		}
		s.kept.Arrives = *arrives
	}
	if everyGiven {
		rule, err := ruleOf(in)
		if err != nil {
			return err
		}
		s.kept.Rule = rule.Words()
	}
	if given := in.Given["from"]; given != "" {
		s.kept.Start = given
	}
	return e.keptSeries(was, s.kept)
}

func (e *Engine) keptSeries(was, now sqlite.Series) error {
	if err := e.store.UpdateSeries(now); err != nil {
		return err
	}
	for _, name := range now.Labels {
		if err := e.store.KeepLabel(name); err != nil {
			return err
		}
	}
	for _, f := range []struct{ field, was, now string }{
		{"title", was.Title, now.Title},
		{"priority", was.Priority, now.Priority},
		{"labels", strings.Join(was.Labels, ", "), strings.Join(now.Labels, ", ")},
		{"steps", strings.Join(was.Steps, "; "), strings.Join(now.Steps, "; ")},
		{"estimate", calendar.FormatDuration(was.Estimate), calendar.FormatDuration(now.Estimate)},
		{"objectives", act.NumbersOrNothing(was.Objectives), act.NumbersOrNothing(now.Objectives)},
		{"rule", was.Rule, now.Rule},
		{"from", was.Start, now.Start},
		{"column", was.Arrives, now.Arrives},
	} {
		if f.was == f.now {
			continue
		}
		if err := e.recordSeries(sqlite.SeriesAct{Series: now.Number, Did: "changed", Field: f.field, Was: f.was, Became: f.now}); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) everyGiven(n int, in Input) error {
	rule, err := ruleOf(in)
	if err != nil {
		return err
	}
	from, err := e.fromOf(in)
	if err != nil {
		return err
	}
	day, err := firstDay(rule, from)
	if err != nil {
		return err
	}
	t, err := e.show(n)
	if err != nil {
		return err
	}
	if t.Carried().Deadline.IsZero() {
		deadline := calendar.FormatDay(day)
		if t, err = e.edit(n, Changes{Deadline: &deadline}); err != nil {
			return err
		}
	}
	return e.beginSeries(t, rule, from, day)
}

func (e *Engine) seriesFollowsTask(n int, in Input) error {
	s, err := e.occurrenceOf(n)
	if err != nil {
		return err
	}
	t, err := e.show(n)
	if err != nil {
		return err
	}
	from := templateOf(t)
	now := s.kept
	for name, take := range map[string]func(){
		"title":     func() { now.Title = from.Title },
		"priority":  func() { now.Priority = from.Priority },
		"label":     func() { now.Labels = from.Labels },
		"estimate":  func() { now.Estimate = from.Estimate },
		"steps":     func() { now.Steps = from.Steps },
		"objective": func() { now.Objectives = from.Objectives },
	} {
		if in.Named(name) {
			take()
		}
	}
	return e.keptSeries(s.kept, now)
}

func (e *Engine) deleteSeries(s aSeries, in Input) (series.Deleted, error) {
	tasks, err := e.everyTask()
	if err != nil {
		return series.Deleted{}, err
	}
	var there []int
	for _, n := range s.made() {
		if _, is := tasks[n]; is {
			there = append(there, n)
		}
	}
	open := s.opened(tasks)
	var chosen []string
	for _, choice := range []string{"keep-tasks", "delete-open", "delete-tasks"} {
		if in.yes(choice) {
			chosen = append(chosen, choice)
		}
	}
	if len(chosen) != 1 {
		return series.Deleted{}, Refusing("series-delete-asks", deletionAsked(s.kept.Number, there, open))
	}
	gone := series.Deleted{Deleted: s.kept.Number, Kept: []int{}, TasksDeleted: []int{}}
	for _, n := range there {
		if chosen[0] == "delete-tasks" || chosen[0] == "delete-open" && slices.Contains(open, n) {
			if err := e.delete(n); err != nil {
				return series.Deleted{}, err
			}
			gone.TasksDeleted = append(gone.TasksDeleted, n)
			continue
		}
		gone.Kept = append(gone.Kept, n)
	}
	if err := e.store.DeleteSeries(s.kept.Number); err != nil {
		return series.Deleted{}, err
	}
	return gone, e.recordSeries(sqlite.SeriesAct{Series: s.kept.Number, Did: "deleted", Became: chosen[0]})
}

func deletionAsked(n int, there, open []int) string {
	openSaid := "deletes none, as none is open"
	if len(open) > 0 {
		openSaid = "keeps the done and deletes " + act.NumbersOrNothing(open)
	}
	return fmt.Sprintf("↻%d made %s. Say what becomes of them: --keep-tasks keeps all %d as ordinary tasks, --delete-open %s, --delete-tasks deletes all %d as well",
		n, item.ManyOf(len(there), "task", "tasks"), len(there), openSaid, len(there))
}

// StandingSaid is where a series stands, in words: active, or stopped since a day.
func StandingSaid(s Series) string {
	if s.State == "stopped" {
		return "stopped since " + series.ShortDays([]string{s.Stopped})
	}
	return "active"
}

// NextSaid is the next days of a series in words, or why there are none.
func NextSaid(s Series) string {
	if s.Ended && len(s.Next) == 0 {
		return "none, it has ended"
	}
	if len(s.Next) == 0 {
		return "none"
	}
	if s.State == "stopped" && len(s.Open) == 0 {
		return "if resumed today: " + series.ShortDays(s.Next)
	}
	return series.ShortDays(s.Next)
}

func seriesListText(all []Series) string {
	if len(all) == 0 {
		return "there is no series\n"
	}
	var out strings.Builder
	for _, s := range all {
		fmt.Fprintf(&out, "↻%d  %s  · %s", s.Number, s.Title, StandingSaid(s))
		if s.State == "active" {
			next := s
			next.Next = s.Next[:min(1, len(s.Next))]
			fmt.Fprintf(&out, "  · next %s", NextSaid(next))
		}
		fmt.Fprintf(&out, "\n    %s · %d done · %d skipped · %d missed\n", strings.SplitN(s.Says, ", ", 2)[0], s.Done, s.Skipped, s.Missed)
	}
	return out.String()
}

// TemplateSaid is what the next occurrence of a series is made from, in a line.
func TemplateSaid(t series.Template) string {
	said := []string{t.Title, t.Priority}
	if len(t.Labels) > 0 {
		said = append(said, strings.Join(t.Labels, ", "))
	}
	if len(t.Steps) > 0 {
		said = append(said, item.ManyOf(len(t.Steps), "step", "steps"))
	}
	if t.Estimate != "" {
		said = append(said, t.Estimate)
	}
	if len(t.Objectives) > 0 {
		said = append(said, "serves "+act.NumbersOrNothing(t.Objectives))
	}
	return strings.Join(said, " · ")
}

func seriesShownText(s Series) string {
	var out strings.Builder
	fmt.Fprintf(&out, "↻%d  %s  %s\n\n", s.Number, s.Title, StandingSaid(s))
	fmt.Fprintf(&out, "rule      %s\nnext      %s\n", s.Says, NextSaid(s))
	if strings.Contains(s.Says, "after the last is done") && !s.Ended {
		rule, _ := series.ParseRule(s.Rule)
		fmt.Fprintf(&out, "then      %s\n", rule.Then())
	}
	fmt.Fprintf(&out, "template  %s\n", TemplateSaid(s.Template))
	soFar := fmt.Sprintf("%d done · %d skipped · %d missed", s.Done, s.Skipped, s.Missed)
	if s.TimeSpent != "" {
		soFar += " · " + s.TimeSpent
	}
	fmt.Fprintf(&out, "so far    %s\nopen      %s\n", soFar, item.OrNothing(act.NumbersOrNothing(s.Open)))
	if len(s.Occurrences) > 0 {
		out.WriteString("\noccurrences\n")
		for _, o := range s.Occurrences {
			what := "—"
			if o.Task != 0 {
				what = item.Ref(o.Task)
			}
			is := o.Is
			if o.Archived {
				is += " · archived"
			}
			fmt.Fprintf(&out, "  %-10s  %-5s  %s", series.ShortDays([]string{o.Day}), what, is)
			if o.TimeSpent != "" {
				out.WriteString("  " + o.TimeSpent)
			}
			out.WriteString("\n")
		}
	}
	if len(s.Acts) > 0 {
		out.WriteString("\ntrace\n")
		for _, a := range s.Acts {
			at, _ := time.Parse(time.RFC3339, a.At)
			line := fmt.Sprintf("  %s  %s", calendar.FormatMoment(at), series.ActSaid(a))
			for _, more := range []string{a.Through, a.By} {
				if more != "" {
					line += " · " + more
				}
			}
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

func seriesActedText(verb string, a Answer) string {
	s := *a.SeriesShown
	switch verb {
	case "stop":
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d stopped; %s stays open, and nothing more is made\n", s.Number, act.NumbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d stopped; nothing more is made\n", s.Number)
	case "resume":
		if len(a.Followed) > 0 {
			return fmt.Sprintf("↻%d resumed · next %s, made as #%d\n", s.Number, series.ShortDays([]string{a.Followed[0].Day}), a.Followed[0].Task)
		}
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d resumed; %s is still open, and the next comes when it is done or skipped\n", s.Number, act.NumbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d resumed; it has ended, and nothing more is made\n", s.Number)
	case "edit":
		return fmt.Sprintf("changed ↻%d %s\nrule      %s\ntemplate  %s\n", s.Number, s.Title, s.Says, TemplateSaid(s.Template))
	}
	return seriesShownText(s)
}

func seriesDeletedText(d series.Deleted) string {
	said := fmt.Sprintf("↻%d was deleted", d.Deleted)
	if len(d.TasksDeleted) > 0 {
		said += " · " + act.NumbersOrNothing(d.TasksDeleted) + " deleted"
	}
	if len(d.Kept) > 0 {
		said += fmt.Sprintf(" · %s kept as ordinary tasks", item.ManyOf(len(d.Kept), "task", "tasks"))
	}
	return said + "\n"
}

var inSeriesArgument = Argument{Name: "series", Kind: Number, Of: "a series", Repeats: true, Bare: "0", Means: "keep only the occurrences of this series, by its number; named alone, or 0, the occurrences of every series"}

var seriesTooArgument = Argument{Name: "series", Kind: Bool, Repeats: true, Means: "change the series it is an occurrence of as well, for the ones to come: whatever of its title, priority, labels, estimate, steps and objectives is named"}

func (e *Engine) beginning(in Input, c *Changes) (*series.Begun, error) {
	every, err := everyNamed(in)
	if err != nil || !every {
		return nil, err
	}
	if err := e.mayRecur(); err != nil {
		return nil, err
	}
	rule, err := ruleOf(in)
	if err != nil {
		return nil, err
	}
	from, err := e.fromOf(in)
	if err != nil {
		return nil, err
	}
	day, err := firstDay(rule, from)
	if err != nil {
		return nil, err
	}
	if c.Deadline == nil {
		deadline := calendar.FormatDay(day)
		c.Deadline = &deadline
	}
	if c.State == nil {
		state := string(item.Todo)
		c.State = &state
	}
	return &series.Begun{Rule: rule, From: from, Day: day}, nil
}

func (e *Engine) repeating(n int, in Input) (bool, error) {
	every, err := everyNamed(in)
	if err != nil {
		return false, err
	}
	if !every && !in.yes("series") {
		return false, nil
	}
	if err := e.mayRecur(); err != nil {
		return false, err
	}
	s, err := e.occurrenceOf(n)
	if in.yes("series") && err != nil {
		return false, err
	}
	if every && err == nil {
		return false, Refusing("already-an-occurrence", fmt.Sprintf("task %d is already an occurrence of ↻%d: cairn series edit %d --every changes its rule", n, s.kept.Number, s.kept.Number))
	}
	if _, err := e.show(n); err != nil {
		return false, err
	}
	return every, nil
}

func (e *Engine) occurrencesIn(in Input) (series.OccurrencesAsked, error) {
	numbered, named := in.whole("series")
	if !named {
		return series.OccurrencesAsked{}, nil
	}
	if err := e.mayRecur(); err != nil {
		return series.OccurrencesAsked{}, err
	}
	if numbered != 0 {
		if _, err := e.seriesAt(numbered); err != nil {
			return series.OccurrencesAsked{}, err
		}
	}
	of, err := e.store.Occurrences()
	return series.OccurrencesAsked{Of: of, Series: numbered}, err
}
