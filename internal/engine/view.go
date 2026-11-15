package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/view"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"cairn/internal/engine/task"
)

// View is how a view crosses to a machine: its name, its place, what it holds and how many tasks it keeps today.
type View struct {
	Name       string           `json:"name"`
	Place      int              `json:"place" jsonschema:"where it stands among the views, from 1; the first nine are opened by a key"`
	Holds      Holds            `json:"holds"`
	Keeps      int              `json:"keeps" jsonschema:"how many tasks it keeps today"`
	Of         int              `json:"of" jsonschema:"how many tasks are on the board, the archived ones left out"`
	Done       int              `json:"done" jsonschema:"how many of the tasks it keeps are done"`
	Gone       []string         `json:"gone" jsonschema:"each condition it holds whose objective or label is gone, which narrows nothing; empty when none is"`
	Outside    []view.Outside   `json:"outside,omitempty" jsonschema:"the tasks outside it that block, wait on or are linked to one it keeps, each once; left out unless the view is shown"`
	Was        string           `json:"was,omitempty" jsonschema:"the name it had before a change renamed it; left out otherwise"`
	Objectives []view.Objective `json:"objectives,omitempty" jsonschema:"each objective it holds that is still there, with its figures as the store counts them, whatever the view keeps; left out unless the view is shown"`
}

// Holds is what a view holds: the filters task list takes, and a search.
type Holds struct {
	State        string   `json:"state,omitempty"`
	Priority     string   `json:"priority,omitempty"`
	Label        []string `json:"label,omitempty"`
	Objective    []int    `json:"objective,omitempty"`
	Itself       bool     `json:"itself,omitempty" jsonschema:"the objectives themselves, leaving out their sub-objectives"`
	DueBefore    string   `json:"due_before,omitempty"`
	Focus        string   `json:"focus,omitempty" jsonschema:"in: only the tasks in focus; out: only the ones that are not"`
	Archived     bool     `json:"archived,omitempty"`
	WithArchived bool     `json:"with_archived,omitempty"`
	Search       string   `json:"search,omitempty"`
}

// Views is how the views cross: every one, in its place.
type Views struct {
	Views []View `json:"views"`
}

var searchArgument = Argument{Name: "search", Kind: Text, Means: "keep only the tasks carrying every word of it, wherever among their facts, the way task search looks"}

var viewFilters = append(slices.Clone(taskNarrowing), searchArgument)

var viewArgument = Argument{Name: "view", Kind: Text, Means: "the name of a view, whose filters keep the tasks first; any filter beside it narrows further, and the tasks outside it touching one it keeps are named under outside"}

func viewNamed(name string) Argument {
	return Argument{Name: "view", Kind: Text, Positional: true, Required: true, Means: "the name of the view, " + name}
}

// ViewNew keeps the filters task list takes, and a search, under a name.
var ViewNew = Operation{
	Thing: "view",
	Verb:  "new",
	Short: "Keep filters under a name",
	Does:  "Keep under a name the filters task list takes, and a search, so task list --view opens them again in one act. A name is one no view has yet, whatever its case or its accents, and a view holds at least one filter or a search. An objective is held by its number and a label by its name, kept as a label until it is deleted, so a rename follows. No task changes, and no trace is written.",
	Example: `cairn view new "Home automation" --objective 4
cairn view new Shopping --label shopping --state todo`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: View{},
	Takes: append([]Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "what the view is called"},
	}, viewFilters...),
	Run: func(on workedOn, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		name, err := viewName(views, in.Given["name"], -1)
		if err != nil {
			return Answer{}, err
		}
		v, err := on.Engine.holding(view.View{Name: name, Objectives: []int{}, Labels: []string{}}, in)
		if err != nil {
			return Answer{}, err
		}
		return on.Engine.keptViews(append(views, v), len(views), "")
	},
}

// ViewList lists every view, in its place, with what it holds and how many tasks it keeps.
var ViewList = Operation{
	Thing:   "view",
	Verb:    "list",
	Short:   "List the views",
	Does:    "List every view in its place, with what it holds, what of it is gone, and how many tasks it keeps today.",
	Example: "cairn view list",
	Level:   Read,
	Needs:   aStore,
	Answers: Views{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		listed := Views{Views: []View{}}
		for at := range views {
			one, _, err := on.Engine.viewOf(views, at)
			if err != nil {
				return Answer{}, err
			}
			listed.Views = append(listed.Views, one)
		}
		return Answer{ViewsKept: &listed}, nil
	},
}

// viewShow says what a view holds, what it keeps, and the tasks outside it touching one it keeps.
var viewShow = Operation{
	Thing:   "view",
	Verb:    "show",
	Short:   "Show what a view holds and keeps",
	Does:    "Show what a view holds, what of it is gone, how many tasks it keeps of how many, and the tasks outside it that block, wait on or are linked to one it keeps, each once.",
	Example: `cairn view show "Home automation"`,
	Level:   Read,
	Needs:   aStore,
	Answers: View{},
	Takes:   []Argument{viewNamed("to show")},
	Run: func(on workedOn, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		at, err := viewAt(views, in.Given["view"])
		if err != nil {
			return Answer{}, err
		}
		shown, kept, err := on.Engine.viewOf(views, at)
		if err != nil {
			return Answer{}, err
		}
		if shown.Outside, err = on.Engine.outside(kept); err != nil {
			return Answer{}, err
		}
		if shown.Objectives, err = on.Engine.viewObjectives(views[at].Objectives); err != nil {
			return Answer{}, err
		}
		return Answer{Viewed: &shown}, nil
	},
}

// ViewEdit renames a view, changes the filters it holds, or moves it to another place.
var ViewEdit = Operation{
	Thing: "view",
	Verb:  "edit",
	Short: "Rename a view, change its filters or move it",
	Does:  "Rename a view, change the filters it holds, or move it to another place, the others moving along. A filter named takes the place of the one the view held, and named empty takes it away; what is not named stays. No task changes, and no trace is written.",
	Example: `cairn view edit Shopping --name Groceries --label shopping --label garden
cairn view edit Groceries --place 1`,
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: View{},
	Takes: append([]Argument{
		viewNamed("to change"),
		{Name: "name", Kind: Text, Means: "the name it takes"},
		{Name: "place", Kind: Number, Of: "a place", Means: "the place it moves to, from 1"},
	}, viewFilters...),
	Run: func(on workedOn, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		at, err := viewAt(views, in.Given["view"])
		if err != nil {
			return Answer{}, err
		}
		if !slices.ContainsFunc(append([]Argument{{Name: "name"}, {Name: "place"}}, viewFilters...), func(a Argument) bool { return in.Named(a.Name) }) {
			return Answer{}, errNothingToChange
		}
		v, was := views[at], ""
		if given, named := in.Given["name"]; named {
			name, err := viewName(views, given, at)
			if err != nil {
				return Answer{}, err
			}
			if name != v.Name {
				was = v.Name
			}
			v.Name = name
		}
		if v, err = on.Engine.holding(v, in); err != nil {
			return Answer{}, err
		}
		views[at] = v
		if place, named := in.whole("place"); named {
			if place < 1 || place > len(views) {
				return Answer{}, Refusing("no-such-place", fmt.Sprintf("a view's place is 1 to %d", len(views)))
			}
			views = slices.Insert(slices.Delete(views, at, at+1), place-1, v)
			at = place - 1
		}
		return on.Engine.keptViews(views, at, was)
	},
}

// ViewDelete deletes a view, and nothing else.
var ViewDelete = Operation{
	Thing:   "view",
	Verb:    "delete",
	Short:   "Delete a view",
	Does:    "Delete a view, the views after it moving up a place. No task, objective or label changes, and no trace is written.",
	Example: "cairn view delete Groceries",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: view.Deleted{},
	Takes:   []Argument{viewNamed("to delete")},
	Run: func(on workedOn, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		at, err := viewAt(views, in.Given["view"])
		if err != nil {
			return Answer{}, err
		}
		name := views[at].Name
		if err := on.Engine.store.KeepViews(slices.Delete(views, at, at+1)); err != nil {
			return Answer{}, err
		}
		return Answer{ViewGone: &name}, nil
	},
}

func viewAt(views []view.View, name string) (int, error) {
	at := slices.IndexFunc(views, func(v view.View) bool { return label.Folded(v.Name) == label.Folded(name) })
	if at < 0 {
		return 0, Refusing("no-such-view", "no view is named "+strings.TrimSpace(name))
	}
	return at, nil
}

func viewName(views []view.View, raw string, except int) (string, error) {
	name, err := item.Title("a view", raw)
	if err != nil {
		return "", err
	}
	for i, v := range views {
		if i != except && label.Folded(v.Name) == label.Folded(name) {
			return "", Refusing("view-name-taken", "a view is already named "+v.Name)
		}
	}
	return name, nil
}

func (e *Engine) holding(v view.View, in Input) (view.View, error) {
	given := maps.Clone(in.Given)
	maps.DeleteFunc(given, func(_, value string) bool { return value == "" })
	for _, argument := range viewFilters {
		if err := accepted(Input{Given: given, Listed: in.Listed}, argument); err != nil {
			return v, err
		}
	}
	for name, into := range map[string]*string{"state": &v.State, "priority": &v.Priority, "deadline-before": &v.DueBefore} {
		if said, named := in.Given[name]; named {
			*into = said
		}
	}
	if said, named := in.Given["search"]; named {
		v.Search = strings.Join(strings.Fields(said), " ")
	}
	for name, into := range map[string]*bool{"itself": &v.Itself, "archived": &v.Archived, "with-archived": &v.WithArchived} {
		if in.Named(name) {
			*into = in.yes(name)
		}
	}
	focus, err := focusHeld(v.Focus, in)
	if err != nil {
		return v, err
	}
	v.Focus = focus
	if objectives := in.changed("objective"); objectives != nil {
		numbers, err := e.objectivesNamed(*objectives)
		if err != nil {
			return v, err
		}
		v.Objectives = numbers
	}
	if labels := in.changed("label"); labels != nil {
		written, err := label.ParseLabels(*labels)
		if err != nil {
			return v, err
		}
		v.Labels = written
	}
	if !view.NarrowsAnything(v) {
		return v, Refusing("view-holds-nothing", "a view holds at least one filter or a search")
	}
	for _, name := range v.Labels {
		if err := e.store.KeepLabel(name); err != nil {
			return v, err
		}
	}
	return v, nil
}

func (e *Engine) objectivesNamed(given []string) ([]int, error) {
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	numbers := []int{}
	for _, one := range given {
		if strings.TrimSpace(one) == "" {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(one))
		if err != nil {
			return nil, fmt.Errorf("%q: %w", one, errObjectiveRequired)
		}
		if _, there := t.By[n]; !there {
			return nil, fmt.Errorf("objective %d: %w", n, objective.ErrNoSuchObjective)
		}
		numbers = append(numbers, n)
	}
	slices.Sort(numbers)
	return slices.Compact(numbers), nil
}

func focusHeld(was string, in Input) (string, error) {
	if in.yes("focus") && in.yes("no-focus") {
		return was, Refusing("focus-and-not", "--focus and --no-focus keep nothing together: name one")
	}
	if in.yes("focus") {
		return "in", nil
	}
	if in.yes("no-focus") {
		return "out", nil
	}
	if in.Named("focus") && was == "in" || in.Named("no-focus") && was == "out" {
		return "", nil
	}
	return was, nil
}

func (e *Engine) keptViews(views []view.View, at int, was string) (Answer, error) {
	if err := e.store.KeepViews(views); err != nil {
		return Answer{}, err
	}
	kept, _, err := e.viewOf(views, at)
	kept.Was = was
	return Answer{Viewed: &kept}, err
}

func (e *Engine) viewOf(views []view.View, at int) (View, []task.Task, error) {
	v := views[at]
	kept, s, err := e.viewing(v)
	if err != nil {
		return View{}, nil, err
	}
	board, err := e.narrowed(&filter.Narrowing{})
	if err != nil {
		return View{}, nil, err
	}
	done := 0
	for _, t := range kept {
		if t.Carried().State == item.Done {
			done++
		}
	}
	return View{
		Name: v.Name, Place: at + 1, Keeps: len(kept), Of: len(board), Done: done, Gone: listed(s.Lost),
		Holds: Holds{State: v.State, Priority: v.Priority, Label: v.Labels, Objective: v.Objectives, Itself: v.Itself,
			DueBefore: v.DueBefore, Focus: v.Focus, Archived: v.Archived, WithArchived: v.WithArchived, Search: v.Search},
	}, kept, nil
}

func (e *Engine) viewObjectives(numbers []int) ([]view.Objective, error) {
	var out []view.Objective
	for _, n := range numbers {
		o, err := e.showObjective(n)
		if errors.Is(err, objective.ErrNoSuchObjective) {
			continue
		}
		if err != nil {
			return nil, err
		}
		p, err := e.objectiveProgress(n)
		if err != nil {
			return nil, err
		}
		under, err := e.tasksUnder(n)
		if err != nil {
			return nil, err
		}
		out = append(out, view.Objective{Number: n, Title: o.Carried().Title, Progress: item.OfProgress(p), TimeSpent: calendar.FormatDuration(reviewOf(under).Spent)})
	}
	return out, nil
}

func (e *Engine) viewing(v view.View) ([]task.Task, view.Seen, error) {
	kept := &filter.Narrowing{}
	s, err := e.seeing(v, kept)
	if err != nil {
		return nil, view.Seen{}, err
	}
	tasks, err := e.narrowed(kept)
	if err != nil {
		return nil, view.Seen{}, err
	}
	return s.Kept(tasks), s, nil
}

func (e *Engine) seeing(v view.View, kept *filter.Narrowing) (view.Seen, error) {
	s := view.Seen{Name: v.Name, Words: strings.Fields(v.Search)}
	t, err := e.tree()
	if err != nil {
		return s, err
	}
	held, err := e.everyLabel()
	if err != nil {
		return s, err
	}
	conditions := len(v.Objectives) + len(v.Labels)
	tell := func(named bool, told func()) {
		if named {
			conditions++
			told()
		}
	}
	tell(v.State != "", func() { kept.InColumn(v.State) })
	tell(v.Priority != "", func() { kept.Carrying(v.Priority) })
	tell(v.DueBefore != "", func() { kept.DueBefore(v.DueBefore) })
	tell(v.Focus != "", func() { kept.InFocus(v.Focus == "in") })
	tell(v.Archived, kept.Archived)
	tell(len(s.Words) > 0, func() {})
	if v.WithArchived {
		kept.WithArchived()
	}
	for _, n := range v.Objectives {
		if _, there := t.By[n]; !there {
			s.Lost = append(s.Lost, fmt.Sprintf("objective %d", n))
			continue
		}
		reached := t.Reach(n)
		if v.Itself {
			reached = []int{n}
		}
		kept.Serving(reached)
	}
	var live []string
	for _, name := range v.Labels {
		if _, there := held[name]; !there {
			s.Lost = append(s.Lost, "label "+name)
			continue
		}
		live = append(live, name)
	}
	if len(live) > 0 {
		kept.Labelled(live)
	}
	s.Nothing = len(s.Lost) == conditions
	return s, nil
}

func (e *Engine) inView(in Input, kept *filter.Narrowing) (*view.Seen, error) {
	name, named := in.Given["view"]
	if !named {
		return nil, nil
	}
	views, err := e.store.Views()
	if err != nil {
		return nil, err
	}
	at, err := viewAt(views, name)
	if err != nil {
		return nil, err
	}
	s, err := e.seeing(views[at], kept)
	return &s, err
}

func (e *Engine) outside(kept []task.Task) ([]view.Outside, error) {
	board, err := e.narrowed(&filter.Narrowing{})
	if err != nil {
		return nil, err
	}
	inside := make([]task.Carried, 0, len(kept))
	for _, t := range kept {
		inside = append(inside, t.Carried())
	}
	out := []view.Outside{}
	for _, t := range board {
		o := t.Carried()
		if slices.ContainsFunc(inside, func(c task.Carried) bool { return c.Number == o.Number }) {
			continue
		}
		var why []view.Touch
		for _, is := range view.Touching {
			for _, c := range inside {
				if view.Touches(is, o, c) {
					why = append(why, view.Touch{Is: is, Of: c.Number})
				}
			}
		}
		if len(why) > 0 {
			out = append(out, view.Outside{Number: o.Number, Title: o.Title, State: string(o.State), Why: why})
		}
	}
	slices.SortFunc(out, func(a, b view.Outside) int { return a.Number - b.Number })
	return out, nil
}

func (e *Engine) viewsHolding(name string) (int, error) {
	views, err := e.store.Views()
	if err != nil {
		return 0, err
	}
	held := 0
	for _, v := range views {
		if slices.Contains(v.Labels, name) {
			held++
		}
	}
	return held, nil
}

func (e *Engine) relabelViews(from, to string) error {
	views, err := e.store.Views()
	if err != nil {
		return err
	}
	for i, v := range views {
		if at := slices.Index(v.Labels, from); at >= 0 {
			v.Labels[at] = to
			slices.Sort(v.Labels)
			views[i].Labels = slices.Compact(v.Labels)
		}
	}
	return e.store.KeepViews(views)
}

func (e *Engine) unview(goners []int, how string) error {
	views, err := e.store.Views()
	if err != nil {
		return err
	}
	kept := []view.View{}
	for _, v := range views {
		if len(view.Naming([]view.View{v}, goners)) == 0 {
			kept = append(kept, v)
			continue
		}
		if how == "delete" {
			continue
		}
		left := v
		left.Objectives = slices.DeleteFunc(slices.Clone(v.Objectives), func(n int) bool { return slices.Contains(goners, n) })
		if view.NarrowsAnything(left) {
			v = left
		}
		kept = append(kept, v)
	}
	return e.store.KeepViews(kept)
}

func viewKeptText(v View, said string) string {
	if v.Was != "" {
		said += " (was " + v.Was + ")"
	}
	return fmt.Sprintf("%s · %s · keeps %d of %s\n", said, view.CmdKey(v.Place), v.Keeps, item.ManyOf(v.Of, "task", "tasks"))
}

// HoldsSays is what a view holds, the way a person reads it: each filter and what it is on, what of it is gone said so.
func HoldsSays(v View) string {
	var said []string
	gone := func(what string) string {
		if slices.Contains(v.Gone, what) {
			return what + ", gone"
		}
		return what
	}
	for _, n := range v.Holds.Objective {
		what := gone(fmt.Sprintf("objective %d", n))
		if v.Holds.Itself {
			what += " itself"
		}
		said = append(said, what)
	}
	for _, name := range v.Holds.Label {
		said = append(said, gone("label "+name))
	}
	for _, one := range []struct {
		named bool
		what  string
	}{
		{v.Holds.State != "", "state " + v.Holds.State},
		{v.Holds.Priority != "", "priority " + v.Holds.Priority},
		{v.Holds.DueBefore != "", "deadline before " + v.Holds.DueBefore},
		{v.Holds.Focus == "in", "in focus"},
		{v.Holds.Focus == "out", "not in focus"},
		{v.Holds.Archived, "archived"},
		{v.Holds.WithArchived, "with archived"},
		{v.Holds.Search != "", fmt.Sprintf("search %q", v.Holds.Search)},
	} {
		if one.named {
			said = append(said, one.what)
		}
	}
	return strings.Join(said, " · ")
}

func viewsText(views []View) string {
	if len(views) == 0 {
		return "there is no view\n"
	}
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "#\tname\tholds\tkeeps")
	for _, v := range views {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%d\n", v.Place, v.Name, HoldsSays(v), v.Keeps)
	}
	_ = tw.Flush()
	return out.String()
}

func viewShownText(v View) string {
	lines := []string{
		fmt.Sprintf("view     %s  (%s)", v.Name, view.CmdKey(v.Place)),
		"holds    " + HoldsSays(v),
		fmt.Sprintf("keeps    %d of %s · %d done", v.Keeps, item.ManyOf(v.Of, "task", "tasks"), v.Done),
	}
	if len(v.Gone) > 0 {
		lines = append(lines, "gone     "+strings.Join(v.Gone, " · "))
	}
	if len(v.Outside) > 0 {
		var outside []string
		for _, o := range v.Outside {
			var why []string
			for _, w := range o.Why {
				why = append(why, fmt.Sprintf("%s #%d", w.Is, w.Of))
			}
			outside = append(outside, fmt.Sprintf("#%d %s", o.Number, strings.Join(why, ", ")))
		}
		lines = append(lines, "outside  "+strings.Join(outside, " · "))
	}
	for _, o := range v.Objectives {
		lines = append(lines, fmt.Sprintf("objective #%d %s  %s", o.Number, o.Title, view.StoreFigures(o)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func lostText(a Answer) string {
	if a.InView == "" || len(a.Lost) == 0 {
		return ""
	}
	name := view.TypedName(a.InView)
	hint := fmt.Sprintf("  cairn view edit %s, or cairn view delete %s\n", name, name)
	if a.KeepsNothing {
		return fmt.Sprintf("every condition view %s holds is gone, so it keeps nothing: %s\n", a.InView, strings.Join(a.Lost, ", ")) + hint
	}
	verb := "is gone and narrows nothing"
	if len(a.Lost) > 1 {
		verb = "are gone and narrow nothing"
	}
	return fmt.Sprintf("view %s: %s %s\n", a.InView, strings.Join(a.Lost, " and "), verb) + hint
}
