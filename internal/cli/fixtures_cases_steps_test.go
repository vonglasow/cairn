package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

type objectiveListed struct {
	Number   int      `json:"number"`
	Labels   []string `json:"labels"`
	Parent   *int     `json:"parent"`
	Children []int    `json:"children"`
}

type seriesListed struct {
	Says    string `json:"says"`
	Ended   bool   `json:"ended"`
	Skipped int    `json:"skipped"`
}

func (w *world) answered(inTheStore func(string) error, line string, into any) error {
	if err := inTheStore(line); err != nil {
		return err
	}
	return json.Unmarshal([]byte(w.Out.Stdout), into)
}

func (w *world) registerFixturesCases(sc *godog.ScenarioContext, inTheStore func(string) error) {
	objectives := func() ([]objectiveListed, error) {
		var listed struct {
			Objectives []objectiveListed `json:"objectives"`
		}
		err := w.answered(inTheStore, "cairn objective list --json", &listed)
		return listed.Objectives, err
	}
	tasks := func() (serving, none int, err error) {
		var listed struct {
			Tasks []struct {
				Objectives []int `json:"objectives"`
			} `json:"tasks"`
		}
		err = w.answered(inTheStore, "cairn task list --with-archived --json --fields objectives", &listed)
		for _, t := range listed.Tasks {
			if len(t.Objectives) > 0 {
				serving++
			} else {
				none++
			}
		}
		return serving, none, err
	}
	series := func() ([]seriesListed, error) {
		var listed struct {
			Series []seriesListed `json:"series"`
		}
		err := w.answered(inTheStore, "cairn series list --json", &listed)
		return listed.Series, err
	}
	anySeries := func(what string, holds func(seriesListed) bool) error {
		listed, err := series()
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(listed, holds) {
			return fmt.Errorf("no series %s:\n%s", what, w.Out.Stdout)
		}
		return nil
	}
	sc.Step(`^most tasks in the store it made serve an objective$`, func() error {
		serving, none, err := tasks()
		if err != nil {
			return err
		}
		if serving < 3*none {
			return fmt.Errorf("%d tasks serve an objective and %d none; want three in four at least", serving, none)
		}
		return nil
	})
	sc.Step(`^the store it made has tasks serving no objective$`, func() error {
		_, none, err := tasks()
		if err != nil {
			return err
		}
		if none < 2 {
			return fmt.Errorf("%d tasks serve no objective; want a few", none)
		}
		return nil
	})
	sc.Step(`^the store it made has an objective labelled "([^"]*)" holding sub-objectives labelled "([^"]*)"$`, func(parent, child string) error {
		listed, err := objectives()
		if err != nil {
			return err
		}
		labels := map[int][]string{}
		for _, o := range listed {
			labels[o.Number] = o.Labels
		}
		for _, o := range listed {
			if slices.Contains(o.Labels, parent) && len(o.Children) > 1 && !slices.ContainsFunc(o.Children, func(n int) bool { return !slices.Contains(labels[n], child) }) {
				return nil
			}
		}
		return fmt.Errorf("no objective labelled %s holds sub-objectives all labelled %s:\n%s", parent, child, w.Out.Stdout)
	})
	sc.Step(`^the store it made has a label for tasks only, one for objectives only, and one carried by both$`, func() error {
		var listed struct {
			Labels []struct {
				Scope      string `json:"scope"`
				Tasks      int    `json:"tasks"`
				Objectives int    `json:"objectives"`
			} `json:"labels"`
		}
		if err := w.answered(inTheStore, "cairn label list --json", &listed); err != nil {
			return err
		}
		scopes := map[string]bool{}
		for _, l := range listed.Labels {
			scopes[l.Scope] = scopes[l.Scope] || l.Tasks+l.Objectives > 0
			scopes["carried by both"] = scopes["carried by both"] || l.Tasks > 0 && l.Objectives > 0
		}
		for _, want := range []string{"tasks", "objectives", "carried by both"} {
			if !scopes[want] {
				return fmt.Errorf("no label carried %s:\n%s", want, w.Out.Stdout)
			}
		}
		return nil
	})
	sc.Step(`^the store it made has an objective proposed as reached$`, func() error {
		listed, err := objectives()
		if err != nil {
			return err
		}
		for _, o := range listed {
			if len(o.Children) == 0 {
				continue
			}
			var shown struct {
				Proposed *int `json:"proposed_reached"`
			}
			if err := w.answered(inTheStore, "cairn objective show "+strconv.Itoa(o.Number)+" --json", &shown); err != nil {
				return err
			}
			if shown.Proposed != nil {
				return nil
			}
		}
		return fmt.Errorf("no objective is proposed as reached")
	})
	sc.Step(`^the store it made has a view naming a sub-objective$`, func() error {
		listed, err := objectives()
		if err != nil {
			return err
		}
		under := map[int]bool{}
		for _, o := range listed {
			under[o.Number] = o.Parent != nil
		}
		var views struct {
			Views []struct {
				Holds struct {
					Objective []int `json:"objective"`
				} `json:"holds"`
			} `json:"views"`
		}
		if err := w.answered(inTheStore, "cairn view list --json", &views); err != nil {
			return err
		}
		for _, v := range views.Views {
			if slices.ContainsFunc(v.Holds.Objective, func(n int) bool { return under[n] }) {
				return nil
			}
		}
		return fmt.Errorf("no view names a sub-objective:\n%s", w.Out.Stdout)
	})
	sc.Step(`^the store it made has a series counted after done$`, func() error {
		return anySeries("counted after done", func(s seriesListed) bool { return strings.Contains(s.Says, "after the last is done") })
	})
	sc.Step(`^the store it made has a series ended by its count$`, func() error {
		return anySeries("ended by its count", func(s seriesListed) bool { return s.Ended && strings.Contains(s.Says, "times") })
	})
	sc.Step(`^the store it made has a series with an occurrence skipped$`, func() error {
		return anySeries("with an occurrence skipped", func(s seriesListed) bool { return s.Skipped > 0 })
	})
	sc.Step(`^the configuration of the store it made says its colour is set in the store's settings$`, func() error {
		if err := inTheStore("cairn config show"); err != nil {
			return err
		}
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.HasPrefix(line, "colour") && strings.Contains(line, "set in the store's settings") {
				return nil
			}
		}
		return fmt.Errorf("no colour set in the store's settings:\n%s", w.Out.Stdout)
	})
}
