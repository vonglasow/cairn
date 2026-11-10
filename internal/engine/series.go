package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
)

// Series is how a series crosses to a machine: its rule, what its next occurrence is made from, what it did so far, and, shown, every occurrence and its trace.
type Series struct {
	Number      int          `json:"number"`
	Title       string       `json:"title"`
	State       string       `json:"state" jsonschema:"active, or stopped"`
	Stopped     string       `json:"stopped,omitempty" jsonschema:"the day it was stopped; left out while it is active"`
	Rule        string       `json:"rule" jsonschema:"how often it comes back, in the words cairn reads back"`
	Says        string       `json:"says" jsonschema:"the whole rule the way a person reads it: how often, what it counts from, from which day, and when it ends"`
	From        string       `json:"from" jsonschema:"the day it counts from"`
	Next        []string     `json:"next" jsonschema:"the days of its next occurrences, the open one first, or the ones it would make if resumed today; empty once it has ended"`
	Ended       bool         `json:"ended" jsonschema:"true once its end is reached: it makes nothing more"`
	Template    Template     `json:"template"`
	Open        []int        `json:"open" jsonschema:"the numbers of its occurrences neither done nor archived"`
	Done        int          `json:"done"`
	Skipped     int          `json:"skipped"`
	Missed      int          `json:"missed"`
	TimeSpent   string       `json:"time_spent" jsonschema:"how long was spent over its occurrences, as 1h30m; empty when nothing was"`
	Occurrences []Occurrence `json:"occurrences,omitempty" jsonschema:"every day it made a task for or missed, the latest first, archived ones included; left out unless it is shown"`
	Acts        []SeriesAct  `json:"acts,omitempty" jsonschema:"its trace: everything that happened to it, oldest first; left out unless it is shown"`
}

// Template is what the next occurrence of a series is made from.
type Template struct {
	Title      string   `json:"title"`
	Priority   string   `json:"priority"`
	Labels     []string `json:"labels"`
	Steps      []string `json:"steps"`
	Estimate   string   `json:"estimate"`
	Objectives []int    `json:"objectives"`
}

// Occurrence is one day of a series: the task made for it, or none when it was missed.
type Occurrence struct {
	Day       string `json:"day"`
	Task      int    `json:"task" jsonschema:"the number of the task made for that day; 0 when none was"`
	Is        string `json:"is" jsonschema:"the column the task stands in, or skipped, missed or deleted"`
	Archived  bool   `json:"archived"`
	TimeSpent string `json:"time_spent"`
}

// SeriesAct is one thing that happened to a series, as it crosses to a machine.
type SeriesAct struct {
	Did     string `json:"did" jsonschema:"made, missed, skipped, changed, stopped, resumed or deleted"`
	Task    int    `json:"task,omitempty"`
	Day     string `json:"day,omitempty"`
	Field   string `json:"field,omitempty"`
	Was     string `json:"was,omitempty"`
	Became  string `json:"became,omitempty"`
	Through string `json:"through,omitempty"`
	By      string `json:"by,omitempty"`
	At      string `json:"at"`
}

// SeriesListed is how a listing of series crosses.
type SeriesListed struct {
	Series []Series `json:"series"`
}

// SeriesDeleted is how deleting a series answers: which, and what became of the tasks it made.
type SeriesDeleted struct {
	Deleted      int   `json:"deleted"`
	Kept         []int `json:"kept" jsonschema:"the tasks it made that stay, ordinary tasks now"`
	TasksDeleted []int `json:"tasks_deleted" jsonschema:"the tasks it made that were deleted with it"`
}

// Occurs is how a task says it is an occurrence of a series.
type Occurs struct {
	Series  int      `json:"series"`
	Rule    string   `json:"rule" jsonschema:"how often it comes back, in the words cairn reads back"`
	Often   string   `json:"often" jsonschema:"how often it comes back, the way a person reads it"`
	Says    string   `json:"says" jsonschema:"the whole rule the way a person reads it"`
	Day     string   `json:"day" jsonschema:"the day the series made it for"`
	At      string   `json:"at,omitempty" jsonschema:"the time of day it is for, as 10:30; left out when the rule names none"`
	Then    []string `json:"then" jsonschema:"the next three days of the rule after its own, on the calendar; empty when it counts from the last one done"`
	Stopped bool     `json:"stopped"`
	after   string
}

// MadeNext is one occurrence a series made because an act left it with none open, and the days it missed first.
type MadeNext struct {
	Series int      `json:"series"`
	Task   int      `json:"task"`
	Day    string   `json:"day"`
	Missed []string `json:"missed,omitempty"`
}

// ErrRecurrenceOff is what anything about a series is refused with while recurrence is off.
var ErrRecurrenceOff = Refusing("recurrence-off", "recurrence is off: [recurrence] on = true, in the configuration or in the store's settings, turns it on")

var everyArguments = []Argument{
	{Name: "every", Kind: Text, Repeats: true, Means: "a rule it comes back by, in plain words: every weekday 10:30, every other thursday 14:00, every 3 days after done, every monday 10 times, every day until 2026-12-31; the task is its first occurrence"},
	{Name: "after-done", Kind: Bool, Repeats: true, Means: "count the rule from the day the last one is done, rather than on the calendar"},
	{Name: "from", Kind: Day, Repeats: true, Means: "the day the rule counts from; today when it is not said"},
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
	Needs:   AStore,
	Repeats: true,
	Answers: SeriesListed{},
	Takes: []Argument{
		{Name: "state", Kind: Text, OneOf: []string{"active", "stopped"}, Means: "keep only the series in this state"},
		{Name: "search", Kind: Text, Means: "keep only the series carrying every word of it in their title or their rule"},
	},
	Run: func(on On, in Input) (Answer, error) {
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
		listed := SeriesListed{Series: []Series{}}
		for _, kept := range every {
			s, err := on.Engine.aSeries(kept)
			if err != nil {
				return Answer{}, err
			}
			shown := s.shape(tasks, on.Engine.today())
			if state, named := in.Given["state"]; named && state != "" && shown.State != state {
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
	Needs:   AStore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to show")},
	Run: func(on On, in Input) (Answer, error) {
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
	Level:      Write,
	Needs:      AStore,
	Again:      ChangesNothingMore,
	Overwrites: true,
	Repeats:    true,
	Answers:    Series{},
	Takes: append([]Argument{
		seriesNumber("to change"),
		{Name: "title", Kind: Text, Means: "what its occurrences are called"},
		{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Outside: domain.ErrNotAPriority, Means: "the priority its occurrences carry"},
		{Name: "label", Kind: List, Means: "a label its occurrences carry; x or +x adds one, -x takes one off, =x replaces them all"},
		{Name: "estimate", Kind: Duration, Means: "how long each occurrence should take"},
		{Name: "objective", Kind: List, Of: "an objective", Means: "the number of an objective its occurrences serve; 4 or +4 adds one, -1 takes one away, =4 replaces them all"},
		stepsArgument,
	}, everyArguments...),
	Run: func(on On, in Input) (Answer, error) {
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to stop")},
	Run: func(on On, in Input) (Answer, error) {
		s, err := on.Engine.recurringSeries(in)
		if err != nil {
			return Answer{}, err
		}
		if !s.active() {
			return Answer{}, Refusing("series-stopped", fmt.Sprintf("↻%d is already stopped, since %s", s.kept.Number, domain.ShortDay(s.stopped())))
		}
		s.kept.Stopped = domain.FormatDay(on.Engine.today())
		if err := on.Engine.store.UpdateSeries(s.kept); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "stopped"}); err != nil {
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Series{},
	Takes:   []Argument{seriesNumber("to resume")},
	Run: func(on On, in Input) (Answer, error) {
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
		if err := on.Engine.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "resumed"}); err != nil {
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
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: SeriesDeleted{},
	Takes: []Argument{
		seriesNumber("to delete"),
		{Name: "keep-tasks", Kind: Bool, Means: "keep every task it made, as an ordinary task"},
		{Name: "delete-open", Kind: Bool, Means: "delete the occurrence open, and keep the others"},
		{Name: "delete-tasks", Kind: Bool, Means: "delete every task it made as well"},
	},
	Run: func(on On, in Input) (Answer, error) {
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Repeats: true,
	Answers: Task{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the occurrence"},
	},
	Run: func(on On, in Input) (Answer, error) {
		if err := on.Engine.mayRecur(); err != nil {
			return Answer{}, err
		}
		n, _ := in.Whole("number")
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
		if !isOpen(found) {
			return Answer{}, Refusing("occurrence-not-open", fmt.Sprintf("task %d is done or archived already: only an open occurrence is skipped", n))
		}
		if err := on.Engine.archive(n); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "skipped", Task: n, Day: s.dayOf(n)}); err != nil {
			return Answer{}, err
		}
		return shownTask(on, n)
	},
}

type aSeries struct {
	kept  store.Series
	rule  domain.Rule
	start time.Time
	acts  []store.SeriesAct
}

func (e *Engine) aSeries(kept store.Series) (aSeries, error) {
	rule, err := domain.ParseRule(kept.Rule)
	if err != nil {
		return aSeries{}, err
	}
	start, err := domain.ParseDeadline(kept.Start)
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
	if errors.Is(err, store.ErrNoSuchSeries) {
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
	series, is := of[n]
	if !is {
		return aSeries{}, Refusing("no-occurrence", fmt.Sprintf("task %d is no occurrence of a series", n))
	}
	return e.seriesAt(series)
}

func (s aSeries) active() bool { return s.kept.Stopped == "" }

func (s aSeries) stopped() time.Time {
	day, _ := domain.ParseDeadline(s.kept.Stopped)
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
		if day, err := domain.ParseDeadline(a.Day); err == nil && day.After(last) {
			last = day
		}
	}
	return last
}

func (s aSeries) room(more int) bool {
	return s.rule.Times() == 0 || s.count("made")+s.count("missed")+more < s.rule.Times()
}

func (s aSeries) ended() bool { return !s.room(0) }

func isOpen(t task.Task) bool {
	f := factsOf(t)
	return !f.Archived && f.State != domain.Done
}

func (s aSeries) opened(tasks map[int]task.Task) []int {
	var open []int
	for _, n := range s.made() {
		if t, there := tasks[n]; there && isOpen(t) {
			open = append(open, n)
		}
	}
	return open
}

func (e *Engine) everyTask() (map[int]task.Task, error) {
	every := domain.Narrowing{}
	every.WithArchived()
	all, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	out := make(map[int]task.Task, len(all))
	for _, t := range all {
		out[factsOf(t).Number] = t
	}
	return out, nil
}

func (e *Engine) today() time.Time { return domain.DayOf(today(e.home)) }

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

func (e *Engine) recordSeries(a store.SeriesAct) error {
	a.Through, a.By, a.At = e.acting.Through, e.acting.By, e.home.now()
	return e.store.RecordSeries(a)
}

func ruleOf(in Input) (domain.Rule, error) {
	said := in.Given["every"]
	if in.Yes("after-done") {
		said += " after done"
	}
	return domain.ParseRule(said)
}

func (e *Engine) fromOf(in Input) (time.Time, error) {
	if given := in.Given["from"]; given != "" {
		return domain.ParseDeadline(given)
	}
	return e.today(), nil
}

func everyNamed(in Input) (bool, error) {
	named := in.Named("every")
	if !named && (in.Yes("after-done") || in.Given["from"] != "") {
		return false, Refusing("every-required", "--after-done and --from go with --every, the rule they belong to")
	}
	return named, nil
}

func templateOf(t task.Task) store.Series {
	f := factsOf(t)
	steps := []string{}
	for _, s := range f.Steps {
		steps = append(steps, s.Title)
	}
	return store.Series{Title: f.Title, Priority: string(f.Priority), Estimate: f.Estimate, Steps: steps,
		Labels: listed(f.Labels), Objectives: listed(f.Objectives)}
}

func (e *Engine) beginSeries(t task.Task, rule domain.Rule, from, day time.Time) error {
	kept := templateOf(t)
	kept.Rule, kept.Start = rule.Words(), domain.FormatDay(from)
	made, err := e.store.KeepSeries(kept)
	if err != nil {
		return err
	}
	for _, label := range made.Labels {
		if err := e.store.KeepLabel(label); err != nil {
			return err
		}
	}
	return e.recordSeries(store.SeriesAct{Series: made.Number, Did: "made", Task: factsOf(t).Number, Day: domain.FormatDay(day)})
}

func firstDay(rule domain.Rule, from time.Time) (time.Time, error) {
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
	labels := slices.DeleteFunc(slices.Clone(s.kept.Labels), func(label string) bool { return scopes.of(label).Refuses(label, domain.OnTasks) != nil })
	state, deadline, estimate := string(domain.Todo), domain.FormatDay(day), domain.FormatDuration(s.kept.Estimate)
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
	n := factsOf(made).Number
	if err := e.record(n, MadeFromSeries, "", "", "↻"+strconv.Itoa(s.kept.Number)); err != nil {
		return 0, err
	}
	return n, e.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "made", Task: n, Day: deadline})
}

// MadeFromSeries is the act written in the trace of a task a series made, naming the series.
const MadeFromSeries = "made from series"

func (e *Engine) openSeries() (map[int]bool, error) {
	if !e.recurring() {
		return nil, nil
	}
	every, err := e.store.EverySeries()
	if err != nil || len(every) == 0 {
		return nil, err
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

func (e *Engine) madeAfter(before map[int]bool) ([]MadeNext, error) {
	var made []MadeNext
	after, err := e.openSeries()
	if err != nil {
		return nil, err
	}
	for n, was := range before {
		if !was || after[n] {
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
	slices.SortFunc(made, func(a, b MadeNext) int { return a.Series - b.Series })
	return made, nil
}

func (e *Engine) follow(s aSeries) (MadeNext, bool, error) {
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
	next := MadeNext{Series: s.kept.Number}
	for _, d := range missed {
		next.Missed = append(next.Missed, domain.FormatDay(d))
		if err := e.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "missed", Day: domain.FormatDay(d)}); err != nil {
			return MadeNext{}, false, err
		}
	}
	if !ok || !s.room(len(missed)) {
		return next, len(missed) > 0, nil
	}
	n, err := e.occur(s, day)
	if err != nil {
		return MadeNext{}, false, err
	}
	next.Task, next.Day = n, domain.FormatDay(day)
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

func (e *Engine) resumed(n int) ([]MadeNext, error) {
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
	return []MadeNext{{Series: n, Task: made, Day: domain.FormatDay(day)}}, nil
}

func (e *Engine) seriesDue() error {
	if !e.recurring() {
		return nil
	}
	every, err := e.store.EverySeries()
	if err != nil || len(every) == 0 {
		return err
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
	return e.store.Within(func(kept *store.Store) error {
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
		if day, err := domain.ParseDeadline(s.dayOf(open[0])); err == nil {
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
		days = append(days, domain.FormatDay(d))
	}
	return days
}

func (s aSeries) shape(tasks map[int]task.Task, today time.Time) Series {
	said := Series{
		Number: s.kept.Number, Title: s.kept.Title, State: "active", Stopped: s.kept.Stopped,
		Rule: s.rule.Words(), Says: s.rule.Sentence(s.start), From: s.kept.Start, Ended: s.ended(),
		Next: s.nextDays(tasks, today), Open: listed(s.opened(tasks)),
		Skipped: s.count("skipped"), Missed: s.count("missed"),
		Template: Template{Title: s.kept.Title, Priority: s.kept.Priority, Labels: listed(s.kept.Labels), Steps: listed(s.kept.Steps),
			Estimate: domain.FormatDuration(s.kept.Estimate), Objectives: listed(s.kept.Objectives)},
	}
	if !s.active() {
		said.State = "stopped"
	}
	var spent time.Duration
	for _, n := range s.made() {
		if t, there := tasks[n]; there {
			f := factsOf(t)
			spent += f.Spent
			if f.State == domain.Done {
				said.Done++
			}
		}
	}
	said.TimeSpent = domain.FormatDuration(spent)
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

func (s aSeries) occurrences(tasks map[int]task.Task) []Occurrence {
	skipped := map[int]bool{}
	for _, a := range s.acts {
		if a.Did == "skipped" {
			skipped[a.Task] = true
		}
	}
	out := []Occurrence{}
	for _, a := range s.acts {
		switch a.Did {
		case "missed":
			out = append(out, Occurrence{Day: a.Day, Is: "missed"})
		case "made":
			one := Occurrence{Day: a.Day, Task: a.Task, Is: "deleted"}
			if t, there := tasks[a.Task]; there {
				f := factsOf(t)
				one.Is, one.Archived, one.TimeSpent = string(f.State), f.Archived, domain.FormatDuration(f.Spent)
				if skipped[a.Task] {
					one.Is = "skipped"
				}
			}
			out = append(out, one)
		}
	}
	slices.Reverse(out)
	slices.SortStableFunc(out, func(a, b Occurrence) int { return strings.Compare(b.Day, a.Day) })
	return out
}

func (s aSeries) trace() []SeriesAct {
	out := make([]SeriesAct, 0, len(s.acts))
	for _, a := range s.acts {
		out = append(out, SeriesAct{Did: a.Did, Task: a.Task, Day: a.Day, Field: a.Field, Was: a.Was, Became: a.Became,
			Through: a.Through, By: a.By, At: a.At.UTC().Format(time.RFC3339)})
	}
	return out
}

func (e *Engine) occurring(numbers []int) (map[int]Occurs, error) {
	if !e.recurring() || len(numbers) == 0 {
		return nil, nil
	}
	of, err := e.store.Occurrences()
	if err != nil {
		return nil, err
	}
	known := map[int]aSeries{}
	out := map[int]Occurs{}
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

func (s aSeries) occurs(n int) Occurs {
	o := Occurs{Series: s.kept.Number, Rule: s.rule.Words(), Often: s.rule.Short(), Says: s.rule.Sentence(s.start),
		Day: s.dayOf(n), At: s.rule.At(), Then: []string{}, Stopped: !s.active()}
	if s.rule.AfterDone() {
		o.after = s.rule.Then()
		return o
	}
	if day, err := domain.ParseDeadline(o.Day); err == nil {
		for _, d := range s.rule.Dates(day, s.start, 3) {
			o.Then = append(o.Then, domain.FormatDay(d))
		}
	}
	return o
}

func (e *Engine) editSeries(s aSeries, in Input) error {
	changes, err := taskChanges(in)
	if err != nil {
		return err
	}
	changes.Title = in.Change("title")
	everyGiven, err := everyNamed(in)
	if err != nil {
		return err
	}
	if !changes.named() && !everyGiven {
		return ErrNothingToChange
	}
	was := s.kept
	if changes.named() {
		steps := []task.Step{}
		for _, title := range was.Steps {
			steps = append(steps, task.Step{Title: title})
		}
		t := task.Made(task.Carried{Title: was.Title, State: domain.Backlog, Priority: domain.Priority(was.Priority),
			Labels: was.Labels, Estimate: was.Estimate, Steps: steps, Objectives: was.Objectives})
		if t, err = apply(t, changes, changing); err != nil {
			return err
		}
		if err := e.inScope(was.Labels, factsOf(t).Labels, domain.OnTasks); err != nil {
			return err
		}
		if err := e.canServe(t); err != nil {
			return err
		}
		template := templateOf(t)
		template.Number, template.Rule, template.Start, template.Stopped = was.Number, was.Rule, was.Start, was.Stopped
		s.kept = template
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

func (e *Engine) keptSeries(was, now store.Series) error {
	if err := e.store.UpdateSeries(now); err != nil {
		return err
	}
	for _, label := range now.Labels {
		if err := e.store.KeepLabel(label); err != nil {
			return err
		}
	}
	for _, f := range []struct{ field, was, now string }{
		{"title", was.Title, now.Title},
		{"priority", was.Priority, now.Priority},
		{"labels", strings.Join(was.Labels, ", "), strings.Join(now.Labels, ", ")},
		{"steps", strings.Join(was.Steps, "; "), strings.Join(now.Steps, "; ")},
		{"estimate", domain.FormatDuration(was.Estimate), domain.FormatDuration(now.Estimate)},
		{"objectives", numbersOrNothing(was.Objectives), numbersOrNothing(now.Objectives)},
		{"rule", was.Rule, now.Rule},
		{"from", was.Start, now.Start},
	} {
		if f.was == f.now {
			continue
		}
		if err := e.recordSeries(store.SeriesAct{Series: now.Number, Did: "changed", Field: f.field, Was: f.was, Became: f.now}); err != nil {
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
	if factsOf(t).Deadline.IsZero() {
		deadline := domain.FormatDay(day)
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

func (e *Engine) deleteSeries(s aSeries, in Input) (SeriesDeleted, error) {
	tasks, err := e.everyTask()
	if err != nil {
		return SeriesDeleted{}, err
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
		if in.Yes(choice) {
			chosen = append(chosen, choice)
		}
	}
	if len(chosen) != 1 {
		return SeriesDeleted{}, Refusing("series-delete-asks", deletionAsked(s.kept.Number, there, open))
	}
	gone := SeriesDeleted{Deleted: s.kept.Number, Kept: []int{}, TasksDeleted: []int{}}
	for _, n := range there {
		if chosen[0] == "delete-tasks" || chosen[0] == "delete-open" && slices.Contains(open, n) {
			if err := e.delete(n); err != nil {
				return SeriesDeleted{}, err
			}
			gone.TasksDeleted = append(gone.TasksDeleted, n)
			continue
		}
		gone.Kept = append(gone.Kept, n)
	}
	if err := e.store.DeleteSeries(s.kept.Number); err != nil {
		return SeriesDeleted{}, err
	}
	return gone, e.recordSeries(store.SeriesAct{Series: s.kept.Number, Did: "deleted", Became: chosen[0]})
}

func deletionAsked(n int, there, open []int) string {
	openSaid := "deletes none, as none is open"
	if len(open) > 0 {
		openSaid = "keeps the done and deletes " + numbersOrNothing(open)
	}
	return fmt.Sprintf("↻%d made %s. Say what becomes of them: --keep-tasks keeps all %d as ordinary tasks, --delete-open %s, --delete-tasks deletes all %d as well",
		n, manyOf(len(there), "task", "tasks"), len(there), openSaid, len(there))
}

// ShortDays writes days the short way, Mon 9 Nov, one after another.
func ShortDays(days []string) string {
	said := make([]string, 0, len(days))
	for _, one := range days {
		if d, err := domain.ParseDeadline(one); err == nil {
			said = append(said, domain.ShortDay(d))
		}
	}
	return strings.Join(said, " · ")
}

func followedText(made []MadeNext) string {
	var out strings.Builder
	for _, m := range made {
		if len(m.Missed) > 0 {
			fmt.Fprintf(&out, "↻%d missed %s\n", m.Series, strings.ReplaceAll(ShortDays(m.Missed), " · ", ", "))
		}
		if m.Task != 0 {
			fmt.Fprintf(&out, "↻%d made #%d for %s\n", m.Series, m.Task, ShortDays([]string{m.Day}))
		}
	}
	return out.String()
}

func occursText(o Occurs) string {
	day := ShortDays([]string{o.Day})
	if o.At != "" {
		day += " " + o.At
	}
	then := ShortDays(o.Then)
	switch {
	case o.Stopped:
		then = "nothing: the series is stopped"
	case o.after != "":
		then = o.after
	case then == "":
		then = "nothing: the series ends"
	}
	return fmt.Sprintf("\nrepeats  ↻%d %s\nday      %s\nthen     %s\n", o.Series, o.Says, day, then)
}

func repeatsOf(of map[int]Occurs) map[int]string {
	if of == nil {
		return nil
	}
	out := make(map[int]string, len(of))
	for n, o := range of {
		out[n] = fmt.Sprintf("↻%d %s", o.Series, o.Often)
	}
	return out
}

// StandingSaid is where a series stands, in words: active, or stopped since a day.
func StandingSaid(s Series) string {
	if s.State == "stopped" {
		return "stopped since " + ShortDays([]string{s.Stopped})
	}
	return "active"
}

// NextSaid is the next days of a series in words, or why there are none.
func NextSaid(s Series) string {
	switch {
	case s.Ended && len(s.Next) == 0:
		return "none, it has ended"
	case len(s.Next) == 0:
		return "none"
	case s.State == "stopped" && len(s.Open) == 0:
		return "if resumed today: " + ShortDays(s.Next)
	}
	return ShortDays(s.Next)
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
func TemplateSaid(t Template) string {
	said := []string{t.Title, t.Priority}
	if len(t.Labels) > 0 {
		said = append(said, strings.Join(t.Labels, ", "))
	}
	if len(t.Steps) > 0 {
		said = append(said, manyOf(len(t.Steps), "step", "steps"))
	}
	if t.Estimate != "" {
		said = append(said, t.Estimate)
	}
	if len(t.Objectives) > 0 {
		said = append(said, "serves "+numbersOrNothing(t.Objectives))
	}
	return strings.Join(said, " · ")
}

func seriesShownText(s Series) string {
	var out strings.Builder
	fmt.Fprintf(&out, "↻%d  %s  %s\n\n", s.Number, s.Title, StandingSaid(s))
	fmt.Fprintf(&out, "rule      %s\nnext      %s\n", s.Says, NextSaid(s))
	if strings.Contains(s.Says, "after the last is done") && !s.Ended {
		rule, _ := domain.ParseRule(s.Rule)
		fmt.Fprintf(&out, "then      %s\n", rule.Then())
	}
	fmt.Fprintf(&out, "template  %s\n", TemplateSaid(s.Template))
	soFar := fmt.Sprintf("%d done · %d skipped · %d missed", s.Done, s.Skipped, s.Missed)
	if s.TimeSpent != "" {
		soFar += " · " + s.TimeSpent
	}
	fmt.Fprintf(&out, "so far    %s\nopen      %s\n", soFar, domain.OrNothing(numbersOrNothing(s.Open)))
	if len(s.Occurrences) > 0 {
		out.WriteString("\noccurrences\n")
		for _, o := range s.Occurrences {
			what := "—"
			if o.Task != 0 {
				what = domain.Ref(o.Task)
			}
			is := o.Is
			if o.Archived {
				is += " · archived"
			}
			fmt.Fprintf(&out, "  %-10s  %-5s  %s", ShortDays([]string{o.Day}), what, is)
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
			line := fmt.Sprintf("  %s  %s", domain.FormatMoment(at), actSaid(a))
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

func actSaid(a SeriesAct) string {
	switch a.Did {
	case "made":
		return fmt.Sprintf("made #%d for %s", a.Task, ShortDays([]string{a.Day}))
	case "missed":
		return "missed " + ShortDays([]string{a.Day})
	case "skipped":
		return fmt.Sprintf("skipped #%d, made for %s", a.Task, ShortDays([]string{a.Day}))
	case "changed":
		return fmt.Sprintf("changed %s from %q to %q", a.Field, a.Was, a.Became)
	case "deleted":
		return "deleted, choosing " + a.Became
	}
	return a.Did
}

func seriesActedText(verb string, a Answer) string {
	s := *a.SeriesShown
	switch verb {
	case "stop":
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d stopped; %s stays open, and nothing more is made\n", s.Number, numbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d stopped; nothing more is made\n", s.Number)
	case "resume":
		if len(a.Followed) > 0 {
			return fmt.Sprintf("↻%d resumed · next %s, made as #%d\n", s.Number, ShortDays([]string{a.Followed[0].Day}), a.Followed[0].Task)
		}
		if len(s.Open) > 0 {
			return fmt.Sprintf("↻%d resumed; %s is still open, and the next comes when it is done or skipped\n", s.Number, numbersOrNothing(s.Open))
		}
		return fmt.Sprintf("↻%d resumed; it has ended, and nothing more is made\n", s.Number)
	case "edit":
		return fmt.Sprintf("changed ↻%d %s\nrule      %s\ntemplate  %s\n", s.Number, s.Title, s.Says, TemplateSaid(s.Template))
	}
	return seriesShownText(s)
}

func seriesDeletedText(d SeriesDeleted) string {
	said := fmt.Sprintf("↻%d was deleted", d.Deleted)
	if len(d.TasksDeleted) > 0 {
		said += " · " + numbersOrNothing(d.TasksDeleted) + " deleted"
	}
	if len(d.Kept) > 0 {
		said += fmt.Sprintf(" · %s kept as ordinary tasks", manyOf(len(d.Kept), "task", "tasks"))
	}
	return said + "\n"
}

var repeatsArgument = Argument{Name: "repeats", Kind: Bool, Repeats: true, Means: "keep only the occurrences of a series"}

var inSeriesArgument = Argument{Name: "series", Kind: Number, Of: "a series", Repeats: true, Means: "keep only the occurrences of this series, by its number"}

var seriesTooArgument = Argument{Name: "series", Kind: Bool, Repeats: true, Means: "change the series it is an occurrence of as well, for the ones to come: whatever of its title, priority, labels, estimate, steps and objectives is named"}

type begun struct {
	rule      domain.Rule
	from, day time.Time
}

func (e *Engine) beginning(in Input, c *Changes) (*begun, error) {
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
		deadline := domain.FormatDay(day)
		c.Deadline = &deadline
	}
	if c.State == nil {
		state := string(domain.Todo)
		c.State = &state
	}
	return &begun{rule: rule, from: from, day: day}, nil
}

func (e *Engine) repeating(n int, in Input) (bool, error) {
	every, err := everyNamed(in)
	if err != nil {
		return false, err
	}
	if !every && !in.Yes("series") {
		return false, nil
	}
	if err := e.mayRecur(); err != nil {
		return false, err
	}
	s, err := e.occurrenceOf(n)
	if in.Yes("series") && err != nil {
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

type occurrencesAsked struct {
	of     map[int]int
	series int
}

func (e *Engine) occurrencesIn(in Input) (occurrencesAsked, error) {
	series, named := in.Whole("series")
	if !in.Yes("repeats") && !named {
		return occurrencesAsked{}, nil
	}
	if err := e.mayRecur(); err != nil {
		return occurrencesAsked{}, err
	}
	if named {
		if _, err := e.seriesAt(series); err != nil {
			return occurrencesAsked{}, err
		}
	}
	of, err := e.store.Occurrences()
	return occurrencesAsked{of: of, series: series}, err
}

func (o occurrencesAsked) kept(tasks []task.Task) []task.Task {
	if o.of == nil {
		return tasks
	}
	return slices.DeleteFunc(tasks, func(t task.Task) bool {
		series, is := o.of[factsOf(t).Number]
		return !is || o.series != 0 && series != o.series
	})
}
