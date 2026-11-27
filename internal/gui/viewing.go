package gui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/board"
	"cairn/internal/gui/geom"

	"gioui.org/font"
	"gioui.org/text"
)

// TabSaid is what a tab is called: the Tasks tab names the view in force after it.
func (w Window) TabSaid(page board.Tab) string {
	if v, in := w.InView(); in && page == board.Tasks {
		return Titled(Pages[page]) + " · " + v.Name
	}
	return Titled(Pages[page])
}

func tabArrowRoom(page board.Tab) int {
	if page == board.Tasks {
		return markWide + padding + dotSide + between
	}
	return 0
}

// TabArrow is the arrow in the Tasks tab that drops the views under it; nothing when the tab is not shown.
func (w Window) TabArrow() geom.Box {
	at := slices.Index(w.Shown(), board.Tasks)
	if at < 0 {
		return geom.Box{}
	}
	return FieldMark(w.TabsAt()[at])
}

func (w Window) openingViews() Window {
	w.Screen, w.viewing = board.OnViewing, viewingState{}
	return w
}

// ViewingLine is one line of the views dropped under the Tasks tab: the whole board or a view, how many tasks it keeps, the key opening it, where it is drawn, and whether the keyboard is on it.
type ViewingLine struct {
	At               geom.Box
	Name, Count, Key string
	Place            int
	On               bool
}

type viewingState struct {
	typed board.Written
	on    int
}

func (w Window) viewingChoices() []ViewingLine {
	sought := strings.ToLower(w.viewing.typed.Reads())
	all := []ViewingLine{{Name: WholeBoardSays, Count: strconv.Itoa(len(board.Laid(w.Snap.Cards, w.Archived))), Key: viewKey(0)}}
	for _, v := range w.Snap.Views {
		all = append(all, ViewingLine{Name: v.Name, Count: strconv.Itoa(len(v.Kept)), Key: viewKey(v.Place), Place: v.Place})
	}
	return slices.DeleteFunc(all, func(line ViewingLine) bool { return !strings.Contains(strings.ToLower(line.Name), sought) })
}

// ViewingBox is the views dropped under the Tasks tab: its field, its head, the whole board and each view.
func (w Window) ViewingBox() geom.Box {
	at := slices.Index(w.Shown(), board.Tasks)
	if at < 0 {
		return geom.Box{}
	}
	tab := w.TabsAt()[at]
	rows := 2 + len(w.viewingChoices()) + 1 + len(w.viewingActs())
	top := tab.Y + tab.H + between
	most := max(4, (w.height-geom.Margin-top-2*inset)/menuRow)
	return geom.Box{X: tab.X, Y: top, W: min(switchWide, w.width-2*geom.Margin), H: min(rows, most)*menuRow + 2*inset}
}

// ViewingField is the line at the head of the views, where a name narrows them.
func (w Window) ViewingField() geom.Box { return switcherRow(w.ViewingBox(), 0) }

// ViewingLines is the lines the views dropped under the Tasks tab list, where each is drawn and which one the keyboard is on; those past its room are not drawn.
func (w Window) ViewingLines() []ViewingLine {
	box := w.ViewingBox()
	room := w.viewingRows() - 3 - len(w.viewingActs())
	var shown []ViewingLine
	for i, line := range w.viewingChoices() {
		if i >= room {
			break
		}
		line.On, line.At = i == w.viewing.on, switcherRow(box, 2+i)
		shown = append(shown, line)
	}
	return shown
}

func (w Window) choosingViewing(line ViewingLine) Window {
	w = w.openingView(line.Place)
	w.Screen = board.OnTheBoard
	return w
}

func (w Window) onViewing(name string) Window {
	lines := w.viewingChoices()
	switch name {
	case "esc":
		w.Screen = board.OnTheBoard
	case "up":
		w.viewing.on = max(0, w.viewing.on-1)
	case "down":
		w.viewing.on = min(len(lines)-1, w.viewing.on+1)
	case "enter":
		if w.viewing.on < len(lines) {
			return w.choosingViewing(lines[w.viewing.on])
		}
	case "backspace":
		w.viewing.typed, w.viewing.on = w.viewing.typed.Back(), 0
	default:
		if opened, ok := w.openingByFigure(name); ok {
			opened.Screen = board.OnTheBoard
			return opened
		}
	}
	return w
}

func (w Window) clickingViewing(p geom.Point) Window {
	if !w.ViewingBox().In(p) {
		w.Screen = board.OnTheBoard
		return w
	}
	for _, line := range w.ViewingLines() {
		if line.At.In(p) {
			return w.choosingViewing(line)
		}
	}
	for _, act := range w.ViewingActs() {
		if act.At.In(p) && !act.Off {
			return w.actingInViewing(act)
		}
	}
	return w
}

func drawViewing(k Kit, w Window) {
	box := w.ViewingBox()
	Popover{}.Draw(k, box)
	in, clipped := k.inside("popover").within(box)
	defer clipped.Pop()
	m := w.measured()
	line := Inside(Field{Lit: true}.Draw(in, w.ViewingField()))
	if w.viewing.typed.Reads() == "" {
		say(k.gtx, k.th, k.skin.Colour.Ink.Faint, line, ViewingSays, reads)
	} else {
		writing(in, m, line, w.viewing.typed, true, reads, 1, Follow)
	}
	Head{Said: ViewsSays}.Draw(in, Inside(switcherRow(box, 1)))
	for _, one := range w.ViewingLines() {
		row := Row{Picked: one.On}
		row.Draw(in, one.At)
		inside := Inside(one.At)
		key := w.keyboard.Spelled(Pretty(one.Key))
		right := inside.X + inside.W
		if one.Key != "" {
			right = row.Chip(in, KeyChip{Key: key}, geom.Box{X: right - KeyChipWide(key), Y: one.At.Y, H: one.At.H}).X - between
		}
		count := ends(m, one.Count, small) + padding
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, geom.Box{X: right - count, Y: inside.Y, W: count, H: inside.H}, one.Count, small, k.skin.Type.Mono, text.End)
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, geom.Box{X: inside.X, Y: inside.Y, W: max(0, right-count-padding-inside.X), H: inside.H}, one.Name, reads, font.Medium, 1, text.Start)
	}
	acts := w.ViewingActs()
	Rule{}.Draw(in, geom.Box{X: box.X, Y: acts[0].At.Y - 1, W: box.W})
	for _, act := range acts {
		inside := Inside(act.At)
		if act.Off {
			row := Row{}
			chip := row.Chip(in, KeyChip{Key: act.Key}, geom.Box{X: inside.X + inside.W - KeyChipWide(act.Key), Y: act.At.Y, H: act.At.H})
			say(k.gtx, k.th, k.skin.Colour.Ink.Low, inside, act.Said, reads)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, geom.Box{X: inside.X, Y: inside.Y, W: chip.X - between - inside.X, H: inside.H}, act.Tag, small, font.Font{}, text.End)
			continue
		}
		Row{Said: act.Said}.Draw(in, act.At)
	}
}

// ViewingAct is one act at the foot of the views: what it says, why it is greyed and the key to press first, where it is drawn, and whether it is greyed.
type ViewingAct struct {
	Said, Tag, Key string
	At             geom.Box
	Off            bool
}

func (w Window) viewingActs() []ViewingAct {
	var acts []ViewingAct
	if w.ViewChanged() {
		acts = append(acts, ViewingAct{Said: SaveToViewSays})
	}
	keep := ViewingAct{Said: KeepAsViewSays}
	if !w.Keepable() {
		keep.Off, keep.Tag, keep.Key = true, FilterFirstSays, w.keyboard.Spelled(Pretty("shift+f"))
	}
	return append(acts, keep, ViewingAct{Said: EditViewsSays})
}

func (w Window) viewingRows() int { return (w.ViewingBox().H - 2*inset) / menuRow }

// ViewingActs is the foot of the views: the filters in force saved to the view they stand on, or as a new one, greyed with what to do first when nothing is on; and the panel of the views.
func (w Window) ViewingActs() []ViewingAct {
	acts := w.viewingActs()
	box, rows := w.ViewingBox(), w.viewingRows()
	for i := range acts {
		acts[i].At = switcherRow(box, rows-len(acts)+i)
	}
	return acts
}

func (w Window) actingInViewing(act ViewingAct) Window {
	switch act.Said {
	case SaveToViewSays:
		return w.savingToView()
	case KeepAsViewSays:
		w.Screen = board.OnTheBoard
		return w.startingKeep()
	}
	return w.listingViews()
}

// TabChanged says whether a tab marks the view in force changed: the Tasks tab, with a filter or a search on top of the view.
func (w Window) TabChanged(page board.Tab) bool { return page == board.Tasks && w.ViewChanged() }
