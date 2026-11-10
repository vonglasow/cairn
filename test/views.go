package test

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
)

// IKeptTheViewHolding keeps a view holding one filter: a label or a priority.
func (g Givens) IKeptTheViewHolding(name, filter, value string) error {
	given, listed := map[string]string{"name": name}, map[string][]string{}
	if filter == "label" {
		listed["label"] = []string{value}
	} else {
		given[filter] = value
	}
	_, err := g.ask(engine.ViewNew, given, listed)
	return err
}

// ILinkedTaskTo links a task to another.
func (g Givens) ILinkedTaskTo(number, other int) error {
	return g.editTask(number, nil, map[string][]string{"link": {strconv.Itoa(other)}})
}

// IDeletedTheLabel takes a label off everything.
func (g Givens) IDeletedTheLabel(label string) error {
	_, err := g.ask(engine.LabelDelete, map[string]string{"label": label}, nil)
	return err
}

// View is a view as the store keeps it, read by its name.
func (g Givens) View(name string) (engine.View, error) {
	views, err := g.Views()
	if err != nil {
		return engine.View{}, err
	}
	for _, v := range views {
		if strings.EqualFold(v.Name, name) {
			return v, nil
		}
	}
	return engine.View{}, fmt.Errorf("no view is named %q among %v", name, views)
}

// Views is every view the store keeps, in its place.
func (g Givens) Views() ([]engine.View, error) {
	listed, err := g.ask(engine.ViewList, nil, nil)
	if err != nil || listed.ViewsKept == nil {
		return nil, err
	}
	return listed.ViewsKept.Views, nil
}

// ViewHolds checks a view holds one value of a filter: a label or a priority.
func (g Givens) ViewHolds(name, filter, value string) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	held := map[string]bool{"priority": v.Holds.Priority == value, "label": slices.Contains(v.Holds.Label, value)}
	if !held[filter] {
		return fmt.Errorf("the view %s holds %+v, not the %s %q", name, v.Holds, filter, value)
	}
	return nil
}

// ViewHoldsObjective checks a view holds an objective.
func (g Givens) ViewHoldsObjective(name string, n int) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	if !slices.Contains(v.Holds.Objective, n) {
		return fmt.Errorf("the view %s holds the objectives %v, not %d", name, v.Holds.Objective, n)
	}
	return nil
}

// ViewHoldsNoLabel checks a view holds no label.
func (g Givens) ViewHoldsNoLabel(name string) error {
	v, err := g.View(name)
	if err != nil {
		return err
	}
	if len(v.Holds.Label) > 0 {
		return fmt.Errorf("the view %s holds the labels %v", name, v.Holds.Label)
	}
	return nil
}

// ViewsRead checks the views, in their places, are named as listed.
func (g Givens) ViewsRead(list string) error {
	views, err := g.Views()
	if err != nil {
		return err
	}
	var names []string
	for _, v := range views {
		names = append(names, strconv.Quote(v.Name))
	}
	if got := strings.Join(names, ", "); got != list {
		return fmt.Errorf("the views read %s, not %s", got, list)
	}
	return nil
}

func (g Givens) registerViews(sc *godog.ScenarioContext) {
	sc.Step(`^I kept the view "([^"]*)" holding the (label|priority) "([^"]*)"$`, g.IKeptTheViewHolding)
	sc.Step(`^I linked task (\d+) to task (\d+)$`, g.ILinkedTaskTo)
	sc.Step(`^I deleted the label "([^"]*)"$`, g.IDeletedTheLabel)
	sc.Then(`^the view "([^"]*)" holds the (label|priority) "([^"]*)"$`, g.ViewHolds)
	sc.Then(`^the view "([^"]*)" holds objective (\d+)$`, g.ViewHoldsObjective)
	sc.Then(`^the view "([^"]*)" holds no label$`, g.ViewHoldsNoLabel)
	sc.Then(`^the views read (".*")$`, g.ViewsRead)
}
