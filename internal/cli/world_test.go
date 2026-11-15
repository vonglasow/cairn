package cli

import (
	"bytes"
	"cairn/internal/engine/item"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/test"
)

type world struct {
	*test.World
	file string

	window   opensWindow
	terminal opensBoard
	stamped  func() build.Info
}

func newWorld(t *testing.T) *world {
	w := &world{}
	w.World = test.NewWorld(t, func(_ *testing.T, h engine.Home, args ...string) test.Result {
		return runWith(w.context(), h, args...)
	})
	return w
}

func (w *world) context() context.Context {
	ctx := context.Background()
	if w.window != nil {
		ctx = withWindow(ctx, w.window)
	}
	if w.terminal != nil {
		ctx = withBoard(ctx, w.terminal)
	}
	if w.stamped != nil {
		ctx = withStamps(ctx, w.stamped)
	}
	return ctx
}

func (w *world) field(key string) string {
	for _, l := range strings.Split(w.Out.Stdout, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), key); ok {
			value := strings.TrimSpace(rest)
			if value == item.Absent {
				return ""
			}
			return value
		}
	}
	return ""
}

func (w *world) heading() string {
	return strings.TrimSpace(strings.SplitN(w.Out.Stdout, "\n", 2)[0])
}

func listedUnder(out, key string) ([]map[string]any, map[string]any, error) {
	var envelope map[string]any
	if err := json.Unmarshal([]byte(out), &envelope); err != nil {
		return nil, nil, fmt.Errorf("the answer is not one object (%w):\n%s", err, out)
	}
	raw, carried := envelope[key]
	if !carried {
		named := make([]string, 0, len(envelope))
		for k := range envelope {
			named = append(named, k)
		}
		sort.Strings(named)
		return nil, envelope, fmt.Errorf("the answer carries %v, and not %q", named, key)
	}
	items, list := raw.([]any)
	if !list && raw != nil {
		return nil, envelope, fmt.Errorf("%q is not a list but a %T", key, raw)
	}
	rows := make([]map[string]any, 0, len(items))
	for _, one := range items {
		row, _ := one.(map[string]any)
		rows = append(rows, row)
	}
	return rows, envelope, nil
}

var numbersIn = regexp.MustCompile(`\d+`)

func flagFor(attribute string) string {
	return map[string]string{
		"title": "--title", "description": "--desc", "deadline": "--deadline",
		"estimate": "--estimate",
		"priority": "--priority", "state": "--state",
		"time spent": "--spent", "objective": "--objective", "blocker": "--blocker",
		"because": "--because", "label": "--label", "link": "--link",
	}[attribute]
}

func (w *world) register(sc *godog.ScenarioContext) {
	w.Register(sc)
	w.registerStep(sc)
	w.registerSession(sc)

	sc.Step(`^it is (\d{2}):(\d{2}) on (\d{4})-(\d{2})-(\d{2})$`, func(hour, minute, year, month, day int) error {
		at := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
		w.Setenv(engine.NowVariable, at.Format(time.RFC3339))
		return nil
	})

	sc.Step(`^it says the (\w+) went from "([^"]*)" to "([^"]*)"$`,
		func(field, was, became string) error {
			for _, said := range []string{field, was, became} {
				if err := test.WantContains(w.Out.Stdout, said, "the change"); err != nil {
					return err
				}
			}
			return nil
		})

	sc.Step(`^it carries (\d+) acts?$`, func(want int) error {
		if got := actsIn(w.Out.Stdout); len(got) != want {
			return fmt.Errorf("%d acts, want %d:\n%s", len(got), want, w.Out.Stdout)
		}
		return nil
	})

	sc.Step("^`([^`]*)` says it sits under nothing$", func(line string) error {
		w.Run(line)
		if got := w.field("parent"); got != "" {
			return fmt.Errorf("it sits under %q, want nothing:\n%s%s", got, w.Out.Stdout, w.Out.Stderr)
		}
		return test.WantContains(w.Out.Stdout, "\nparent", "the answer")
	})

	sc.Step(`^the answer says "([^"]*)"$`, func(said string) error {
		return test.WantContains(w.Out.Stdout, said, "the answer")
	})

	sc.Step(`^the answer does not say "([^"]*)"$`, func(said string) error {
		if strings.Contains(w.Out.Stdout, said) {
			return fmt.Errorf("the answer says %q:\n%s", said, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^I added a task titled "([^"]*)" with the (description|deadline|estimate|priority|state|time spent|objective) "([^"]*)"$`,
		func(title, attribute, value string) error {
			return w.Seed(fmt.Sprintf(`cairn task new "%s" %s "%s"`, title, flagFor(attribute), value))
		})

	sc.Step(`^I added a task titled "([^"]*)" serving objective (\d+)$`, func(title string, n int) error {
		return w.Seed(fmt.Sprintf(`cairn task new "%s" --objective %d`, title, n))
	})

	sc.Step(`^I edited task (\d+) with the labels "([^"]*)" and "([^"]*)"$`,
		func(n int, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task edit %d --label "=%s" --label "%s"`, n, first, second))
		})

	sc.Step(`^I added a task titled "([^"]*)" with the objectives "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task new "%s" --objective "%s" --objective "%s"`,
				title, first, second))
		})

	sc.Step(`^I added a task titled "([^"]*)" with the labels "([^"]*)" and "([^"]*)"$`,
		func(title, first, second string) error {
			return w.Seed(fmt.Sprintf(`cairn task new "%s" --label "%s" --label "%s"`, title, first, second))
		})

	sc.Step(`^I edited task (\d+) with the (title|description|deadline|estimate|priority|state|time spent|blocker|because|objective|label|link) "([^"]*)"$`,
		func(n int, attribute, value string) error {
			if slices.Contains([]string{"objective", "blocker", "label", "link"}, attribute) && !strings.ContainsAny(value[:min(1, len(value))], "+-") {
				value = "=" + value
			}
			return w.Seed(fmt.Sprintf(`cairn task edit %d %s "%s"`, n, flagFor(attribute), value))
		})

	sc.Step(`^I deleted task (\d+)$`, func(n int) error {
		return w.Seed(fmt.Sprintf(`cairn task delete %d --yes`, n))
	})

	sc.Step(`^cairn has exited$`, func() error {
		return nil
	})

	sc.Step(`^it does not crash$`, func() error {
		if strings.Contains(w.Out.Stderr, "panic:") {
			return fmt.Errorf("it panicked:\n%s", w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it exits with code (\d+)$`, func(want int) error {
		if w.Out.Code != want {
			return fmt.Errorf("exit %d, want %d (%s)", w.Out.Code, want, w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it writes nothing to standard error$`, func() error {
		if w.Out.Stderr != "" {
			return fmt.Errorf("stderr = %q, want nothing", w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it writes the reason to standard error$`, func() error {
		if w.Out.Stderr == "" {
			return fmt.Errorf("stderr is empty, want the reason")
		}
		return nil
	})

	sc.Step(`^it writes the task to standard output$`, func() error {
		if w.Out.Stdout == "" {
			return fmt.Errorf("stdout is empty, want the task")
		}
		return nil
	})

	for step, want := range map[string]string{
		`^it says a title is one line$`:                                           "a title is one line",
		`^it says a task needs a title$`:                                          "a task needs a title",
		`^it says a label needs a name$`:                                          "a label needs a name",
		`^it says a label has no family$`:                                         "a label has no family",
		`^it says what a label may hold$`:                                         "a label holds only",
		`^it says there is no such task$`:                                         "there is no such task",
		`^it says a task number is expected$`:                                     "a task number is expected",
		`^it says how to confirm$`:                                                "--yes",
		`^it says the flag is unknown$`:                                           "unknown flag",
		`^it says the store cannot be read$`:                                      "the store cannot be read",
		`^it says what it can change$`:                                            "--title",
		`^it says the labels are either replaced or changed, not both$`:           "either replaced or changed",
		`^it says a label is added or taken away only when changing a task$`:      "only when changing a task",
		`^it says a task cannot block itself$`:                                    "cannot block itself",
		`^it says a task cannot be linked to itself$`:                             "cannot be linked to itself",
		`^it says the objectives are either replaced or changed, not both$`:       "either replaced or changed",
		`^it says an objective is added or taken away only when changing a task$`: "only when changing a task",
		`^it says what is in the way and how to take it away$`:                    "take them away first: task ",
		`^it says nothing would ever be able to start$`:                           "nothing would ever be able to start",
		`^it says a task cannot be moved behind itself$`:                          "cannot be moved behind itself",
		`^it says a task is moved inside its own column$`:                         "moved inside its own column",
		`^it says where a task can be moved$`:                                     "behind another",
		`^it says a task that is done cannot be in focus$`:                        "done cannot be in focus",
	} {
		wanted := want
		sc.Step(step, func() error {
			return test.WantContains(w.Out.Stderr, wanted, "the reason")
		})
	}

	sc.Step(`^it shows the number it could not find$`, func() error {
		asked := numbersIn.FindAllString(w.Line, -1)
		if len(asked) == 0 {
			return fmt.Errorf("the question named no number: %q", w.Line)
		}
		for _, number := range asked {
			if strings.Contains(w.Out.Stderr, number) {
				return nil
			}
		}
		return fmt.Errorf("the reason quotes none of the numbers in %q:\n%s", w.Line, w.Out.Stderr)
	})

	sc.Step(`^it shows the title I typed$`, func() error {
		return test.WantContains(w.Out.Stderr, "got ", "the reason")
	})

	sc.Step(`^task (\d+) is in the column "([^"]*)"$`, func(n int, column string) error {
		w.Run(fmt.Sprintf(`cairn task show %d --json`, n))
		shown := w.Out
		var said struct {
			State string `json:"state"`
		}
		if err := json.Unmarshal([]byte(shown.Stdout), &said); err != nil {
			return fmt.Errorf("reading what task %d carries: %w", n, err)
		}
		if said.State != column {
			return fmt.Errorf("task %d is in %q, not %q", n, said.State, column)
		}
		return nil
	})

	sc.Step(`^task (\d+) is archived$`, func(n int) error { return w.archived(n, true) })
	sc.Step(`^task (\d+) is not archived$`, func(n int) error { return w.archived(n, false) })

	sc.Step(`^it says what it needs to go on$`, func() error {
		return test.WantContains(w.Out.Stderr, "narrow", "the refusal")
	})

	sc.Step(`^it says the task is on the board$`, func() error {
		return test.WantContains(w.Out.Stderr, "on the board", "the refusal")
	})

	sc.Step(`^it says there is no task$`, func() error {
		return test.WantContains(w.Out.Stdout, "there is no task", "the answer")
	})

	sc.Step(`^it says nothing is in focus$`, func() error {
		return test.WantContains(w.Out.Stdout, "nothing is in focus", "the answer")
	})

	sc.Step(`^it says task (\d+) was deleted$`, func(n int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("task %d was deleted", n), "the answer")
	})

	sc.Step(`^it reports (?:a|the) task titled "([^"]*)"$`, func(title string) error {
		return test.WantContains(w.heading(), title, "the heading")
	})

	sc.Step(`^it numbers the task (\d+)$`, func(n int) error {
		return test.WantContains(w.heading(), "#"+strconv.Itoa(n), "the heading")
	})

	sc.Step(`^the task carries the (description|deadline|estimate|priority|column|time spent) "([^"]*)"$`,
		func(attribute, want string) error {
			key := map[string]string{
				"description": "description", "deadline": "deadline",
				"estimate": "estimate", "priority": "priority",
				"column": "state", "time spent": "time spent",
			}[attribute]
			if got := w.field(key); got != want {
				return fmt.Errorf("%s = %q, want %q", key, got, want)
			}
			return nil
		})

	sc.Step(`^the task carries the labels "([^"]*)" and "([^"]*)"$`, func(first, second string) error {
		want := first + ", " + second
		if got := w.field("labels"); got != want {
			return fmt.Errorf("labels = %q, want %q", got, want)
		}
		return nil
	})

	sc.Step(`^the task carries exactly (\d+) labels?$`, func(n int) error {
		labels := w.field("labels")
		got := 0
		if labels != "" {
			got = len(strings.Split(labels, ", "))
		}
		if got != n {
			return fmt.Errorf("%d labels (%q), want %d", got, labels, n)
		}
		return nil
	})

	sc.Step("^`([^`]*)` says it serves objectives (\\d+) and (\\d+)$",
		func(line string, first, second int) error {
			return w.servesExactly(line, first, second)
		})

	sc.Step("^`([^`]*)` says it serves objective (\\d+) alone$",
		func(line string, only int) error {
			return w.servesExactly(line, only)
		})

	sc.Step(`^the task is in the "([^"]*)" column$`, func(want string) error {
		if got := w.field("state"); got != want {
			return fmt.Errorf("state = %q, want %q", got, want)
		}
		return nil
	})

	sc.Step(`^the task has no (description|labels|deadline|estimate|time spent)$`, func(attribute string) error {
		key := map[string]string{
			"description": "description", "labels": "labels", "deadline": "deadline",
			"estimate": "estimate", "time spent": "time spent",
		}[attribute]
		if got := w.field(key); got != "" {
			return fmt.Errorf("%s = %q, want none", key, got)
		}
		return nil
	})

	sc.Step(`^it lists (\d+) tasks?$`, func(n int) error {
		return test.CountsTasks(w.Out.Stdout, n)
	})

	sc.Step("^`([^`]*)` lists task (\\d+) titled \"([^\"]*)\"$", func(line string, n int, title string) error {
		w.Run(line)
		return test.ListsTask(w.Out.Stdout, n, title)
	})

	sc.Step(`^it lists task (\d+) titled "([^"]*)"$`, func(n int, title string) error {
		return test.ListsTask(w.Out.Stdout, n, title)
	})

	sc.Step("^`([^`]*)` says there is no task$", func(line string) error {
		w.Run(line)
		return test.WantContains(w.Out.Stdout, "there is no task", "the answer")
	})

	sc.Step(`^it writes JSON to standard output$`, func() error {
		var any any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &any); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		return nil
	})

	sc.Step("^I run `([^`]*)` with \"([^\"]*)\" and \"([^\"]*)\" on standard input$", func(line, first, second string) error {
		var out, errs bytes.Buffer
		stdin := strings.NewReader(first + "\n\n" + second + "\n")
		code := MainIn(w.context(), w.Home(), test.SplitArgs(line)[1:], stdin, &out, &errs)
		w.Line, w.Out = line, test.Result{Stdout: out.String(), Stderr: errs.String(), Code: code}
		return nil
	})

	sc.Step(`^it carries the title "([^"]*)"$`, func(title string) error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["title"] != title {
			return fmt.Errorf("title = %v, want %q", doc["title"], title)
		}
		return nil
	})

	sc.Step(`^it carries an identifier$`, func() error {
		var doc map[string]any
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return err
		}
		if doc["number"] == nil {
			return fmt.Errorf("no identifier in %s", w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says no task in it is in focus$`, func() error {
		_, envelope, err := listedUnder(w.Out.Stdout, "tasks")
		if err != nil {
			return err
		}
		held := envelope["in_focus"]
		if held == nil {
			return fmt.Errorf("the listing says nothing about the focus, and it was asked")
		}
		if numbers, ok := held.([]any); !ok || len(numbers) != 0 {
			return fmt.Errorf("it says %v is in focus, and the answer does not carry it", held)
		}
		return nil
	})

	sc.Step(`^it says (\d+) of (\d+) (tasks|steps) are done$`,
		func(done, total int, kind string) error {
			return test.WantContains(w.Out.Stdout,
				fmt.Sprintf("%d of %d %s are done", done, total, kind), "the answer")
		})

}

func (w *world) archived(number int, want bool) error {
	w.Run(fmt.Sprintf("cairn task show %d --json", number))
	var said struct {
		Archived bool `json:"archived"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
		return fmt.Errorf("reading what task %d carries: %w", number, err)
	}
	if said.Archived != want {
		return fmt.Errorf("task %d archived = %v, want %v", number, said.Archived, want)
	}
	return nil
}
