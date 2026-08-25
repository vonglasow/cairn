package cli

// The command tree is built once, from engine.All(), rather than declared by
// hand for each of the nineteen things cairn can be asked to do: one group per
// distinct Operation.Thing, in the order All() gives them, one verb under it
// per Operation. A surface here parses, calls engine.Do, and prints — it
// decides nothing.

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"cairn/internal/engine"
)

// errNotConfirmed is returned when something that loses work was asked for
// without saying so out loud. What it was is said by naming op.Thing at the
// point it is returned, not by the sentinel itself: every scenario that
// checks this only looks for "--yes" in the reason, so one sentinel serves
// a task deleted, a backup deleted, a backup put back and a change rolled
// back alike.
var errNotConfirmed = errors.New("needs --yes to confirm")

// groupShort is the one sentence cobra shows for each group of commands.
// Operation carries a Short and a Does for a verb, but nothing for the thing
// its verbs are grouped under, so the four groups are named here, once.
//
// It stays here rather than moving to engine: a group is a shape the command
// line needs to hang verbs under, and the MCP server has no such grouping —
// every tool it exposes is named and described on its own, flat. Putting
// this where only the surface that needs it can see it is not an oversight.
var groupShort = map[string]string{
	"task":      "Add, list, read, change and delete tasks",
	"objective": "Add, list, read, change and delete objectives",
	"backup":    "Make, list, read and put back backups of the store",
	"migration": "Read and undo what has been done to the shape of the store",
}

// storePath is where the store lives, according to the environment and the
// default place. Read by every command that needs it, and by no other: a
// group and `cairn --version` must never create a file just by being run.
func storePath() (string, error) {
	path, err := engine.StoreAddress(os.LookupEnv, os.UserHomeDir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", engine.ErrStore, err)
	}
	return path, nil
}

// exactly says how many arguments a command takes, and says how to use it
// when it was given the wrong number. Cobra's own message names the count
// and not the command, which leaves someone who mistyped with nothing to
// correct.
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

// newActionGroups builds one command group per distinct Operation.Thing,
// in the order engine.All() gives them, each carrying the verbs that belong
// to it.
func newActionGroups() []*cobra.Command {
	var groups []*cobra.Command
	byThing := map[string]*cobra.Command{}
	for _, op := range engine.All() {
		group, ok := byThing[op.Thing]
		if !ok {
			group = newGroupCmd(op.Thing)
			byThing[op.Thing] = group
			groups = append(groups, group)
		}
		group.AddCommand(newActionCmd(op))
	}
	return groups
}

// newGroupCmd builds the group a thing's verbs are added under.
func newGroupCmd(thing string) *cobra.Command {
	return &cobra.Command{
		Use:   thing,
		Short: groupShort[thing],
		Args:  noneAtAll(),
		// Without a RunE the group would accept a stray argument, print its
		// help and exit zero — so a typo would look like a success.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("%s needs a command — usage: %s", thing, cmd.UseLine())
		},
	}
}

// positionalsOf is what an operation takes on the command line rather than
// as a flag, in the order it takes them.
func positionalsOf(op engine.Operation) []engine.Argument {
	var out []engine.Argument
	for _, arg := range op.Takes {
		if arg.Positional {
			out = append(out, arg)
		}
	}
	return out
}

// flagsOf is what an operation takes as a flag rather than as a positional
// argument.
func flagsOf(op engine.Operation) []engine.Argument {
	var out []engine.Argument
	for _, arg := range op.Takes {
		if !arg.Positional {
			out = append(out, arg)
		}
	}
	return out
}

// allRequired says whether every positional argument must be named for the
// operation to make sense.
func allRequired(positionals []engine.Argument) bool {
	for _, p := range positionals {
		if !p.Required {
			return false
		}
	}
	return true
}

// argsFor is how many positional arguments a command accepts: none, exactly
// as many as it has, or at most one when its one positional is not required
// — `backup restore` and `migration apply`, where a number left out means
// "the most recent" or "everything waiting" rather than a mistake.
func argsFor(positionals []engine.Argument) cobra.PositionalArgs {
	if len(positionals) == 0 {
		return noneAtAll()
	}
	if allRequired(positionals) {
		return exactly(len(positionals))
	}
	return cobra.MaximumNArgs(1)
}

// useLine is how a verb reads on the command line: its name, then each of its
// positional arguments, required ones bracketed like cobra's own convention
// and the rest in square brackets.
func useLine(op engine.Operation, positionals []engine.Argument) string {
	use := op.Verb
	for _, p := range positionals {
		if p.Required {
			use += " <" + p.Name + ">"
		} else {
			use += " [" + p.Name + "]"
		}
	}
	return use
}

// newActionCmd builds one verb: its flags and positional arguments from what
// the operation takes, a --yes gate when losing work needs one, and a RunE
// that builds an engine.Input from what cobra says was actually named,
// resolves the store's path, calls engine.Do, and prints what it answered.
func newActionCmd(op engine.Operation) *cobra.Command {
	positionals := positionalsOf(op)
	given := map[string]*string{}
	listed := map[string]*[]string{}
	var confirmed bool

	cmd := &cobra.Command{
		Use:   useLine(op, positionals),
		Short: op.Short,
		Long:  op.Does,
		Args:  argsFor(positionals),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Checked before the store is opened: removing work on a typo is
			// not something to be helpful about, and a request refused for
			// lack of confirmation has no business touching anything first.
			if op.Level == engine.Full && !confirmed {
				return fmt.Errorf("%s %w", op.Thing, errNotConfirmed)
			}
			path, err := storePath()
			if err != nil {
				return err
			}
			answer, err := engine.Do(op, path, inputFrom(cmd, positionals, args, given, listed))
			if err != nil {
				return err
			}
			return write(cmd, op, answer)
		},
	}

	for _, arg := range flagsOf(op) {
		if arg.Kind == engine.List {
			var v []string
			listed[arg.Name] = &v
			cmd.Flags().StringArrayVar(&v, arg.Name, nil, arg.Means)
			continue
		}
		var v string
		given[arg.Name] = &v
		cmd.Flags().StringVar(&v, arg.Name, "", arg.Means)
	}
	if op.Level == engine.Full {
		cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm, since this loses work")
	}
	return cmd
}

// inputFrom builds what an operation was actually asked for: its positional
// arguments, always named since cobra already checked how many there had to
// be, and its flags, named only when cobra says they were changed — a flag
// left alone is a change nobody asked for, which is the whole reason
// engine.Input tells "not named" from "named as nothing" apart.
func inputFrom(
	cmd *cobra.Command, positionals []engine.Argument, args []string,
	given map[string]*string, listed map[string]*[]string,
) engine.Input {
	in := engine.Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for i, p := range positionals {
		if i < len(args) {
			in.Given[p.Name] = args[i]
		}
	}
	for name, v := range given {
		if cmd.Flags().Changed(name) {
			in.Given[name] = *v
		}
	}
	for name, v := range listed {
		if cmd.Flags().Changed(name) {
			in.Listed[name] = *v
		}
	}
	return in
}
