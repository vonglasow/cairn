package engine

import (
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/exchange"
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

type exchanged struct {
	Cairn      string                    `json:"cairn"`
	Version    int                       `json:"version"`
	Objectives []exchange.ObjectiveFiled `json:"objectives"`
	Labels     []exchange.LabelFiled     `json:"labels,omitempty"`
	Tasks      []exchange.Filed          `json:"tasks"`
	Series     []exchange.SeriesFiled    `json:"series,omitempty"`
	Views      []viewFiled               `json:"views,omitempty"`
}

// TaskExport writes every task to a file, with every objective.
var TaskExport = Operation{
	Thing:   "task",
	Verb:    "export",
	Short:   "Write every task to a file, with every objective",
	Does:    "Write every task, archived ones included, to a JSON file of your choosing, with every objective, served by a task or not, and the views, for task import to read into another store. A file already there is never written over.",
	Example: "cairn task export ~/Desktop/tasks.json",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: exchange.Exported{},
	Takes: []Argument{
		{Name: "to", Kind: Text, Positional: true, Required: true, Means: "the file to write the tasks to"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		to := in.Given["to"]
		out, err := on.Engine.exporting()
		if err != nil {
			return Answer{}, err
		}
		if err := writeNew(to, out); err != nil {
			return Answer{}, err
		}
		held := backup.Held{Tasks: len(out.Tasks), Objectives: len(out.Objectives)}
		return Answer{Exported: &to, Holds: &held}, nil
	},
}

// TaskImport adds the tasks of a file written by TaskExport, as new tasks.
var TaskImport = Operation{
	Thing:   "task",
	Verb:    "import",
	Short:   "Add the tasks of a file written by task export",
	Does:    "Add the tasks of a file written by task export, as new tasks numbered after the ones already here, with the objectives they serve as new objectives and the views after the ones here. Nothing already here is replaced or merged, so a file read twice is added twice, and a view whose name one here has is left out. A file that is not one, that holds nothing to add, or whose tasks point outside it, is refused before anything is added.",
	Example: "cairn task import ~/Downloads/tasks.json",
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: exchange.ImportedFile{},
	Takes: []Argument{
		{Name: "from", Kind: Text, Positional: true, Required: true, Means: "the file holding the tasks to add"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
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
	var every filter.Narrowing
	every.WithArchived()
	tasks, err := e.narrowed(&every)
	if err != nil {
		return exchanged{}, err
	}
	out := exchanged{Cairn: exchange.Kind, Version: exchange.Version, Tasks: []exchange.Filed{}, Objectives: []exchange.ObjectiveFiled{}}
	spent, err := e.spentByDay()
	if err != nil {
		return exchanged{}, err
	}
	held, err := e.focus()
	if err != nil {
		return exchanged{}, err
	}
	labels, err := e.labelsCarried()
	if err != nil {
		return exchanged{}, err
	}
	for _, l := range labels {
		filed := exchange.LabelFiled{Label: l.Label, Colour: l.Colour}
		if l.Scope != string(label.ForBoth) {
			filed.Scope = l.Scope
		}
		out.Labels = append(out.Labels, filed)
	}
	for _, t := range tasks {
		shape := task.OfTask(t)
		out.Tasks = append(out.Tasks, exchange.Filed{Shape: shape, Spent: spent[shape.Number], Focused: slices.Contains(held, shape.Number)})
	}
	slices.SortFunc(out.Tasks, func(a, b exchange.Filed) int { return a.Number - b.Number })
	if out.Series, err = e.seriesWritten(); err != nil {
		return exchanged{}, err
	}
	t, err := e.tree()
	if err != nil {
		return exchanged{}, err
	}
	if out.Views, err = e.viewsWritten(t); err != nil {
		return exchanged{}, err
	}
	for _, n := range slices.Sorted(maps.Keys(t.By)) {
		c := t.By[n].Carried()
		out.Objectives = append(out.Objectives, exchange.ObjectiveFiled{Shape: objective.OfObjective(t.By[n]), Parent: c.Parent, WaitsOn: c.WaitsOn})
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
	_, err = file.Write(append(written, '\n'))
	return cmp.Or(err, file.Close())
}

func readExport(from string) (exchanged, error) {
	held, err := os.ReadFile(filepath.Clean(from))
	if err != nil {
		return exchanged{}, err
	}
	var read exchanged
	if err := json.Unmarshal(held, &read); err != nil || read.Cairn != exchange.Kind {
		return exchanged{}, exchange.ErrNotAnExport
	}
	if read.Version < exchange.BeforeSteps || read.Version > exchange.Version {
		return exchanged{}, fmt.Errorf("%w: it is of version %d, and this cairn reads versions %d to %d", exchange.ErrNotAnExport, read.Version, exchange.BeforeSteps, exchange.Version)
	}
	if len(read.Tasks)+len(read.Objectives)+len(read.Labels)+len(read.Series)+len(read.Views) == 0 {
		return exchanged{}, fmt.Errorf("%w: it holds nothing to add", exchange.ErrNotAnExport)
	}
	return read, nil
}

func stepsOfSubTasks(read exchanged) (exchanged, error) {
	byNumber := map[int]bool{}
	for _, t := range read.Tasks {
		byNumber[t.Number] = true
	}
	var left, done = map[int][]task.Step{}, map[int][]task.Step{}
	var work []exchange.Filed
	for _, t := range read.Tasks {
		if t.Parent == 0 {
			work = append(work, t)
			continue
		}
		if !byNumber[t.Parent] {
			return exchanged{}, fmt.Errorf("%w: task %d names task %d — task %d is not in the file", exchange.ErrExportPointsOutside, t.Number, t.Parent, t.Parent)
		}
		step := task.Step{Title: t.Title, Done: t.State == string(item.Done)}
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

func (e *Engine) importing(read exchanged) (backup.Held, error) {
	read, err := stepsOfSubTasks(read)
	if err != nil {
		return backup.Held{}, err
	}
	if read, err = countedFromToday(read, time.Now()); err != nil {
		return backup.Held{}, err
	}
	order, err := checked(read)
	if err != nil {
		return backup.Held{}, err
	}
	if err := e.labelsFiled(read.Labels); err != nil {
		return backup.Held{}, err
	}
	objectives, err := e.objectivesFiled(read.Objectives)
	if err != nil {
		return backup.Held{}, err
	}
	tasks := map[int]int{}
	for _, t := range order {
		c := fromFile(t)
		serves := exchange.Numbered(t.Objectives, objectives)
		c.Objective = &serves
		made, err := e.add(t.Title, c)
		if err != nil {
			return backup.Held{}, err
		}
		tasks[t.Number] = task.OfTask(made).Number
		if err := e.spentOnDays(tasks[t.Number], t.Spent); err != nil {
			return backup.Held{}, err
		}
	}
	for _, t := range slices.Backward(order) {
		if err := e.standing(t, tasks); err != nil {
			return backup.Held{}, err
		}
	}
	for _, t := range order {
		if t.Archived {
			if err := e.archive(tasks[t.Number]); err != nil {
				return backup.Held{}, err
			}
		}
	}
	for _, t := range order {
		if t.Focused {
			if _, err := e.putInFocus(tasks[t.Number]); err != nil {
				return backup.Held{}, err
			}
		}
	}
	if err := e.seriesFiled(read.Series, tasks, objectives); err != nil {
		return backup.Held{}, err
	}
	if err := e.viewsFiled(read.Views, objectives); err != nil {
		return backup.Held{}, err
	}
	return backup.Held{Tasks: len(order), Objectives: len(read.Objectives)}, nil
}

func (e *Engine) objectivesFiled(filedAll []exchange.ObjectiveFiled) (map[int]int, error) {
	objectives := map[int]int{}
	for _, o := range filedAll {
		made, err := e.addObjective(o.Title, objectiveChangesOf(o.Shape))
		if err != nil {
			return nil, err
		}
		objectives[o.Number] = exchange.NumberOf(made)
	}
	for _, o := range filedAll {
		var c ObjectiveChanges
		if o.Parent != 0 {
			parent := strconv.Itoa(objectives[o.Parent])
			c.Parent = &parent
		}
		if len(o.WaitsOn) != 0 {
			on := exchange.Numbered(o.WaitsOn, objectives)
			c.WaitsOn = &on
		}
		if !c.named() {
			continue
		}
		if _, err := e.editObjective(objectives[o.Number], c); err != nil {
			return nil, fmt.Errorf("objective %d: %w", o.Number, err)
		}
	}
	return objectives, nil
}

func (e *Engine) standing(t exchange.Filed, tasks map[int]int) error {
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
	if t.BlockedBy != nil {
		blockers := exchange.Numbered(t.BlockedBy, tasks)
		c.Blocker = &blockers
	}
	if len(t.Linked) > 0 {
		links := exchange.PlusAll(exchange.Numbered(t.Linked, tasks))
		c.Link = &links
	}
	if !c.named() {
		return nil
	}
	_, err = e.edit(number, c)
	return err
}

func checked(read exchanged) ([]exchange.Filed, error) {
	for _, l := range read.Labels {
		if _, err := label.OneLabel(l.Label); err != nil {
			return nil, fmt.Errorf("label %q: %w", l.Label, err)
		}
		if l.Colour != "" && !slices.Contains(label.LabelColours, l.Colour) {
			return nil, fmt.Errorf("label %q: %w", l.Label, label.ErrNoSuchColour)
		}
		if _, err := label.ParseScope(l.Scope); err != nil {
			return nil, fmt.Errorf("label %q: %w", l.Label, err)
		}
	}
	objectives := map[int]bool{}
	for _, o := range read.Objectives {
		if objectives[o.Number] {
			return nil, fmt.Errorf("%w: objective %d is in the file twice", exchange.ErrNotAnExport, o.Number)
		}
		objectives[o.Number] = true
		made, err := objective.NewObjective(o.Title)
		if err == nil {
			_, err = applyObjective(made, objectiveChangesOf(o.Shape), creating)
		}
		if err != nil {
			return nil, fmt.Errorf("objective %d: %w", o.Number, err)
		}
	}
	for _, o := range read.Objectives {
		for _, n := range append(slices.Clone(o.WaitsOn), o.Parent) {
			if n != 0 && !objectives[n] {
				return nil, fmt.Errorf("%w: objective %d names objective %d — objective %d is not in the file", exchange.ErrExportPointsOutside, o.Number, n, n)
			}
		}
	}
	if err := checkedSeries(read.Series, objectives); err != nil {
		return nil, err
	}
	if err := checkedViews(read.Views, objectives); err != nil {
		return nil, err
	}
	byNumber := map[int]exchange.Filed{}
	for _, t := range read.Tasks {
		if _, twice := byNumber[t.Number]; twice {
			return nil, fmt.Errorf("%w: task %d is in the file twice", exchange.ErrNotAnExport, t.Number)
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

func checkedTask(t exchange.Filed, byNumber map[int]exchange.Filed, objectives map[int]bool) error {
	outside := func(what string, n int) error {
		return fmt.Errorf("%w: task %d names %s %d — %s %d is not in the file", exchange.ErrExportPointsOutside, t.Number, what, n, what, n)
	}
	for _, b := range slices.Concat(t.BlockedBy, t.Linked) {
		if _, in := byNumber[b]; !in {
			return outside("task", b)
		}
	}
	for _, o := range t.Objectives {
		if !objectives[o] {
			return outside("objective", o)
		}
	}
	if _, err := item.ParseState(t.State); err != nil {
		return fmt.Errorf("task %d: %w", t.Number, err)
	}
	made, err := task.NewTask(t.Title)
	if err == nil {
		_, err = apply(made, fromFile(t), creating)
	}
	for day, length := range t.Spent {
		if err == nil {
			_, err = calendar.ParseDeadline(day)
		}
		if err == nil {
			_, err = calendar.ParseEstimate(length)
		}
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

func fromFile(f exchange.Filed) Changes {
	t := f.Shape
	c := Changes{Description: &t.Description, Priority: &t.Priority, Because: &t.Because}
	if len(f.Spent) == 0 && t.TimeSpent != "" {
		c.Spent = &t.TimeSpent
	}
	labels := slices.Clone(t.Labels)
	c.Labels = &labels
	if t.Deadline != "" {
		c.Deadline = &t.Deadline
	}
	if t.Estimate != "" {
		c.Estimate = &t.Estimate
	}
	steps := slices.Clone(t.Steps)
	c.Steps = &steps
	return c
}

func objectiveChangesOf(o objective.Shape) ObjectiveChanges {
	c := ObjectiveChanges{Description: &o.Description, State: &o.State, Priority: &o.Priority}
	labels := slices.Clone(o.Labels)
	c.Labels = &labels
	if o.Deadline != "" {
		c.Deadline = &o.Deadline
	}
	if o.Estimate != "" {
		c.Estimate = &o.Estimate
	}
	return c
}
