package engine

import (
	"cairn/internal/engine/filter"
	"cairn/internal/engine/label"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/tabwriter"
)

func (e *Engine) everyLabel() (map[string]label.Carriers, error) {
	every := filter.Narrowing{}
	every.WithArchived()
	tasks, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	goals, err := e.listObjectives()
	if err != nil {
		return nil, err
	}
	held := map[string]label.Carriers{}
	kept, err := e.store.KeptLabels()
	if err != nil {
		return nil, err
	}
	for _, name := range kept {
		held[name] = label.Carriers{}
	}
	for _, t := range tasks {
		f := t.Carried()
		for _, name := range f.Labels {
			c := held[name]
			c.Tasks = append(c.Tasks, f.Number)
			held[name] = c
		}
	}
	for _, g := range goals {
		f := g.Carried()
		for _, name := range f.Labels {
			c := held[name]
			c.Objectives = append(c.Objectives, f.Number)
			held[name] = c
		}
	}
	return held, nil
}

func (e *Engine) labelsCarried() ([]label.Carried, error) {
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
	out := []label.Carried{}
	for _, name := range slices.Sorted(maps.Keys(held)) {
		colour, ok := chosen[name]
		if !ok {
			colour = label.ColourOf(name)
		}
		out = append(out, label.Carried{Label: name, Tasks: len(held[name].Tasks), Objectives: len(held[name].Objectives), Colour: colour, Scope: string(scopes.Of(name))})
	}
	return out, nil
}

func (e *Engine) scopes() (label.Scoped, error) {
	said, err := e.store.LabelScopes()
	return label.Scoped(said), err
}

func (e *Engine) inScope(was, now []string, on label.Bearer) error {
	scopes, err := e.scopes()
	if err != nil {
		return err
	}
	for _, name := range now {
		if slices.Contains(was, name) {
			continue
		}
		if err := scopes.Of(name).Refuses(name, on); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) scope(given, named string) error {
	name, found, err := e.carriersOf(given)
	if err != nil {
		return err
	}
	scope, err := label.ParseScope(named)
	if err != nil {
		return err
	}
	if !scope.Admits(label.OnTasks) && len(found.Tasks) > 0 {
		return fmt.Errorf("%q is carried by %s; %w", name, manyOf(len(found.Tasks), "task", "tasks"), label.ErrOutOfScope)
	}
	if !scope.Admits(label.OnObjectives) && len(found.Objectives) > 0 {
		return fmt.Errorf("%q is carried by %s; %w", name, manyOf(len(found.Objectives), "objective", "objectives"), label.ErrOutOfScope)
	}
	return e.store.ScopeLabel(name, string(scope))
}

func (e *Engine) carriersOf(given string) (string, label.Carriers, error) {
	name, err := label.OneLabel(given)
	if err != nil {
		return "", label.Carriers{}, err
	}
	held, err := e.everyLabel()
	if err != nil {
		return "", label.Carriers{}, err
	}
	found, there := held[name]
	if !there {
		return "", label.Carriers{}, fmt.Errorf("%q: %w", name, label.ErrNoSuchLabel)
	}
	return name, found, nil
}

func (e *Engine) relabel(c label.Carriers, change []string) error {
	for _, n := range c.Tasks {
		if _, err := e.edit(n, Changes{Labels: &change}); err != nil {
			return err
		}
	}
	for _, n := range c.Objectives {
		if _, err := e.editObjective(n, ObjectiveChanges{Labels: &change}); err != nil {
			return err
		}
	}
	return nil
}

func listedLabels(on workedOn) (Answer, error) { return offeredLabels(on, "") }

func offeredLabels(on workedOn, onto string) (Answer, error) {
	listed, err := on.Engine.labelsCarried()
	if err != nil {
		return Answer{}, err
	}
	if onto != "" {
		listed = slices.DeleteFunc(listed, func(l label.Carried) bool { return !label.Scope(l.Scope).Admits(label.Bearer(onto)) })
	}
	return Answer{Labels: &listed}, nil
}

var scopeArgument = Argument{Name: "scope", Kind: Text, OneOf: oneOf(label.Scopes), Outside: label.ErrNoSuchScope, Means: "what it may be put on: tasks, objectives, or both, which a label nobody scoped is for"}

// LabelList lists every label, with how many tasks and objectives carry it.
var LabelList = Operation{
	Thing: "label",
	Verb:  "list",
	Short: "List every label, with what carries it",
	Does:  "List every label some task or objective carries, and every one made on its own, in alphabetical order, with how many tasks and how many objectives carry it, archived tasks included, and what it may be put on. Given what the labels are for, only those that may be put on it.",
	Example: `cairn label list
cairn label list --for objectives`,
	Level:   Read,
	Needs:   aStore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "for", Kind: Text, OneOf: oneOf([]label.Bearer{label.OnTasks, label.OnObjectives}), Outside: label.ErrNoSuchScope, Means: "keep only the labels that may be put on tasks, or on objectives"},
	},
	Run: func(on workedOn, in Input) (Answer, error) { return offeredLabels(on, in.Given["for"]) },
}

// LabelDelete takes a label off every task and every objective carrying it.
var LabelDelete = Operation{
	Thing:   "label",
	Verb:    "delete",
	Short:   "Take a label off everything carrying it",
	Does:    "Take a label off every task and every objective carrying it, each change written in its task's trace. Answers the labels left.",
	Example: "cairn label delete sp26-20 --yes",
	Level:   Full,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to take off everything"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		name, found, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.relabel(found, []string{"-" + name}); err != nil {
			return Answer{}, err
		}
		if err := on.Engine.store.ForgetLabel(name); err != nil {
			return Answer{}, err
		}
		held, err := on.Engine.viewsHolding(name)
		if err != nil {
			return Answer{}, err
		}
		answer, err := listedLabels(on)
		if held > 0 {
			answer.Warnings = []string{fmt.Sprintf("%s the label %s; it narrows nothing there now", manyOf(held, "view holds", "views hold"), name)}
		}
		return answer, err
	},
}

// LabelRename writes another name for a label on every task and every objective carrying it.
var LabelRename = Operation{
	Thing:   "label",
	Verb:    "rename",
	Short:   "Rename a label on everything carrying it",
	Does:    "Rename a label on every task and every objective carrying it, the new name written the one way a label is; renamed to one already there, the two become one. Answers the labels there are.",
	Example: "cairn label rename sp26-20 sprint-20",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to rename"},
		{Name: "to", Kind: Text, Positional: true, Required: true, Means: "the name it takes"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		name, found, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		to, err := label.OneLabel(in.Given["to"])
		if err != nil {
			return Answer{}, err
		}
		if to != name {
			if err := on.Engine.relabel(found, []string{"-" + name, "+" + to}); err != nil {
				return Answer{}, err
			}
			if err := on.Engine.store.RenameKeptLabel(name, to); err != nil {
				return Answer{}, err
			}
			if err := on.Engine.relabelViews(name, to); err != nil {
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
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to make"},
		scopeArgument,
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		name, err := label.OneLabel(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.store.KeepLabel(name); err != nil {
			return Answer{}, err
		}
		if scope, named := in.Given["scope"]; named {
			if err := on.Engine.scope(name, scope); err != nil {
				return Answer{}, err
			}
		}
		return listedLabels(on)
	},
}

// labelScope says what a label may be put on: tasks, objectives, or both.
var labelScope = Operation{
	Thing:   "label",
	Verb:    "scope",
	Short:   "Say what a label may be put on",
	Does:    "Say what a label may be put on: tasks, objectives, or both, which a label nobody scoped is for. Narrowed to one, it is refused while the other carries it, and putting it on the other is refused after. Answers the labels there are.",
	Example: "cairn label scope area objectives",
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to scope"},
		{Name: "scope", Kind: Text, Positional: true, Required: true, OneOf: scopeArgument.OneOf, Outside: label.ErrNoSuchScope, Means: scopeArgument.Means},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
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
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: label.List{},
	Takes: []Argument{
		{Name: "label", Kind: Text, Positional: true, Required: true, Means: "the label to colour"},
		{Name: "colour", Kind: Text, Positional: true, Required: true, Means: "the colour it wears"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		name, _, err := on.Engine.carriersOf(in.Given["label"])
		if err != nil {
			return Answer{}, err
		}
		colour, err := label.ParseColour(in.Given["colour"])
		if err != nil {
			return Answer{}, err
		}
		if err := on.Engine.store.ColourLabel(name, colour); err != nil {
			return Answer{}, err
		}
		return listedLabels(on)
	},
}

func labelsText(listed []label.Carried) string {
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
