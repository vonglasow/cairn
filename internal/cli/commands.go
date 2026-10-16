package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cairn/internal/engine"
)

var groupShort = map[string]string{
	"task":      "Add, list, read, change and delete tasks",
	"objective": "Add, list, read, change and delete objectives",
	"backup":    "New, list, show, restore and delete backups of the store",
	"migration": "Read and undo what has been done to the shape of the store",
	"config":    "Show which places cairn uses, and where it got them",
	"trace":     "Read what was done to a task, or over a period",
	"label":     "List every label, and delete or rename one on everything carrying it",
	"timesheet": "Say how long the work took, day by day, and add what was done without focus",
	"plugin":    "Write the Claude Code plugin this cairn carries",
}

func home() engine.Home {
	dir, err := os.UserHomeDir()
	return engine.Home{Dir: dir, Missing: err, Lookup: os.LookupEnv}
}

func exactly(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("%s takes %d argument(s), not %d — usage: %s",
				cmd.CommandPath(), n, len(args), cmd.UseLine())
		}
		return nil
	}
}

func noneAtAll() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("%s takes no argument — usage: %s", cmd.CommandPath(), cmd.UseLine())
		}
		return nil
	}
}

func atLeastOne() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%s needs at least one argument — usage: %s", cmd.CommandPath(), cmd.UseLine())
		}
		return nil
	}
}

func newActionGroups(pages engine.Pages) []*cobra.Command {
	var groups []*cobra.Command
	byThing := map[string]*cobra.Command{}
	for _, op := range engine.All() {
		if !pages.Offers(op) {
			continue
		}
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

func newGroupCmd(thing string) *cobra.Command {
	return &cobra.Command{
		Use:   thing,
		Short: groupShort[thing],
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return engine.Misread(fmt.Errorf("%s needs a command — usage: %s", thing, cmd.UseLine()))
		},
	}
}

func positionalsOf(op engine.Operation) []engine.Argument {
	var out []engine.Argument
	for _, arg := range op.Takes {
		if arg.Positional {
			out = append(out, arg)
		}
	}
	return out
}

func flagsOf(op engine.Operation) []engine.Argument {
	var out []engine.Argument
	for _, arg := range op.Takes {
		if !arg.Positional {
			out = append(out, arg)
		}
	}
	return out
}

func allRequired(positionals []engine.Argument) bool {
	for _, p := range positionals {
		if !p.Required {
			return false
		}
	}
	return true
}

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

func newActionCmd(op engine.Operation) *cobra.Command {
	positionals := positionalsOf(op)
	given := map[string]*string{}
	listed := map[string]*[]string{}
	asked := map[string]*bool{}
	var confirmed bool

	cmd := &cobra.Command{
		Use:     useLine(op, positionals),
		Short:   op.Short,
		Long:    op.Does,
		Example: indented(op.Example),
		Args:    argsFor(positionals),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := inputFrom(cmd, positionals, args, given, listed, asked)
			in.Confirmed = confirmed
			if err := readDash(op, in, cmd.InOrStdin()); err != nil {
				return err
			}
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
		if arg.Kind == engine.Bool {
			var v bool
			asked[arg.Name] = &v
			cmd.Flags().BoolVar(&v, arg.Name, false, arg.Means)
			continue
		}
		var v string
		given[arg.Name] = &v
		cmd.Flags().StringVar(&v, arg.Name, "", helpOf(arg))
	}
	if op.Level == engine.Full || op.Sweeps != nil || op.Finishes {
		cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm, since this loses work, reaches everything a narrowing does, or puts in done what hangs under a task")
	}
	return cmd
}

func indented(examples string) string {
	return "  " + strings.ReplaceAll(examples, "\n", "\n  ")
}

func helpOf(arg engine.Argument) string {
	if arg.Kind != engine.Text || len(arg.OneOf) > 0 {
		return arg.Means
	}
	return arg.Means + "; " + fromStandardInput + " reads it from standard input"
}

const fromStandardInput = "-"

func readDash(op engine.Operation, in engine.Input, stdin io.Reader) error {
	named := ""
	for _, arg := range op.Takes {
		if arg.Kind != engine.Text || in.Given[arg.Name] != fromStandardInput {
			continue
		}
		if named != "" {
			return engine.Misread(fmt.Errorf("%q and %q: only one text can be read from standard input", named, arg.Name))
		}
		named = arg.Name
	}
	if named == "" {
		return nil
	}
	read, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("reading %q from standard input: %w", named, err)
	}
	in.Given[named] = strings.TrimSuffix(string(read), "\n")
	return nil
}

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
