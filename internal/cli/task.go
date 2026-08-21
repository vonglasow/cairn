package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// errNotConfirmed is returned when a deletion was asked for without saying so
// out loud. Removing work on a typo is not something to be helpful about.
var errNotConfirmed = errors.New("deleting a task needs --yes")

// attributes are the flags that describe a task. They are declared once and
// bound to both `add` and `edit`, because the same word must mean the same
// thing whichever command carries it.
type attributes struct {
	title       string
	description string
	state       string
	priority    string
	labels      []string
	deadline    string
	estimate    string
	spent       string
	parent      string
	objective   string
}

// bind declares the attribute flags on a command. title is only offered where
// it can be changed: `add` takes it as an argument.
func (a *attributes) bind(cmd *cobra.Command, withTitle bool) {
	if withTitle {
		cmd.Flags().StringVar(&a.title, "title", "", "what the task is called")
	}
	cmd.Flags().StringVar(&a.description, "desc", "", "what the task is about")
	cmd.Flags().StringVar(&a.state, "state", "", "the column it sits in")
	cmd.Flags().StringVar(&a.priority, "priority", "", "the priority it carries")
	cmd.Flags().StringArrayVar(&a.labels, "label", nil, "a label it carries")
	cmd.Flags().StringVar(&a.deadline, "deadline", "", "the day it is due")
	cmd.Flags().StringVar(&a.estimate, "estimate", "", "how long it should take")
	cmd.Flags().StringVar(&a.spent, "spent", "", "the time already spent on it; +30m adds, -15m takes off")
	cmd.Flags().StringVar(&a.parent, "parent", "", "the number of the task it hangs under")
	cmd.Flags().StringVar(&a.objective, "objective", "", "the number of the objective it serves")
}

// changes turns the flags a user actually typed into changes. A flag left
// alone is a change nobody asked for, which is why this reads Changed rather
// than the values themselves.
func (a *attributes) changes(cmd *cobra.Command) (engine.Changes, error) {
	var c engine.Changes
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
	if cmd.Flags().Changed("spent") {
		c.Spent = &a.spent
	}
	if cmd.Flags().Changed("parent") {
		// Empty is how a task is set free, so it is read here rather than sent
		// on as a number nobody could tell apart from "hang under task zero".
		under := domain.NoParent
		if a.parent != "" {
			var err error
			if under, err = number(a.parent); err != nil {
				return engine.Changes{}, err
			}
		}
		c.Parent = &under
	}
	if cmd.Flags().Changed("objective") {
		// Empty is how a task is set loose, read here for the same reason as
		// the parent: "serves objective zero" is not a thing anyone can mean.
		serves := domain.NoObjective
		if a.objective != "" {
			var err error
			if serves, err = objectiveNumber(a.objective); err != nil {
				return engine.Changes{}, err
			}
		}
		c.Objective = &serves
	}
	return c, nil
}

// openEngine opens the store where the environment says it lives. It is called
// by the commands that need it and by no one else, so `cairn --version` never
// creates a file.
func openEngine(_ *cobra.Command) (*engine.Engine, error) {
	path, err := storePath()
	if err != nil {
		return nil, err
	}
	eng, err := engine.Open(path)
	if errors.Is(err, engine.ErrStoreBehind) {
		// The store knows it is behind; only a surface knows what somebody has
		// to type about it, so the instruction is added here rather than buried
		// in a package that has never heard of a command line.
		return nil, fmt.Errorf("%w — run `cairn migration apply`", err)
	}
	return eng, err
}

// exactly says how many arguments a command takes, and says how to use it when
// it was given the wrong number. Cobra's own message names the count and not
// the command, which leaves someone who mistyped with nothing to correct.
func exactly(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("%s takes %d argument(s), not %d — usage: %s",
				cmd.CommandPath(), n, len(args), cmd.UseLine())
		}
		return nil
	}
}

// noneAtAll refuses a stray argument, saying how the command is used.
func noneAtAll() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("%s takes no argument — usage: %s", cmd.CommandPath(), cmd.UseLine())
		}
		return nil
	}
}

// number reads a task number, saying so when what was given is not one.
func number(arg string) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("%q: a task number is expected", arg)
	}
	return n, nil
}

// newTaskCmd builds the `task` group and everything under it.
func newTaskCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "task",
		Short: "Add, list, read, change and delete tasks",
		Args:  noneAtAll(),
		// Without a RunE the group would accept a stray argument, print its help
		// and exit zero — so a typo would look like a success.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("task needs a command — usage: %s", cmd.UseLine())
		},
	}
	group.AddCommand(newAddCmd(), newListCmd(), newShowCmd(), newEditCmd(), newDeleteCmd())
	return group
}

func newAddCmd() *cobra.Command {
	var attrs attributes
	cmd := &cobra.Command{
		Use:   `add "a title"`,
		Short: "Add a task",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			asked, err := attrs.changes(cmd)
			if err != nil {
				return err
			}
			task, err := eng.Add(args[0], asked)
			if err != nil {
				return withTypedTitle(err, args[0])
			}
			// Nothing hangs under a task that has just been created, so the
			// progress is asked of nobody rather than of the store.
			return write(cmd, task, domain.Progress{}, card(task, domain.Progress{}))
		},
	}
	attrs.bind(cmd, false)
	return cmd
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every task",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			tasks, err := eng.List()
			if err != nil {
				return err
			}
			return writeList(cmd, tasks)
		},
	}
}

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <number>",
		Short: "Show what a task carries",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			task, err := eng.Show(n)
			if err != nil {
				return err
			}
			hanging, err := eng.Under(n)
			if err != nil {
				return err
			}
			p, err := eng.TaskProgress(n)
			if err != nil {
				return err
			}
			return write(cmd, task, p, card(task, p)+under(hanging))
		},
	}
}

func newEditCmd() *cobra.Command {
	var attrs attributes
	cmd := &cobra.Command{
		Use:   "edit <number>",
		Short: "Change what a task carries",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			asked, err := attrs.changes(cmd)
			if err != nil {
				return err
			}
			task, err := eng.Edit(n, asked)
			if err != nil {
				return editReason(err, cmd)
			}
			p, err := eng.TaskProgress(n)
			if err != nil {
				return err
			}
			return write(cmd, task, p, card(task, p))
		},
	}
	attrs.bind(cmd, true)
	return cmd
}

func newDeleteCmd() *cobra.Command {
	var confirmed bool
	cmd := &cobra.Command{
		Use:   "delete <number>",
		Short: "Delete a task",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("%w: add --yes to confirm", errNotConfirmed)
			}
			eng, err := openEngine(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			if err := eng.Delete(n); err != nil {
				return err
			}
			return writeDeleted(cmd, n)
		},
	}
	cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm the deletion")
	return cmd
}

// withTypedTitle puts the title back in the reason when the title is what was
// wrong, so a user sees what cairn read — including nothing at all.
func withTypedTitle(err error, typed string) error {
	if errors.Is(err, domain.ErrTitleRequired) {
		return fmt.Errorf("%w, got %q", err, typed)
	}
	return err
}

// editReason says what an edit can change, when it was asked to change nothing.
func editReason(err error, cmd *cobra.Command) error {
	if errors.Is(err, engine.ErrNothingToChange) {
		return fmt.Errorf("%w: name one of --title, --desc, --state, --priority, --label, --deadline, --estimate (%s)",
			err, cmd.UseLine())
	}
	return err
}
