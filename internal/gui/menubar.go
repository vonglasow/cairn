package gui

import (
	"slices"
	"strings"
	"sync"

	"cairn/internal/engine"
)

// MenuGroup is one menu of the bar: its name and its lines.
type MenuGroup struct {
	Name  string
	Items []string
}

const menuButtonWide = 40

var menuBar = []MenuGroup{
	{Name: "File", Items: []string{"New", "New backup", "Backups", "Labels", "Export", "Import", "Configuration…"}},
	{Name: "Edit", Items: []string{"Edit", "Focus", "Archive", "Delete", "Search"}},
	{Name: "View", Items: []string{"Tasks", "Objectives", "Time", "Archived", "Filter", "Reset sizes"}},
	{Name: "Help", Items: []string{"Help", "About cairn"}},
}

var actsOf = map[string][]engine.Operation{
	"New":        {engine.TaskNew, engine.ObjectiveNew},
	"New backup": {engine.BackupNew},
	"Backups":    {engine.BackupList},
	"Labels":     {engine.LabelList},
	"Export":     {engine.TaskExport},
	"Import":     {engine.TaskImport},
	"Edit":       {engine.TaskEdit, engine.ObjectiveEdit},
	"Focus":      {engine.TaskFocus},
	"Archive":    {engine.TaskArchive},
	"Delete":     {engine.TaskDelete, engine.ObjectiveDelete},
	"Search":     {engine.TaskSearch, engine.ObjectiveSearch},
	"Archived":   {engine.TaskList},
	"Filter":     {engine.TaskList},
}

var pageOfLine = map[string]string{"Tasks": engine.PageTasks, "Objectives": engine.PageObjectives, "Time": engine.PageTime}

func offered(pages engine.Pages, line string) bool {
	if page, isOne := pageOfLine[line]; isOne {
		return pages.Shows(page)
	}
	acts, acting := actsOf[line]
	return !acting || slices.ContainsFunc(acts, pages.Offers)
}

// MenuBar is every menu of the bar, in the order a window lays them out, with only the lines the pages shown offer: a line acting on a page turned off is gone, a menu left with no line goes too — the same rule as the command line's and the MCP server's (engine.Pages) — and with every page off there is none, the window holding nothing to act on.
func MenuBar(pages engine.Pages) []MenuGroup {
	if pages.None() {
		return nil
	}
	var bar []MenuGroup
	for _, group := range menuBar {
		var lines []string
		for _, line := range group.Items {
			if offered(pages, line) {
				lines = append(lines, line)
			}
		}
		if len(lines) > 0 {
			bar = append(bar, MenuGroup{Name: group.Name, Items: lines})
		}
	}
	return bar
}

// Picked does what a line of the menu bar offers, from the board.
func (w Window) Picked(item string) Window {
	if !w.Doable(item) {
		return w
	}
	w.screen = OnTheBoard
	if item == "Focus" {
		return w.working()
	}
	return w.Choose(item)
}

var tabOfLine = map[string]Tab{"Tasks": Tasks, "Objectives": Objectives, "Time": Time}

// Doable says whether a line of the menu bar does something now: offered by the pages shown, on the board, and with what it acts on in sight — a card chosen to edit, a task to focus or archive, another page to go to. The native bars grey what is not, the button's menu leaves it out, and picking it does nothing.
func (w Window) Doable(line string) bool {
	if w.Blank() || w.screen != OnTheBoard && w.screen != OnSearch || !offered(w.pages, line) {
		return false
	}
	onTasks := w.tab == Tasks
	chosenTask := onTasks && w.chosen != 0 && !w.chosenGoal
	switch line {
	case "Tasks", "Objectives", "Time":
		return w.tab != tabOfLine[line]
	case "New":
		if w.tab == Objectives {
			return w.pages.Offers(engine.ObjectiveNew)
		}
		return w.pages.Offers(engine.TaskNew)
	case "Edit":
		return w.chosen != 0
	case "Delete":
		return w.chosen != 0 || len(w.Taken()) > 0
	case "Focus":
		card, found := w.snap.Card(w.chosen)
		return chosenTask && found && !engine.Finished(card)
	case "Archive":
		return chosenTask || onTasks && len(w.Taken()) > 0
	case "Archived", "Reset sizes":
		return onTasks
	}
	return true
}

var doableNow struct {
	sync.Mutex
	lines map[string]bool
}

// tellingTheBars hands the native menu bars what each line can do now; they read it, from the system's thread, as a menu opens.
func (w Window) tellingTheBars() {
	lines := map[string]bool{}
	for _, group := range menuBar {
		for _, line := range group.Items {
			lines[line] = w.Doable(line)
		}
	}
	doableNow.Lock()
	defer doableNow.Unlock()
	doableNow.lines = lines
}

func lineDoable(line string) bool {
	doableNow.Lock()
	defer doableNow.Unlock()
	return doableNow.lines[line]
}

// Native says the system gave the window a menu bar of its own, so no button stands in for one.
func (w Window) Native(given bool) Window {
	w.native = given
	return w
}

// MenuButton is the button standing in for a menu bar, at the end of the tabs.
func MenuButton(l Layout) Box {
	return Box{X: l.Tabs.X + l.Tabs.W - menuButtonWide, Y: l.Tabs.Y + between, W: menuButtonWide, H: l.Tabs.H - between}
}

// MenuStrokes are the three strokes the button is drawn as, across its middle.
func MenuStrokes(button Box) [3]Box {
	wide, mid := button.W/2, button.Y+button.H/2
	at := func(y int) Box { return Box{X: button.X + button.W/4, Y: y, W: wide, H: 2} }
	return [3]Box{at(mid - 6), at(mid - 1), at(mid + 4)}
}

func (w Window) openingMenuBar() Window {
	button := MenuButton(w.Layout())
	w.menu = nil
	for _, group := range MenuBar(w.pages) {
		for _, line := range group.Items {
			if w.Doable(line) {
				w.menu = append(w.menu, line)
			}
		}
	}
	w.menuAt, w.menuOn, w.screen = Point{X: button.X + button.W - menuWide, Y: button.Y + button.H}, 0, OnMenu
	return w
}

// KeyOf is the key that does what a line of the menu bar does, spelled for this keyboard, or nothing when no key does.
func (k Keyboard) KeyOf(item string) string {
	key := Shortcut(item)
	if key == "" {
		return ""
	}
	return k.Spelled(Pretty(key))
}

func made(handle uintptr) bool { return handle != 0 }

func handedOver(line string, picks chan<- string, wake func()) {
	select {
	case picks <- line:
	default:
	}
	go wake()
}

func lineAt(lines []string, at int) (string, bool) {
	if at < 0 || at >= len(lines) {
		return "", false
	}
	return lines[at], true
}

func commandOf(i int) uintptr { return uintptr(i + 1) }

func commandLine(msg, wParam uintptr) (int, bool) {
	const command = 0x0111
	if msg != command || wParam>>16 != 0 {
		return 0, false
	}
	return int(wParam&0xffff) - 1, true
}

func winTitle(item string) string {
	if key := KeyboardOf("windows").KeyOf(item); key != "" {
		return item + "\t" + key
	}
	return item
}

func winMenu(name string) string { return "&" + name }

func macKey(item string) string {
	chord, _ := strings.CutPrefix(Shortcut(item), "cmd+")
	if chord == Shortcut(item) {
		return ""
	}
	return chord
}
