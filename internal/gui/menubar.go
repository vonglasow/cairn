package gui

// The window's menu bar: every act the window has, grouped the way a
// window's menus are, so a person who looks for them in the menu bar finds
// them there. Nothing in it is new — each line is an offer the bar, the help
// or a key already makes, done the same way. On macOS and Windows it is the
// system's own menu bar (menu_darwin.go, menu_windows.go); where the system
// gives a window none, a button at the end of the tabs opens the same lines.
// See docs/gui.md.

import "strings"

// MenuGroup is one menu of the bar: its name and its lines.
type MenuGroup struct {
	Name  string
	Items []string
}

// menuButtonWide is how wide the button standing in for a menu bar is.
const menuButtonWide = 40

// MenuBar is every menu of the bar, in the order a window lays them out.
func MenuBar() []MenuGroup {
	return []MenuGroup{
		{Name: "File", Items: []string{"New", "Back up", "Backups", "Export", "Import"}},
		{Name: "Edit", Items: []string{"Edit", "Focus", "Archive", "Delete", "Search"}},
		{Name: "View", Items: []string{"Tasks", "Objectives", "Time", "Archived"}},
		{Name: "Help", Items: []string{"Help"}},
	}
}

// Picked does what a line of the menu bar offers, from the board. Over a
// panel it does nothing, as a key the panel does not take does nothing: a
// form left behind for a page picked in the menu would be a form lost.
func (w Window) Picked(item string) Window {
	if w.screen != OnTheBoard && w.screen != OnSearch {
		return w
	}
	w.screen = OnTheBoard
	if item == "Focus" {
		return w.working()
	}
	return w.Choose(item)
}

// Native says the system gave the window a menu bar of its own, so no button
// stands in for one.
func (w Window) Native(given bool) Window {
	w.native = given
	return w
}

// MenuButton is the button standing in for a menu bar, at the end of the tabs.
func MenuButton(l Layout) Box {
	return Box{X: l.Tabs.X + l.Tabs.W - menuButtonWide, Y: l.Tabs.Y + between, W: menuButtonWide, H: l.Tabs.H - between}
}

// MenuStrokes are the three strokes the button is drawn as, across its
// middle.
func MenuStrokes(button Box) [3]Box {
	wide, mid := button.W/2, button.Y+button.H/2
	at := func(y int) Box { return Box{X: button.X + button.W/4, Y: y, W: wide, H: 2} }
	return [3]Box{at(mid - 6), at(mid - 1), at(mid + 4)}
}

// openingMenuBar opens every line of the menu bar as one menu, under its
// button.
func (w Window) openingMenuBar() Window {
	button := MenuButton(w.Layout())
	w.menu = nil
	for _, group := range MenuBar() {
		w.menu = append(w.menu, group.Items...)
	}
	// Its right edge under the button's, so it opens into the window.
	w.menuAt, w.menuOn, w.screen = Point{X: button.X + button.W - menuWide, Y: button.Y + button.H}, 0, OnMenu
	return w
}

// KeyOf is the key that does what a line of the menu bar does, spelled for
// this keyboard, or nothing when no key does.
func (k Keyboard) KeyOf(item string) string {
	key := Shortcut(item)
	if key == "" {
		return ""
	}
	return k.Spelled(Pretty(key))
}

// What the system's menu bars need decided, kept here where a test reaches
// it: the code that calls the system only calls it.

// made says whether a handle the system handed back names something.
func made(handle uintptr) bool { return handle != 0 }

// handedOver gives the loop a line picked in the system's menu bar, and asks
// for the frame that takes it on the side. The system calls on its own
// thread, and on macOS asking for a frame from there is answered there and
// then, inside the lock the asking holds: asked in place, it never returned.
// A loop with picks still to take loses the line rather than hold the system.
func handedOver(line string, picks chan<- string, wake func()) {
	select {
	case picks <- line:
	default:
	}
	go wake()
}

// lineAt is the line of the menu bar a number names, if any.
func lineAt(lines []string, at int) (string, bool) {
	if at < 0 || at >= len(lines) {
		return "", false
	}
	return lines[at], true
}

// commandOf is the number Win32 hands back for the i-th line: never 0,
// which it hands back for nothing.
func commandOf(i int) uintptr { return uintptr(i + 1) }

// commandLine is the line a Win32 message names, when it is a command from
// the menu bar: the low word of its argument, and a high word of nought.
func commandLine(msg, wParam uintptr) (int, bool) {
	const command = 0x0111
	if msg != command || wParam>>16 != 0 {
		return 0, false
	}
	return int(wParam&0xffff) - 1, true
}

// winTitle is a line as a Win32 menu writes it: the key that does the same
// after a tab, which Windows lines up on the right.
func winTitle(item string) string {
	if key := KeyboardOf("windows").KeyOf(item); key != "" {
		return item + "\t" + key
	}
	return item
}

// winMenu is a menu's name as a Win32 menu bar writes it, its first letter
// the one alt reaches it by.
func winMenu(name string) string { return "&" + name }

// macKey is the key a line answers in a macOS menu: a ⌘ chord's letter, and
// nothing else — a plain letter taken by the menu bar would never reach a
// field typed into.
func macKey(item string) string {
	chord, _ := strings.CutPrefix(Shortcut(item), "cmd+")
	if chord == Shortcut(item) {
		return ""
	}
	return chord
}
