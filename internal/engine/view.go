package engine

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
)

// View is how a view crosses to a machine: its name, its place, what it holds and how many tasks it keeps today.
type View struct {
	Name       string          `json:"name"`
	Place      int             `json:"place" jsonschema:"where it stands among the views, from 1; the first nine are opened by a key"`
	Holds      Holds           `json:"holds"`
	Keeps      int             `json:"keeps" jsonschema:"how many tasks it keeps today"`
	Of         int             `json:"of" jsonschema:"how many tasks are on the board, the archived ones left out"`
	Done       int             `json:"done" jsonschema:"how many of the tasks it keeps are done"`
	Gone       []string        `json:"gone" jsonschema:"each condition it holds whose objective or label is gone, which narrows nothing; empty when none is"`
	Outside    []Outside       `json:"outside,omitempty" jsonschema:"the tasks outside it that block, wait on or are linked to one it keeps, each once; left out unless the view is shown"`
	Was        string          `json:"was,omitempty" jsonschema:"the name it had before a change renamed it; left out otherwise"`
	Objectives []ViewObjective `json:"objectives,omitempty" jsonschema:"each objective it holds that is still there, with its figures as the store counts them, whatever the view keeps; left out unless the view is shown"`
}

// ViewObjective is an objective a view holds, with the store's figures for it: the same in a view and out of one.
type ViewObjective struct {
	Number    int                   `json:"number"`
	Title     string                `json:"title"`
	Progress  *domain.ProgressShape `json:"progress" jsonschema:"how far along it is over its whole subtree, each task once; null when nothing serves it"`
	TimeSpent string                `json:"time_spent" jsonschema:"how long has been spent on the tasks serving it, as 1h30m; empty when nothing has"`
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

// ViewDeleted is how deleting a view answers: the name of the view deleted.
type ViewDeleted struct {
	Deleted string `json:"deleted"`
}

// Outside is a task outside a view that touches one it keeps, and every way it does.
type Outside struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	State  string  `json:"state"`
	Why    []Touch `json:"why"`
}

// Touch is one way a task outside a view touches one inside it: it blocks it, waits on it, or is linked to it.
type Touch struct {
	Is string `json:"is" jsonschema:"blocks, waits on or linked to"`
	Of int    `json:"of" jsonschema:"the number of the task inside the view"`
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: View{},
	Takes: append([]Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "what the view is called"},
	}, viewFilters...),
	Run: func(on On, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		name, err := viewName(views, in.Given["name"], -1)
		if err != nil {
			return Answer{}, err
		}
		v, err := on.Engine.holding(store.View{Name: name, Objectives: []int{}, Labels: []string{}}, in)
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
	Needs:   AStore,
	Answers: Views{},
	Run: func(on On, _ Input) (Answer, error) {
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

// ViewShow says what a view holds, what it keeps, and the tasks outside it touching one it keeps.
var ViewShow = Operation{
	Thing:   "view",
	Verb:    "show",
	Short:   "Show what a view holds and keeps",
	Does:    "Show what a view holds, what of it is gone, how many tasks it keeps of how many, and the tasks outside it that block, wait on or are linked to one it keeps, each once.",
	Example: `cairn view show "Home automation"`,
	Level:   Read,
	Needs:   AStore,
	Answers: View{},
	Takes:   []Argument{viewNamed("to show")},
	Run: func(on On, in Input) (Answer, error) {
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: View{},
	Takes: append([]Argument{
		viewNamed("to change"),
		{Name: "name", Kind: Text, Means: "the name it takes"},
		{Name: "place", Kind: Number, Of: "a place", Means: "the place it moves to, from 1"},
	}, viewFilters...),
	Run: func(on On, in Input) (Answer, error) {
		views, err := on.Engine.store.Views()
		if err != nil {
			return Answer{}, err
		}
		at, err := viewAt(views, in.Given["view"])
		if err != nil {
			return Answer{}, err
		}
		if !slices.ContainsFunc(append([]Argument{{Name: "name"}, {Name: "place"}}, viewFilters...), func(a Argument) bool { return in.Named(a.Name) }) {
			return Answer{}, ErrNothingToChange
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
		if place, named := in.Whole("place"); named {
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
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: ViewDeleted{},
	Takes:   []Argument{viewNamed("to delete")},
	Run: func(on On, in Input) (Answer, error) {
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

func viewAt(views []store.View, name string) (int, error) {
	at := slices.IndexFunc(views, func(v store.View) bool { return domain.Folded(v.Name) == domain.Folded(name) })
	if at < 0 {
		return -1, Refusing("no-such-view", "no view is named "+strings.TrimSpace(name))
	}
	return at, nil
}

func viewName(views []store.View, raw string, except int) (string, error) {
	name, err := domain.Title("a view", raw)
	if err != nil {
		return "", err
	}
	for i, v := range views {
		if i != except && domain.Folded(v.Name) == domain.Folded(name) {
			return "", Refusing("view-name-taken", "a view is already named "+v.Name)
		}
	}
	return name, nil
}

func (e *Engine) holding(v store.View, in Input) (store.View, error) {
	given := maps.Clone(in.Given)
	maps.DeleteFunc(given, func(_, value string) bool { return value == "" })
	for _, argument := range viewFilters {
		if err := accepted(Input{Given: given, Listed: in.Listed}, argument); err != nil {
			return v, err
		}
	}
	for name, into := range map[string]*string{"state": &v.State, "priority": &v.Priority, "due-before": &v.DueBefore} {
		if said, named := in.Given[name]; named {
			*into = said
		}
	}
	if said, named := in.Given["search"]; named {
		v.Search = strings.Join(strings.Fields(said), " ")
	}
	for name, into := range map[string]*bool{"itself": &v.Itself, "archived": &v.Archived, "with-archived": &v.WithArchived} {
		if in.Named(name) {
			*into = in.Yes(name)
		}
	}
	focus, err := focusHeld(v.Focus, in)
	if err != nil {
		return v, err
	}
	v.Focus = focus
	if objectives := in.Changed("objective"); objectives != nil {
		numbers, err := e.objectivesNamed(*objectives)
		if err != nil {
			return v, err
		}
		v.Objectives = numbers
	}
	if labels := in.Changed("label"); labels != nil {
		written, err := domain.ParseLabels(*labels)
		if err != nil {
			return v, err
		}
		v.Labels = written
	}
	if !narrowsAnything(v) {
		return v, Refusing("view-holds-nothing", "a view holds at least one filter or a search")
	}
	for _, label := range v.Labels {
		if err := e.store.KeepLabel(label); err != nil {
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
			return nil, fmt.Errorf("%q: %w", one, ErrObjectiveRequired)
		}
		if _, there := t.by[n]; !there {
			return nil, fmt.Errorf("objective %d: %w", n, ErrNoSuchObjective)
		}
		numbers = append(numbers, n)
	}
	slices.Sort(numbers)
	return slices.Compact(numbers), nil
}

func focusHeld(was string, in Input) (string, error) {
	if in.Yes("focus") && in.Yes("no-focus") {
		return was, Refusing("focus-and-not", "--focus and --no-focus keep nothing together: name one")
	}
	switch {
	case in.Yes("focus"):
		return "in", nil
	case in.Yes("no-focus"):
		return "out", nil
	case in.Named("focus") && was == "in", in.Named("no-focus") && was == "out":
		return "", nil
	}
	return was, nil
}

func narrowsAnything(v store.View) bool {
	return v.Focus != "" || v.State != "" || v.Priority != "" || v.DueBefore != "" || v.Archived || v.Search != "" || len(v.Objectives) > 0 || len(v.Labels) > 0
}

func (e *Engine) keptViews(views []store.View, at int, was string) (Answer, error) {
	if err := e.store.KeepViews(views); err != nil {
		return Answer{}, err
	}
	kept, _, err := e.viewOf(views, at)
	kept.Was = was
	return Answer{Viewed: &kept}, err
}

func (e *Engine) viewOf(views []store.View, at int) (View, []task.Task, error) {
	v := views[at]
	kept, s, err := e.viewing(v)
	if err != nil {
		return View{}, nil, err
	}
	board, err := e.narrowed(&domain.Narrowing{})
	if err != nil {
		return View{}, nil, err
	}
	done := 0
	for _, t := range kept {
		if factsOf(t).State == domain.Done {
			done++
		}
	}
	return View{
		Name: v.Name, Place: at + 1, Keeps: len(kept), Of: len(board), Done: done, Gone: listed(s.lost),
		Holds: Holds{State: v.State, Priority: v.Priority, Label: v.Labels, Objective: v.Objectives, Itself: v.Itself,
			DueBefore: v.DueBefore, Focus: v.Focus, Archived: v.Archived, WithArchived: v.WithArchived, Search: v.Search},
	}, kept, nil
}

func (e *Engine) viewObjectives(numbers []int) ([]ViewObjective, error) {
	var out []ViewObjective
	for _, n := range numbers {
		o, err := e.showObjective(n)
		if errors.Is(err, ErrNoSuchObjective) {
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
		out = append(out, ViewObjective{Number: n, Title: o.Carried().Title, Progress: domain.OfProgress(p), TimeSpent: domain.FormatDuration(reviewOf(under).Spent)})
	}
	return out, nil
}

type seen struct {
	name    string
	lost    []string
	nothing bool
	words   []string
}

func (e *Engine) viewing(v store.View) ([]task.Task, seen, error) {
	kept := &domain.Narrowing{}
	s, err := e.seeing(v, kept)
	if err != nil {
		return nil, seen{}, err
	}
	tasks, err := e.narrowed(kept)
	if err != nil {
		return nil, seen{}, err
	}
	return s.kept(tasks), s, nil
}

func (e *Engine) seeing(v store.View, kept *domain.Narrowing) (seen, error) {
	s := seen{name: v.Name, words: strings.Fields(v.Search)}
	t, err := e.tree()
	if err != nil {
		return s, err
	}
	held, err := e.everyLabel()
	if err != nil {
		return s, err
	}
	conditions := len(v.Objectives) + len(v.Labels)
	for _, one := range []struct {
		named bool
		tell  func()
	}{
		{v.State != "", func() { kept.InColumn(v.State) }},
		{v.Priority != "", func() { kept.Carrying(v.Priority) }},
		{v.DueBefore != "", func() { kept.DueBefore(v.DueBefore) }},
		{v.Focus != "", func() { kept.InFocus(v.Focus == "in") }},
		{v.Archived, kept.Archived},
		{len(s.words) > 0, func() {}},
	} {
		if one.named {
			conditions++
			one.tell()
		}
	}
	if v.WithArchived {
		kept.WithArchived()
	}
	for _, n := range v.Objectives {
		if _, there := t.by[n]; !there {
			s.lost = append(s.lost, fmt.Sprintf("objective %d", n))
			continue
		}
		reached := t.reach(n)
		if v.Itself {
			reached = []int{n}
		}
		kept.Serving(reached)
	}
	var live []string
	for _, label := range v.Labels {
		if _, there := held[label]; !there {
			s.lost = append(s.lost, "label "+label)
			continue
		}
		live = append(live, label)
	}
	if len(live) > 0 {
		kept.Labelled(live)
	}
	s.nothing = len(s.lost) == conditions
	return s, nil
}

func (s seen) kept(tasks []task.Task) []task.Task {
	if s.nothing {
		return []task.Task{}
	}
	return slices.DeleteFunc(tasks, func(t task.Task) bool { return len(s.words) > 0 && !task.Matching(t, s.words) })
}

func (e *Engine) inView(in Input, kept *domain.Narrowing) (*seen, error) {
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

var touching = []string{"blocks", "waits on", "linked to"}

func touches(is string, outside, inside task.Carried) bool {
	switch is {
	case "blocks":
		return slices.Contains(inside.Blockers, outside.Number)
	case "waits on":
		return slices.Contains(outside.Blockers, inside.Number)
	}
	return slices.Contains(outside.Linked, inside.Number) || slices.Contains(inside.Linked, outside.Number)
}

func (e *Engine) outside(kept []task.Task) ([]Outside, error) {
	board, err := e.narrowed(&domain.Narrowing{})
	if err != nil {
		return nil, err
	}
	inside := make([]task.Carried, 0, len(kept))
	for _, t := range kept {
		inside = append(inside, t.Carried())
	}
	out := []Outside{}
	for _, t := range board {
		o := t.Carried()
		if slices.ContainsFunc(inside, func(c task.Carried) bool { return c.Number == o.Number }) {
			continue
		}
		var why []Touch
		for _, is := range touching {
			for _, c := range inside {
				if touches(is, o, c) {
					why = append(why, Touch{Is: is, Of: c.Number})
				}
			}
		}
		if len(why) > 0 {
			out = append(out, Outside{Number: o.Number, Title: o.Title, State: string(o.State), Why: why})
		}
	}
	slices.SortFunc(out, func(a, b Outside) int { return a.Number - b.Number })
	return out, nil
}

func (e *Engine) viewsHolding(label string) (int, error) {
	views, err := e.store.Views()
	if err != nil {
		return 0, err
	}
	held := 0
	for _, v := range views {
		if slices.Contains(v.Labels, label) {
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

func viewsNaming(views []store.View, goners []int) []string {
	var naming []string
	for _, v := range views {
		if slices.ContainsFunc(v.Objectives, func(n int) bool { return slices.Contains(goners, n) }) {
			naming = append(naming, v.Name)
		}
	}
	return naming
}

func (e *Engine) unview(goners []int, how string) error {
	views, err := e.store.Views()
	if err != nil {
		return err
	}
	kept := []store.View{}
	for _, v := range views {
		if len(viewsNaming([]store.View{v}, goners)) == 0 {
			kept = append(kept, v)
			continue
		}
		if how == "delete" {
			continue
		}
		left := v
		left.Objectives = slices.DeleteFunc(slices.Clone(v.Objectives), func(n int) bool { return slices.Contains(goners, n) })
		if narrowsAnything(left) {
			v = left
		}
		kept = append(kept, v)
	}
	return e.store.KeepViews(kept)
}

func cmdKey(place int) string {
	if place <= 9 {
		return "⌘" + strconv.Itoa(place)
	}
	return "place " + strconv.Itoa(place)
}

func viewKeptText(v View, said string) string {
	if v.Was != "" {
		said += " (was " + v.Was + ")"
	}
	return fmt.Sprintf("%s · %s · keeps %d of %s\n", said, cmdKey(v.Place), v.Keeps, manyOf(v.Of, "task", "tasks"))
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
	for _, label := range v.Holds.Label {
		said = append(said, gone("label "+label))
	}
	for _, one := range []struct {
		named bool
		what  string
	}{
		{v.Holds.State != "", "state " + v.Holds.State},
		{v.Holds.Priority != "", "priority " + v.Holds.Priority},
		{v.Holds.DueBefore != "", "due before " + v.Holds.DueBefore},
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
		fmt.Sprintf("view     %s  (%s)", v.Name, cmdKey(v.Place)),
		"holds    " + HoldsSays(v),
		fmt.Sprintf("keeps    %d of %s · %d done", v.Keeps, manyOf(v.Of, "task", "tasks"), v.Done),
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
		lines = append(lines, fmt.Sprintf("objective #%d %s  %s", o.Number, o.Title, storeFigures(o)))
	}
	return strings.Join(lines, "\n") + "\n"
}

func storeFigures(o ViewObjective) string {
	if o.Progress == nil {
		return "nothing serves it yet"
	}
	var said []string
	if t := o.Progress.Tasks; t != nil {
		said = append(said, fmt.Sprintf("%d/%d", t.Done, t.Total))
	}
	if s := o.Progress.Steps; s != nil {
		said = append(said, fmt.Sprintf("%d/%d steps", s.Done, s.Total))
	}
	said = append(said, fmt.Sprintf("%d%%", o.Progress.Percent))
	if o.TimeSpent != "" {
		said = append(said, o.TimeSpent)
	}
	return strings.Join(said, " · ") + " (the store's)"
}

func typedName(name string) string {
	if strings.ContainsAny(name, " \t'\"") {
		return strconv.Quote(name)
	}
	return name
}

func lostText(a Answer) string {
	if a.InView == "" || len(a.Lost) == 0 {
		return ""
	}
	name := typedName(a.InView)
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

func outsideText(outside []Outside) string {
	if len(outside) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\noutside this view, never counted:\n")
	for _, o := range outside {
		said := []string{fmt.Sprintf("#%d  %s ", o.Number, o.Title), o.State}
		for _, w := range o.Why {
			said = append(said, fmt.Sprintf("%s #%d", w.Is, w.Of))
		}
		out.WriteString(strings.Join(said, " · ") + "\n")
	}
	return out.String()
}
