package cli

import (
	"reflect"
	"strings"
	"testing"
)

// splitArgs turns the command line a scenario wrote into the arguments a shell
// would have handed over: words separated by spaces, unless quoted. It is
// deliberately small — a scenario writes commands a person types, not shell
// programs.
func splitArgs(line string) []string {
	var (
		args    []string
		current strings.Builder
		quoted  bool
		started bool
	)
	for _, r := range line {
		switch {
		case r == '"':
			quoted = !quoted
			started = true
		case r == ' ' && !quoted:
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(r)
			started = true
		}
	}
	if started {
		args = append(args, current.String())
	}
	return joinFlagValues(args)
}

// joinFlagValues puts back together a flag value that carried a space. A
// scenario's Examples table cannot quote its cells, so `--desc the one` arrives
// as three words where a shell would have been handed two. Everything between a
// flag and the next one is that flag's value.
func joinFlagValues(args []string) []string {
	joined := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		joined = append(joined, args[i])
		if !strings.HasPrefix(args[i], "--") || args[i] == "--" {
			continue
		}
		value := []string{}
		for i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			value = append(value, args[i+1])
			i++
		}
		if len(value) > 0 {
			joined = append(joined, strings.Join(value, " "))
		}
	}
	return joined
}

func TestSplitArgs_readsACommandLineTheWayAShellWould(t *testing.T) {
	for _, tc := range []struct {
		line string
		want []string
	}{
		{`cairn task add "Write the spec"`, []string{"cairn", "task", "add", "Write the spec"}},
		{`cairn task add ""`, []string{"cairn", "task", "add", ""}},
		{`cairn task add "   "`, []string{"cairn", "task", "add", "   "}},
		{`cairn task add Write the spec`, []string{"cairn", "task", "add", "Write", "the", "spec"}},
		{`cairn task add -- "--not-a-flag"`, []string{"cairn", "task", "add", "--", "--not-a-flag"}},
		{`cairn task add "a" "b"`, []string{"cairn", "task", "add", "a", "b"}},
		{`cairn task list`, []string{"cairn", "task", "list"}},
		{`cairn task add "Spec" --desc the one`, []string{"cairn", "task", "add", "Spec", "--desc", "the one"}},
		{`cairn task add "Spec" --priority high --state todo`, []string{"cairn", "task", "add", "Spec", "--priority", "high", "--state", "todo"}},
	} {
		if got := splitArgs(tc.line); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("splitArgs(%q) = %#v, want %#v", tc.line, got, tc.want)
		}
	}
}
