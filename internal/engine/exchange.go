package engine

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

const (
	exchangeKind    = "tasks"
	exchangeVersion = 2
	beforeSteps     = 1
)

// ErrNotAnExport is what reading a file that does not hold tasks written by task export is refused with.
var ErrNotAnExport = errors.New("the file holds no tasks written by cairn task export")

// ErrExportPointsOutside is what reading a file whose tasks point at a task or an objective it does not hold is refused with.
var ErrExportPointsOutside = errors.New("the file points outside itself")

type exchanged struct {
	Cairn      string            `json:"cairn"`
	Version    int               `json:"version"`
	Objectives []objective.Shape `json:"objectives"`
	Tasks      []filed           `json:"tasks"`
}

type filed struct {
	task.Shape
	Parent int `json:"parent,omitempty"`
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

func (e *Engine) exporting() (exchanged, error) {
	var every domain.Narrowing
	every.WithArchived()
	tasks, err := e.narrowed(&every)
	if err != nil {
		return exchanged{}, err
	}
	out := exchanged{Cairn: exchangeKind, Version: exchangeVersion, Tasks: []filed{}, Objectives: []objective.Shape{}}
	served := map[int]bool{}
	for _, t := range tasks {
		shape := task.OfTask(t)
		out.Tasks = append(out.Tasks, filed{Shape: shape})
		for _, o := range shape.Objectives {
			served[o] = true
		}
	}
	slices.SortFunc(out.Tasks, func(a, b filed) int { return a.Number - b.Number })
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

func readExport(from string) (exchanged, error) {
	held, err := os.ReadFile(filepath.Clean(from))
	if err != nil {
		return exchanged{}, err
	}
	var read exchanged
	if err := json.Unmarshal(held, &read); err != nil || read.Cairn != exchangeKind {
		return exchanged{}, ErrNotAnExport
	}
	if read.Version != exchangeVersion && read.Version != beforeSteps {
		return exchanged{}, fmt.Errorf("%w: it is of version %d, and this cairn reads versions %d and %d", ErrNotAnExport, read.Version, beforeSteps, exchangeVersion)
	}
	return read, nil
}

func stepsOfSubTasks(read exchanged) (exchanged, error) {
	byNumber := map[int]bool{}
	for _, t := range read.Tasks {
		byNumber[t.Number] = true
	}
	var left, done = map[int][]task.Step{}, map[int][]task.Step{}
	var work []filed
	for _, t := range read.Tasks {
		if t.Parent == 0 {
			work = append(work, t)
			continue
		}
		if !byNumber[t.Parent] {
			return exchanged{}, fmt.Errorf("%w: task %d names task %d — task %d is not in the file", ErrExportPointsOutside, t.Number, t.Parent, t.Parent)
		}
		step := task.Step{Title: t.Title, Done: t.State == string(domain.Done)}
		if step.Done {
			done[t.Parent] = append(done[t.Parent], step)
		} else {
			left[t.Parent] = append(left[t.Parent], step)
		}
	}
	for i, t := range work {
		work[i].Steps = append(append(slices.Clone(t.Steps), left[t.Number]...), done[t.Number]...)
	}
	read.Tasks = work
	return read, nil
}

func (e *Engine) importing(read exchanged) (domain.Held, error) {
	read, err := stepsOfSubTasks(read)
	if err != nil {
		return domain.Held{}, err
	}
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
		c := fromFile(t.Shape)
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
		if t.Archived {
			if err := e.archive(tasks[t.Number]); err != nil {
				return domain.Held{}, err
			}
		}
	}
	return domain.Held{Tasks: len(order), Objectives: len(read.Objectives)}, nil
}

func (e *Engine) standing(t filed, tasks map[int]int) error {
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

func checked(read exchanged) ([]filed, error) {
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
	byNumber := map[int]filed{}
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
	return read.Tasks, nil
}

func checkedTask(t filed, byNumber map[int]filed, objectives map[int]bool) error {
	outside := func(what string, n int) error {
		return fmt.Errorf("%w: task %d names %s %d — %s %d is not in the file", ErrExportPointsOutside, t.Number, what, n, what, n)
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
		_, err = apply(made, fromFile(t.Shape), creating)
	}
	for _, s := range t.Steps {
		if err == nil {
			_, err = task.NewStep(s.Title, s.Done)
		}
	}
	if err != nil {
		return fmt.Errorf("task %d: %w", t.Number, err)
	}
	return nil
}

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
	if len(t.Steps) > 0 {
		steps := slices.Clone(t.Steps)
		c.Steps = &steps
	}
	return c
}

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

func numbered(old []int, now map[int]int) []string {
	out := make([]string, 0, len(old))
	for _, n := range old {
		out = append(out, strconv.Itoa(now[n]))
	}
	return out
}

func numberOf(o objective.Objective) int { return objective.OfObjective(o).Number }
