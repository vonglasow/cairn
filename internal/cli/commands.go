package cli

// The command tree is built once, from engine.All(), rather than declared by
// hand for each of the things cairn can be asked to do: one group per
// distinct Operation.Thing, in the order All() gives them, one verb under it
// per Operation. A surface here parses, calls engine.DoFrom, and prints — it
// decides nothing.

import (
	"errors"
	"fmt"
	"os"
	"strconv"

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
// its verbs are grouped under, so the groups are named here, once — every
// thing engine.All() names has one, which TestEveryThingHasAGroupSentence holds.
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
	"config":    "Show which places cairn uses, and where it got them",
	"trace":     "Read what was done to a task, or over a period",
	"timesheet": "Say how long the work took, day by day, and add what was done without focus",
}

// home is what this process knows about where it runs: its home directory, or
// why there is none, and its environment. Where the store and the notes live
// is worked out from it by the engine, not here.
func home() engine.Home {
	dir, err := os.UserHomeDir()
	return engine.Home{Dir: dir, Missing: err, Lookup: os.LookupEnv}
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

// atLeastOne refuses being given nothing at all, saying how the command is
// used — for a required positional of Kind: List, which takes any number of
// arguments but has to take at least one.
func atLeastOne() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%s needs at least one argument — usage: %s", cmd.CommandPath(), cmd.UseLine())
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
// as many as it has, at least one when its one positional is a required
// Kind: List — `task search`, which takes any number of things to look for
// but has to be given something — or at most one when its one positional is
// not required — `backup restore`, `migration apply`, `task focus`, `task
// unfocus` and `task archive`, where a number left out means "the most
// recent", "everything waiting", "what is in focus" or "what the narrowing
// reaches" rather than a mistake.
func argsFor(positionals []engine.Argument) cobra.PositionalArgs {
	if len(positionals) == 0 {
		return noneAtAll()
	}
	if len(positionals) == 1 && positionals[0].Kind == engine.List && positionals[0].Required {
		return atLeastOne()
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
// and hands it to engine.DoFrom with this process's home, and prints what it
// answered.
func newActionCmd(op engine.Operation) *cobra.Command {
	positionals := positionalsOf(op)
	given := map[string]*string{}
	listed := map[string]*[]string{}
	asked := map[string]*bool{}
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
			in := inputFrom(cmd, positionals, args, given, listed, asked)
			answer, err := engine.DoFrom(op, home(), in, engine.ByHand)
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
		// A Kind: Bool argument is a flag that carries nothing, so it is
		// declared as one rather than as a string somebody has to type "true"
		// into. What reaches the engine is still a string, since Input speaks
		// one language for every kind.
		if arg.Kind == engine.Bool {
			var v bool
			asked[arg.Name] = &v
			cmd.Flags().BoolVar(&v, arg.Name, false, arg.Means)
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
	given map[string]*string, listed map[string]*[]string, asked map[string]*bool,
) engine.Input {
	in := engine.Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for i, p := range positionals {
		if p.Kind == engine.List {
			in.Listed[p.Name] = args
			continue
		}
		if i < len(args) {
			in.Given[p.Name] = args[i]
		}
	}
	for name, v := range given {
		if cmd.Flags().Changed(name) {
			in.Given[name] = *v
		}
	}
	for name, v := range asked {
		if cmd.Flags().Changed(name) {
			in.Given[name] = strconv.FormatBool(*v)
		}
	}
	for name, v := range listed {
		if cmd.Flags().Changed(name) {
			in.Listed[name] = *v
		}
	}
	return in
}
