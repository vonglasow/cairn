package engine

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"

	"cairn/internal/domain"
)

// ErrNoSuchLabel is returned when a label named is carried by no task and no objective.
var ErrNoSuchLabel = errors.New("nothing carries that label")

// LabelCarried is one label and how many tasks and objectives carry it.
type LabelCarried struct {
	Label      string `json:"label"`
	Tasks      int    `json:"tasks"`
	Objectives int    `json:"objectives"`
}

// Labels is how the labels answer: every one, in alphabetical order, with what carries it.
type Labels struct {
	Labels []LabelCarried `json:"labels"`
}

type carriers struct {
	tasks, objectives []int
}

func (e *Engine) everyLabel() (map[string]carriers, error) {
	every := domain.Narrowing{}
	every.WithArchived()
	tasks, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	goals, err := e.listObjectives()
	if err != nil {
		return nil, err
	}
	held := map[string]carriers{}
	for _, t := range tasks {
		f := factsOf(t)
		for _, label := range f.Labels {
			c := held[label]
			c.tasks = append(c.tasks, f.Number)
			held[label] = c
		}
	}
	for _, g := range goals {
		f := factsOfObjective(g)
		for _, label := range f.Labels {
			c := held[label]
			c.objectives = append(c.objectives, f.Number)
			held[label] = c
		}
	}
	return held, nil
}

func (e *Engine) labelsCarried() ([]LabelCarried, error) {
	held, err := e.everyLabel()
	if err != nil {
		return nil, err
	}
	out := []LabelCarried{}
	for _, label := range slices.Sorted(maps.Keys(held)) {
		out = append(out, LabelCarried{Label: label, Tasks: len(held[label].tasks), Objectives: len(held[label].objectives)})
	}
	return out, nil
}

func (e *Engine) carriersOf(given string) (string, carriers, error) {
	written, err := domain.ParseLabels([]string{given})
	if err != nil {
		return "", carriers{}, err
	}
	held, err := e.everyLabel()
	if err != nil {
		return "", carriers{}, err
	}
	label := written[0]
	found, there := held[label]
	if !there {
		return "", carriers{}, fmt.Errorf("%q: %w", label, ErrNoSuchLabel)
	}
	return label, found, nil
}

func (e *Engine) relabel(c carriers, change []string) error {
	for _, n := range c.tasks {
		if _, err := e.edit(n, Changes{Labels: &change}); err != nil {
			return err
		}
	}
	for _, n := range c.objectives {
		if _, err := e.editObjective(n, ObjectiveChanges{Labels: &change}); err != nil {
			return err
		}
	}
	return nil
}

func listedLabels(on On) (Answer, error) {
	listed, err := on.Engine.labelsCarried()
	if err != nil {
		return Answer{}, err
	}
	return Answer{Labels: &listed}, nil
}

// LabelList lists every label, with how many tasks and objectives carry it.
var LabelList = Operation{
	Thing:   "label",
	Verb:    "list",
	Short:   "List every label, with what carries it",
	Does:    "List every label some task or objective carries, in alphabetical order, with how many tasks and how many objectives carry it, archived tasks included.",
	Example: "cairn label list",
	Level:   Read,
	Needs:   AStore,
	Answers: Labels{},
	Run:     func(on On, _ Input) (Answer, error) { return listedLabels(on) },
}

// LabelDelete takes a label off every task and every objective carrying it.
var LabelDelete = Operation{
	Thing:   "label",
	Verb:    "delete",
	Short:   "Take a label off everything carrying it",
	Does:    "Take a label off every task and every objective carrying it, each change written in its task's trace. Answers the labels left.",
	Example: "cairn label delete sp26-20 --yes",
	Level:   Full,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to take off everything"},
	},
	Run: func(on On, in Input) (Answer, error) {
		label, found, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.relabel(found, []string{"-" + label}); err != nil {
			return Answer{}, err
		}
		return listedLabels(on)
	},
}

// LabelRename writes another name for a label on every task and every objective carrying it.
var LabelRename = Operation{
	Thing:   "label",
	Verb:    "rename",
	Short:   "Rename a label on everything carrying it",
	Does:    "Rename a label on every task and every objective carrying it, the new name written the one way a label is; renamed to one already there, the two become one. Answers the labels there are.",
	Example: "cairn label rename sp26-20 sprint-20",
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to rename"},
		{Name: "to", Kind: Text, Positional: true, Required: true, Means: "the name it takes"},
	},
	Run: func(on On, in Input) (Answer, error) {
		label, found, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		to, err := domain.ParseLabels([]string{in.Given["to"]})
		if err != nil {
			return Answer{}, err
		}
		if to[0] != label {
			if err := on.Engine.relabel(found, []string{"-" + label, "+" + to[0]}); err != nil {
				return Answer{}, err
			}
		}
		return listedLabels(on)
	},
}

func labelsText(listed []LabelCarried) string {
	if len(listed) == 0 {
		return "no label is carried\n"
	}
	var out strings.Builder
	tw := tabwriter.NewWriter(&out, 0, 0, 2, ' ', 0)
	for _, one := range listed {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", one.Label, manyOf(one.Tasks, "task", "tasks"), manyOf(one.Objectives, "objective", "objectives"))
	}
	_ = tw.Flush()
	return out.String()
}
