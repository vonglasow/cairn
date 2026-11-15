package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// The words the window says about its views.
const (
	ViewsSays        = "Views"
	EditViewsSays    = "Edit the views…"
	KeepAsViewSays   = "Save as a view…"
	KeepTheViewSays  = "Keep the view"
	SaveToViewSays   = "Save to the view"
	KeepAsNewSays    = "Keep as a new view…"
	WholeBoardSays   = "Whole board"
	ViewsNoneSays    = "No view yet: filter the board, then keep it from the filters"
	viewsThenSays    = "e opens it to change it, ⌘⌫ deletes it"
	viewsAtMostByKey = 9
)

func viewKey(place int) string {
	if place > viewsAtMostByKey {
		return ""
	}
	return "cmd+" + strconv.Itoa(place)
}

func (w Window) openingView(place int) Window {
	if place == 0 {
		return w.leavingView()
	}
	if place > len(w.Snap.Views) {
		return w
	}
	w.Viewing = w.Snap.Views[place-1].Name
	w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
	w.Screen, w.FilterOpen, w.FilterFromPill = board.OnTheBoard, false, false
	return w
}

func (w Window) leavingView() Window {
	if w.Viewing == "" {
		return w
	}
	w.Viewing = ""
	w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
	return w
}

func (w Window) openingByFigure(name string) (Window, bool) {
	figure, ok := strings.CutPrefix(name, "cmd+")
	if !ok || !aFigure(figure) {
		return w, false
	}
	place, _ := strconv.Atoi(figure)
	return w.openingView(place), true
}

func (w Window) startingKeep() Window {
	if !w.Keepable() {
		return w
	}
	w = w.filteringBoard()
	proposed := w.Filters.Proposed(w.Snap, w.Looking.Reads())
	if v, in := w.InView(); in {
		proposed = strings.Join(slices.DeleteFunc([]string{v.Name, proposed}, func(s string) bool { return s == "" }), " · ")
	}
	w.Screen, w.OfView, w.KeepName = board.OnKeeping, false, board.Ending(proposed).All()
	return w
}

func (w Window) openingViewPopover() Window {
	w = w.filteringBoard()
	w.Screen, w.OfView = board.OnKeeping, true
	return w
}

func (w Window) keepingView(name string) Window {
	in := w.KeepInput()
	in.Given["name"] = name
	kept, err := w.work.Ask(engine.ViewNew, in)
	if err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.Viewing = kept.Viewed.Name
	w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
	w.Screen, w.OfView = board.OnTheBoard, false
	return w
}

func (w Window) savingToView() Window {
	v, in := w.InView()
	if !in || !w.ViewChanged() {
		return w
	}
	if _, err := w.work.Ask(engine.ViewEdit, w.Filters.Saving(v.View, w.Looking.Reads())); err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.Filters, w.Looking = w.Filters.Cleared(), board.Written{}
	w.Screen, w.OfView = board.OnTheBoard, false
	return w
}

func (w Window) onKeeping(name string) Window {
	if w.OfView {
		return w
	}
	if edit, ok := textKey(name); ok {
		w.KeepName = edit(w.KeepName)
		return w
	}
	switch name {
	case "backspace":
		w.KeepName = w.KeepName.Back()
	case "cmd+a":
		w.KeepName = w.KeepName.All()
	}
	return w
}

// KeepingBox is the popover dropped under the pill it opened from: a view about to be kept, its name typed and what it would keep said, or the view in force, what it holds and what it shows against the store.
func (w Window) KeepingBox() geom.Box {
	box := w.droppedUnder(w.filterAnchor(), max(keepingWide, w.footWide()), 0)
	box.H = dropHead + w.keepingFacts().High() + dropFoot + 2*padding
	if !w.OfView {
		box.H += menuRow + padding
	}
	return box
}

// KeepingField is where the name of a view about to be kept is typed.
func (w Window) KeepingField() geom.Box {
	box := w.KeepingBox()
	return geom.Box{X: box.X + padding, Y: box.Y + dropHead + padding/2, W: box.W - 2*padding, H: menuRow}
}

// KeepingFactsAt is where the popover says what the view holds and keeps.
func (w Window) KeepingFactsAt() geom.Box {
	box := w.KeepingBox()
	y := box.Y + dropHead + padding/2
	if !w.OfView {
		y = w.KeepingField().Y + menuRow + padding
	}
	return geom.Box{X: box.X + padding, Y: y, W: box.W - 2*padding, H: w.keepingFacts().High()}
}

// KeepingHead is what heads the popover: keeping a view, or the name of the view in force.
func (w Window) KeepingHead() string {
	if v, in := w.InView(); in && w.OfView {
		return "view " + v.Name
	}
	return KeepAsViewSays
}

func (w Window) keepingFacts() Facts {
	of := len(board.Laid(w.Snap.Cards, w.Archived))
	if v, in := w.InView(); in && w.OfView {
		lines := []Fact{
			{Name: "holds", Value: engine.HoldsSays(v.View)},
			{Name: "shows", Value: strconv.Itoa(v.Keeps) + " of " + countOf(v.Of, "task") + " · " + strconv.Itoa(v.Done) + " done"},
		}
		for _, n := range v.Holds.Objective {
			if g, found := w.Snap.Goal(n); found {
				lines = append(lines, Fact{Name: "#" + strconv.Itoa(n), Value: strconv.Itoa(g.Done) + "/" + countOf(g.Total, "task") + " done, the store's", Aside: g.Title})
			}
		}
		if gone := board.GoneSays(v.View); gone != "" {
			lines = append(lines, Fact{Name: "gone", Value: gone})
		}
		return Facts{Lines: lines}
	}
	lines := []Fact{{Name: "keeps", Value: strconv.Itoa(w.Holds(board.Tasks)) + " of " + countOf(of, "task")}}
	if said := w.Filters.Proposed(w.Snap, w.Looking.Reads()); said != "" {
		lines = append(lines, Fact{Name: "holds", Value: said})
	}
	if key := viewKey(len(w.Snap.Views) + 1); key != "" {
		lines = append(lines, Fact{Name: "key", Value: w.keyboard.Spelled(Pretty(key))})
	}
	return Facts{Lines: lines}
}

// ViewRow is one view listed at the head of the filter menu, or in the panel of the views: where it is, its name, how many tasks it keeps, the key opening it, what it holds, whether it is the one in force or the one chosen, whether its name is being typed, and whether something it holds is gone.
type ViewRow struct {
	At                geom.Box
	Name, Holds, Key  string
	Keeps             int
	On, Taken, Typing bool
	Late              bool
}

func rowOf(v board.Viewing) ViewRow {
	return ViewRow{Name: v.Name, Keeps: len(v.Kept), Key: viewKey(v.Place), Holds: engine.HoldsSays(v.View), Late: len(v.Gone) > 0}
}

// ViewRows is the views listed at the head of the open filter menu, the one in force taken.
func (w Window) ViewRows() []ViewRow {
	if w.Screen != board.OnFiltering || w.ViewsAbove() == 0 {
		return nil
	}
	box := w.filtersColumn()
	if box.W == 0 {
		return nil
	}
	var rows []ViewRow
	for i, v := range w.Snap.Views {
		row := rowOf(v)
		row.At, row.Taken, row.On = DropRowAt(box, i), v.Name == w.Viewing, i+1 == w.ViewWalked
		rows = append(rows, row)
	}
	return rows
}

// EditViewsAt is the line under the views of the filter menu that opens their panel.
func (w Window) EditViewsAt() geom.Box { return DropRowAt(w.filtersColumn(), len(w.Snap.Views)) }

// FilterHeadAt is the line naming the filters under the views of the filter menu.
func (w Window) FilterHeadAt() geom.Box { return DropRowAt(w.filtersColumn(), len(w.Snap.Views)+1) }

func (w Window) clickingViews(p geom.Point) (Window, bool) {
	for _, row := range w.ViewRows() {
		if v, found := w.Snap.View(row.Name); found && row.At.In(p) {
			return w.openingView(v.Place), true
		}
	}
	if w.ViewsAbove() > 0 && w.EditViewsAt().In(p) {
		return w.listingViews(), true
	}
	return w, false
}

func (w Window) listingViews() Window {
	w.Screen, w.RenamingView = board.OnViews, false
	w.ViewOn = max(0, min(w.ViewOn, len(w.Snap.Views)-1))
	return w
}

func (w Window) viewsOpens() geom.Box {
	var said []string
	for _, v := range w.Snap.Views {
		said = append(said, v.Name+"   "+engine.HoldsSays(v.View)+"   "+strconv.Itoa(len(v.Kept))+"   ⌘9")
	}
	if len(said) == 0 {
		said = append(said, ViewsNoneSays)
	}
	if _, _, warns := w.viewsWarned(); warns {
		said = append(said, "", "")
	}
	return w.listBox(said)
}

// ViewsPanelRows is the lines of the panel of the views, the one chosen on, its name a field while it is renamed.
func (w Window) ViewsPanelRows() []ViewRow {
	if w.Screen != board.OnViews {
		return nil
	}
	box := w.ViewsBox()
	var rows []ViewRow
	for i, v := range w.Snap.Views {
		row := rowOf(v)
		row.At, row.On, row.Taken = ListRowAt(box, i), i == w.ViewOn, v.Name == w.Viewing
		row.Typing = row.On && w.RenamingView
		rows = append(rows, row)
	}
	return rows
}

func (w Window) viewsWarned() (string, geom.Box, bool) {
	if w.ViewOn >= len(w.Snap.Views) {
		return "", geom.Box{}, false
	}
	gone := board.GoneSays(w.Snap.Views[w.ViewOn].View)
	return gone, geom.Box{}, gone != ""
}

// ViewsWarning is what the panel of the views warns of the view chosen, where, and whether it warns at all: what it holds that is gone.
func (w Window) ViewsWarning() (string, geom.Box, bool) {
	said, _, warns := w.viewsWarned()
	if !warns || w.Screen != board.OnViews {
		return "", geom.Box{}, false
	}
	box := w.ViewsBox()
	row := ListRowAt(box, len(w.Snap.Views))
	return said, geom.Box{X: row.X, Y: row.Y + between, W: row.W, H: BannerHigh(w.measured(), said, viewsThenSays, row.W)}, true
}

func (w Window) onViews(name string) Window {
	if w.RenamingView {
		switch name {
		case "backspace":
			w.ViewRenamed = w.ViewRenamed.Back()
		case "cmd+a":
			w.ViewRenamed = w.ViewRenamed.All()
		}
		return w
	}
	switch name {
	case "down", "j":
		w.ViewOn = min(len(w.Snap.Views)-1, w.ViewOn+1)
	case "up", "k":
		w.ViewOn = max(0, w.ViewOn-1)
	case "shift+down":
		return w.movingView(1)
	case "shift+up":
		return w.movingView(-1)
	}
	return w
}

func (w Window) movingView(by int) Window {
	v, ok := w.ChosenView()
	to := v.Place + by
	if !ok || to < 1 || to > len(w.Snap.Views) {
		return w
	}
	in := engine.Input{Given: map[string]string{"view": v.Name, "place": strconv.Itoa(to)}}
	if _, err := w.work.Ask(engine.ViewEdit, in); err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.ViewOn = to - 1
	return w
}

func (w Window) renamingTheView() Window {
	v, ok := w.ChosenView()
	if !ok {
		return w
	}
	in := engine.Input{Given: map[string]string{"view": v.Name, "name": w.ViewRenamed.Reads()}}
	renamed, err := w.work.Ask(engine.ViewEdit, in)
	if err != nil {
		return w.failing(err)
	}
	if w.Viewing == v.Name {
		w.Viewing = renamed.Viewed.Name
	}
	w = w.Look()
	w.RenamingView = false
	return w
}

func (w Window) deletingView() Window {
	v, ok := w.ChosenView()
	if !ok {
		return w
	}
	if _, err := w.work.Ask(engine.ViewDelete, engine.Input{Given: map[string]string{"view": v.Name}}); err != nil {
		return w.failing(err)
	}
	if w.Viewing == v.Name {
		w = w.leavingView()
	}
	w = w.Look().listingViews()
	return w.telling("the view " + v.Name + " was deleted, and no task with it")
}

func (w Window) choosingOnViews(offer string) (Window, bool) {
	switch strings.ToLower(offer) {
	case strings.ToLower(KeepTheViewSays):
		return w.keepingView(strings.TrimSpace(w.KeepName.Reads())), true
	case strings.ToLower(SaveToViewSays):
		return w.savingToView(), true
	case strings.ToLower(KeepAsNewSays):
		w.Screen = board.OnTheBoard
		return w.startingKeep(), true
	case "open":
		v, ok := w.ChosenView()
		if !ok {
			return w, true
		}
		return w.openingView(v.Place), true
	case "rename":
		if v, ok := w.ChosenView(); ok {
			w.RenamingView, w.ViewRenamed = true, board.Ending(v.Name).All()
		}
		return w, true
	case "save":
		return w.renamingTheView(), true
	case "delete":
		return w.deletingView(), true
	case "cancel", "close":
		if w.RenamingView {
			w.RenamingView = false
			return w, true
		}
		w.Screen, w.OfView = board.OnTheBoard, false
		return w, true
	}
	return w, false
}

func (w Window) clickingKeeping(p geom.Point) Window {
	if !w.KeepingBox().In(p) {
		w.Screen, w.OfView = board.OnTheBoard, false
	}
	return w
}

func (w Window) clickingViewsPanel(p geom.Point) Window {
	for i, row := range w.ViewsPanelRows() {
		if row.At.In(p) && !row.Typing {
			w.ViewOn, w.RenamingView = i, false
			return w
		}
	}
	if !w.ViewsBox().In(p) {
		w.Screen, w.RenamingView = board.OnTheBoard, false
	}
	return w
}

func (w Window) viewRows() []PaletteRow {
	var rows []PaletteRow
	for _, v := range w.Snap.Views {
		place := v.Place
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: "View: " + v.Name, Key: viewKey(place), does: func(w Window) Window { return w.openingView(place) }})
	}
	if w.Viewing != "" {
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: WholeBoardSays, Key: "cmd+0", does: Window.leavingView})
	}
	if w.Keepable() {
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: KeepAsViewSays, does: Window.startingKeep})
	}
	return append(rows, PaletteRow{Group: paletteActionsGroup, Says: ViewsSays + "…", does: Window.listingViews})
}

func (w Window) viewCommandRows(typed string) []PaletteRow {
	name, ok := strings.CutPrefix(strings.TrimSpace(typed), "view new ")
	if name = strings.TrimSpace(name); !ok || name == "" || !w.Keepable() {
		return nil
	}
	return []PaletteRow{{Group: paletteActionsGroup, Says: "Keep the filters as the view " + name, does: func(w Window) Window {
		return w.keepingView(name)
	}}}
}

// ViewsBox is the panel of the views: across the middle of the window, a line a view, and the warning of the one chosen when something it holds is gone.
func (w Window) ViewsBox() geom.Box { return w.placedPanel(board.OnViews, w.viewsOpens()) }
