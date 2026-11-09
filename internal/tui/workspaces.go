package tui

import (
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"cairn/internal/engine"
)

// Switcher opens the store of another workspace in place of the one a board holds, and says which pages the settings of the store now held show.
type Switcher func(workspace string) (engine.Pages, error)

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

// Listed is every workspace the configuration names, read when the switcher opens.
type Listed struct {
	Spaces []engine.Workspace
	Err    error
}

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
		return Listed{Spaces: answer.Workspaces.Workspaces}
	}
}

func (b Board) switchingTo(name string) tea.Cmd {
	return func() tea.Msg {
		if b.switcher == nil {
			return Switched{Err: errNoSwitching}
		}
		pages, err := b.switcher(name)
		if err != nil {
			return Switched{Err: err}
		}
		found, err := engine.Look(b.work)
		return Switched{Pages: pages, Looked: Looked{Snapshot: found, Err: err}, Placed: placedOf(b.work)}
	}
}

func (b Board) listedAll(m Listed) Board {
	if m.Err != nil {
		b.err = m.Err
		return b
	}
	b.spaces, b.sought, b.screen = m.Spaces, "", OnWorkspaces
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
	fresh.opened = openedOver(m.Placed.name(), b.place.name(), b.snap.Focus)
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

func (p Placed) name() string {
	if p.Shown.Workspace != "" {
		return p.Shown.Workspace
	}
	if p.Shown.Store == "" {
		return ""
	}
	return filepath.Dir(p.Shown.Store)
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

func (b Board) inWorkspaces(key string) (tea.Model, tea.Cmd) {
	shown := b.listed()
	switch key {
	case "esc":
		b.screen, b.spaces, b.sought = OnTheBoard, nil, ""
		return b, nil
	case goUp:
		b.spaceAt = max(0, b.spaceAt-1)
		return b, nil
	case goDown:
		b.spaceAt = min(len(shown)-1, b.spaceAt+1)
		return b, nil
	case "enter":
		if b.spaceAt < len(shown) {
			return b.choosingSpace(shown[b.spaceAt])
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
		b.screen, b.spaces, b.sought = OnTheBoard, nil, ""
		return b, nil
	}
	return b, b.switchingTo(one.Name)
}

func (b Board) placeMark() string {
	if b.place.Shown.Workspace != "" {
		return b.skin.Crumb.Render("● "+b.place.Shown.Workspace) + "  "
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
	row(" " + b.skin.Prompt.Render("› "+b.sought) + b.skin.Key.Render("▏"))
	row("")
	row(" " + b.skin.Head.Render("WORKSPACES"))
	shown := b.listed()
	if len(shown) == 0 {
		row("   " + b.skin.Faint.Render("none is named "+b.sought))
	}
	room := max(1, b.height-11)
	from := max(0, b.spaceAt-room+1)
	for i := from; i < min(len(shown), from+room); i++ {
		row(b.spaceLine(shown[i], i == b.spaceAt, inner))
	}
	lines = append(lines, edge.Render("└"+strings.Repeat("─", inner)+"┘"))
	return strings.Join(lines, "\n")
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
	name := b.skin.Crumb.Render("● " + pad(one.Name, wide))
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
