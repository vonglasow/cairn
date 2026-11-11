package gui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

type viewsState struct {
	viewing      string
	keepName     Written
	ofView       bool
	viewOn       int
	renamingView bool
	viewRenamed  Written
}

// Viewing is the name of the view in force, and nothing on the whole board.
func (w Window) Viewing() string { return w.viewing }

// Naming is the name being typed for a view about to be kept.
func (w Window) Naming() Written { return w.keepName }

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

func (w Window) inView() (engine.Viewing, bool) {
	if w.viewing == "" {
		return engine.Viewing{}, false
	}
	return w.snap.View(w.viewing)
}

// ViewChanged says whether a filter or a search stands on top of the view in force.
func (w Window) ViewChanged() bool {
	_, in := w.inView()
	return in && (w.filters.Narrows() || !w.looking.Empty())
}

func (w Window) keepable() bool {
	return w.filters.Narrows() || !w.looking.Empty()
}

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
	if place > len(w.snap.Views) {
		return w
	}
	w.viewing = w.snap.Views[place-1].Name
	w.filters, w.looking = w.filters.Cleared(), Written{}
	w.screen, w.filterOpen, w.filterFromPill = OnTheBoard, false, false
	return w
}

func (w Window) leavingView() Window {
	if w.viewing == "" {
		return w
	}
	w.viewing = ""
	w.filters, w.looking = w.filters.Cleared(), Written{}
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
	if !w.keepable() {
		return w
	}
	w = w.filteringBoard()
	proposed := w.filters.Proposed(w.snap, w.looking.Reads())
	if v, in := w.inView(); in {
		proposed = strings.Join(slices.DeleteFunc([]string{v.Name, proposed}, func(s string) bool { return s == "" }), " · ")
	}
	w.screen, w.ofView, w.keepName = OnKeeping, false, Ending(proposed).All()
	return w
}

func (w Window) openingViewPopover() Window {
	w = w.filteringBoard()
	w.screen, w.ofView = OnKeeping, true
	return w
}

func (w Window) keepInput() engine.Input {
	v, _ := w.inView()
	return w.filters.Holding(v.View, w.looking.Reads())
}

func (w Window) keepingView(name string) Window {
	in := w.keepInput()
	in.Given["name"] = name
	kept, err := w.work.Ask(engine.ViewNew, in)
	if err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.viewing = kept.Viewed.Name
	w.filters, w.looking = w.filters.Cleared(), Written{}
	w.screen, w.ofView = OnTheBoard, false
	return w
}

func (w Window) savingToView() Window {
	v, in := w.inView()
	if !in || !w.ViewChanged() {
		return w
	}
	if _, err := w.work.Ask(engine.ViewEdit, w.filters.Saving(v.View, w.looking.Reads())); err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.filters, w.looking = w.filters.Cleared(), Written{}
	w.screen, w.ofView = OnTheBoard, false
	return w
}

func (w Window) onKeeping(name string) Window {
	if w.ofView {
		return w
	}
	if edit, ok := textKey(name); ok {
		w.keepName = edit(w.keepName)
		return w
	}
	switch name {
	case "backspace":
		w.keepName = w.keepName.Back()
	case "cmd+a":
		w.keepName = w.keepName.All()
	}
	return w
}

// KeepingBox is the popover dropped under the pill it opened from: a view about to be kept, its name typed and what it would keep said, or the view in force, what it holds and what it shows against the store.
func (w Window) KeepingBox() Box {
	box := w.droppedUnder(w.filterAnchor(), max(keepingWide, w.footWide()), 0)
	box.H = dropHead + w.keepingFacts().High() + dropFoot + 2*padding
	if !w.ofView {
		box.H += menuRow + padding
	}
	return box
}

// KeepingField is where the name of a view about to be kept is typed.
func (w Window) KeepingField() Box {
	box := w.KeepingBox()
	return Box{X: box.X + padding, Y: box.Y + dropHead + padding/2, W: box.W - 2*padding, H: menuRow}
}

// KeepingFactsAt is where the popover says what the view holds and keeps.
func (w Window) KeepingFactsAt() Box {
	box := w.KeepingBox()
	y := box.Y + dropHead + padding/2
	if !w.ofView {
		y = w.KeepingField().Y + menuRow + padding
	}
	return Box{X: box.X + padding, Y: y, W: box.W - 2*padding, H: w.keepingFacts().High()}
}

// KeepingHead is what heads the popover: keeping a view, or the name of the view in force.
func (w Window) KeepingHead() string {
	if v, in := w.inView(); in && w.ofView {
		return "view " + v.Name
	}
	return KeepAsViewSays
}

func (w Window) keepingFacts() Facts {
	of := len(engine.OnTheBoard(w.snap.Cards, w.archived))
	if v, in := w.inView(); in && w.ofView {
		lines := []Fact{
			{Name: "holds", Value: engine.HoldsSays(v.View)},
			{Name: "shows", Value: strconv.Itoa(v.Keeps) + " of " + countOf(v.Of, "task") + " · " + strconv.Itoa(v.Done) + " done"},
		}
		for _, n := range v.Holds.Objective {
			if g, found := w.snap.Goal(n); found {
				lines = append(lines, Fact{Name: "#" + strconv.Itoa(n), Value: strconv.Itoa(g.Done) + "/" + countOf(g.Total, "task") + " done, the store's", Aside: g.Title})
			}
		}
		if gone := engine.GoneSays(v.View); gone != "" {
			lines = append(lines, Fact{Name: "gone", Value: gone})
		}
		return Facts{Lines: lines}
	}
	lines := []Fact{{Name: "keeps", Value: strconv.Itoa(w.Holds(Tasks)) + " of " + countOf(of, "task")}}
	if said := w.filters.Proposed(w.snap, w.looking.Reads()); said != "" {
		lines = append(lines, Fact{Name: "holds", Value: said})
	}
	if key := viewKey(len(w.snap.Views) + 1); key != "" {
		lines = append(lines, Fact{Name: "key", Value: w.keyboard.Spelled(Pretty(key))})
	}
	return Facts{Lines: lines}
}

// KeepingFoot is the keys the foot of the popover says, each with what it does.
func (w Window) KeepingFoot() [][2]string {
	if !w.ofView {
		return [][2]string{{"⏎", "keep"}, {"esc", "close"}}
	}
	if w.ViewChanged() {
		return [][2]string{{"⏎", "save"}, {"v", "keep as new"}, {"esc", "close"}}
	}
	return [][2]string{{"esc", "close"}}
}

// ViewRow is one view listed at the head of the filter menu, or in the panel of the views: where it is, its name, how many tasks it keeps, the key opening it, what it holds, whether it is the one in force or the one chosen, whether its name is being typed, and whether something it holds is gone.
type ViewRow struct {
	At                Box
	Name, Holds, Key  string
	Keeps             int
	On, Taken, Typing bool
	Late              bool
}

func (w Window) viewsAbove() int {
	if len(w.snap.Views) == 0 {
		return 0
	}
	return len(w.snap.Views) + 2
}

func rowOf(v engine.Viewing) ViewRow {
	return ViewRow{Name: v.Name, Keeps: len(v.Kept), Key: viewKey(v.Place), Holds: engine.HoldsSays(v.View), Late: len(v.Gone) > 0}
}

// ViewRows is the views listed at the head of the open filter menu, the one in force taken.
func (w Window) ViewRows() []ViewRow {
	if w.screen != OnFiltering || w.viewsAbove() == 0 {
		return nil
	}
	box := w.filtersColumn()
	if box.W == 0 {
		return nil
	}
	var rows []ViewRow
	for i, v := range w.snap.Views {
		row := rowOf(v)
		row.At, row.Taken, row.On = DropRowAt(box, i), v.Name == w.viewing, i+1 == w.viewWalked
		rows = append(rows, row)
	}
	return rows
}

// EditViewsAt is the line under the views of the filter menu that opens their panel.
func (w Window) EditViewsAt() Box { return DropRowAt(w.filtersColumn(), len(w.snap.Views)) }

// FilterHeadAt is the line naming the filters under the views of the filter menu.
func (w Window) FilterHeadAt() Box { return DropRowAt(w.filtersColumn(), len(w.snap.Views)+1) }

func (w Window) clickingViews(p Point) (Window, bool) {
	for _, row := range w.ViewRows() {
		if v, found := w.snap.View(row.Name); found && row.At.In(p) {
			return w.openingView(v.Place), true
		}
	}
	if w.viewsAbove() > 0 && w.EditViewsAt().In(p) {
		return w.listingViews(), true
	}
	return w, false
}

func (w Window) listingViews() Window {
	w.screen, w.renamingView = OnViews, false
	w.viewOn = max(0, min(w.viewOn, len(w.snap.Views)-1))
	return w
}

func (w Window) viewsOpens() Box {
	var said []string
	for _, v := range w.snap.Views {
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
	if w.screen != OnViews {
		return nil
	}
	box := w.ViewsBox()
	var rows []ViewRow
	for i, v := range w.snap.Views {
		row := rowOf(v)
		row.At, row.On, row.Taken = ListRowAt(box, i), i == w.viewOn, v.Name == w.viewing
		row.Typing = row.On && w.renamingView
		rows = append(rows, row)
	}
	return rows
}

func (w Window) viewsWarned() (string, Box, bool) {
	if w.viewOn >= len(w.snap.Views) {
		return "", Box{}, false
	}
	gone := engine.GoneSays(w.snap.Views[w.viewOn].View)
	return gone, Box{}, gone != ""
}

// ViewsWarning is what the panel of the views warns of the view chosen, where, and whether it warns at all: what it holds that is gone.
func (w Window) ViewsWarning() (string, Box, bool) {
	said, _, warns := w.viewsWarned()
	if !warns || w.screen != OnViews {
		return "", Box{}, false
	}
	box := w.ViewsBox()
	row := ListRowAt(box, len(w.snap.Views))
	return said, Box{X: row.X, Y: row.Y + between, W: row.W, H: BannerHigh(w.measured(), said, viewsThenSays, row.W)}, true
}

func (w Window) chosenView() (engine.Viewing, bool) {
	if w.viewOn >= len(w.snap.Views) {
		return engine.Viewing{}, false
	}
	return w.snap.Views[w.viewOn], true
}

func (w Window) onViews(name string) Window {
	if w.renamingView {
		switch name {
		case "backspace":
			w.viewRenamed = w.viewRenamed.Back()
		case "cmd+a":
			w.viewRenamed = w.viewRenamed.All()
		}
		return w
	}
	switch name {
	case "down", "j":
		w.viewOn = min(len(w.snap.Views)-1, w.viewOn+1)
	case "up", "k":
		w.viewOn = max(0, w.viewOn-1)
	case "shift+down":
		return w.movingView(1)
	case "shift+up":
		return w.movingView(-1)
	}
	return w
}

func (w Window) movingView(by int) Window {
	v, ok := w.chosenView()
	to := v.Place + by
	if !ok || to < 1 || to > len(w.snap.Views) {
		return w
	}
	in := engine.Input{Given: map[string]string{"view": v.Name, "place": strconv.Itoa(to)}}
	if _, err := w.work.Ask(engine.ViewEdit, in); err != nil {
		return w.failing(err)
	}
	w = w.Look()
	w.viewOn = to - 1
	return w
}

func (w Window) renamingTheView() Window {
	v, ok := w.chosenView()
	if !ok {
		return w
	}
	in := engine.Input{Given: map[string]string{"view": v.Name, "name": w.viewRenamed.Reads()}}
	renamed, err := w.work.Ask(engine.ViewEdit, in)
	if err != nil {
		return w.failing(err)
	}
	if w.viewing == v.Name {
		w.viewing = renamed.Viewed.Name
	}
	w = w.Look()
	w.renamingView = false
	return w
}

func (w Window) deletingView() Window {
	v, ok := w.chosenView()
	if !ok {
		return w
	}
	if _, err := w.work.Ask(engine.ViewDelete, engine.Input{Given: map[string]string{"view": v.Name}}); err != nil {
		return w.failing(err)
	}
	if w.viewing == v.Name {
		w = w.leavingView()
	}
	w = w.Look().listingViews()
	return w.telling("the view " + v.Name + " was deleted, and no task with it")
}

func (w Window) choosingOnViews(offer string) (Window, bool) {
	switch strings.ToLower(offer) {
	case strings.ToLower(KeepTheViewSays):
		return w.keepingView(strings.TrimSpace(w.keepName.Reads())), true
	case strings.ToLower(SaveToViewSays):
		return w.savingToView(), true
	case strings.ToLower(KeepAsNewSays):
		w.screen = OnTheBoard
		return w.startingKeep(), true
	case "open":
		v, ok := w.chosenView()
		if !ok {
			return w, true
		}
		return w.openingView(v.Place), true
	case "rename":
		if v, ok := w.chosenView(); ok {
			w.renamingView, w.viewRenamed = true, Ending(v.Name).All()
		}
		return w, true
	case "save":
		return w.renamingTheView(), true
	case "delete":
		return w.deletingView(), true
	case "cancel", "close":
		if w.renamingView {
			w.renamingView = false
			return w, true
		}
		w.screen, w.ofView = OnTheBoard, false
		return w, true
	}
	return w, false
}

func (w Window) clickingKeeping(p Point) Window {
	if !w.KeepingBox().In(p) {
		w.screen, w.ofView = OnTheBoard, false
	}
	return w
}

func (w Window) clickingViewsPanel(p Point) Window {
	for i, row := range w.ViewsPanelRows() {
		if row.At.In(p) && !row.Typing {
			w.viewOn, w.renamingView = i, false
			return w
		}
	}
	if !w.ViewsBox().In(p) {
		w.screen, w.renamingView = OnTheBoard, false
	}
	return w
}

func (w Window) viewRows() []PaletteRow {
	var rows []PaletteRow
	for _, v := range w.snap.Views {
		place := v.Place
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: "View: " + v.Name, Key: viewKey(place), does: func(w Window) Window { return w.openingView(place) }})
	}
	if w.viewing != "" {
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: WholeBoardSays, Key: "cmd+0", does: Window.leavingView})
	}
	if w.keepable() {
		rows = append(rows, PaletteRow{Group: paletteActionsGroup, Says: KeepAsViewSays, does: Window.startingKeep})
	}
	return append(rows, PaletteRow{Group: paletteActionsGroup, Says: ViewsSays + "…", does: Window.listingViews})
}

func (w Window) viewCommandRows(typed string) []PaletteRow {
	name, ok := strings.CutPrefix(strings.TrimSpace(typed), "view new ")
	if name = strings.TrimSpace(name); !ok || name == "" || !w.keepable() {
		return nil
	}
	return []PaletteRow{{Group: paletteActionsGroup, Says: "Keep the filters as the view " + name, does: func(w Window) Window {
		return w.keepingView(name)
	}}}
}

// ViewsBox is the panel of the views: across the middle of the window, a line a view, and the warning of the one chosen when something it holds is gone.
func (w Window) ViewsBox() Box { return w.placedPanel(OnViews, w.viewsOpens()) }
