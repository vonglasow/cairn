package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Switcher opens the store of another workspace, a folder or a store opened before in place of the one a board holds, and says which pages the settings of the store now held show.
type Switcher func(to engine.Elsewhere) (engine.Pages, error)

// Switching gives the board the way to open another workspace in place of the one it shows; without one, ctrl+o does nothing and is not offered.
func (b Board) Switching(s Switcher) Board {
	b.switcher = s
	return b
}

var errNoSwitching = engine.Refusing("no-switching", "this board holds one store and cannot open another")

// Placed is where the store a board shows is: the workspace that names it, or its folder when none does.
type Placed struct {
	Shown engine.WorkspaceShown
	Err   error
}

// Listed is every workspace the configuration names, and the folders a board opened by path that none names, read when the switcher opens.
type Listed struct {
	Spaces []engine.Workspace
	Recent []engine.RecentFolder
	Err    error
}

type spaceTyping int

const (
	narrowing spaceTyping = iota
	aFolder
	aName
	itsFolder
	keepingAs
)

// Switched is the board read again once another workspace's store is held in place of the one it showed.
type Switched struct {
	Pages  engine.Pages
	Looked Looked
	Placed Placed
	Err    error
}

const (
	switcherKey  = "ctrl+o"
	maxPlace     = 9
	switcherWide = 76
)

func (b Board) roams() bool { return b.switcher != nil }

func (b Board) placing() tea.Cmd {
	return func() tea.Msg { return placedOf(b.work) }
}

func placedOf(w engine.Work) Placed {
	answer, err := w.Ask(engine.WorkspaceShow, engine.Input{})
	if err != nil || answer.WorkspaceShown == nil {
		return Placed{Err: err}
	}
	return Placed{Shown: *answer.WorkspaceShown}
}

func (b Board) listing() tea.Cmd {
	return func() tea.Msg {
		answer, err := b.work.Ask(engine.WorkspaceList, engine.Input{})
		if err != nil || answer.Workspaces == nil {
			return Listed{Err: err}
		}
		return Listed{Spaces: answer.Workspaces.Workspaces, Recent: answer.Workspaces.Recent}
	}
}

func (b Board) switchingTo(to engine.Elsewhere) tea.Cmd {
	return func() tea.Msg { return b.opening(to) }
}

func (b Board) opening(to engine.Elsewhere) Switched {
	if b.switcher == nil {
		return Switched{Err: errNoSwitching}
	}
	pages, err := b.switcher(to)
	if err != nil {
		return Switched{Err: err}
	}
	found, err := engine.Look(b.work)
	return Switched{Pages: pages, Looked: Looked{Snapshot: found, Err: err}, Placed: placedOf(b.work)}
}

func namedAt(name, folder string) engine.Input {
	return engine.Input{Given: map[string]string{"name": name, "folder": folder}}
}

func (b Board) making(name, folder string) tea.Cmd {
	return func() tea.Msg {
		if _, err := b.work.Ask(engine.WorkspaceAdd, namedAt(name, folder)); err != nil {
			return Switched{Err: err}
		}
		return b.opening(engine.Elsewhere{Workspace: name})
	}
}

func (b Board) keepingIt(name, folder string) tea.Cmd {
	return func() tea.Msg {
		if _, err := b.work.Ask(engine.WorkspaceAdd, namedAt(name, folder)); err != nil {
			return Listed{Err: err}
		}
		return b.listing()()
	}
}

func (b Board) listedAll(m Listed) Board {
	if m.Err != nil {
		b.err = m.Err
		return b
	}
	b.spaces, b.recent, b.sought, b.screen = m.Spaces, m.Recent, "", OnWorkspaces
	b.spaceAsk, b.spaceIn = narrowing, ""
	b.spaceAt = b.firstElsewhere()
	return b
}

func (b Board) firstElsewhere() int {
	for i, one := range b.listed() {
		if !b.isHere(one) {
			return i
		}
	}
	return 0
}

func (b Board) switched(m Switched) (tea.Model, tea.Cmd) {
	if m.Err != nil {
		b.err = m.Err
		return b, nil
	}
	fresh := New(b.work, b.glyphs, b.clock, b.width, b.height).Paged(m.Pages).Noting(b.noted).Switching(b.switcher)
	fresh.skin, fresh.place = b.skin, m.Placed
	fresh.opened = openedOver(m.Placed.Shown.Name(), b.place.Shown.Name(), b.snap.Focus)
	return fresh.update(m.Looked)
}

func openedOver(now, was string, focus []int) string {
	said := "opened " + now
	if was == "" {
		return said
	}
	said += " in place of " + was
	if len(focus) == 0 {
		return said
	}
	held := make([]string, 0, len(focus))
	for _, n := range focus {
		held = append(held, "#"+strconv.Itoa(n))
	}
	return said + " · " + was + "'s focus on " + strings.Join(held, ", ") + " is still held"
}

func (b Board) isHere(one engine.Workspace) bool {
	return one.Name == b.place.Shown.Workspace
}

func (b Board) listed() []engine.Workspace {
	var out []engine.Workspace
	for _, one := range b.spaces {
		if strings.Contains(one.Name, strings.ToLower(b.sought)) {
			out = append(out, one)
		}
	}
	return out
}

func (b Board) placeOf(one engine.Workspace) int {
	place := 0
	for _, other := range b.spaces {
		if b.isHere(other) {
			continue
		}
		place++
		if other.Name == one.Name {
			return place
		}
	}
	return 0
}

func (b Board) recentListed() []engine.RecentFolder {
	var out []engine.RecentFolder
	for _, one := range b.recent {
		if strings.Contains(one.Shown(), b.sought) {
			out = append(out, one)
		}
	}
	return out
}

func (b Board) recentChosen() (engine.RecentFolder, bool) {
	at := b.spaceAt - len(b.listed())
	recent := b.recentListed()
	if at < 0 || at >= len(recent) {
		return engine.RecentFolder{}, false
	}
	return recent[at], true
}

func (b Board) isHereStore(one engine.RecentFolder) bool {
	return b.place.Shown.Workspace == "" && one.Store == b.place.Shown.Store
}

func (b Board) leavingSwitcher() Board {
	b.screen, b.spaces, b.recent, b.sought = OnTheBoard, nil, nil, ""
	b.spaceAsk, b.spaceIn = narrowing, ""
	return b
}

func (b Board) typingIn(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		b.spaceAsk, b.spaceIn = narrowing, ""
		return b, nil
	case "backspace":
		b.spaceIn = trimLast(b.spaceIn)
		return b, nil
	case "enter":
		switch b.spaceAsk {
		case aFolder:
			return b, b.switchingTo(engine.Elsewhere{Folder: b.spaceIn})
		case aName:
			b.newSpace, b.spaceAsk, b.spaceIn = b.spaceIn, itsFolder, ""
			return b, nil
		case itsFolder:
			return b, b.making(b.newSpace, b.spaceIn)
		default:
			return b, b.keepingIt(b.spaceIn, b.keeping.Folder)
		}
	}
	if letter, ok := typed(key); ok {
		b.spaceIn += letter
	}
	return b, nil
}

func (b Board) inWorkspaces(key string) (tea.Model, tea.Cmd) {
	if b.spaceAsk != narrowing {
		return b.typingIn(key)
	}
	shown := b.listed()
	switch key {
	case "esc":
		return b.leavingSwitcher(), nil
	case switcherKey:
		b.spaceAsk, b.spaceIn = aFolder, ""
		return b, nil
	case "ctrl+n":
		b.spaceAsk, b.spaceIn = aName, ""
		return b, nil
	case "ctrl+k":
		if one, ok := b.recentChosen(); ok && one.Folder != "" {
			b.spaceAsk, b.spaceIn, b.keeping = keepingAs, "", one
		}
		return b, nil
	case goUp:
		b.spaceAt = max(0, b.spaceAt-1)
		return b, nil
	case goDown:
		b.spaceAt = min(len(shown)+len(b.recentListed())-1, b.spaceAt+1)
		return b, nil
	case "enter":
		if b.spaceAt < len(shown) {
			return b.choosingSpace(shown[b.spaceAt])
		}
		if one, ok := b.recentChosen(); ok {
			if b.isHereStore(one) {
				return b.leavingSwitcher(), nil
			}
			return b, b.switchingTo(engine.Elsewhere{Store: one.Store})
		}
		return b, nil
	case "backspace":
		b.sought = trimLast(b.sought)
		b.spaceAt = b.firstElsewhere()
		return b, nil
	}
	if place, err := strconv.Atoi(key); err == nil {
		for _, one := range shown {
			if b.placeOf(one) == place && place > 0 && place <= maxPlace {
				return b.choosingSpace(one)
			}
		}
		return b, nil
	}
	if letter, ok := typed(key); ok {
		b.sought += letter
		b.spaceAt = b.firstElsewhere()
	}
	return b, nil
}

func (b Board) choosingSpace(one engine.Workspace) (tea.Model, tea.Cmd) {
	if b.isHere(one) {
		return b.leavingSwitcher(), nil
	}
	return b, b.switchingTo(engine.Elsewhere{Workspace: one.Name})
}

func (b Board) wearing(colour string) lipgloss.Style {
	ink, ok := domain.LabelInks[colour]
	if !ok {
		return b.skin.Crumb
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", ink[0], ink[1], ink[2])))
}

func (b Board) placeMark() string {
	if b.place.Shown.Workspace != "" {
		return b.wearing(b.place.Shown.Colour).Render("● ") + b.skin.Crumb.Render(b.place.Shown.Workspace) + "  "
	}
	if b.place.Shown.Store == "" {
		return ""
	}
	return b.skin.Faint.Render("◌ "+clipStart(filepath.Dir(b.place.Shown.Store), b.width/3)) + "  "
}

func clipStart(s string, width int) string {
	runes := []rune(s)
	if len(runes) <= width || width < 2 {
		return s
	}
	return ellipsis + string(runes[len(runes)-width+1:])
}

func (b Board) drawSwitcher() string {
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(b.drawBoard()),
		lipgloss.NewLayer(b.switcherBox()).X(margin+1).Y(2).Z(1),
	).Render()
}

func (b Board) switcherBox() string {
	wide := min(switcherWide, b.width-margin-2)
	inner := wide - 2
	edge := b.skin.Rule
	title := "─ OPEN A WORKSPACE "
	top := edge.Render("┌") + b.skin.Title.Render(title) + edge.Render(strings.Repeat("─", max(0, inner-lipgloss.Width(title)))+"┐")
	lines := []string{top}
	row := func(s string) {
		s = Clip(s, inner)
		lines = append(lines, edge.Render("│")+s+strings.Repeat(" ", inner-lipgloss.Width(s))+edge.Render("│"))
	}
	row(" " + b.skin.Prompt.Render(b.typingLabel()+"› "+b.typedLine()) + b.skin.Key.Render("▏"))
	row("")
	body, at := b.switcherBody(inner)
	room := max(1, b.height-14)
	from := max(0, at-room+1)
	for _, line := range body[from:min(len(body), from+room)] {
		row(line)
	}
	row("")
	row(" " + b.switcherActions())
	lines = append(lines, edge.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(lines, "\n")
}

func (b Board) typingLabel() string {
	switch b.spaceAsk {
	case aFolder:
		return "folder "
	case aName:
		return "new workspace "
	case itsFolder:
		return b.newSpace + "'s folder "
	case keepingAs:
		return "keep as "
	}
	return ""
}

func (b Board) typedLine() string {
	if b.spaceAsk == narrowing {
		return b.sought
	}
	return b.spaceIn
}

func (b Board) switcherBody(inner int) ([]string, int) {
	body := []string{" " + b.skin.Head.Render("WORKSPACES")}
	shown := b.listed()
	if len(shown) == 0 {
		body = append(body, "   "+b.skin.Faint.Render("none is named "+b.sought))
	}
	at := 0
	for i, one := range shown {
		if i == b.spaceAt {
			at = len(body)
		}
		body = append(body, b.spaceLine(one, i == b.spaceAt, inner))
	}
	recent := b.recentListed()
	if len(recent) == 0 {
		return body, at
	}
	body = append(body, "", " "+b.skin.Head.Render("RECENT FOLDERS"))
	for i, one := range recent {
		chosen := len(shown)+i == b.spaceAt
		if chosen {
			at = len(body)
		}
		body = append(body, b.recentLine(one, chosen, inner))
	}
	return body, at
}

func (b Board) recentLine(one engine.RecentFolder, chosen bool, inner int) string {
	cursor := "   "
	if chosen {
		cursor = " " + b.skin.Key.Render("›") + " "
	}
	marks := []string{one.Said(b.clock())}
	if b.isHereStore(one) {
		marks = append([]string{"here"}, marks...)
	}
	right := b.skin.Faint.Render(strings.Join(marks, " · ")) + " "
	room := inner - 3 - 2 - lipgloss.Width(right) - 1
	return Spread(cursor+b.skin.Faint.Render("◌ "+clipStart(one.Shown(), max(2, room))), right, inner)
}

func (b Board) switcherActions() string {
	if b.spaceAsk == aFolder || b.spaceAsk == itsFolder {
		return b.skin.Faint.Render("absolute, or under ~") + "   " + b.offer("esc", "back to the list")
	}
	if b.spaceAsk == keepingAs {
		return b.offer("⏎", "keep "+clipStart(b.keeping.Folder, max(2, b.width/3))) + "   " + b.offer("esc", "back to the list")
	}
	if b.spaceAsk != narrowing {
		return b.offer("⏎", "next") + "   " + b.offer("esc", "back to the list")
	}
	said := b.offer("^O", "open a folder") + "   " + b.offer("^N", "new workspace")
	if one, ok := b.recentChosen(); ok && one.Folder != "" {
		said += "   " + b.offer("^K", "keep as a workspace")
	}
	return said
}

func (b Board) spaceLine(one engine.Workspace, chosen bool, inner int) string {
	wide := 0
	for _, other := range b.spaces {
		wide = max(wide, len(other.Name))
	}
	cursor := "   "
	if chosen {
		cursor = " " + b.skin.Key.Render("›") + " "
	}
	name := b.wearing(one.Colour).Render("● ") + b.skin.Crumb.Render(pad(one.Name, wide))
	if !one.Found() {
		name = b.skin.Faint.Render("● " + pad(one.Name, wide))
	}
	var marks []string
	if b.isHere(one) {
		marks = append(marks, "here")
	}
	if one.Default {
		marks = append(marks, "default")
	}
	if !one.Found() {
		marks = append(marks, "not found")
	}
	if place := b.placeOf(one); place > 0 && place <= maxPlace {
		marks = append(marks, strconv.Itoa(place))
	}
	right := b.skin.Faint.Render(strings.Join(marks, " · ")) + " "
	room := inner - 3 - 2 - wide - 2 - lipgloss.Width(right) - 1
	folder := b.skin.Faint.Render(clipStart(one.Folder, room))
	if room < 2 {
		folder = ""
	}
	return Spread(cursor+name+"  "+folder, right, inner)
}

func (b Board) switcherBar() string {
	return b.bar(b.offer("↑↓", "choose"), b.offer("⏎", "open here"), b.offer("1-9", "by place"), b.offer("esc", "back"))
}
