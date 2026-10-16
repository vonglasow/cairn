package cli

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"cairn/internal/config"
	"cairn/internal/engine"
)

func TestConfigShow_isNotStoppedByAStoreItCannotResolve(t *testing.T) {
	t.Setenv("HOME", "")
	if err := os.Unsetenv("HOME"); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.StoreVariable, "")
	if err := os.Unsetenv(config.StoreVariable); err != nil {
		t.Fatal(err)
	}

	r := run(t, "config", "show")
	if r.Code != exitOK {
		t.Fatalf("config show exit = %d, want %d (%s)", r.Code, exitOK, r.Stderr)
	}
}

func TestTree_carriesEveryActionAndWhatItSays(t *testing.T) {
	root := NewRootCmd()

	groups := map[string]*cobra.Command{}
	for _, g := range root.Commands() {
		groups[g.Name()] = g
	}

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
			if flag.Usage != helpOf(arg) {
				t.Errorf("%s --%s: help = %q, want %q (its Means)", op.Line(), arg.Name, flag.Usage, helpOf(arg))
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
		mayAsk := op.Level == engine.Full || op.Sweeps != nil || op.Finishes
		if mayAsk && yes == nil {
			t.Errorf("%s: no --yes flag, and it may ask first", op.Line())
		}
		if !mayAsk && yes != nil {
			t.Errorf("%s: has a --yes flag, and it never asks first", op.Line())
		}
	}

	for thing, verbs := range wanted {
		for _, c := range groups[thing].Commands() {
			if !verbs[c.Name()] {
				t.Errorf("%s %s: in the tree, but not declared by any action", thing, c.Name())
			}
		}
	}
}

func TestEveryExampleIsACommandTheTreeReads(t *testing.T) {
	for _, op := range engine.All() {
		cmd, _, err := NewRootCmd().Find([]string{op.Thing, op.Verb})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(cmd.Example, "cairn "+op.Line()) {
			t.Errorf("%s: --help shows no example: %q", op.Line(), cmd.Example)
		}
		for _, line := range strings.Split(op.Example, "\n") {
			words, err := shellWords(strings.TrimSpace(strings.Split(line, " < ")[0]))
			if err != nil {
				t.Fatalf("%q: %v", line, err)
			}
			root := NewRootCmd()
			found, rest, err := root.Find(words[1:])
			if err != nil {
				t.Errorf("%q: %v", line, err)
				continue
			}
			if err := found.ParseFlags(rest); err != nil {
				t.Errorf("%q: %v", line, err)
				continue
			}
			if err := found.ValidateArgs(found.Flags().Args()); err != nil {
				t.Errorf("%q: %v", line, err)
			}
		}
	}
}

func shellWords(line string) ([]string, error) {
	var words []string
	var word strings.Builder
	quoted, started := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '\\' && quoted && i+1 < len(line):
			i++
			word.WriteByte(line[i])
		case c == '"':
			quoted, started = !quoted, true
		case c == ' ' && !quoted:
			if started {
				words = append(words, word.String())
				word.Reset()
				started = false
			}
		default:
			word.WriteByte(c)
			started = true
		}
	}
	if quoted {
		return nil, errors.New("a quote is left open")
	}
	if started {
		words = append(words, word.String())
	}
	return words, nil
}
