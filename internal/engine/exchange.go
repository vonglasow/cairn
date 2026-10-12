package engine

// Tasks carried from one store to another as a file of their own. Export
// writes every task, archived ones included, with the objectives they serve,
// in the shapes a machine already reads them in; import adds what such a file
// holds as new tasks, numbered by the store they land in, and replaces and
// merges nothing. What was done to a task — its trace, its time spent — stays
// with the store it was done in. See bdd/exchange.feature.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// exchangeKind and exchangeVersion are what a file of tasks says it is: a
// file saying anything else is not one, and one of a later version is not
// one this cairn can read.
const (
	exchangeKind    = "tasks"
	exchangeVersion = 1
)

// ErrNotAnExport is what reading a file that does not hold tasks written by
// task export is refused with.
var ErrNotAnExport = errors.New("the file holds no tasks written by cairn task export")

// ErrExportPointsOutside is what reading a file whose tasks point at a task or
// an objective it does not hold is refused with.
var ErrExportPointsOutside = errors.New("the file points outside itself")

// exchanged is a file of tasks.
type exchanged struct {
	Cairn      string            `json:"cairn"`
	Version    int               `json:"version"`
	Objectives []objective.Shape `json:"objectives"`
	Tasks      []task.Shape      `json:"tasks"`
}

// Exported says which file the tasks were written to, and what it holds.
type Exported struct {
	Exported string       `json:"exported"`
	Holds    *domain.Held `json:"holds"`
}

// Imported says which file was read in, and how much of it was added.
type Imported struct {
	Imported string      `json:"imported"`
	Added    domain.Held `json:"added"`
}

// TaskExport writes every task to a file, with the objectives they serve.
var TaskExport = Operation{
	Thing:   "task",
	Verb:    "export",
	Short:   "Write every task to a file, with the objectives they serve",
	Does:    "Write every task, archived ones included, to a JSON file of your choosing, with the objectives they serve, for task import to read into another store. A file already there is never written over.",
	Example: "cairn task export ~/Desktop/tasks.json",
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Exported{},
	Takes: []Argument{
		{Name: "to", Kind: Text, Positional: true, Required: true, Means: "the file to write the tasks to"},
	},
	Run: func(on On, in Input) (Answer, error) {
		to := in.Given["to"]
		out, err := on.Engine.exporting()
		if err != nil {
			return Answer{}, err
		}
		if err := writeNew(to, out); err != nil {
			return Answer{}, err
		}
		held := domain.Held{Tasks: len(out.Tasks), Objectives: len(out.Objectives)}
		return Answer{Exported: &to, Holds: &held}, nil
	},
}

// TaskImport adds the tasks of a file written by TaskExport, as new tasks.
var TaskImport = Operation{
	Thing:   "task",
	Verb:    "import",
	Short:   "Add the tasks of a file written by task export",
	Does:    "Add the tasks of a file written by task export, as new tasks numbered after the ones already here, with the objectives they serve as new objectives. Nothing already here is replaced or merged, so a file read twice is added twice. A file that is not one, or whose tasks point outside it, is refused before anything is added.",
	Example: "cairn task import ~/Downloads/tasks.json",
	Level:   Write,
	Needs:   AStore,
	Again:   DoesItAgain,
	Answers: Imported{},
	Takes: []Argument{
		{Name: "from", Kind: Text, Positional: true, Required: true, Means: "the file holding the tasks to add"},
	},
	Run: func(on On, in Input) (Answer, error) {
		from := in.Given["from"]
		read, err := readExport(from)
		if err != nil {
			return Answer{}, err
		}
		added, err := on.Engine.importing(read)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Imported: &from, Holds: &added}, nil
	},
}

// exporting is every task, archived ones included, and the objectives any of
// them serves, in the order they were made.
func (e *Engine) exporting() (exchanged, error) {
	var every domain.Narrowing
	every.WithArchived()
	tasks, err := e.narrowed(&every)
	if err != nil {
		return exchanged{}, err
	}
	out := exchanged{Cairn: exchangeKind, Version: exchangeVersion, Tasks: []task.Shape{}, Objectives: []objective.Shape{}}
	served := map[int]bool{}
	for _, t := range tasks {
		shape := task.OfTask(t)
		out.Tasks = append(out.Tasks, shape)
		for _, o := range shape.Objectives {
			served[o] = true
		}
	}
	slices.SortFunc(out.Tasks, func(a, b task.Shape) int { return a.Number - b.Number })
	objectives, err := e.listObjectives()
	if err != nil {
		return exchanged{}, err
	}
	for _, o := range objectives {
		if shape := objective.OfObjective(o); served[shape.Number] {
			out.Objectives = append(out.Objectives, shape)
		}
	}
	return out, nil
}

// writeNew writes what was exported to a file nobody wrote before, making the
// folder it goes in when there is none.
func writeNew(to string, out exchanged) error {
	written, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o750); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Clean(to), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(written, '\n')); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// readExport reads a file written by TaskExport, and refuses one that is not.
func readExport(from string) (exchanged, error) {
	held, err := os.ReadFile(filepath.Clean(from))
	if err != nil {
		return exchanged{}, err
	}
	var read exchanged
	if err := json.Unmarshal(held, &read); err != nil || read.Cairn != exchangeKind {
		return exchanged{}, ErrNotAnExport
	}
	if read.Version != exchangeVersion {
		return exchanged{}, fmt.Errorf("%w: it is of version %d, and this cairn reads version %d", ErrNotAnExport, read.Version, exchangeVersion)
	}
	return read, nil
}

// importing adds what a file holds: checked whole first, so a file refused
// adds nothing; then the objectives, then the tasks — each before its parts —
// then the column each stands in and what is in its way, parts first, and
// last what was archived.
func (e *Engine) importing(read exchanged) (domain.Held, error) {
	order, err := checked(read)
	if err != nil {
		return domain.Held{}, err
	}
	objectives := map[int]int{}
	for _, o := range read.Objectives {
		made, err := e.addObjective(o.Title, objectiveChangesOf(o))
		if err != nil {
			return domain.Held{}, err
		}
		objectives[o.Number] = numberOf(made)
	}
	tasks := map[int]int{}
	for _, t := range order {
		c := fromFile(t)
		if t.Parent != 0 {
			parent := tasks[t.Parent]
			c.Parent = &parent
		}
		if len(t.Objectives) > 0 {
			serves := numbered(t.Objectives, objectives)
			c.Objective = &serves
		}
		made, err := e.add(t.Title, c)
		if err != nil {
			return domain.Held{}, err
		}
		tasks[t.Number] = task.OfTask(made).Number
	}
	for _, t := range slices.Backward(order) {
		if err := e.standing(t, tasks); err != nil {
			return domain.Held{}, err
		}
	}
	for _, t := range order {
		if t.Archived && t.Parent == 0 {
			if err := e.archive(tasks[t.Number]); err != nil {
				return domain.Held{}, err
			}
		}
	}
	return domain.Held{Tasks: len(order), Objectives: len(read.Objectives)}, nil
}

// standing puts a task read in where it stood: its column, and what was in
// its way, pointing at the new numbers. Nothing is asked when both are as
// the task already has them.
func (e *Engine) standing(t task.Shape, tasks map[int]int) error {
	number := tasks[t.Number]
	now, err := e.show(number)
	if err != nil {
		return err
	}
	var c Changes
	if task.OfTask(now).State != t.State {
		state := t.State
		c.State = &state
	}
	if len(t.BlockedBy) > 0 {
		blockers := numbered(t.BlockedBy, tasks)
		c.Blocker = &blockers
	}
	if !c.named() {
		return nil
	}
	_, err = e.edit(number, c)
	return err
}

// checked refuses a file that would be added only in part — a title, a
// column or an attribute this store refuses, a task or an objective it points
// at and does not hold, a task hanging under itself — and hands back its
// tasks in an order where each comes after the task it hangs under.
func checked(read exchanged) ([]task.Shape, error) {
	objectives := map[int]bool{}
	for _, o := range read.Objectives {
		if objectives[o.Number] {
			return nil, fmt.Errorf("%w: objective %d is in the file twice", ErrNotAnExport, o.Number)
		}
		objectives[o.Number] = true
		made, err := objective.NewObjective(o.Title)
		if err == nil {
			_, err = applyObjective(made, objectiveChangesOf(o), creating)
		}
		if err != nil {
			return nil, fmt.Errorf("objective %d: %w", o.Number, err)
		}
	}
	byNumber := map[int]task.Shape{}
	for _, t := range read.Tasks {
		if _, twice := byNumber[t.Number]; twice {
			return nil, fmt.Errorf("%w: task %d is in the file twice", ErrNotAnExport, t.Number)
		}
		byNumber[t.Number] = t
	}
	for _, t := range read.Tasks {
		if err := checkedTask(t, byNumber, objectives); err != nil {
			return nil, err
		}
	}
	return parentsFirst(read.Tasks, byNumber)
}

// checkedTask refuses one task of a file, by the number the file gave it.
func checkedTask(t task.Shape, byNumber map[int]task.Shape, objectives map[int]bool) error {
	outside := func(what string, n int) error {
		return fmt.Errorf("%w: task %d names %s %d — %s %d is not in the file", ErrExportPointsOutside, t.Number, what, n, what, n)
	}
	if _, in := byNumber[t.Parent]; t.Parent != 0 && !in {
		return outside("task", t.Parent)
	}
	for _, b := range t.BlockedBy {
		if _, in := byNumber[b]; !in {
			return outside("task", b)
		}
	}
	for _, o := range t.Objectives {
		if !objectives[o] {
			return outside("objective", o)
		}
	}
	if _, err := domain.ParseState(t.State); err != nil {
		return fmt.Errorf("task %d: %w", t.Number, err)
	}
	made, err := task.NewTask(t.Title)
	if err == nil {
		_, err = apply(made, fromFile(t), creating)
	}
	if err != nil {
		return fmt.Errorf("task %d: %w", t.Number, err)
	}
	return nil
}

// parentsFirst orders the tasks of a file so each comes after the task it
// hangs under, keeping the file's order otherwise; a task that hangs under
// itself, however far round, is refused.
func parentsFirst(tasks []task.Shape, byNumber map[int]task.Shape) ([]task.Shape, error) {
	placed := map[int]bool{}
	var order []task.Shape
	var place func(t task.Shape, seen map[int]bool) error
	place = func(t task.Shape, seen map[int]bool) error {
		if placed[t.Number] {
			return nil
		}
		if seen[t.Number] {
			return fmt.Errorf("%w: task %d hangs under itself", ErrNotAnExport, t.Number)
		}
		seen[t.Number] = true
		if t.Parent != 0 {
			if err := place(byNumber[t.Parent], seen); err != nil {
				return err
			}
		}
		placed[t.Number] = true
		order = append(order, t)
		return nil
	}
	for _, t := range tasks {
		if err := place(t, map[int]bool{}); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// fromFile is what a task of a file carries that points at nothing else: its
// column and what is in its way are put after every task is in.
func fromFile(t task.Shape) Changes {
	c := Changes{Description: &t.Description, Priority: &t.Priority, Because: &t.Because}
	if len(t.Labels) > 0 {
		labels := slices.Clone(t.Labels)
		c.Labels = &labels
	}
	if t.Deadline != "" {
		c.Deadline = &t.Deadline
	}
	if t.Estimate != "" {
		c.Estimate = &t.Estimate
	}
	return c
}

// objectiveChangesOf is what an objective of a file carries.
func objectiveChangesOf(o objective.Shape) ObjectiveChanges {
	c := ObjectiveChanges{Description: &o.Description, State: &o.State, Priority: &o.Priority}
	if len(o.Labels) > 0 {
		labels := slices.Clone(o.Labels)
		c.Labels = &labels
	}
	if o.Deadline != "" {
		c.Deadline = &o.Deadline
	}
	if o.Estimate != "" {
		c.Estimate = &o.Estimate
	}
	return c
}

// numbered is the new numbers of what a file numbered, as the engine reads
// numbers it is given.
func numbered(old []int, now map[int]int) []string {
	out := make([]string, 0, len(old))
	for _, n := range old {
		out = append(out, strconv.Itoa(now[n]))
	}
	return out
}

// numberOf is an objective's number.
func numberOf(o objective.Objective) int { return objective.OfObjective(o).Number }
