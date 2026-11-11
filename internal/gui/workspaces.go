package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
)

type spaceAsking int

const (
	narrowing spaceAsking = iota
	aFolder
	aName
	itsFolder
	keepingAs
)

type switcherState struct {
	place     engine.WorkspaceShown
	placeRead bool
	spaces    []engine.Workspace
	recent    []engine.RecentFolder
	typed     Written
	on        int
	asking    spaceAsking
	newSpace  string
	keeping   string
	apart     *engine.Elsewhere
	named     string
	seeing    func(string) bool
}

const (
	// OpenFolderSays is the line of the switcher asking for a folder to open in a window of its own.
	OpenFolderSays = "Open a folder…"
	// NewWorkspaceSays is the line of the switcher making a workspace, then opening it in a window of its own.
	NewWorkspaceSays = "New workspace…"
	// KeepAsWorkspaceSays is the line of the switcher naming the recent folder the keyboard is on as a workspace.
	KeepAsWorkspaceSays = "Keep as a workspace…"
	// SwitcherSays is what the switcher's line says before anything is typed in it.
	SwitcherSays = "Open a workspace or a folder…"
	// WorkspacesHead and RecentHead head the two groups of the switcher.
	WorkspacesHead = "Workspaces"
	// RecentHead heads the stores opened by path that no workspace names.
	RecentHead  = "Recent folders"
	maxPlace    = 9
	placeWide   = 128
	placeLeast  = 70
	placeWords  = 24
	placeSlack  = 4
	placeCaret  = 10
	switchWide  = 520
	thisWindow  = "this window"
	openHere    = "open · brings it forward"
	openApart   = "opens a new window"
	notFoundNow = "not found"
)

// Seeing gives the window the way to know whether another window shows a store; without one, none does.
func (w Window) Seeing(seeing func(string) bool) Window {
	w.switcher.seeing = seeing
	return w
}

func (w Window) readingPlace() Window {
	if w.switcher.placeRead {
		return w
	}
	w.switcher.placeRead = true
	if got, err := w.work.Ask(engine.WorkspaceShow, engine.Input{}); err == nil && got.WorkspaceShown != nil {
		w.switcher.place = *got.WorkspaceShown
	}
	return w
}

// Place is the workspace the window shows, or the store no workspace names.
func (w Window) Place() engine.WorkspaceShown { return w.switcher.place }

// Title is what the window is called on the screen: cairn, and the workspace it shows, or the folder of a store no workspace names.
func (w Window) Title() string {
	if name := w.switcher.place.Name(); name != "" {
		return "cairn — " + name
	}
	return "cairn"
}

// PlaceSays is what the pill opening the top bar says: the workspace, or the folder of the store.
func (w Window) PlaceSays() string { return w.switcher.place.Name() }

// Loose says whether the store the window shows is one no workspace names.
func (w Window) Loose() bool { return w.switcher.place.Workspace == "" }

func (w Window) placeAt(l Layout) Box {
	return Box{X: l.Tabs.X, Y: l.Tabs.Y + between, W: placeRoom + w.placeTextWide(), H: l.Tabs.H - between}
}

const placeRoom = 2*padding + dotSide + 2*between + placeCaret

func (w Window) placeTextWide() int {
	return min(w.placeText(), ends(w.measured(), w.PlaceShown(), reads)+placeSlack)
}

func (w Window) placeText() int {
	room := min(placeWide, placeLeast+max(0, w.width-Narrowest)) - placeRoom
	if room < placeWords {
		return 0
	}
	return room
}

// PlaceShown is what the pill writes: the short name of the place, never its whole path; nothing when the pill has no room for words.
func (w Window) PlaceShown() string {
	if w.placeText() == 0 {
		return ""
	}
	return w.PlaceSays()
}

func (w Window) switcherPathAt(p Point) ([]string, Box, int) {
	if w.screen != OnWorkspaces {
		return nil, Box{}, 0
	}
	for _, line := range w.SwitcherLines() {
		if line.At.In(p) && line.Folder != "" {
			return []string{line.Folder}, line.At, line.At.W
		}
	}
	return nil, Box{}, 0
}

// PlaceParts is where the pill draws its dot, its name and its caret.
func PlaceParts(pill Box) (dot, name, caret Box) {
	dot = Box{X: pill.X + padding, Y: pill.Y + (pill.H-dotSide)/2, W: dotSide, H: dotSide}
	caret = Box{X: pill.X + pill.W - padding - placeCaret, Y: pill.Y + (pill.H-placeCaret)/2, W: placeCaret, H: placeCaret}
	x := dot.X + dot.W + between
	name = Box{X: x, Y: pill.Y + (pill.H-lineHigh)/2, W: caret.X - between - x, H: lineHigh}
	return dot, name, caret
}

func (w Window) openingSwitcher() Window {
	got, err := w.work.Ask(engine.WorkspaceList, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	w.switcher.spaces, w.switcher.recent = nil, nil
	if got.Workspaces != nil {
		w.switcher.spaces, w.switcher.recent = got.Workspaces.Workspaces, got.Workspaces.Recent
	}
	w.switcher.typed, w.switcher.on, w.switcher.asking = Written{}, 0, narrowing
	w.screen = OnWorkspaces
	return w
}

// SwitcherLine is one line of the switcher: a workspace, or a recent folder, what choosing it does, the key choosing it from anywhere, and where it is drawn.
type SwitcherLine struct {
	Name, Folder, Colour, Tag, Key string
	Recent, Here, Gone, On         bool
	At                             Box
	to                             engine.Elsewhere
	keeps                          string
}

func (w Window) isHere(store string) bool { return store == w.switcher.place.Store }

func (w Window) shownElsewhere(store string) bool {
	return w.switcher.seeing != nil && w.switcher.seeing(store)
}

func (w Window) spaceLine(place int, one engine.Workspace) SwitcherLine {
	line := SwitcherLine{Name: one.Name, Folder: one.Folder, Colour: one.Colour, Gone: one.NotFound, to: engine.Elsewhere{Workspace: one.Name}}
	var tags []string
	switch {
	case w.isHere(one.Store()):
		line.Here = true
		tags = append(tags, thisWindow)
	case one.NotFound:
		tags = append(tags, notFoundNow)
	case w.shownElsewhere(one.Store()):
		tags = append(tags, openHere)
	default:
		tags = append(tags, openApart)
	}
	if one.Default {
		tags = append(tags, "default")
	}
	line.Tag = strings.Join(tags, " · ")
	if place <= maxPlace {
		line.Key = Pretty("cmd+shift+" + strconv.Itoa(place))
	}
	return line
}

func (w Window) recentLine(one engine.RecentFolder) SwitcherLine {
	line := SwitcherLine{Name: one.Name(), Folder: one.Shown(), Recent: true, Gone: one.NotFound, Tag: one.Said(w.now), to: engine.Elsewhere{Store: one.Store}}
	if !one.NotFound {
		line.keeps = one.Folder
	}
	if w.isHere(one.Store) {
		line.Here, line.Tag = true, thisWindow+" · "+line.Tag
	}
	return line
}

func (w Window) switcherSought() string { return strings.ToLower(w.switcher.typed.Reads()) }

func (w Window) switcherChoices() []SwitcherLine {
	var lines []SwitcherLine
	for i, one := range w.switcher.spaces {
		if strings.Contains(one.Name, w.switcherSought()) {
			lines = append(lines, w.spaceLine(i+1, one))
		}
	}
	for _, one := range w.switcher.recent {
		if strings.Contains(strings.ToLower(one.Shown()), w.switcherSought()) {
			lines = append(lines, w.recentLine(one))
		}
	}
	return lines
}

// SwitcherBox is the switcher, dropped under the pill: its line, the workspaces, the recent folders, and its three acts.
func (w Window) SwitcherBox() Box {
	l := w.Layout()
	pill := l.Place
	lines := w.switcherChoices()
	rows := 2 + len(lines) + 1 + len(w.actsSaid())
	if w.hasRecent(lines) {
		rows++
	}
	top := pill.Y + pill.H + between
	most := max(4, (w.height-margin-top-2*inset)/menuRow)
	return Box{X: pill.X, Y: top, W: min(switchWide, w.width-2*margin), H: min(rows, most)*menuRow + 2*inset}
}

func (w Window) hasRecent(lines []SwitcherLine) bool {
	for _, line := range lines {
		if line.Recent {
			return true
		}
	}
	return false
}

func switcherRow(box Box, i int) Box {
	return Box{X: box.X + inset, Y: box.Y + inset + i*menuRow, W: box.W - 2*inset, H: menuRow}
}

// SwitcherField is the line at the switcher's head, where a name narrows it or a folder is typed.
func (w Window) SwitcherField() Box { return switcherRow(w.SwitcherBox(), 0) }

// SwitcherHeads is where the switcher's two groups are headed, the second only when a recent folder is listed.
func (w Window) SwitcherHeads() []Box {
	box := w.SwitcherBox()
	heads := []Box{switcherRow(box, 1)}
	lines := w.SwitcherLines()
	for _, line := range lines {
		if line.Recent {
			heads = append(heads, Box{X: line.At.X, Y: line.At.Y - menuRow, W: line.At.W, H: menuRow})
			break
		}
	}
	return heads
}

func (w Window) switcherRoom() int {
	box := w.SwitcherBox()
	headed := 3
	if w.hasRecent(w.switcherChoices()) {
		headed++
	}
	return max(1, (box.H-2*inset)/menuRow-headed-len(w.actsSaid()))
}

// SwitcherLines is every line the switcher shows, under its heads, where each is drawn and which one the keyboard is on; those past its room are not drawn.
func (w Window) SwitcherLines() []SwitcherLine {
	box := w.SwitcherBox()
	lines := w.switcherChoices()
	room := w.switcherRoom()
	from := max(0, min(w.switcher.on-room+1, len(lines)-room))
	row, headed := 2, false
	var shown []SwitcherLine
	for i, line := range lines {
		if line.Recent && !headed {
			row, headed = row+1, true
		}
		if i < from || i >= from+room {
			continue
		}
		line.On = i == w.switcher.on
		line.At = switcherRow(box, row)
		shown = append(shown, line)
		row++
	}
	return shown
}

// SwitcherAct is one of the three acts at the switcher's foot: what it says, the key doing it, and where it is drawn.
type SwitcherAct struct {
	Said, Key string
	At        Box
}

func (w Window) chosenLine() (SwitcherLine, bool) {
	lines := w.switcherChoices()
	if w.switcher.on < 0 || w.switcher.on >= len(lines) {
		return SwitcherLine{}, false
	}
	return lines[w.switcher.on], true
}

func (w Window) actsSaid() []SwitcherAct {
	acts := []SwitcherAct{{Said: OpenFolderSays}, {Said: NewWorkspaceSays}}
	if line, ok := w.chosenLine(); ok && line.keeps != "" {
		acts = append(acts, SwitcherAct{Said: KeepAsWorkspaceSays})
	}
	if w.switcher.place.Workspace != "" {
		acts = append(acts, SwitcherAct{Said: "Settings of " + w.switcher.place.Workspace + "…", Key: Pretty("cmd+,")})
	} else {
		acts = append(acts, SwitcherAct{Said: "Configuration…", Key: Pretty("cmd+,")})
	}
	return acts
}

// SwitcherActs is the switcher's foot: a folder opened, a workspace made, the recent folder the keyboard is on kept as one, the settings of this one.
func SwitcherActs(w Window) []SwitcherAct {
	acts := w.actsSaid()
	box := w.SwitcherBox()
	last := (box.H-2*inset)/menuRow - 1
	for i := range acts {
		acts[i].At = switcherRow(box, last-len(acts)+1+i)
	}
	return acts
}

// SwitcherAsks is what the switcher's line is asking for, said before what is typed: nothing while it narrows.
func (w Window) SwitcherAsks() string {
	switch w.switcher.asking {
	case aFolder:
		return "Folder"
	case aName:
		return "Name"
	case itsFolder:
		return w.switcher.newSpace + "'s folder"
	case keepingAs:
		return "Keep " + w.switcher.keeping + " as"
	}
	return ""
}

// SwitcherTyped is what is typed in the switcher's line.
func (w Window) SwitcherTyped() Written { return w.switcher.typed }

func (w Window) onWorkspaces(name string) Window {
	if w.switcher.asking != narrowing {
		return w.askingInSwitcher(name)
	}
	lines := w.switcherChoices()
	switch name {
	case "esc":
		w.screen = OnTheBoard
	case "up":
		w.switcher.on = max(0, w.switcher.on-1)
	case "down":
		w.switcher.on = min(len(lines)-1, w.switcher.on+1)
	case "enter":
		if w.switcher.on < len(lines) {
			return w.choosingLine(lines[w.switcher.on])
		}
	case "backspace":
		w.switcher.typed, w.switcher.on = w.switcher.typed.Back(), 0
	case "cmd+,":
		return w.openingConfiguration()
	default:
		if chosen, ok := w.byPlace(name); ok {
			return chosen
		}
	}
	return w
}

func (w Window) byPlace(name string) (Window, bool) {
	figure, ok := strings.CutPrefix(name, "cmd+shift+")
	place, err := strconv.Atoi(figure)
	if !ok || err != nil || place < 1 || place > min(maxPlace, len(w.switcher.spaces)) {
		return w, false
	}
	return w.choosingLine(w.spaceLine(place, w.switcher.spaces[place-1])), true
}

func (w Window) choosingByPlace(name string) Window {
	opened := w.openingSwitcher()
	if opened.err != nil {
		return opened
	}
	if chosen, ok := opened.byPlace(name); ok {
		return chosen
	}
	return w
}

func (w Window) choosingLine(line SwitcherLine) Window {
	if line.Here {
		w.screen = OnTheBoard
		return w
	}
	return w.askingApart(line.to, line.Name)
}

func (w Window) askingApart(to engine.Elsewhere, named string) Window {
	w.switcher.apart, w.switcher.named = &to, named
	return w
}

func (w Window) askingInSwitcher(name string) Window {
	switch name {
	case "esc":
		w.switcher.asking, w.switcher.typed = narrowing, Written{}
	case "backspace":
		w.switcher.typed = w.switcher.typed.Back()
	case "enter":
		typed := strings.TrimSpace(w.switcher.typed.Reads())
		switch w.switcher.asking {
		case aFolder:
			return w.askingApart(engine.Elsewhere{Folder: typed}, typed)
		case aName:
			w.switcher.newSpace, w.switcher.asking, w.switcher.typed = typed, itsFolder, Written{}
		case itsFolder:
			return w.makingWorkspace(w.switcher.newSpace, typed)
		case keepingAs:
			return w.keepingFolder(typed, w.switcher.keeping)
		}
	}
	return w
}

func (w Window) addingWorkspace(name, folder string) error {
	_, err := w.work.Ask(engine.WorkspaceAdd, engine.Input{Given: map[string]string{"name": name, "folder": folder}})
	return err
}

func (w Window) makingWorkspace(name, folder string) Window {
	if err := w.addingWorkspace(name, folder); err != nil {
		return w.failing(err)
	}
	return w.askingApart(engine.Elsewhere{Workspace: name}, name)
}

func (w Window) keepingFolder(name, folder string) Window {
	if err := w.addingWorkspace(name, folder); err != nil {
		return w.failing(err)
	}
	return w.openingSwitcher()
}

func (w Window) actingInSwitcher(said string) Window {
	switch said {
	case OpenFolderSays:
		w.switcher.asking, w.switcher.typed = aFolder, Written{}
		w.fileAsk = &FileAsk{Folder: true}
	case NewWorkspaceSays:
		w.switcher.asking, w.switcher.typed = aName, Written{}
	case KeepAsWorkspaceSays:
		line, _ := w.chosenLine()
		w.switcher.asking, w.switcher.typed, w.switcher.keeping = keepingAs, Written{}, line.keeps
	default:
		return w.openingConfiguration()
	}
	return w
}

func (w Window) clickingWorkspaces(p Point) Window {
	if !w.SwitcherBox().In(p) {
		w.screen = OnTheBoard
		return w
	}
	for _, line := range w.SwitcherLines() {
		if line.At.In(p) {
			return w.choosingLine(line)
		}
	}
	for _, act := range SwitcherActs(w) {
		if act.At.In(p) {
			return w.actingInSwitcher(act.Said)
		}
	}
	return w
}

// ApartAsked is the workspace, folder or store the window wants shown in a window of its own, if any.
func (w Window) ApartAsked() (engine.Elsewhere, bool) {
	if w.switcher.apart == nil {
		return engine.Elsewhere{}, false
	}
	return *w.switcher.apart, true
}

// ApartTaken says the loop has taken the window asked for in hand.
func (w Window) ApartTaken() Window {
	w.switcher.apart = nil
	return w
}

// Went takes what came of asking for a window elsewhere: said, and the switcher closed; a refusal is said and the switcher left as it was, what was typed kept.
func (w Window) Went(g Went) Window {
	if g.Err != nil {
		return w.failing(g.Err)
	}
	if w.screen == OnWorkspaces {
		w.screen = OnTheBoard
	}
	if g.Raised {
		return w.telling(w.switcher.named + " is brought forward")
	}
	return w.telling(w.switcher.named + " opens in a window of its own")
}
