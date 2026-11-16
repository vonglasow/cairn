package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/series"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
)

func depthOf(b board.Branch) int {
	if b.Under {
		return 1
	}
	return 0
}

// TaskLine is one task the detail lists: the task, what among the sub-objectives it serves, and where it stands.
type TaskLine struct {
	Card          board.Card
	Serves, Right string
	At            geom.Box
}

// TilePlaced is one sub-objective the detail shows as a tile, and where.
type TilePlaced struct {
	Number int
	Tile   GoalTile
	At     geom.Box
}

// Detail is the right of the objectives tab: the objective shown and where each of its parts stands.
type Detail struct {
	Goal                                  board.Goal
	Box, Heading, Line, Progress, Figures geom.Box
	Buttons                               []geom.Box
	Note, Proposed, WaitsAt               geom.Box
	Waits                                 Facts
	SubsHead, OwnHead, InSubsHead, MoreAt geom.Box
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

// DetailOffers is what the detail of an objective offers beside its heading; one of level one offers a sub-objective under it too.
func DetailOffers(g board.Goal) []string {
	if g.Parent == 0 {
		return []string{ShowItsTasks, AddSubSays, "Edit"}
	}
	return []string{ShowItsTasks, "Edit"}
}

// ShowItsTasks is the offer that opens the tasks board on an objective's subtree.
const ShowItsTasks = "Show its tasks"

// AddSubSays is the offer that opens the form of a new objective under the one shown.
const AddSubSays = "+ Sub-objective"

const (
	detailPad = 20
	lineRoom  = geom.SubHeight
)

// Detail lays out the right of the objectives tab for the objective it shows.
func (w Window) Detail() Detail {
	_, box := SplitOf(w.Layout())
	g, ok := w.Detailed()
	d := Detail{Box: box, Goal: g, Offers: DetailOffers(g)}
	if !ok {
		return d
	}
	inside := geom.Box{X: box.X + detailPad, Y: box.Y + detailPad, W: box.W - 2*detailPad, H: box.H - 2*detailPad}
	m := w.measured()
	if w.widthsWide(d.Offers)+len(d.Offers)*geom.Gap > inside.W-headingLeast {
		d.Offers = slices.DeleteFunc(d.Offers, func(o string) bool { return o == AddSubSays })
	}
	end := inside.X + inside.W
	widths := w.widthsOf(d.Offers)
	for i := len(widths) - 1; i >= 0; i-- {
		end -= widths[i]
		d.Buttons = append([]geom.Box{{X: end, Y: inside.Y, W: widths[i], H: buttonHigh}}, d.Buttons...)
		end -= geom.Gap
	}
	y := inside.Y
	d.Heading = geom.Box{X: inside.X, Y: y, W: max(0, end-inside.X), H: buttonHigh}
	y += buttonHigh + between
	d.Line = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	d.Facts = w.goalFacts(g)
	y += lineHigh + padding
	d.Subs = len(g.Children) > 0
	if g.Total > 0 {
		d.Progress = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		d.Counted = strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " tasks · " + strconv.Itoa(g.Percent()) + "%"
		y += lineHigh + inset
		d.Figures = geom.Box{X: inside.X, Y: y, W: inside.W, H: FiguresHigh}
		y += FiguresHigh + inset
	}
	d.Spent, d.Left, d.LeftSays = w.SpentSaid(calendar.FormatDuration(g.Spent)), "0", "left"
	if left := w.Snap.Left(g.Number); left.Tasks > 0 {
		d.Left = strconv.Itoa(left.Tasks)
		if !left.Due.IsZero() {
			d.LeftSays += ", the next deadline " + calendar.FormatDay(left.Due) + " (#" + strconv.Itoa(left.On) + ")"
		}
	}
	if d.Subs {
		d.NoteSays = "Over its whole subtree, each task once."
		d.Note = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + between
	}
	if g.Proposed {
		d.Proposes, d.ProposedSays = true, "every sub-objective is reached: it may be put reached too"
		d.Proposed = geom.Box{X: inside.X, Y: y, W: inside.W, H: BannerHigh(m, d.ProposedSays, "", inside.W)}
		y += d.Proposed.H + between
	}
	d.Waits = w.waitsFacts(g)
	if high := d.Waits.High(); high > 0 {
		d.WaitsAt = geom.Box{X: inside.X, Y: y, W: inside.W, H: high}
		y += high + between
	}
	foot := inside.Y + inside.H - lineHigh
	cards := board.Laid(w.Snap.Cards, w.Archived)
	own := ranked(board.Serving(cards, g.Number))
	var under []board.Card
	if d.Subs {
		d.SubsHead = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + inset
		across := max(1, min(len(g.Children), (inside.W+geom.Gap)/(tileWide+geom.Gap)))
		wide := (inside.W - (across-1)*geom.Gap) / across
		for i, c := range g.Children {
			at := geom.Box{X: inside.X + i*(wide+geom.Gap), Y: y, W: wide, H: TileHigh}
			if i >= across || at.Y+at.H > foot {
				break
			}
			sub, _ := w.Snap.Goal(c)
			d.Tiles = append(d.Tiles, TilePlaced{Number: c, Tile: w.tileOf(sub), At: at})
		}
		y += TileHigh + inset
		under = ranked(board.Serving(cards, g.Children...))
		d.OwnSaid = "On " + g.Title + " itself"
	} else {
		d.OwnSaid = "Its tasks"
	}
	d.OwnCount = strconv.Itoa(len(own))
	shown := 0
	place := func(head *geom.Box, lines []board.Card, into *[]TaskLine, serves bool) {
		if y+lineHigh+lineRoom > foot {
			return
		}
		*head = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
		y += lineHigh + inset
		for _, c := range lines {
			if y+lineRoom > foot {
				break
			}
			line := TaskLine{Card: c, At: geom.Box{X: inside.X, Y: y, W: inside.W, H: lineRoom}, Right: c.State}
			if serves {
				line.Serves = w.ServedAmong(c, g.Children)
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
		d.More = strconv.Itoa(more) + " more · `t` shows them all on the board"
		d.MoreAt = geom.Box{X: inside.X, Y: min(y, foot), W: inside.W, H: lineHigh}
	}
	return d
}

const buttonHigh = 30

const headingLeast = 140

func (w Window) widthsWide(offers []string) int {
	wide := 0
	for _, one := range w.widthsOf(offers) {
		wide += one
	}
	return wide
}

func ranked(cards []board.Card) []board.Card {
	out := slices.Clone(cards)
	rank := func(c board.Card) int {
		switch {
		case c.State == "current":
			return 0
		case board.Finished(c):
			return 2
		}
		return 1
	}
	slices.SortStableFunc(out, func(a, b board.Card) int { return rank(a) - rank(b) })
	return out
}

func (w Window) goalFacts(g board.Goal) string {
	var facts []string
	if g.Parent != 0 {
		parent, _ := w.Snap.Goal(g.Parent)
		facts = append(facts, "a sub-objective of #"+strconv.Itoa(parent.Number)+" "+parent.Title, "counted there too")
	} else {
		facts = append(facts, "level one")
		if n := len(g.Children); n > 0 {
			facts = append(facts, strconv.Itoa(n)+" sub-objective"+plural(n))
		}
	}
	if !g.Deadline.IsZero() {
		facts = append(facts, "deadline "+calendar.FormatDay(g.Deadline))
	}
	return strings.Join(facts, " · ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (w Window) waitsFacts(g board.Goal) Facts {
	var f Facts
	if len(g.WaitsOn) > 0 {
		f.Lines = append(f.Lines, Fact{Name: "waits on", Value: w.Named(g.WaitsOn)})
	}
	if len(g.WaitedOnBy) > 0 {
		f.Lines = append(f.Lines, Fact{Name: "waited on by", Value: w.Named(g.WaitedOnBy)})
	}
	return f
}

func (w Window) tileOf(g board.Goal) GoalTile {
	t := GoalTile{Number: "#" + strconv.Itoa(g.Number), Title: g.Title, State: g.State, Priority: g.Priority, Done: g.Done, Total: g.Total}
	if len(g.Labels) > 0 {
		t.Chip, t.ChipColour = g.Labels[0], w.LabelColour(g.Labels[0])
	}
	t.Counts = strconv.Itoa(g.Done) + " of " + strconv.Itoa(g.Total) + " · " + strconv.Itoa(g.Percent()) + "%"
	if g.Spent > 0 {
		t.Counts += " · " + calendar.FormatDuration(g.Spent)
	}
	for _, c := range board.Serving(board.Laid(w.Snap.Cards, w.Archived), g.Number) {
		var also []string
		for _, other := range w.Snap.Served(c) {
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
		if on, ok := w.Snap.Goal(n); ok && on.State != "reached" {
			waits = append(waits, on.Title)
		}
	}
	if len(waits) > 0 {
		t.Warn = "waits on " + strings.Join(waits, ", ")
	}
	if g.LateOn(w.Today()) {
		t.Warn = "late · its deadline was " + series.ShortDay(g.Deadline)
	}
	return t
}

func (w Window) showingItsTasks() Window {
	g, ok := w.Detailed()
	if !ok {
		return w
	}
	filters := board.FiltersOf(w.Snap, w.Today()).Keeping(w.Filters)
	w = w.onPage(board.Tasks)
	w.Filters = filters.On("objective", strconv.Itoa(g.Number))
	return w
}

func (w Window) detailing(offer string) Window {
	g, ok := w.Detailed()
	if !ok {
		return w
	}
	w.Chosen, w.ChosenGoal = g.Number, true
	switch offer {
	case ShowItsTasks:
		return w.showingItsTasks()
	case AddSubSays:
		return w.addingUnder(g)
	}
	return w.Edit()
}

func (w Window) addingUnder(g board.Goal) Window {
	if g.Parent != 0 {
		return w
	}
	w.Form, w.formMoved = board.AsOpened(board.Form{Adding: true, ForObjective: true, Fields: board.NewSubGoalFields(g.Number)}), geom.Point{}
	w = w.beside()
	w.Screen = board.OnForm
	return w
}

func (w Window) buttonLook(offer string, at geom.Box) buttonLook {
	if w.pointed == at {
		return litLook(buttonPlain)
	}
	return buttonPlain
}

func (w Window) goalSlotAt(p geom.Point) geom.Box {
	l := w.Layout()
	if GoalsRun(l).In(p) {
		for at := range w.Tree() {
			if item := GoalAt(l, at, w.goals); item.In(p) {
				return item
			}
		}
	}
	d := w.Detail()
	for _, at := range append(slices.Clone(d.Buttons), tilesOf(d)...) {
		if at.In(p) {
			return at
		}
	}
	return geom.Box{}
}

func tilesOf(d Detail) []geom.Box {
	var out []geom.Box
	for _, t := range d.Tiles {
		out = append(out, t.At)
	}
	return out
}

func (w Window) clickingDetail(p geom.Point) (Window, bool) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Objectives {
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
			return w.choosingGoal(t.Number), true
		}
	}
	for _, line := range append(append([]TaskLine{}, d.Own...), d.InSubs...) {
		if line.At.In(p) {
			w.DetailGoal = d.Goal.Number
			return w.openingTask(line.Card.Number), true
		}
	}
	for at, b := range w.Tree() {
		if len(b.Goal.Children) > 0 && !b.Under && FoldAt(GoalAt(w.Layout(), at, w.goals)).In(p) && GoalsRun(w.Layout()).In(p) {
			return w.togglingFold(b.Goal.Number), true
		}
	}
	return w, false
}
