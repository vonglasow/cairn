package engine

// The task operations: add, list, show, edit and delete a task, described
// once so both surfaces build the same command from the same facts.

import (
	"errors"
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// Name is how an operation is named as a command or a tool: "task_add".
func (o Operation) Name() string { return o.Thing + "_" + o.Verb }

// Line is how an operation reads on the command line: "task add".
func (o Operation) Line() string { return o.Thing + " " + o.Verb }

// All is the twenty-one things cairn can be asked to do, in a stable order.
func All() []Operation {
	return []Operation{
		TaskAdd, TaskList, TaskShow, TaskEdit, TaskDelete, TaskSearch,
		ObjectiveAdd, ObjectiveList, ObjectiveShow, ObjectiveEdit, ObjectiveDelete, ObjectiveSearch,
		BackupAdd, BackupList, BackupShow, BackupDelete, BackupRestore,
		MigrationList, MigrationShow, MigrationApply, MigrationRollback,
	}
}

// TaskAdd records a new task.
var TaskAdd = Operation{
	Thing: "task",
	Verb:  "add",
	Short: "Add a task",
	Does:  "Add a task, carrying whatever of its attributes were given.",
	Level: Write,
	Needs: AStore,
	Takes: append([]Argument{
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the task is called"},
	}, taskAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		title := ""
		if given := in.Change("title"); given != nil {
			title = *given
		}
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		added, err := on.Engine.Add(title, changes)
		if err != nil {
			return Answer{}, withTypedTitle(err, title)
		}
		// Nothing hangs under a task that has just been created, so the
		// progress is asked of nobody rather than of the store.
		return Answer{Task: &added, Progress: domain.Progress{}}, nil
	},
}

// TaskList names every task.
var TaskList = Operation{
	Thing: "task",
	Verb:  "list",
	Short: "List every task",
	Does:  "List every task.",
	Level: Read,
	Needs: AStore,
	Run: func(on On, _ Input) (Answer, error) {
		tasks, err := on.Engine.List()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Tasks: tasks}, nil
	},
}

// TaskShow says what a task carries, what hangs under it and how far along it
// is.
var TaskShow = Operation{
	Thing: "task",
	Verb:  "show",
	Short: "Show what a task carries",
	Does:  "Show what a task carries, what hangs under it and how far along it is.",
	Level: Read,
	Needs: AStore,
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		shown, err := on.Engine.Show(n)
		if err != nil {
			return Answer{}, err
		}
		under, err := on.Engine.Under(n)
		if err != nil {
			return Answer{}, err
		}
		p, err := on.Engine.TaskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &shown, Under: under, Progress: p}, nil
	},
}

// TaskEdit changes what a task carries.
var TaskEdit = Operation{
	Thing: "task",
	Verb:  "edit",
	Short: "Change what a task carries",
	Does:  "Change what a task carries.",
	Level: Write,
	Needs: AStore,
	Takes: append([]Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
		{Name: "title", Kind: Text, Means: "what the task is called"},
	}, taskAttributes...),
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		changes, err := taskChanges(in)
		if err != nil {
			return Answer{}, err
		}
		changes.Title = in.Change("title")
		edited, err := on.Engine.Edit(n, changes)
		if err != nil {
			return Answer{}, editReason(err)
		}
		p, err := on.Engine.TaskProgress(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Task: &edited, Progress: p}, nil
	},
}

// TaskDelete removes a task.
var TaskDelete = Operation{
	Thing: "task",
	Verb:  "delete",
	Short: "Delete a task",
	Does:  "Delete a task.",
	Level: Full,
	Needs: AStore,
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		if err := on.Engine.Delete(n); err != nil {
			return Answer{}, err
		}
		return Answer{Deleted: &n}, nil
	},
}

// TaskSearch names every task carrying everything named, wherever among its
// facts each one happens to sit.
var TaskSearch = Operation{
	Thing: "task",
	Verb:  "search",
	Short: "Look for a task by anything it carries",
	Does:  "List every task carrying everything named, wherever among its facts each one sits.",
	Level: Read,
	Needs: AStore,
	Takes: []Argument{
		{Name: "named", Kind: List, Positional: true, Required: true, Means: "what to look for, read the way it is shown"},
	},
	Run: func(on On, in Input) (Answer, error) {
		named := in.Changed("named")
		tasks, err := on.Engine.List()
		if err != nil {
			return Answer{}, err
		}
		found := []task.Task{}
		for _, t := range tasks {
			if named != nil && task.Matching(t, *named) {
				found = append(found, t)
			}
		}
		return Answer{Found: found}, nil
	},
}

// taskAttributes are the ten attributes a task carries, minus title, which is
// taken as a positional argument where it can be given at all.
var taskAttributes = []Argument{
	{Name: "desc", Kind: Text, Means: "what the task is about"},
	{Name: "state", Kind: Text, OneOf: oneOf(domain.States), Means: "the column it sits in"},
	{Name: "priority", Kind: Text, OneOf: oneOf(domain.Priorities), Means: "the priority it carries"},
	{Name: "label", Kind: List, Means: "a label it carries"},
	{Name: "deadline", Kind: Day, Means: "the day it is due"},
	{Name: "estimate", Kind: Duration, Means: "how long it should take"},
	{Name: "spent", Kind: Duration, Means: "the time already spent on it; +30m adds, -15m takes off"},
	{Name: "parent", Kind: Number, Of: "a task", Means: "the number of the task it hangs under"},
	{Name: "objective", Kind: Number, Of: "an objective", Means: "the number of the objective it serves"},
}

// oneOf turns a closed set of typed values into the strings Argument.OneOf
// needs, in one place rather than once per attribute that has one.
func oneOf[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}

// withTypedTitle puts the title back in the reason when the title is what was
// wrong, so whoever asked sees what cairn read — including nothing at all.
func withTypedTitle(err error, typed string) error {
	if errors.Is(err, domain.ErrTitleRequired) {
		return fmt.Errorf("%w, got %q", err, typed)
	}
	return err
}

// editReason says what an edit can change, when it was asked to change nothing.
func editReason(err error) error {
	if errors.Is(err, ErrNothingToChange) {
		return fmt.Errorf("%w: name one of --title, --desc, --state, --priority, --label, --deadline, --estimate", err)
	}
	return err
}

// taskChanges turns what was named in in into the changes Add and
// Edit take.
func taskChanges(in Input) (Changes, error) {
	var c Changes
	c.Description = in.Change("desc")
	c.State = in.Change("state")
	c.Priority = in.Change("priority")
	c.Labels = in.Changed("label")
	c.Deadline = in.Change("deadline")
	c.Estimate = in.Change("estimate")
	c.Spent = in.Change("spent")
	if in.Named("parent") {
		// Empty is how a task is set free, and Whole reads it as zero — the
		// same task.NoParent a number nobody could tell apart from "hang
		// under task zero" would otherwise have to spell out here.
		under, _ := in.Whole("parent")
		c.Parent = &under
	}
	if in.Named("objective") {
		// Empty is how a task is set loose, for the same reason as the parent.
		serves, _ := in.Whole("objective")
		c.Objective = &serves
	}
	return c, nil
}
