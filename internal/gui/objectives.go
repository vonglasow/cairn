package gui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Tree is the objectives shown, in the order the tab lists them: each of level one followed by its sub-objectives.
func (w Window) Tree() []engine.Branch { return engine.Branches(w.Showing().Goals) }

func depthOf(b engine.Branch) int {
	if b.Under {
		return 1
	}
	return 0
}

// Detailed is the objective the right of the tab shows: the one chosen, or the first of the tree while none is.
func (w Window) Detailed() (engine.Goal, bool) {
	tree := w.Tree()
	for _, b := range tree {
		if w.chosenGoal && b.Goal.Number == w.chosen {
			return b.Goal, true
		}
	}
	if len(tree) == 0 {
		return engine.Goal{}, false
	}
	return tree[0].Goal, true
}

// TaskLine is one task the detail lists: the task, what among the sub-objectives it serves, and where it stands.
type TaskLine struct {
	Card          engine.Card
	Serves, Right string
	At            Box
}

// TilePlaced is one sub-objective the detail shows as a tile, and where.
type TilePlaced struct {
	Number int
	Tile   GoalTile
	At     Box
}

// Detail is the right of the objectives tab: the objective shown and where each of its parts stands.
type Detail struct {
	Goal                                  engine.Goal
	Box, Heading, Line, Progress, Figures Box
	Buttons                               []Box
	Note, Proposed, WaitsAt               Box
	Waits                                 Facts
	SubsHead, OwnHead, InSubsHead, MoreAt Box
	Tiles                                 []TilePlaced
	Own, InSubs                           []TaskLine
	OwnSaid, OwnCount, InSubsCount, More  string
	Left, LeftSays                        string
	Spent                                 string
	Facts                                 string
	Offers                                []string
	Subs, Proposes                        bool
	Counted, ProposedSays, NoteSays       string
}

// DetailOffers is what the detail of an objective offers beside its heading.
var DetailOffers = []string{ShowItsTasks, "Edit"}

// ShowItsTasks is the offer that opens the tasks board on an objective's subtree.
const ShowItsTasks = "Show its tasks"

const (
	detailPad = 20
	lineRoom  = SubHeight
)

// Detail lays out the right of the objectives tab for the objective it shows.
func (w Window) Detail() Detail {
	_, box := SplitOf(w.Layout())
	g, ok := w.Detailed()
	d := Detail{Box: box, Goal: g, Offers: DetailOffers}
	if !ok {
		return d
	}
	inside := Box{X: box.X + detailPad, Y: box.Y + detailPad, W: box.W - 2*detailPad, H: box.H - 2*detailPad}
	m := w.measured()
	end := inside.X + inside.W
	widths := w.widthsOf(DetailOffers)
	for i := len(widths) - 1; i >= 0; i-- {
		end -= widths[i]
		d.Buttons = append([]Box{{X: end, Y: inside.Y, W: widths[i], H: buttonHigh}}, d.Buttons...)
		end -= gap
	}
	y := inside.Y
	d.Heading = Box{X: inside.X, Y: y, W: max(0, end-inside.X), H: buttonHigh}
	y += buttonHigh + between
	d.Line = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	d.Facts = w.goalFacts(g)
	y += lineHigh + padding
	d.Subs = len(g.Children) > 0
	if g.Total > 0 {
		d.Progress = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		d.Counted = strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " tasks · " + strconv.Itoa(g.Percent()) + "%"
		y += lineHigh + inset
		d.Figures = Box{X: inside.X, Y: y, W: inside.W, H: FiguresHigh}
		y += FiguresHigh + inset
	}
	d.Spent, d.Left, d.LeftSays = "—", "0", "left"
	if g.Spent > 0 {
		d.Spent = domain.FormatDuration(g.Spent)
	}
	if left := w.snap.Left(g.Number); left.Tasks > 0 {
		d.Left = strconv.Itoa(left.Tasks)
		if !left.Due.IsZero() {
			d.LeftSays += ", the next due " + domain.FormatDay(left.Due) + " (#" + strconv.Itoa(left.On) + ")"
		}
	}
	if d.Subs {
		d.NoteSays = "Over its whole subtree, each task once."
		d.Note = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + between
	}
	if g.Proposed {
		d.Proposes, d.ProposedSays = true, "every sub-objective is reached: it may be put reached too"
		d.Proposed = Box{X: inside.X, Y: y, W: inside.W, H: BannerHigh(m, d.ProposedSays, "", inside.W)}
		y += d.Proposed.H + between
	}
	d.Waits = w.waitsFacts(g)
	if high := d.Waits.High(); high > 0 {
		d.WaitsAt = Box{X: inside.X, Y: y, W: inside.W, H: high}
		y += high + between
	}
	foot := inside.Y + inside.H - lineHigh
	cards := engine.OnTheBoard(w.snap.Cards, w.archived)
	own := ranked(engine.Serving(cards, g.Number))
	var under []engine.Card
	if d.Subs {
		d.SubsHead = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + inset
		across := max(1, min(len(g.Children), (inside.W+gap)/(tileWide+gap)))
		wide := (inside.W - (across-1)*gap) / across
		for i, c := range g.Children {
			at := Box{X: inside.X + i*(wide+gap), Y: y, W: wide, H: TileHigh}
			if i >= across || at.Y+at.H > foot {
				break
			}
			sub, _ := w.snap.Goal(c)
			d.Tiles = append(d.Tiles, TilePlaced{Number: c, Tile: w.tileOf(sub), At: at})
		}
		y += TileHigh + inset
		under = ranked(engine.Serving(cards, g.Children...))
		d.OwnSaid = "On " + g.Title + " itself"
	} else {
		d.OwnSaid = "Its tasks"
	}
	d.OwnCount = strconv.Itoa(len(own))
	shown := 0
	place := func(head *Box, lines []engine.Card, into *[]TaskLine, serves bool) {
		if y+lineHigh+lineRoom > foot {
			return
		}
		*head = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + inset
		for _, c := range lines {
			if y+lineRoom > foot {
				break
			}
			line := TaskLine{Card: c, At: Box{X: inside.X, Y: y, W: inside.W, H: lineRoom}, Right: c.State}
			if serves {
				line.Serves = w.servedAmong(c, g.Children)
			} else if c.Spent != "" {
				line.Right += " · " + c.Spent
			}
			*into = append(*into, line)
			y += lineRoom
		}
		y += between
		shown += len(*into)
	}
	place(&d.OwnHead, own, &d.Own, false)
	if d.Subs {
		d.InSubsCount = strconv.Itoa(len(under))
		place(&d.InSubsHead, under, &d.InSubs, true)
	}
	if more := len(own) + len(under) - shown; more > 0 {
		d.More = strconv.Itoa(more) + " more · t shows them all on the board"
		d.MoreAt = Box{X: inside.X, Y: min(y, foot), W: inside.W, H: lineHigh}
	}
	return d
}

const buttonHigh = 30

func ranked(cards []engine.Card) []engine.Card {
	out := slices.Clone(cards)
	rank := func(c engine.Card) int {
		switch {
		case c.State == "current":
			return 0
		case engine.Finished(c):
			return 2
		}
		return 1
	}
	slices.SortStableFunc(out, func(a, b engine.Card) int { return rank(a) - rank(b) })
	return out
}

func (w Window) goalFacts(g engine.Goal) string {
	var facts []string
	if g.Parent != 0 {
		parent, _ := w.snap.Goal(g.Parent)
		facts = append(facts, "a sub-objective of #"+strconv.Itoa(parent.Number)+" "+parent.Title, "counted there too")
	} else {
		facts = append(facts, "level one")
		if n := len(g.Children); n > 0 {
			facts = append(facts, strconv.Itoa(n)+" sub-objective"+plural(n))
		}
	}
	if !g.Deadline.IsZero() {
		facts = append(facts, "due "+domain.FormatDay(g.Deadline))
	}
	return strings.Join(facts, " · ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (w Window) named(numbers []int) string {
	var said []string
	for _, n := range numbers {
		g, _ := w.snap.Goal(n)
		said = append(said, "#"+strconv.Itoa(n)+" "+g.Title)
	}
	return strings.Join(said, ", ")
}

func (w Window) waitsFacts(g engine.Goal) Facts {
	var f Facts
	if len(g.WaitsOn) > 0 {
		f.Lines = append(f.Lines, Fact{Name: "waits on", Value: w.named(g.WaitsOn)})
	}
	if len(g.WaitedOnBy) > 0 {
		f.Lines = append(f.Lines, Fact{Name: "waited on by", Value: w.named(g.WaitedOnBy)})
	}
	return f
}

func (w Window) tileOf(g engine.Goal) GoalTile {
	t := GoalTile{Number: "#" + strconv.Itoa(g.Number), Title: g.Title, State: g.State, Priority: g.Priority, Done: g.Done, Total: g.Total}
	if len(g.Labels) > 0 {
		t.Chip, t.ChipColour = g.Labels[0], w.LabelColour(g.Labels[0])
	}
	t.Counts = strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " · " + strconv.Itoa(g.Percent()) + "%"
	if g.Spent > 0 {
		t.Counts += " · " + domain.FormatDuration(g.Spent)
	}
	for _, c := range engine.Serving(engine.OnTheBoard(w.snap.Cards, w.archived), g.Number) {
		var also []string
		for _, other := range w.snap.Served(c) {
			if other.Number != g.Number && other.Number != g.Parent {
				also = append(also, other.Title)
			}
		}
		if len(also) > 0 {
			t.Notes = append(t.Notes, "#"+strconv.Itoa(c.Number)+" also in "+strings.Join(also, ", "))
		}
	}
	var waits []string
	for _, n := range g.WaitsOn {
		if on, ok := w.snap.Goal(n); ok && on.State != "reached" {
			waits = append(waits, on.Title)
		}
	}
	if len(waits) > 0 {
		t.Warn = "waits on " + strings.Join(waits, ", ")
	}
	return t
}

func (w Window) servedAmong(c engine.Card, among []int) string {
	var said []string
	for _, g := range w.snap.Served(c) {
		if slices.Contains(among, g.Number) {
			said = append(said, g.Title)
		}
	}
	return strings.Join(said, ", ")
}

func (w Window) showingItsTasks() Window {
	g, ok := w.Detailed()
	if !ok {
		return w
	}
	filters := engine.BoardFilters(w.snap, w.today()).Keeping(w.filters)
	w = w.onPage(Tasks)
	w.filters = filters.On("objective", strconv.Itoa(g.Number))
	return w
}

func (w Window) detailing(offer string) Window {
	g, ok := w.Detailed()
	if !ok {
		return w
	}
	w.chosen, w.chosenGoal = g.Number, true
	if offer == ShowItsTasks {
		return w.showingItsTasks()
	}
	return w.Edit()
}

func (w Window) buttonLook(offer string, at Box) buttonLook {
	if w.pointed == at {
		return litLook(buttonPlain)
	}
	return buttonPlain
}

func (w Window) clickingDetail(p Point) (Window, bool) {
	if w.screen != OnTheBoard || w.tab != Objectives {
		return w, false
	}
	d := w.Detail()
	for i, at := range d.Buttons {
		if at.In(p) {
			return w.detailing(d.Offers[i]), true
		}
	}
	for _, t := range d.Tiles {
		if t.At.In(p) {
			w.chosen, w.chosenGoal = t.Number, true
			return w, true
		}
	}
	return w, false
}
