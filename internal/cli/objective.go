package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// errObjectiveNotConfirmed is returned when an objective was asked to be
// deleted without saying so out loud.
var errObjectiveNotConfirmed = errors.New("deleting an objective needs --yes")

// objectiveAttributes are the flags that describe an objective. Declared once
// and bound to both `add` and `edit`, so the same word means the same thing
// whichever command carries it.
//
// It is not the task's attributes with fields left out: an objective has no
// column to sit in, nothing to hang under and no time spent on it, and a
// shared set would offer four flags that could only ever be refused.
type objectiveAttributes struct {
	title       string
	description string
	state       string
	priority    string
	labels      []string
	deadline    string
	estimate    string
}

// bind declares the attribute flags on a command. title is only offered where
// it can be changed: `add` takes it as an argument.
func (a *objectiveAttributes) bind(cmd *cobra.Command, withTitle bool) {
	if withTitle {
		cmd.Flags().StringVar(&a.title, "title", "", "what the objective is called")
	}
	cmd.Flags().StringVar(&a.description, "desc", "", "what the objective is about")
	cmd.Flags().StringVar(&a.state, "state", "", "where the objective stands")
	cmd.Flags().StringVar(&a.priority, "priority", "", "the priority it carries")
	cmd.Flags().StringArrayVar(&a.labels, "label", nil, "a label it carries")
	cmd.Flags().StringVar(&a.deadline, "deadline", "", "the day it is due")
	cmd.Flags().StringVar(&a.estimate, "estimate", "", "how long it should take")
}

// changes turns the flags a user actually typed into changes.
func (a *objectiveAttributes) changes(cmd *cobra.Command) engine.ObjectiveChanges {
	var c engine.ObjectiveChanges
	if cmd.Flags().Changed("title") {
		c.Title = &a.title
	}
	if cmd.Flags().Changed("desc") {
		c.Description = &a.description
	}
	if cmd.Flags().Changed("state") {
		c.State = &a.state
	}
	if cmd.Flags().Changed("priority") {
		c.Priority = &a.priority
	}
	if cmd.Flags().Changed("label") {
		c.Labels = &a.labels
	}
	if cmd.Flags().Changed("deadline") {
		c.Deadline = &a.deadline
	}
	if cmd.Flags().Changed("estimate") {
		c.Estimate = &a.estimate
	}
	return c
}

// objectiveNumber reads an objective number, saying so when what was given is
// not one. Worded apart from a task's, so somebody who typed the wrong kind of
// number is told which kind was expected.
func objectiveNumber(arg string) (int, error) {
	n, err := number(arg)
	if err != nil {
		return 0, fmt.Errorf("%q: an objective number is expected", arg)
	}
	return n, nil
}

// newObjectiveCmd builds the `objective` group and everything under it.
func newObjectiveCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "objective",
		Short: "Add, list, read, change and delete objectives",
		Args:  noneAtAll(),
		// Without a RunE the group would accept a stray argument, print its help
		// and exit zero — so a typo would look like a success.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("objective needs a command — usage: %s", cmd.UseLine())
		},
	}
	group.AddCommand(
		newObjectiveAddCmd(), newObjectiveListCmd(), newObjectiveShowCmd(),
		newObjectiveEditCmd(), newObjectiveDeleteCmd(),
	)
	return group
}

func newObjectiveAddCmd() *cobra.Command {
	var attrs objectiveAttributes
	cmd := &cobra.Command{
		Use:   `add "a title"`,
		Short: "Add an objective",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			o, err := eng.AddObjective(args[0], attrs.changes(cmd))
			if err != nil {
				return withTypedTitle(err, args[0])
			}
			// Nothing is under an objective that has just been created.
			return writeObjective(cmd, o, nil, domain.Progress{},
				objectiveCard(o, domain.Progress{})+serving(nil))
		},
	}
	attrs.bind(cmd, false)
	return cmd
}

func newObjectiveListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every objective",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			all, err := eng.ListObjectives()
			if err != nil {
				return err
			}
			return writeObjectives(cmd, all)
		},
	}
}

func newObjectiveShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <number>",
		Short: "Show what an objective carries, and how far along it is",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := objectiveNumber(args[0])
			if err != nil {
				return err
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			o, under, p, err := readObjective(eng, n)
			if err != nil {
				return err
			}
			return writeObjective(cmd, o, under, p, objectiveCard(o, p)+serving(under))
		},
	}
}

func newObjectiveEditCmd() *cobra.Command {
	var attrs objectiveAttributes
	cmd := &cobra.Command{
		Use:   "edit <number>",
		Short: "Change what an objective carries",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := objectiveNumber(args[0])
			if err != nil {
				return err
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			if _, err := eng.EditObjective(n, attrs.changes(cmd)); err != nil {
				return withTypedTitle(editReason(err, cmd), attrs.title)
			}
			o, under, p, err := readObjective(eng, n)
			if err != nil {
				return err
			}
			return writeObjective(cmd, o, under, p, objectiveCard(o, p)+serving(under))
		},
	}
	attrs.bind(cmd, true)
	return cmd
}

func newObjectiveDeleteCmd() *cobra.Command {
	var confirmed bool
	cmd := &cobra.Command{
		Use:   "delete <number>",
		Short: "Delete an objective",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := objectiveNumber(args[0])
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("%w: add --yes to confirm", errObjectiveNotConfirmed)
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			if err := eng.DeleteObjective(n); err != nil {
				return err
			}
			return writeObjectiveDeleted(cmd, n)
		},
	}
	cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm the deletion")
	return cmd
}

// readObjective asks for everything one objective is shown with. Three calls
// that always travel together, so the two commands that show an objective
// cannot each remember a different two of them.
func readObjective(eng *engine.Engine, n int) (domain.Objective, []domain.Task, domain.Progress, error) {
	o, err := eng.ShowObjective(n)
	if err != nil {
		return domain.Objective{}, nil, domain.Progress{}, err
	}
	under, err := eng.TasksUnder(n)
	if err != nil {
		return domain.Objective{}, nil, domain.Progress{}, err
	}
	p, err := eng.ObjectiveProgress(n)
	if err != nil {
		return domain.Objective{}, nil, domain.Progress{}, err
	}
	return o, under, p, nil
}
