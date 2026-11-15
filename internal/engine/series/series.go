package series

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Template is what the next occurrence of a series is made from.
type Template struct {
	Title      string   `json:"title"`
	Priority   string   `json:"priority"`
	Labels     []string `json:"labels"`
	Steps      []string `json:"steps"`
	Estimate   string   `json:"estimate"`
	Objectives []int    `json:"objectives"`
	Arrives    string   `json:"arrives" jsonschema:"the column its occurrences arrive in: backlog, todo or current"`
}

// Arrivals are the columns a series' occurrences may arrive in: a column that needs a reason, or done, is no place for work still to do.
var Arrivals = []string{string(item.Backlog), string(item.Todo), string(item.Current)}

// ErrNotAnArrival is returned when a series is told to put its occurrences in a column they cannot arrive in.
var ErrNotAnArrival = errors.New("an occurrence arrives in backlog, todo or current")

// ArrivalOf is the column an occurrence arrives in: the state given when it is one, todo otherwise.
func ArrivalOf(state item.State) string {
	if slices.Contains(Arrivals, string(state)) {
		return string(state)
	}
	return string(item.Todo)
}

// Occurrence is one day of a series: the task made for it, or none when it was missed.
type Occurrence struct {
	Day       string `json:"day"`
	Task      int    `json:"task" jsonschema:"the number of the task made for that day; 0 when none was"`
	Is        string `json:"is" jsonschema:"the column the task stands in, or skipped, missed or deleted"`
	Archived  bool   `json:"archived"`
	TimeSpent string `json:"time_spent"`
}

// Act is one thing that happened to a series, as it crosses to a machine.
type Act struct {
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

// Deleted is how deleting a series answers: which, and what became of the tasks it made.
type Deleted struct {
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
	After   string   `json:"-"`
}

// MadeNext is one occurrence a series made because an act left it with none open, and the days it missed first.
type MadeNext struct {
	Series int      `json:"series"`
	Task   int      `json:"task"`
	Day    string   `json:"day"`
	Missed []string `json:"missed,omitempty"`
}

// IsOpen says whether a task is neither archived nor done.
func IsOpen(t task.Task) bool {
	f := t.Carried()
	return !f.Archived && f.State != item.Done
}

// MadeFromSeries is the act written in the trace of a task a series made, naming the series.
const MadeFromSeries = "made from series"

// ShortDays writes days the short way, Mon 9 Nov, one after another.
func ShortDays(days []string) string {
	said := make([]string, 0, len(days))
	for _, one := range days {
		if d, err := calendar.ParseDeadline(one); err == nil {
			said = append(said, ShortDay(d))
		}
	}
	return strings.Join(said, " · ")
}

// FollowedText is, for a person, each occurrence a series made, and the days it missed first.
func FollowedText(made []MadeNext) string {
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

// OccursText is, for a person, the series a task is an occurrence of.
func OccursText(o Occurs) string {
	day := ShortDays([]string{o.Day})
	if o.At != "" {
		day += " " + o.At
	}
	then := ShortDays(o.Then)
	if o.Stopped {
		then = "nothing: the series is stopped"
	} else if o.After != "" {
		then = o.After
	} else if then == "" {
		then = "nothing: the series ends"
	}
	return fmt.Sprintf("\nrepeats  ↻%d %s\nday      %s\nthen     %s\n", o.Series, o.Says, day, then)
}

// RepeatsOf is how often each task that is an occurrence comes back.
func RepeatsOf(of map[int]Occurs) map[int]string {
	if of == nil {
		return nil
	}
	out := make(map[int]string, len(of))
	for n, o := range of {
		out[n] = fmt.Sprintf("↻%d %s", o.Series, o.Often)
	}
	return out
}

// ActSaid is an act of a series, for a person.
func ActSaid(a Act) string {
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

// Begun is a series begun: its rule, the day it counts from, and its first day.
type Begun struct {
	Rule      Rule
	From, Day time.Time
}

// OccurrencesAsked is the occurrences a listing asked for: every task, or those of one series.
type OccurrencesAsked struct {
	Of     map[int]int
	Series int
}

// Kept is the tasks the listing keeps.
func (o OccurrencesAsked) Kept(tasks []task.Task) []task.Task {
	if o.Of == nil {
		return tasks
	}
	return slices.DeleteFunc(tasks, func(t task.Task) bool {
		series, is := o.Of[t.Carried().Number]
		return !is || o.Series != 0 && series != o.Series
	})
}
