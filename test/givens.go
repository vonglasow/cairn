package test

// The givens a board's scenarios share. A scenario says what is in the store,
// not how it got there, and the two boards — the terminal's and the window's —
// say it in the same words. Each used to carry its own copy of every one of
// these steps, and each reached past the operations into the engine's methods
// to write them; written once here, they go through the same door everything
// else does.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Asker is somebody else at the store: a second connection, a command line,
// a server. What a board's givens are written through, because a board notices
// another's writes and not its own.
type Asker interface {
	Ask(engine.Operation, engine.Input) (engine.Answer, error)
}

// Givens writes a board's givens through whoever Other answers at the moment a
// step runs — a scenario opens a fresh store before each one.
type Givens struct {
	Other func() Asker
}

// ask runs one operation as somebody else, handing over what a person would
// have typed.
func (g Givens) ask(op engine.Operation, given map[string]string, listed map[string][]string) (engine.Answer, error) {
	return g.Other().Ask(op, engine.Input{Given: given, Listed: listed})
}

// addTask adds a task carrying what a scenario named about it.
func (g Givens) addTask(title string, with map[string]string, objectives ...string) error {
	given := map[string]string{"title": title}
	for name, value := range with {
		given[name] = value
	}
	// Named always, empty or not: a task added serving nothing is a task
	// added with no objective named, and the engine reads the two alike.
	_, err := g.ask(engine.TaskAdd, given, map[string][]string{"objective": objectives})
	return err
}

// editTask changes what a scenario named about a task.
func (g Givens) editTask(number int, given map[string]string, listed map[string][]string) error {
	if given == nil {
		given = map[string]string{}
	}
	given["number"] = strconv.Itoa(number)
	_, err := g.ask(engine.TaskEdit, given, listed)
	return err
}

// IAddedATaskTitled adds a task carrying nothing but its title.
func (g Givens) IAddedATaskTitled(title string) error { return g.addTask(title, nil) }

// IAddedATaskWithTheEstimate adds a task carrying an estimate.
func (g Givens) IAddedATaskWithTheEstimate(title, estimate string) error {
	return g.addTask(title, map[string]string{"estimate": estimate})
}

// IAddedATaskWithTheState adds a task straight into a column.
func (g Givens) IAddedATaskWithTheState(title, state string) error {
	return g.addTask(title, map[string]string{"state": state})
}

// IAddedATaskWithTheDescription adds a task with something written about it.
func (g Givens) IAddedATaskWithTheDescription(title, description string) error {
	return g.addTask(title, map[string]string{"desc": description})
}

// IAddedATaskWithTheParagraphs adds a task whose description is two paragraphs.
func (g Givens) IAddedATaskWithTheParagraphs(title, first, second string) error {
	return g.IAddedATaskWithTheDescription(title, first+"\n"+second)
}

// IAddedATaskWithADescriptionOfLines adds a task whose description is long
// enough to scroll, ending on a line the scenario names.
func (g Givens) IAddedATaskWithADescriptionOfLines(title string, lines int, end string) error {
	said := make([]string, 0, lines)
	for i := range lines - 1 {
		said = append(said, fmt.Sprintf("line %d of a description long enough to scroll", i+1))
	}
	return g.IAddedATaskWithTheDescription(title, strings.Join(append(said, end), "\n"))
}

// IAddedATaskServingObjective adds a task towards an objective.
func (g Givens) IAddedATaskServingObjective(title string, objective int) error {
	return g.addTask(title, nil, strconv.Itoa(objective))
}

// IAddedATaskWithTheStateServingObjective adds a task in a column, towards an
// objective.
func (g Givens) IAddedATaskWithTheStateServingObjective(title, state string, objective int) error {
	return g.addTask(title, map[string]string{"state": state}, strconv.Itoa(objective))
}

// IAddedATaskHangingUnderTask adds a part of another task's work.
func (g Givens) IAddedATaskHangingUnderTask(title string, under int) error {
	return g.addTask(title, map[string]string{"parent": strconv.Itoa(under)})
}

// IAddedTasksTo adds n tasks to one column, numbered in their titles.
func (g Givens) IAddedTasksTo(n int, state string) error {
	for i := 1; i <= n; i++ {
		if err := g.IAddedATaskWithTheState(fmt.Sprintf("Task %d", i), state); err != nil {
			return err
		}
	}
	return nil
}

// IAddedAnObjectiveTitled adds an objective carrying nothing but its title.
func (g Givens) IAddedAnObjectiveTitled(title string) error {
	_, err := g.ask(engine.ObjectiveAdd, map[string]string{"title": title}, nil)
	return err
}

// IArchivedTask takes a task off the board.
func (g Givens) IArchivedTask(number int) error {
	_, err := g.ask(engine.TaskArchive, map[string]string{"number": strconv.Itoa(number)}, nil)
	return err
}

// IEditedTaskWithTheBlocker puts a task in another's way.
func (g Givens) IEditedTaskWithTheBlocker(number int, blocker string) error {
	return g.editTask(number, nil, map[string][]string{"blocker": {blocker}})
}

// IEditedTaskWithTheState moves a task to another column.
func (g Givens) IEditedTaskWithTheState(number int, state string) error {
	return g.editTask(number, map[string]string{"state": state}, nil)
}

// IEditedTaskWithTheLabels puts two labels on a task.
func (g Givens) IEditedTaskWithTheLabels(number int, first, second string) error {
	return g.editTask(number, nil, map[string][]string{"label": {first, second}})
}

// IEditedTaskWithTheDeadline gives a task a day it is due.
func (g Givens) IEditedTaskWithTheDeadline(number int, day string) error {
	return g.editTask(number, map[string]string{"deadline": day}, nil)
}

// IDeletedTask deletes a task.
func (g Givens) IDeletedTask(number int) error {
	_, err := g.ask(engine.TaskDelete, map[string]string{"number": strconv.Itoa(number)}, nil)
	return err
}

// IPutTaskInFocus starts working on a task.
func (g Givens) IPutTaskInFocus(number int) error {
	_, err := g.ask(engine.TaskFocus, map[string]string{"number": strconv.Itoa(number)}, nil)
	return err
}

// ISpentOnTask records time spent on a task on a given day, the way the
// timesheet is told about work done without the focus.
func (g Givens) ISpentOnTask(length string, number int, day string) error {
	_, err := g.ask(engine.TimesheetAdd, map[string]string{"number": strconv.Itoa(number), "time": length, "on": day}, nil)
	return err
}

// Task reads a task back as somebody else would see it.
func (g Givens) Task(number int) (task.Task, error) {
	answer, err := g.ask(engine.TaskShow, map[string]string{"number": strconv.Itoa(number)}, nil)
	if err != nil {
		return task.Task{}, err
	}
	return *answer.Task, nil
}

// Tasks reads back every task on the board, as somebody else would see them.
func (g Givens) Tasks() ([]task.Task, error) {
	answer, err := g.ask(engine.TaskList, nil, nil)
	return answer.Tasks, err
}

// Objective reads an objective back as somebody else would see it.
func (g Givens) Objective(number int) (objective.Objective, error) {
	answer, err := g.ask(engine.ObjectiveShow, map[string]string{"number": strconv.Itoa(number)}, nil)
	if err != nil {
		return objective.Objective{}, err
	}
	return *answer.Objective, nil
}

// Register declares the givens both boards share. A store made before the
// scenario began is the store of its own a scenario asks for, so the two
// steps that say so have nothing left to do.
func (g Givens) Register(sc *godog.ScenarioContext) {
	sc.Given(`^a store of my own, somewhere that is not the default place$`, func() error { return nil })
	sc.Given(`^nothing in it yet$`, func() error { return nil })
	sc.Step(`^I added a task titled "([^"]*)"$`, g.IAddedATaskTitled)
	sc.Step(`^I added a task titled "([^"]*)" with the estimate "([^"]*)"$`, g.IAddedATaskWithTheEstimate)
	sc.Step(`^I added a task titled "([^"]*)" with the state "([^"]*)"$`, g.IAddedATaskWithTheState)
	sc.Step(`^I added a task titled "([^"]*)" with the description "([^"]*)"$`, g.IAddedATaskWithTheDescription)
	sc.Step(`^I added a task titled "([^"]*)" with the paragraphs "([^"]*)" and "([^"]*)"$`, g.IAddedATaskWithTheParagraphs)
	sc.Step(`^I added a task titled "([^"]*)" with a description of (\d+) lines ending "([^"]*)"$`,
		g.IAddedATaskWithADescriptionOfLines)
	sc.Step(`^I added a task titled "([^"]*)" serving objective (\d+)$`, g.IAddedATaskServingObjective)
	sc.Step(`^I added a task titled "([^"]*)" with the state "([^"]*)" serving objective (\d+)$`,
		g.IAddedATaskWithTheStateServingObjective)
	sc.Step(`^I added a task titled "([^"]*)" hanging under task (\d+)$`, g.IAddedATaskHangingUnderTask)
	sc.Step(`^I added (\d+) tasks to "([^"]*)"$`, g.IAddedTasksTo)
	sc.Step(`^I added an objective titled "([^"]*)"$`, g.IAddedAnObjectiveTitled)
	sc.Step(`^I spent "([^"]*)" on task (\d+) on "([^"]*)"$`, g.ISpentOnTask)
	sc.Step(`^I archived task (\d+)$`, g.IArchivedTask)
	sc.Step(`^I edited task (\d+) with the blocker "([^"]*)"$`, g.IEditedTaskWithTheBlocker)
	sc.Step(`^I edited task (\d+) with the state "([^"]*)"$`, g.IEditedTaskWithTheState)
	sc.Step(`^I edited task (\d+) with the labels "([^"]*)" and "([^"]*)"$`, g.IEditedTaskWithTheLabels)
	sc.Step(`^I edited task (\d+) with the deadline "([^"]*)"$`, g.IEditedTaskWithTheDeadline)
	sc.Step(`^I deleted task (\d+)$`, g.IDeletedTask)
	sc.Step(`^I put task (\d+) in focus$`, g.IPutTaskInFocus)
}
