package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// ErrObjectiveHoldsSubObjectives is why an objective holding sub-objectives is not deleted until it is said what becomes of them.
var ErrObjectiveHoldsSubObjectives = objective.ErrObjectiveHoldsSubObjectives

// ErrNotASubObjective is returned when a deletion names, to keep or to delete, an objective that is not a sub-objective of what is deleted.
var ErrNotASubObjective = errors.New("only a sub-objective of what is deleted is kept or deleted on its own")

type tree struct {
	by       map[int]objective.Objective
	children map[int][]int
}

func (e *Engine) tree() (tree, error) {
	all, err := e.objectives.AllObjectives()
	if err != nil {
		return tree{}, err
	}
	t := tree{by: map[int]objective.Objective{}, children: map[int][]int{}}
	for _, o := range all {
		c := o.Carried()
		t.by[c.Number] = o
		if c.Parent != 0 {
			t.children[c.Parent] = append(t.children[c.Parent], c.Number)
		}
	}
	return t, nil
}

func (t tree) reach(n int) []int { return append([]int{n}, t.children[n]...) }

func (t tree) counted(served []int) []int {
	out := slices.Clone(served)
	for _, n := range served {
		if parent := t.by[n].Carried().Parent; parent != 0 && !slices.Contains(out, parent) {
			out = append(out, parent)
		}
	}
	return out
}

func (t tree) childrenOf(n int) []objective.Objective {
	out := make([]objective.Objective, 0, len(t.children[n]))
	for _, c := range t.children[n] {
		out = append(out, t.by[c])
	}
	return out
}

func (e *Engine) countedOnce(number int, served []int, changed *objective.Objective) error {
	t, err := e.tree()
	if err != nil {
		return err
	}
	if changed != nil {
		t.by[changed.Carried().Number] = *changed
	}
	found := make([]objective.Objective, 0, len(served))
	for _, n := range served {
		if o, ok := t.by[n]; ok {
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
		return 0, fmt.Errorf("%q: %w, or none", said, ErrObjectiveRequired)
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
	if err := objective.MayHang(o, parent, c.Number != 0 && len(t.children[c.Number]) > 0); err != nil {
		return err
	}
	if c.Number == 0 {
		return nil
	}
	serving, err := e.objectives.TasksServing(c.Number)
	if err != nil {
		return err
	}
	for _, one := range serving {
		f := factsOf(one)
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
	o, found := t.by[n]
	if !found {
		return objective.Family{}, fmt.Errorf("objective %d: %w", n, ErrNoSuchObjective)
	}
	var f objective.Family
	if parent := o.Carried().Parent; parent != 0 {
		f.Parent = t.by[parent]
	}
	_, along, err := e.alongEvery(nil, false)
	if err != nil {
		return objective.Family{}, err
	}
	for _, c := range t.childrenOf(n) {
		under, err := e.tasksUnder(c.Carried().Number)
		if err != nil {
			return objective.Family{}, err
		}
		f.Children = append(f.Children, objective.Child{Objective: c, Progress: along[c.Carried().Number], Spent: reviewOf(under).Spent})
	}
	f.Proposed = objective.ProposedReached(o, t.childrenOf(n))
	return f, nil
}

func (e *Engine) proposedBy(was *objective.Objective, now objective.Objective) (*objective.Objective, error) {
	c := now.Carried()
	if c.Parent == 0 || c.State != objective.Reached || was != nil && was.Carried().State == objective.Reached {
		return nil, nil
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	parent := t.by[c.Parent]
	if !objective.ProposedReached(parent, t.childrenOf(c.Parent)) {
		return nil, nil
	}
	return &parent, nil
}

func (e *Engine) spentEvery() (map[int]time.Duration, error) {
	every := domain.Narrowing{}
	every.WithArchived()
	all, err := spentTasks{tasks: e.tasks, history: e.history}.told(e.tasks.Narrowed(&every))
	if err != nil {
		return nil, err
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	spent := map[int]time.Duration{}
	for _, one := range all {
		f := factsOf(one)
		for _, o := range t.counted(f.Objectives) {
			spent[o] += f.Spent
		}
	}
	return spent, nil
}

func (e *Engine) treeMap() (map[int][]int, error) {
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	return t.children, nil
}

// Went is what a deletion took with it: the sub-objectives kept, each standing on its own, those deleted, and the tasks detached or deleted.
type Went struct {
	Kept     []int
	Gone     []int
	Detached []int
	Deleted  []int
}

var deletionChoices = []Argument{
	{Name: "sub-objectives", Kind: Text, OneOf: []string{"keep", "delete"}, Means: "what becomes of its sub-objectives: kept, each standing on its own, or deleted"},
	{Name: "keep", Kind: List, Of: "an objective", Means: "a sub-objective kept, standing on its own, whatever sub-objectives says"},
	{Name: "delete", Kind: List, Of: "an objective", Means: "a sub-objective deleted, whatever sub-objectives says"},
	{Name: "tasks", Kind: Text, OneOf: []string{"detach", "delete"}, Means: "what becomes of the tasks attached to what is deleted: detached from it, or deleted"},
}

type deletion struct {
	subs, tasks  string
	keep, delete []int
}

func deletionOf(in Input) (deletion, error) {
	for _, arg := range deletionChoices {
		if err := accepted(in, arg); err != nil {
			return deletion{}, err
		}
	}
	d := deletion{subs: in.Given["sub-objectives"], tasks: in.Given["tasks"]}
	for name, into := range map[string]*[]int{"keep": &d.keep, "delete": &d.delete} {
		for _, raw := range in.Listed[name] {
			n, err := strconv.Atoi(strings.TrimSpace(raw))
			if err != nil {
				return deletion{}, fmt.Errorf("%q: %w", raw, ErrObjectiveRequired)
			}
			*into = append(*into, n)
		}
	}
	return d, nil
}

func (e *Engine) deleteObjectives(numbers []int, d deletion) (Went, error) {
	t, err := e.tree()
	if err != nil {
		return Went{}, err
	}
	for _, n := range numbers {
		if _, found := t.by[n]; !found {
			return Went{}, fmt.Errorf("objective %d: %w", n, ErrNoSuchObjective)
		}
	}
	if err := d.namesOnly(numbers, t); err != nil {
		return Went{}, err
	}
	goners := slices.Clone(numbers)
	var went Went
	var undecided []objective.Objective
	for _, n := range numbers {
		for _, c := range t.children[n] {
			switch {
			case slices.Contains(numbers, c):
			case slices.Contains(d.keep, c) || !slices.Contains(d.delete, c) && d.subs == "keep":
				went.Kept = append(went.Kept, c)
			case slices.Contains(d.delete, c) || d.subs == "delete":
				goners = append(goners, c)
				went.Gone = append(went.Gone, c)
			default:
				undecided = append(undecided, t.by[c])
			}
		}
	}
	touched, err := e.objectives.TasksServing(goners...)
	if err != nil {
		return Went{}, err
	}
	if len(undecided) > 0 || len(touched) > 0 && d.tasks == "" {
		return Went{}, e.asking(numbers, undecided, touched, d)
	}
	return went, e.carryOut(goners, went.Kept, touched, d.tasks, &went)
}

func (d deletion) namesOnly(numbers []int, t tree) error {
	held := func(c int) bool {
		parent := t.by[c].Carried().Parent
		return parent != 0 && slices.Contains(numbers, parent)
	}
	for _, c := range append(slices.Clone(d.keep), d.delete...) {
		if !held(c) {
			return fmt.Errorf("#%d is not a sub-objective of %s; %w", c, domain.Ref(numbers[0]), ErrNotASubObjective)
		}
		if slices.Contains(d.keep, c) && slices.Contains(d.delete, c) {
			return fmt.Errorf("#%d is both kept and deleted; %w", c, ErrNotASubObjective)
		}
	}
	return nil
}

func (e *Engine) carryOut(goners, kept []int, touched []task.Task, tasks string, went *Went) error {
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
		f := factsOf(one)
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

type askingWhatGoes struct {
	said           string
	carries, holds bool
}

func (a *askingWhatGoes) Error() string { return a.said }

func (a *askingWhatGoes) Is(target error) bool {
	return target == ErrObjectiveCarriesTasks && a.carries || target == ErrObjectiveHoldsSubObjectives && a.holds
}

func (e *Engine) asking(numbers []int, undecided []objective.Objective, touched []task.Task, d deletion) error {
	a := &askingWhatGoes{holds: len(undecided) > 0, carries: len(touched) > 0}
	var lines []string
	for _, c := range undecided {
		cn := c.Carried()
		lines = append(lines, fmt.Sprintf("  sub-objective #%d  %s", cn.Number, cn.Title))
		theirs, err := e.objectives.TasksServing(cn.Number)
		if err != nil {
			return err
		}
		for _, one := range theirs {
			f := factsOf(one)
			lines = append(lines, fmt.Sprintf("    task #%d  %s", f.Number, f.Title))
		}
		a.carries = a.carries || len(theirs) > 0
	}
	for _, one := range touched {
		f := factsOf(one)
		lines = append(lines, fmt.Sprintf("  task #%d  %s", f.Number, f.Title))
	}
	what := ErrObjectiveHoldsSubObjectives.Error()
	if a.carries {
		what = ErrObjectiveCarriesTasks.Error()
	}
	var flags []string
	if a.holds {
		flags = append(flags, "--sub-objectives keep|delete, or one by one --keep N --delete N")
		if a.carries {
			what += " and holds sub-objectives"
		}
	}
	if a.carries && d.tasks == "" {
		flags = append(flags, "--tasks detach|delete")
	}
	a.said = fmt.Sprintf("objective %d: %s, and deleting it asks what goes with them:\n%s\nsay it with %s",
		numbers[0], what, strings.Join(lines, "\n"), strings.Join(flags, "; and "))
	return a
}

func wentText(w *Went) string {
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
