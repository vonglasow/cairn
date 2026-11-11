package engine

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// Viewing is one view as a board holds it: the view, and the numbers of the tasks it keeps today.
type Viewing struct {
	View
	Kept []int
}

func lookAtViews(w Work) ([]Viewing, error) {
	listed, err := w.Ask(ViewList, Input{})
	if err != nil || listed.ViewsKept == nil {
		return nil, err
	}
	var out []Viewing
	for _, v := range listed.ViewsKept.Views {
		kept, err := w.Ask(TaskList, Input{Given: map[string]string{"view": v.Name, "with-archived": "true"}})
		if err != nil {
			return nil, err
		}
		seen := Viewing{View: v, Kept: []int{}}
		for _, t := range kept.Matched {
			seen.Kept = append(seen.Kept, t.Carried().Number)
		}
		out = append(out, seen)
	}
	return out, nil
}

// View is the view of that name, whatever its case or its accents, and whether there is one.
func (s Snapshot) View(name string) (Viewing, bool) {
	at := slices.IndexFunc(s.Views, func(v Viewing) bool { return domain.Folded(v.Name) == domain.Folded(name) })
	if at < 0 {
		return Viewing{}, false
	}
	return s.Views[at], true
}

// Viewed is the cards the view of that name keeps, or every card when no view of that name is in force.
func (s Snapshot) Viewed(name string, cards []Card) []Card {
	v, found := s.View(name)
	if name == "" || !found {
		return cards
	}
	return slices.DeleteFunc(slices.Clone(cards), func(c Card) bool { return !slices.Contains(v.Kept, c.Number) })
}

// GoneSays is what a view says of the conditions it holds whose objective or label is gone, or nothing when none is.
func GoneSays(v View) string {
	switch len(v.Gone) {
	case 0:
		return ""
	case 1:
		return v.Gone[0] + " is gone and narrows nothing"
	}
	return strings.Join(v.Gone, " and ") + " are gone and narrow nothing"
}

// Narrows says whether a view holds the filter standing for an argument of task list.
func (h Holds) Narrows(arg string) bool {
	switch arg {
	case "state":
		return h.State != ""
	case "priority":
		return h.Priority != ""
	case "label":
		return len(h.Label) > 0
	case "objective":
		return len(h.Objective) > 0
	case "focus":
		return h.Focus != ""
	}
	return h.DueBefore != ""
}

// Reference is a task a board leaves out that touches one it shows, and every way it does.
type Reference struct {
	Card Card
	Why  []Touch
}

// Touching is the tasks left out that block, wait on or are linked to one shown, each once with every way it touches them: what a narrowed board keeps in sight as a reference.
func Touching(shown, all []Card) []Reference {
	var out []Reference
	for _, c := range all {
		if slices.ContainsFunc(shown, func(s Card) bool { return s.Number == c.Number }) {
			continue
		}
		r := Reference{Card: c}
		for _, is := range touching {
			for _, s := range shown {
				if touches(is, carriedOf(c), carriedOf(s)) {
					r.Why = append(r.Why, Touch{Is: is, Of: s.Number})
				}
			}
		}
		if len(r.Why) > 0 {
			out = append(out, r)
		}
	}
	return out
}

func carriedOf(c Card) task.Carried {
	return task.Carried{Number: c.Number, Blockers: c.Blockers, Linked: c.Linked}
}

// Says is what a reference says of itself: that the board leaves it out, and every way it touches what is shown.
func (r Reference) Says() string { return "outside this view · " + r.Touches("#") }

// Touches is every way a reference touches what is shown, each kind once with every task it touches, its number after the mark a board writes before one.
func (r Reference) Touches(mark string) string {
	var said []string
	for _, is := range touching {
		var of []string
		for _, w := range r.Why {
			if w.Is == is {
				of = append(of, mark+strconv.Itoa(w.Of))
			}
		}
		if len(of) > 0 {
			said = append(said, is+" "+strings.Join(of, ", "))
		}
	}
	return strings.Join(said, " · ")
}

func objectiveOf(value string) (string, bool) {
	number, itself := strings.CutSuffix(value, " itself")
	return number, itself
}

func (fs Filters) holding(in Input) Input {
	for _, f := range fs {
		taken := f.Taken()
		switch {
		case len(taken) == 0:
		case f.Arg == "objective":
			for _, value := range taken {
				number, itself := objectiveOf(value)
				if !slices.Contains(in.Listed["objective"], number) {
					in.Listed["objective"] = append(in.Listed["objective"], number)
				}
				if itself {
					in.Given["itself"] = "true"
				}
			}
		case f.Arg == "focus":
			delete(in.Given, "focus")
			delete(in.Given, "no-focus")
			in.Given[focusAsked(taken[0])] = "true"
		case f.Arg == "label":
			for _, value := range taken {
				if !slices.Contains(in.Listed["label"], value) {
					in.Listed["label"] = append(in.Listed["label"], value)
				}
			}
		default:
			in.Given[f.Arg] = taken[0]
		}
	}
	return in
}

func searched(words ...string) string {
	return strings.Join(strings.Fields(strings.Join(words, " ")), " ")
}

// Holding is what keeps these filters and a search, standing over a view or over the whole board, as a new view: the input view new is given beside a name, holding what the view holds and what stands on top of it.
func (fs Filters) Holding(over View, search string) Input {
	in := fs.Saving(over, search)
	delete(in.Given, "view")
	held := over.Holds
	for name, value := range map[string]string{"state": held.State, "priority": held.Priority, "deadline-before": held.DueBefore, "search": held.Search} {
		if _, named := in.Given[name]; !named && value != "" {
			in.Given[name] = value
		}
	}
	if _, named := in.Listed["label"]; !named && len(held.Label) > 0 {
		in.Listed["label"] = slices.Clone(held.Label)
	}
	if _, named := in.Listed["objective"]; !named {
		for _, n := range held.Objective {
			in.Listed["objective"] = append(in.Listed["objective"], strconv.Itoa(n))
		}
	}
	if held.Itself {
		in.Given["itself"] = "true"
	}
	if _, named := in.Given["focus"]; !named && held.Focus != "" {
		if _, named := in.Given["no-focus"]; !named {
			in.Given[focusAsked(held.Focus)] = "true"
		}
	}
	return in
}

// Saving is what puts these filters and a search, added on top of a view, into it: a label or an objective joins those it holds, any other filter takes the place of the one held, and the search's words follow its own.
func (fs Filters) Saving(v View, search string) Input {
	in := Input{Given: map[string]string{"view": v.Name}, Listed: map[string][]string{}}
	if fs.NarrowsBy("label") {
		in.Listed["label"] = slices.Clone(v.Holds.Label)
	}
	if fs.NarrowsBy("objective") {
		for _, n := range v.Holds.Objective {
			in.Listed["objective"] = append(in.Listed["objective"], strconv.Itoa(n))
		}
	}
	in = fs.holding(in)
	if said := searched(v.Holds.Search, search); said != "" && search != "" {
		in.Given["search"] = said
	}
	return in
}

// Proposed is the name a view of these filters and a search is offered under: what each is on, an objective by its title, joined by a middle dot.
func (fs Filters) Proposed(s Snapshot, search string) string {
	var said []string
	for _, f := range fs {
		if len(f.Taken()) == 0 {
			continue
		}
		if f.Arg != "objective" {
			said = append(said, f.TakenSays())
			continue
		}
		for _, value := range f.Taken() {
			number, _ := objectiveOf(value)
			n, _ := strconv.Atoi(number)
			g, _ := s.Goal(n)
			said = append(said, g.Title)
		}
	}
	if words := searched(search); words != "" {
		said = append(said, words)
	}
	return strings.Join(said, " · ")
}
