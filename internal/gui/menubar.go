package gui

import "strings"

// MenuGroup is one menu of the bar: its name and its lines.
type MenuGroup struct {
	Name  string
	Items []string
}

const menuButtonWide = 40

// MenuBar is every menu of the bar, in the order a window lays them out.
func MenuBar() []MenuGroup {
	return []MenuGroup{
		{Name: "File", Items: []string{"New", "New backup", "Backups", "Labels", "Export", "Import"}},
		{Name: "Edit", Items: []string{"Edit", "Focus", "Archive", "Delete", "Search"}},
		{Name: "View", Items: []string{"Tasks", "Objectives", "Time", "Archived"}},
		{Name: "Help", Items: []string{"Help", "About cairn"}},
	}
}

// Picked does what a line of the menu bar offers, from the board.
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
	for _, group := range MenuBar() {
		w.menu = append(w.menu, group.Items...)
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
