package objective

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// ErrTooDeep is returned when an objective would sit three levels deep.
var ErrTooDeep = errors.New("an objective goes two levels deep at most")

// ErrUnderItself is returned when an objective is put under itself.
var ErrUnderItself = errors.New("an objective does not sit under itself")

// ErrCountedTwice is returned when a task would be counted twice in one tree: on a parent and through one of its sub-objectives.
var ErrCountedTwice = errors.New("nothing was changed")

// MayHang says whether o may sit under parent, given whether o holds sub-objectives of its own.
func MayHang(o, parent Objective, holds bool) error {
	if parent.number == o.number {
		return fmt.Errorf("#%d %s: %w", o.number, o.title, ErrUnderItself)
	}
	if parent.parent != 0 {
		return fmt.Errorf("#%d %s already sits under #%d; %w", parent.number, parent.title, parent.parent, ErrTooDeep)
	}
	if holds {
		return fmt.Errorf("#%d %s holds sub-objectives of its own; %w", o.number, o.title, ErrTooDeep)
	}
	return nil
}

// CountedOnce refuses a set of objectives a task would serve in which one sits under another, naming the first tree it would be counted twice in.
func CountedOnce(task int, served []Objective) error {
	by := map[int]Objective{}
	for _, o := range served {
		by[o.number] = o
	}
	through := map[int][]Objective{}
	for _, o := range served {
		if _, both := by[o.parent]; o.parent != 0 && both {
			through[o.parent] = append(through[o.parent], o)
		}
	}
	if len(through) == 0 {
		return nil
	}
	first := slices.Min(slicesOfKeys(through))
	under := through[first]
	slices.SortFunc(under, func(a, b Objective) int { return a.number - b.number })
	names := make([]string, 0, len(under))
	for _, o := range under {
		names = append(names, refOf(o))
	}
	who := "the task"
	if task != 0 {
		who = item.Ref(task)
	}
	return fmt.Errorf("%s is already counted in %s through %s; %w", who, refOf(by[first]), joined(names), ErrCountedTwice)
}

func slicesOfKeys(m map[int][]Objective) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func refOf(o Objective) string { return item.Ref(o.number) + " " + o.title }

func joined(names []string) string {
	if len(names) == 1 {
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// ProposedReached says whether a parent is proposed as reached: it holds sub-objectives, every one of them is reached, and it is not.
func ProposedReached(parent Objective, children []Objective) bool {
	if len(children) == 0 || parent.state == Reached {
		return false
	}
	for _, c := range children {
		if c.state != Reached {
			return false
		}
	}
	return true
}

// Family is where an objective sits: the one it sits under, none when it stands on its own, and the sub-objectives under it with what each counts.
type Family struct {
	Parent   Objective
	Children []Child
	Proposed bool
	WaitsOn  []Objective
	Awaited  []Objective
}

// Child is one sub-objective, with how far along it is and the time spent on its tasks.
type Child struct {
	Objective Objective
	Progress  item.Progress
	Spent     time.Duration
}

func (f Family) waitsLine() string {
	names := make([]string, 0, len(f.WaitsOn))
	for _, o := range f.WaitsOn {
		names = append(names, refOf(o))
	}
	return strings.Join(names, ", ")
}

func (f Family) parentLine() string {
	if f.Parent.number == 0 {
		return ""
	}
	return refOf(f.Parent)
}

// Children names each sub-objective on a line of its own, with where it stands, how many of its tasks are done and the time spent on them; nothing when it holds none.
func Children(f Family) string {
	if len(f.Children) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\nsub-objectives\n")
	for _, c := range f.Children {
		fmt.Fprintf(&out, "  #%d  %s%s\n", c.Objective.number, c.Objective.title,
			dotted(string(c.Objective.state), counts(c.Progress), calendar.FormatDuration(c.Spent)))
	}
	return out.String()
}

// Proposal says a parent may be put reached, every sub-objective of it being reached — said on the parent itself, or naming it from the act that reached the last one.
func Proposal(parent Objective, itself bool) string {
	if itself {
		return "\nevery sub-objective is reached: it may be put reached too\n"
	}
	return fmt.Sprintf("\nevery sub-objective of %s is reached: it may be put reached too\n", refOf(parent))
}

// Apart renders a parent's tasks in two: those put on it directly, and those it counts through its sub-objectives.
func Apart(own, through []string) string {
	return block("tasks on it directly", "nothing is on it directly", own) +
		block("tasks in its sub-objectives", "nothing is in its sub-objectives", through)
}

func block(heading, none string, lines []string) string {
	if len(lines) == 0 {
		return "\n" + none + "\n"
	}
	var out strings.Builder
	out.WriteString("\n" + heading + "\n")
	for _, l := range lines {
		fmt.Fprintf(&out, "  %s\n", l)
	}
	return out.String()
}

func counts(p item.Progress) string {
	if p.Tasks.Total == 0 {
		return ""
	}
	return fmt.Sprintf("%d of %d tasks", p.Tasks.Done, p.Tasks.Total)
}

func dotted(said ...string) string {
	var out string
	for _, s := range said {
		if s != "" {
			out += "  · " + s
		}
	}
	return out
}

// ErrNotASubObjective is returned when a deletion names, to keep or to delete, an objective that is not a sub-objective of what is deleted.
var ErrNotASubObjective = errors.New("only a sub-objective of what is deleted is kept or deleted on its own")

// WentWith is what a deletion took with it: the sub-objectives kept, each standing on its own, those deleted, and the tasks detached or deleted.
type WentWith struct {
	Kept     []int
	Gone     []int
	Detached []int
	Deleted  []int
}

// Deletion is what a deletion of objectives was told becomes of their sub-objectives, tasks and views.
type Deletion struct {
	Subs, Tasks  string
	Views        string
	Keep, Delete []int
}

// NamesOnly refuses a sub-objective kept or deleted that is not one of what is deleted.
func (d Deletion) NamesOnly(numbers []int, t Tree) error {
	held := func(c int) bool {
		parent := t.By[c].Carried().Parent
		return parent != 0 && slices.Contains(numbers, parent)
	}
	for _, c := range append(slices.Clone(d.Keep), d.Delete...) {
		if !held(c) {
			return fmt.Errorf("#%d is not a sub-objective of %s; %w", c, item.Ref(numbers[0]), ErrNotASubObjective)
		}
		if slices.Contains(d.Keep, c) && slices.Contains(d.Delete, c) {
			return fmt.Errorf("#%d is both kept and deleted; %w", c, ErrNotASubObjective)
		}
	}
	return nil
}

// AskingWhatGoes is a deletion refused until it is said what goes with what is deleted.
type AskingWhatGoes struct {
	Said                  string
	Carries, Holds, Named bool
}

func (a *AskingWhatGoes) Error() string { return a.Said }

// Is says which of the reasons a deletion was refused for it carries.
func (a *AskingWhatGoes) Is(target error) bool {
	return target == ErrObjectiveCarriesTasks && a.Carries || target == ErrObjectiveHoldsSubObjectives && a.Holds || target == ErrObjectiveNamedByAView && a.Named
}

// ErrObjectiveNamedByAView is why an objective a view names is not deleted until it is said what becomes of the view.
var ErrObjectiveNamedByAView = errors.New("a view names it")

// WentText is what a deletion took with it, for a person.
func WentText(w *WentWith) string {
	if w == nil {
		return ""
	}
	var out strings.Builder
	for _, n := range w.Kept {
		fmt.Fprintf(&out, "objective %d stands on its own now\n", n)
	}
	for _, n := range w.Gone {
		fmt.Fprintf(&out, "objective %d was deleted with it\n", n)
	}
	for _, n := range w.Detached {
		fmt.Fprintf(&out, "task %d was detached from it\n", n)
	}
	for _, n := range w.Deleted {
		fmt.Fprintf(&out, "task %d was deleted with it\n", n)
	}
	return out.String()
}
