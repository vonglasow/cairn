package engine

import (
	"cairn/internal/engine/filter"
	"cairn/internal/engine/view"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func (e *Engine) tree() (objective.Tree, error) {
	all, err := e.objectives.AllObjectives()
	if err != nil {
		return objective.Tree{}, err
	}
	return objective.TreeOf(all), nil
}

func (e *Engine) countedOnce(number int, served []int, changed *objective.Objective) error {
	t, err := e.tree()
	if err != nil {
		return err
	}
	if changed != nil {
		t.By[changed.Carried().Number] = *changed
	}
	found := make([]objective.Objective, 0, len(served))
	for _, n := range served {
		if o, ok := t.By[n]; ok {
			found = append(found, o)
		}
	}
	return objective.CountedOnce(number, found)
}

func parentNamed(said string) (int, error) {
	trimmed := strings.TrimSpace(said)
	if trimmed == "" || trimmed == "none" {
		return 0, nil
	}
	n, err := strconv.Atoi(trimmed)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%q: %w, or none", said, errObjectiveRequired)
	}
	return n, nil
}

func (e *Engine) mayHang(o objective.Objective) error {
	c := o.Carried()
	if c.Parent == 0 {
		return nil
	}
	parent, err := e.objectives.FindObjective(c.Parent)
	if err != nil {
		return err
	}
	t, err := e.tree()
	if err != nil {
		return err
	}
	if err := objective.MayHang(o, parent, c.Number != 0 && len(t.Children[c.Number]) > 0); err != nil {
		return err
	}
	if c.Number == 0 {
		return nil
	}
	serving, err := e.store.TasksServing(c.Number)
	if err != nil {
		return err
	}
	for _, one := range serving {
		f := one.Carried()
		if err := e.countedOnce(f.Number, f.Objectives, &o); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) familyOf(n int) (objective.Family, error) {
	t, err := e.tree()
	if err != nil {
		return objective.Family{}, err
	}
	o, found := t.By[n]
	if !found {
		return objective.Family{}, fmt.Errorf("objective %d: %w", n, objective.ErrNoSuchObjective)
	}
	var f objective.Family
	if parent := o.Carried().Parent; parent != 0 {
		f.Parent = t.By[parent]
	}
	_, along, err := e.alongEvery(nil, false)
	if err != nil {
		return objective.Family{}, err
	}
	for _, c := range t.ChildrenOf(n) {
		under, err := e.tasksUnder(c.Carried().Number)
		if err != nil {
			return objective.Family{}, err
		}
		f.Children = append(f.Children, objective.Child{Objective: c, Progress: along[c.Carried().Number], Spent: reviewOf(under).Spent})
	}
	f.Proposed = objective.ProposedReached(o, t.ChildrenOf(n))
	for _, on := range o.Carried().WaitsOn {
		f.WaitsOn = append(f.WaitsOn, t.By[on])
	}
	for _, by := range t.Waiters()[n] {
		f.Awaited = append(f.Awaited, t.By[by])
	}
	return f, nil
}

func (e *Engine) proposedBy(was *objective.Objective, now objective.Objective) (*objective.Objective, error) {
	c := now.Carried()
	if c.Parent == 0 {
		return nil, nil
	}
	if c.State != objective.Reached {
		return nil, nil
	}
	if was != nil && was.Carried().State == objective.Reached {
		return nil, nil
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	parent := t.By[c.Parent]
	if !objective.ProposedReached(parent, t.ChildrenOf(c.Parent)) {
		return nil, nil
	}
	return &parent, nil
}

func (e *Engine) spentEvery() (map[int]time.Duration, error) {
	every := filter.Narrowing{}
	every.WithArchived()
	all, err := spentTasks{Store: e.tasks, history: e.history}.told(e.tasks.Narrowed(&every))
	if err != nil {
		return nil, err
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	spent := map[int]time.Duration{}
	for _, one := range all {
		f := one.Carried()
		for _, o := range t.Counted(f.Objectives) {
			spent[o] += f.Spent
		}
	}
	return spent, nil
}

func (e *Engine) seating() (children, waiters map[int][]int, err error) {
	t, err := e.tree()
	if err != nil {
		return nil, nil, err
	}
	return t.Children, t.Waiters(), nil
}

var deletionChoices = []Argument{
	{Name: "views", Kind: Text, OneOf: []string{"off", "delete"}, Means: "what becomes of the views naming what is deleted: the objective taken off them, a view holding nothing else keeping it gone, or the views deleted"},
	{Name: "sub-objectives", Kind: Text, OneOf: []string{"keep", "delete"}, Means: "what becomes of its sub-objectives: kept, each standing on its own, or deleted"},
	{Name: "keep", Kind: List, Of: "an objective", Means: "a sub-objective kept, standing on its own, whatever sub-objectives says"},
	{Name: "delete", Kind: List, Of: "an objective", Means: "a sub-objective deleted, whatever sub-objectives says"},
	{Name: "tasks", Kind: Text, OneOf: []string{"detach", "delete"}, Means: "what becomes of the tasks attached to what is deleted: detached from it, or deleted"},
}

func deletionOf(in Input) (objective.Deletion, error) {
	for _, arg := range deletionChoices {
		if err := accepted(in, arg); err != nil {
			return objective.Deletion{}, err
		}
	}
	d := objective.Deletion{Subs: in.Given["sub-objectives"], Tasks: in.Given["tasks"], Views: in.Given["views"]}
	for name, into := range map[string]*[]int{"keep": &d.Keep, "delete": &d.Delete} {
		for _, raw := range in.Listed[name] {
			n, err := strconv.Atoi(strings.TrimSpace(raw))
			if err != nil {
				return objective.Deletion{}, fmt.Errorf("%q: %w", raw, errObjectiveRequired)
			}
			*into = append(*into, n)
		}
	}
	return d, nil
}

func (e *Engine) deleteObjectives(numbers []int, d objective.Deletion) (objective.WentWith, error) {
	t, err := e.tree()
	if err != nil {
		return objective.WentWith{}, err
	}
	for _, n := range numbers {
		if _, found := t.By[n]; !found {
			return objective.WentWith{}, fmt.Errorf("objective %d: %w", n, objective.ErrNoSuchObjective)
		}
	}
	if err := d.NamesOnly(numbers, t); err != nil {
		return objective.WentWith{}, err
	}
	goners := slices.Clone(numbers)
	var went objective.WentWith
	var undecided []objective.Objective
	for _, n := range numbers {
		for _, c := range t.Children[n] {
			if slices.Contains(numbers, c) {
				continue
			}
			if slices.Contains(d.Keep, c) || !slices.Contains(d.Delete, c) && d.Subs == "keep" {
				went.Kept = append(went.Kept, c)
			} else if slices.Contains(d.Delete, c) || d.Subs == "delete" {
				goners = append(goners, c)
				went.Gone = append(went.Gone, c)
			} else {
				undecided = append(undecided, t.By[c])
			}
		}
	}
	touched, err := e.store.TasksServing(goners...)
	if err != nil {
		return objective.WentWith{}, err
	}
	views, err := e.store.Views()
	if err != nil {
		return objective.WentWith{}, err
	}
	naming := view.Naming(views, goners)
	if len(undecided) > 0 || len(touched) > 0 && d.Tasks == "" || len(naming) > 0 && d.Views == "" {
		return objective.WentWith{}, e.asking(numbers, undecided, touched, naming, d)
	}
	if err := e.carryOut(goners, went.Kept, touched, d.Tasks, &went); err != nil {
		return objective.WentWith{}, err
	}
	if len(naming) == 0 {
		return went, nil
	}
	return went, e.unview(goners, d.Views)
}

func (e *Engine) carryOut(goners, kept []int, touched []task.Task, tasks string, went *objective.WentWith) error {
	for _, c := range kept {
		o, err := e.objectives.FindObjective(c)
		if err != nil {
			return err
		}
		loose := o.Carried()
		loose.Parent = 0
		if err := e.objectives.UpdateObjective(objective.Made(loose)); err != nil {
			return err
		}
	}
	for _, one := range touched {
		f := one.Carried()
		if tasks == "delete" {
			if err := e.delete(f.Number); err != nil {
				return err
			}
			went.Deleted = append(went.Deleted, f.Number)
			continue
		}
		var off []string
		for _, g := range f.Objectives {
			if slices.Contains(goners, g) {
				off = append(off, "-"+strconv.Itoa(g))
			}
		}
		if _, err := e.edit(f.Number, Changes{Objective: &off}); err != nil {
			return err
		}
		went.Detached = append(went.Detached, f.Number)
	}
	for _, g := range goners {
		if err := e.objectives.DeleteObjective(g); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) asking(numbers []int, undecided []objective.Objective, touched []task.Task, naming []string, d objective.Deletion) error {
	a := &objective.AskingWhatGoes{Holds: len(undecided) > 0, Carries: len(touched) > 0, Named: len(naming) > 0 && d.Views == ""}
	var lines []string
	for _, c := range undecided {
		cn := c.Carried()
		lines = append(lines, fmt.Sprintf("  sub-objective #%d  %s", cn.Number, cn.Title))
		theirs, err := e.store.TasksServing(cn.Number)
		if err != nil {
			return err
		}
		for _, one := range theirs {
			f := one.Carried()
			lines = append(lines, fmt.Sprintf("    task #%d  %s", f.Number, f.Title))
		}
		a.Carries = a.Carries || len(theirs) > 0
	}
	for _, one := range touched {
		f := one.Carried()
		lines = append(lines, fmt.Sprintf("  task #%d  %s", f.Number, f.Title))
	}
	if !a.Holds && d.Tasks != "" {
		a.Carries = false
	}
	what := objective.ErrObjectiveHoldsSubObjectives.Error()
	if a.Carries {
		what = objective.ErrObjectiveCarriesTasks.Error()
	}
	var flags []string
	if a.Holds {
		flags = append(flags, "--sub-objectives keep|delete, or one by one --keep N --delete N")
		if a.Carries {
			what += " and holds sub-objectives"
		}
	}
	if a.Carries && d.Tasks == "" {
		flags = append(flags, "--tasks detach|delete")
	}
	if a.Named {
		for _, name := range naming {
			lines = append(lines, "  view "+name)
		}
		what += " and " + objective.ErrObjectiveNamedByAView.Error()
		if !a.Holds && !a.Carries {
			what = objective.ErrObjectiveNamedByAView.Error()
		}
		flags = append(flags, "--views off|delete")
	}
	a.Said = fmt.Sprintf("objective %d: %s, and deleting it asks what goes with them:\n%s\nsay it with %s",
		numbers[0], what, strings.Join(lines, "\n"), strings.Join(flags, "; and "))
	return a
}
