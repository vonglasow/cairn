package cli

// The command tree is built from engine.All() by newActionGroups, which
// makes "one command per action" true by construction — but only for the
// shape. Nothing else checks that the words survived the trip: a flag's help,
// a command's Short and Long, whether --yes landed on the right verbs. This
// is that check.

import (
	"cairn/internal/engine"
	"testing"

	"github.com/spf13/cobra"
)

func TestTree_carriesEveryActionAndWhatItSays(t *testing.T) {
	root := NewRootCmd()

	groups := map[string]*cobra.Command{}
	for _, g := range root.Commands() {
		groups[g.Name()] = g
	}

	// wanted names every verb engine.All() declares under its thing, so a
	// second pass can tell a stray command from a declared one.
	wanted := map[string]map[string]bool{}

	for _, op := range engine.All() {
		group, ok := groups[op.Thing]
		if !ok {
			t.Fatalf("%s: no %q group in the tree", op.Line(), op.Thing)
		}

		var cmd *cobra.Command
		for _, c := range group.Commands() {
			if c.Name() == op.Verb {
				cmd = c
				break
			}
		}
		if cmd == nil {
			t.Fatalf("%s: no command under %q for it", op.Line(), op.Thing)
		}

		if wanted[op.Thing] == nil {
			wanted[op.Thing] = map[string]bool{}
		}
		wanted[op.Thing][op.Verb] = true

		if cmd.Short != op.Short {
			t.Errorf("%s: Short = %q, want %q", op.Line(), cmd.Short, op.Short)
		}
		if cmd.Long != op.Does {
			t.Errorf("%s: Long = %q, want %q", op.Line(), cmd.Long, op.Does)
		}

		for _, arg := range op.Takes {
			if arg.Positional {
				continue
			}
			flag := cmd.Flags().Lookup(arg.Name)
			if flag == nil {
				t.Errorf("%s: no --%s flag, and %q takes it", op.Line(), arg.Name, arg.Name)
				continue
			}
			if flag.Usage != arg.Means {
				t.Errorf("%s --%s: help = %q, want %q (its Means)", op.Line(), arg.Name, flag.Usage, arg.Means)
			}
			wantType := "string"
			if arg.Kind == engine.List {
				wantType = "stringArray"
			}
			if arg.Kind == engine.Bool {
				wantType = "bool"
			}
			if got := flag.Value.Type(); got != wantType {
				t.Errorf("%s --%s: flag type = %q, want %q", op.Line(), arg.Name, got, wantType)
			}
		}

		yes := cmd.Flags().Lookup("yes")
		if op.Level == engine.Full && yes == nil {
			t.Errorf("%s: no --yes flag, and its Level is Full", op.Line())
		}
		if op.Level != engine.Full && yes != nil {
			t.Errorf("%s: has a --yes flag, and its Level is not Full", op.Line())
		}
	}

	// `cairn mcp serve` is deliberately not an action (see mcp.go: it is the
	// same cairn answering an agent instead of a person, not one of the
	// the operations), so the mcp group is never in wanted and is never
	// walked here. Cobra's own completion and help, and fang's hidden man,
	// are root-level commands rather than verbs under an action group, so
	// this loop — which only walks the things engine.All() named — never
	// reaches them either.
	for thing, verbs := range wanted {
		for _, c := range groups[thing].Commands() {
			if !verbs[c.Name()] {
				t.Errorf("%s %s: in the tree, but not declared by any action", thing, c.Name())
			}
		}
	}
}
