package cli

// The command tree is built once, from engine.All(), rather than declared by
// hand for each of the things cairn can be asked to do: one group per
// distinct Operation.Thing, in the order All() gives them, one verb under it
// per Operation. A surface here parses, calls engine.Do, and prints — it
// decides nothing.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"cairn/internal/config"
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
	"config":    "Show which places cairn uses, and where it got them",
}

// places is where the store and the notes live, according to the environment,
// the configuration file, and the default place under the home directory, in
// that order. Read by every command that needs it, and by no other: a group
// and `cairn --version` must never create a file just by being run.
//
// The notes are only ever named, never opened, so finding them costs nothing
// and refuses nothing. With no home to put the default under, and nothing
// naming them, they are left unsaid: a place relative to wherever the command
// happened to run would send a harness to write somewhere nobody looks.
//
// The home directory is asked for unconditionally, but only required when
// the environment does not already answer: CAIRN_STORE set with no home to
// find — a cron entry, a container — still has to work, so a failed
// os.UserHomeDir is not itself a refusal.
func places() (engine.Places, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		if said, ok := os.LookupEnv(config.StoreVariable); !ok || said == "" {
			return engine.Places{}, fmt.Errorf("%w: %w", engine.ErrStore, err)
		}
	}
	places, err := config.Read(home, os.LookupEnv)
	if err != nil {
		// Only what the person wrote is theirs to fix: everything else about
		// reading the file — permissions, a missing directory, whatever a
		// driver might raise — stays folded into engine.ErrStore, the same
		// answer an unusable store gives everywhere else.
		for _, kind := range []error{config.ErrNotTOML, config.ErrUnknownKey, config.ErrNotAbsolute} {
			if errors.Is(err, kind) {
				return engine.Places{}, err
			}
		}
		return engine.Places{}, fmt.Errorf("%w: %w", engine.ErrStore, err)
	}
	notes := places.Notes()
	if !filepath.IsAbs(notes) {
		notes = ""
	}
	return engine.Places{Store: places.Store(), Notes: notes}, nil
}

// runReading runs an operation that reads a tolerant configuration instead of
// opening a store — config show today — so it stays reachable exactly when a
// broken store would otherwise turn every command away before it ran.
//
// engine.Do never fills On.Reading: nothing it opens needs to, since only this
// one operation reads it. So this calls op.Run directly, after op.Check — the
// same refusal engine.Do would have given a caller that named an argument
// wrong — rather than going through engine.Do just to hand it a path and
// backups this operation has no use for.
//
// The home comes from os.UserHomeDir(), the same as places(). A failed
// UserHomeDir is not a refusal here either: config.NewReading tolerates an
// empty home as it tolerates a missing file, so the command answers with
// what a best-effort reading found rather than joining every other command
// in refusing.
func runReading(op engine.Operation, in engine.Input) (engine.Answer, error) {
	if err := op.Check(in); err != nil {
		return engine.Answer{}, err
	}
	home, _ := os.UserHomeDir()
	reading, err := config.NewReading(home, os.LookupEnv)
	if err != nil {
		return engine.Answer{}, err
	}
	return op.Run(engine.On{Reading: reading}, in)
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
// not required — `backup restore` and `migration apply`, where a number left
// out means "the most recent" or "everything waiting" rather than a mistake.
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
// resolves the store's path, calls engine.Do, and prints what it answered.
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
			// Needs: AConfiguration reads a tolerant configuration instead of
			// a store — it has to stay reachable exactly when places()
			// would refuse, since that is when it is needed. The five backup
			// verbs carry Needs: NoStore instead: they still locate backups
			// beside where the store would be, so they keep going through
			// places() like every operation that opens one.
			if op.Needs == engine.AConfiguration {
				answer, err := runReading(op, in)
				if err != nil {
					return err
				}
				return write(cmd, op, answer)
			}
			at, err := places()
			if err != nil {
				return err
			}
			answer, err := engine.DoAt(op, at, in, engine.ByHand)
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
