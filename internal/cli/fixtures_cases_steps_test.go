package cli

import (
	"encoding/json"
	"fmt"
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
