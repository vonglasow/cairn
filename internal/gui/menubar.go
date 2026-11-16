package gui

import (
	"cairn/internal/board"
	"cairn/internal/cli/render"
	"cairn/internal/gui/geom"
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
	{Name: "cairn", Items: []string{"About cairn", "Settings…", "Upgrade", "Quit"}},
	{Name: "Help", Items: []string{"Help"}},
}

var buttonMenu = [][]string{{"Settings…", "Upgrade"}, {"Help", "About cairn"}, {"Quit"}}

func offered(pages engine.Pages, line string) bool {
	return line != "Upgrade" || pages.Offers(engine.Upgrade)
}

// MenuBar is every menu of the bar, in the order a window lays them out, with only the lines the pages shown offer; with every page off there is only Settings…, the window holding nothing to act on.
func MenuBar(pages engine.Pages) []MenuGroup {
	if pages.None() {
		return []MenuGroup{{Name: "cairn", Items: []string{"Settings…"}}}
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
	w.Screen = board.OnTheBoard
	if item == "Upgrade" {
		return w.upgrading()
	}
	return w.Choose(item)
}

// Doable says whether a line of the menu bar does something now: offered by the pages shown, and on the board. The native bars grey what is not, the button's menu leaves it out, and picking it does nothing.
func (w Window) Doable(line string) bool {
	if w.Blank() {
		return line == "Settings…" && w.Screen == board.OnTheBoard
	}
	return (w.Screen == board.OnTheBoard || w.Screen == board.OnSearch) && offered(w.Pages, line)
}

func (w Window) upgrading() Window {
	answer, err := w.work.Ask(engine.Upgrade, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	return w.telling(strings.TrimSpace(render.Text(engine.Upgrade, answer)))
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
func MenuButton(l geom.Layout) geom.Box {
	return geom.Box{X: l.Tabs.X + l.Tabs.W - menuButtonWide, Y: l.Tabs.Y + between, W: menuButtonWide, H: l.Tabs.H - between}
}

// MenuStrokes are the three strokes the button is drawn as, across its middle.
func MenuStrokes(button geom.Box) [3]geom.Box {
	wide, mid := button.W/2, button.Y+button.H/2
	at := func(y int) geom.Box { return geom.Box{X: button.X + button.W/4, Y: y, W: wide, H: 2} }
	return [3]geom.Box{at(mid - 6), at(mid - 1), at(mid + 4)}
}

func (w Window) openingMenuBar() Window {
	button := MenuButton(w.Layout())
	w.menu, w.rules = nil, nil
	for _, group := range buttonMenu {
		starts := len(w.menu)
		for _, line := range group {
			if w.Doable(line) {
				w.menu = append(w.menu, line)
			}
		}
		if starts > 0 && len(w.menu) > starts {
			w.rules = append(w.rules, starts)
		}
	}
	w.menuAt, w.menuOn, w.Screen = geom.Point{X: button.X + button.W - menuWide, Y: button.Y + button.H}, 0, board.OnMenu
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
	if shifted, ok := strings.CutPrefix(chord, "shift+"); ok {
		return strings.ToUpper(shifted)
	}
	return chord
}
