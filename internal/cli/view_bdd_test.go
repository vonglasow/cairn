package cli

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func initView(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerTimesheet(sc)
		w.registerView(sc)
	}
}

func TestAcceptance_View(t *testing.T) {
	runAcceptance(t, "view.feature", initView(t), "")
}

type listedWithAView struct {
	Tasks []struct {
		Number int `json:"number"`
	} `json:"tasks"`
	Gone    []string `json:"gone"`
	Outside []struct {
		Number int `json:"number"`
		Why    []struct {
			Is string `json:"is"`
			Of int    `json:"of"`
		} `json:"why"`
	} `json:"outside"`
}

func (w *world) listedWithAView() (listedWithAView, error) {
	var listed listedWithAView
	err := json.Unmarshal([]byte(w.Out.Stdout), &listed)
	return listed, err
}

func (w *world) registerView(sc *godog.ScenarioContext) {
	sc.Step(`^it says on standard error "([^"]*)"$`, w.itSaysOnStandardError)
	sc.Step(`^the JSON lists the tasks ([\d, ]+)$`, func(numbers string) error {
		listed, err := w.listedWithAView()
		if err != nil {
			return err
		}
		var got, want []string
		for _, one := range listed.Tasks {
			got = append(got, strconv.Itoa(one.Number))
		}
		for _, one := range strings.Split(numbers, ",") {
			want = append(want, strings.TrimSpace(one))
		}
		if !slices.Equal(got, want) {
			return fmt.Errorf("the tasks listed are %v, want %v:\n%s", got, want, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON lists no task$`, func() error {
		listed, err := w.listedWithAView()
		if err != nil {
			return err
		}
		if len(listed.Tasks) != 0 {
			return fmt.Errorf("it lists %d tasks, want none:\n%s", len(listed.Tasks), w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON says the view lost "([^"]*)"$`, func(gone string) error {
		listed, err := w.listedWithAView()
		if err != nil {
			return err
		}
		if !slices.Contains(listed.Gone, gone) {
			return fmt.Errorf("gone reads %v, want %q:\n%s", listed.Gone, gone, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the JSON names task (\d+) outside, because it "([^"]*)" task (\d+) and is "([^"]*)" task (\d+)$`,
		func(n int, first string, firstOf int, second string, secondOf int) error {
			listed, err := w.listedWithAView()
			if err != nil {
				return err
			}
			if len(listed.Outside) != 1 || listed.Outside[0].Number != n {
				return fmt.Errorf("outside reads %+v, want task %d alone:\n%s", listed.Outside, n, w.Out.Stdout)
			}
			why := listed.Outside[0].Why
			if len(why) != 2 || why[0].Is != first || why[0].Of != firstOf || why[1].Is != second || why[1].Of != secondOf {
				return fmt.Errorf("task %d is outside because %+v", n, why)
			}
			return nil
		})
	sc.Step(`^the view "([^"]*)" is listed in place (\d+), holding "([^"]*)", keeping (\d+)$`, func(name string, place int, holds string, keeps int) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 3 || fields[0] != strconv.Itoa(place) {
				continue
			}
			if fields[1] != name || !strings.Contains(line, holds) || fields[len(fields)-1] != strconv.Itoa(keeps) {
				return fmt.Errorf("place %d reads %q, want %s holding %q keeping %d", place, line, name, holds, keeps)
			}
			return nil
		}
		return fmt.Errorf("no view is listed in place %d:\n%s", place, w.Out.Stdout)
	})
}
