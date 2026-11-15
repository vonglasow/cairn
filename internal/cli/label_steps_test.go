package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) registerLabel(sc *godog.ScenarioContext) {
	sc.Step(`^the label "([^"]*)" is carried by (\d+) tasks? and (\d+) objectives?$`, func(label string, tasks, objectives int) error {
		for _, l := range strings.Split(w.Out.Stdout, "\n") {
			fields := strings.Fields(l)
			if len(fields) == 0 || fields[0] != label {
				continue
			}
			want := []string{strconv.Itoa(tasks), strconv.Itoa(objectives)}
			if got := countsIn(fields[1:]); strings.Join(got, " ") != strings.Join(want, " ") {
				return fmt.Errorf("the label %q reads %q, want %v tasks then objectives", label, l, want)
			}
			return nil
		}
		return fmt.Errorf("no label %q is listed:\n%s", label, w.Out.Stdout)
	})
	sc.Step(`^the JSON gives the label "([^"]*)" the colour "([^"]*)"$`, func(label, colour string) error {
		var listed struct {
			Labels []struct {
				Label  string `json:"label"`
				Colour string `json:"colour"`
			} `json:"labels"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &listed); err != nil {
			return err
		}
		for _, one := range listed.Labels {
			if one.Label == label {
				if one.Colour != colour {
					return fmt.Errorf("the label %q wears %q, not %q", label, one.Colour, colour)
				}
				return nil
			}
		}
		return fmt.Errorf("the JSON lists no label %q:\n%s", label, w.Out.Stdout)
	})
	sc.Step(`^the JSON gives the label "([^"]*)" the scope "([^"]*)"$`, func(label, scope string) error {
		var listed struct {
			Labels []struct {
				Label string `json:"label"`
				Scope string `json:"scope"`
			} `json:"labels"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &listed); err != nil {
			return err
		}
		for _, one := range listed.Labels {
			if one.Label == label {
				if one.Scope != scope {
					return fmt.Errorf("the label %q is for %q, not %q", label, one.Scope, scope)
				}
				return nil
			}
		}
		return fmt.Errorf("the JSON lists no label %q:\n%s", label, w.Out.Stdout)
	})
}

func countsIn(fields []string) []string {
	var out []string
	for _, f := range fields {
		if _, err := strconv.Atoi(f); err == nil {
			out = append(out, f)
		}
	}
	return out
}
