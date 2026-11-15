package label

import (
	"errors"
)

// ErrNoSuchLabel is returned when a label named is carried by no task and no objective.
var ErrNoSuchLabel = errors.New("nothing carries that label")

// Carried is one label and how many tasks and objectives carry it.
type Carried struct {
	Label      string `json:"label"`
	Tasks      int    `json:"tasks"`
	Objectives int    `json:"objectives"`
	Colour     string `json:"colour"`
	Scope      string `json:"scope" jsonschema:"what it may be put on: tasks, objectives or both"`
}

// List is how the labels answer: every one, in alphabetical order, with what carries it.
type List struct {
	Labels   []Carried `json:"labels"`
	Warnings []string  `json:"warnings,omitempty" jsonschema:"what a deletion did and should be known: how many views held the label, which narrows nothing there now; left out when there is nothing to say"`
}

// Carriers is the tasks and the objectives carrying a label.
type Carriers struct {
	Tasks, Objectives []int
}

// Scoped is the scope each label is narrowed to.
type Scoped map[string]string

// Of is the scope of a label: the one it is narrowed to, or both.
func (s Scoped) Of(name string) Scope {
	if scope, narrowed := s[name]; narrowed {
		return Scope(scope)
	}
	return ForBoth
}

// OneLabel reads one label name.
func OneLabel(given string) (string, error) {
	written, err := ParseLabels([]string{given})
	if err != nil {
		return "", err
	}
	if len(written) == 0 {
		return "", ErrLabelRequired
	}
	return written[0], nil
}
