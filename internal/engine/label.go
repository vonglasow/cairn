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
	Colour     string `json:"colour"`
	Scope      string `json:"scope" jsonschema:"what it may be put on: tasks, objectives or both"`
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
	kept, err := e.store.KeptLabels()
	if err != nil {
		return nil, err
	}
	for _, label := range kept {
		held[label] = carriers{}
	}
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
	chosen, err := e.store.LabelColours()
	if err != nil {
		return nil, err
	}
	scopes, err := e.scopes()
	if err != nil {
		return nil, err
	}
	out := []LabelCarried{}
	for _, label := range slices.Sorted(maps.Keys(held)) {
		colour, ok := chosen[label]
		if !ok {
			colour = domain.ColourOf(label)
		}
		out = append(out, LabelCarried{Label: label, Tasks: len(held[label].tasks), Objectives: len(held[label].objectives), Colour: colour, Scope: string(scopes.of(label))})
	}
	return out, nil
}

type scoped map[string]string

func (s scoped) of(label string) domain.Scope {
	if scope, narrowed := s[label]; narrowed {
		return domain.Scope(scope)
	}
	return domain.ForBoth
}

func (e *Engine) scopes() (scoped, error) {
	said, err := e.store.LabelScopes()
	return scoped(said), err
}

func (e *Engine) inScope(was, now []string, on domain.Bearer) error {
	scopes, err := e.scopes()
	if err != nil {
		return err
	}
	for _, label := range now {
		if slices.Contains(was, label) {
			continue
		}
		if err := scopes.of(label).Refuses(label, on); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) scope(given, named string) error {
	label, found, err := e.carriersOf(given)
	if err != nil {
		return err
	}
	scope, err := domain.ParseScope(named)
	if err != nil {
		return err
	}
	if !scope.Admits(domain.OnTasks) && len(found.tasks) > 0 {
		return fmt.Errorf("%q is carried by %s; %w", label, manyOf(len(found.tasks), "task", "tasks"), domain.ErrOutOfScope)
	}
	if !scope.Admits(domain.OnObjectives) && len(found.objectives) > 0 {
		return fmt.Errorf("%q is carried by %s; %w", label, manyOf(len(found.objectives), "objective", "objectives"), domain.ErrOutOfScope)
	}
	return e.store.ScopeLabel(label, string(scope))
}

func oneLabel(given string) (string, error) {
	written, err := domain.ParseLabels([]string{given})
	if err != nil {
		return "", err
	}
	if len(written) == 0 {
		return "", domain.ErrLabelRequired
	}
	return written[0], nil
}

func (e *Engine) carriersOf(given string) (string, carriers, error) {
	label, err := oneLabel(given)
	if err != nil {
		return "", carriers{}, err
	}
	held, err := e.everyLabel()
	if err != nil {
		return "", carriers{}, err
	}
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

func listedLabels(on On) (Answer, error) { return offeredLabels(on, "") }

func offeredLabels(on On, onto string) (Answer, error) {
	listed, err := on.Engine.labelsCarried()
	if err != nil {
		return Answer{}, err
	}
	if onto != "" {
		listed = slices.DeleteFunc(listed, func(l LabelCarried) bool { return !domain.Scope(l.Scope).Admits(domain.Bearer(onto)) })
	}
	return Answer{Labels: &listed}, nil
}

var scopeArgument = Argument{Name: "scope", Kind: Text, OneOf: oneOf(domain.Scopes), Outside: domain.ErrNoSuchScope, Means: "what it may be put on: tasks, objectives, or both, which a label nobody scoped is for"}

// LabelList lists every label, with how many tasks and objectives carry it.
var LabelList = Operation{
	Thing: "label",
	Verb:  "list",
	Short: "List every label, with what carries it",
	Does:  "List every label some task or objective carries, and every one made on its own, in alphabetical order, with how many tasks and how many objectives carry it, archived tasks included, and what it may be put on. Given what the labels are for, only those that may be put on it.",
	Example: `cairn label list
cairn label list --for objectives`,
	Level:   Read,
	Needs:   AStore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "for", Kind: Text, OneOf: oneOf([]domain.Bearer{domain.OnTasks, domain.OnObjectives}), Outside: domain.ErrNoSuchScope, Means: "keep only the labels that may be put on tasks, or on objectives"},
	},
	Run: func(on On, in Input) (Answer, error) { return offeredLabels(on, in.Given["for"]) },
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
		if err := on.Engine.store.ForgetLabel(label); err != nil {
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
		to, err := oneLabel(in.Given["to"])
		if err != nil {
			return Answer{}, err
		}
		if to != label {
			if err := on.Engine.relabel(found, []string{"-" + label, "+" + to}); err != nil {
				return Answer{}, err
			}
			if err := on.Engine.store.RenameKeptLabel(label, to); err != nil {
				return Answer{}, err
			}
		}
		return listedLabels(on)
	},
}

// LabelNew makes a label nothing carries yet, kept until it is deleted.
var LabelNew = Operation{
	Thing: "label",
	Verb:  "new",
	Short: "Make a label nothing carries yet",
	Does:  "Make a label before any task or objective carries it, written the one way a label is, for tasks, objectives or both; it is listed and offered until it is deleted, carried or not. Answers the labels there are.",
	Example: `cairn label new release-train
cairn label new area --scope objectives`,
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to make"},
		scopeArgument,
	},
	Run: func(on On, in Input) (Answer, error) {
		label, err := oneLabel(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.store.KeepLabel(label); err != nil {
			return Answer{}, err
		}
		if scope, named := in.Given["scope"]; named {
			if err := on.Engine.scope(label, scope); err != nil {
				return Answer{}, err
			}
		}
		return listedLabels(on)
	},
}

// LabelScope says what a label may be put on: tasks, objectives, or both.
var LabelScope = Operation{
	Thing:   "label",
	Verb:    "scope",
	Short:   "Say what a label may be put on",
	Does:    "Say what a label may be put on: tasks, objectives, or both, which a label nobody scoped is for. Narrowed to one, it is refused while the other carries it, and putting it on the other is refused after. Answers the labels there are.",
	Example: "cairn label scope area objectives",
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to scope"},
		{Name: "scope", Kind: Text, Positional: true, Required: true, OneOf: scopeArgument.OneOf, Outside: domain.ErrNoSuchScope, Means: scopeArgument.Means},
	},
	Run: func(on On, in Input) (Answer, error) {
		if err := on.Engine.scope(in.Given["label"], in.Given["scope"]); err != nil {
			return Answer{}, err
		}
		return listedLabels(on)
	},
}

// LabelColour chooses the colour a label wears, in place of the one read off its name.
var LabelColour = Operation{
	Thing:   "label",
	Verb:    "colour",
	Short:   "Choose the colour a label wears",
	Does:    "Choose the colour a label wears wherever it is drawn, one of rosewater, flamingo, pink, mauve, red, maroon, peach, yellow, green, teal, sky, sapphire, blue and lavender; a label nobody chose one for wears one read off its name. Answers the labels there are.",
	Example: "cairn label colour sp26-20 peach",
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Labels{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to colour"},
		{Name: "colour", Kind: Text, Positional: true, Required: true, Means: "the colour it wears"},
	},
	Run: func(on On, in Input) (Answer, error) {
		label, _, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		colour, err := domain.ParseColour(in.Given["colour"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.store.ColourLabel(label, colour); err != nil {
			return Answer{}, err
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
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\tfor %s\n", one.Label, manyOf(one.Tasks, "task", "tasks"), manyOf(one.Objectives, "objective", "objectives"), one.Scope)
	}
	_ = tw.Flush()
	return out.String()
}
