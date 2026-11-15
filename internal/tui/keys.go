package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine/migration"
	"strconv"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

func keyName(k tea.KeyPressMsg) string {
	name := k.Keystroke()
	if len([]rune(name)) == 1 {
		r := []rune(name)[0]
		if unicode.IsUpper(r) {
			return "shift+" + strings.ToLower(string(r))
		}
	}
	return name
}

const (
	goLeft  = "left"
	goDown  = "down"
	goUp    = "up"
	goRight = "right"
)

func direction(key string) (string, bool) {
	switch key {
	case "h", goLeft:
		return goLeft, true
	case "j", goDown:
		return goDown, true
	case "k", goUp:
		return goUp, true
	case "l", goRight:
		return goRight, true
	}
	return "", false
}

func shifted(key string) (string, bool) {
	if !strings.HasPrefix(key, "shift+") {
		return "", false
	}
	return direction(strings.TrimPrefix(key, "shift+"))
}

func (b Board) pressed(key string) (tea.Model, tea.Cmd) {
	if key == "ctrl+c" {
		return b, tea.Quit
	}
	if b.pages.None() {
		return b, nil
	}
	if b.asking && b.aboutStore() {
		return b.answeringStore(key)
	}
	if b.deleting != nil {
		return b.choosingFates(key)
	}
	if b.seriesGoing != nil {
		return b.choosingSeriesFate(key)
	}
	if b.wholeSave != nil {
		return b.answeringWhole(key)
	}
	if key == "?" {
		return b.help(), nil
	}
	if answered, model, cmd := b.jumping(key); answered {
		return model, cmd
	}
	b.pending = ""
	if b.asking {
		b.asking = false
		agreed := key == "enter" || key == "return"
		if b.leaving {
			b.leaving = false
			if !agreed {
				b.screen = board.OnForm
				return b, nil
			}
			b.screen = board.OnTheBoard
			return b, nil
		}
		if b.sweeping {
			b.sweeping = false
			if !agreed {
				return b, nil
			}
			return b, b.sweepingColumn()
		}
		if b.going != nil {
			going := *b.going
			b.going = nil
			if !agreed {
				return b, nil
			}
			return b, b.carry(engine.StepDelete, going.In)
		}
		if b.agreeing != nil {
			held := *b.agreeing
			b.agreeing = nil
			if !agreed {
				b.column, b.inBand = b.leftFrom.column, b.leftFrom.inBand
				return b.follow(), b.look()
			}
			return b, b.carry(held.Op, held.In)
		}
		if !agreed {
			return b, nil
		}
		if b.tab == board.Objectives {
			return b, b.removeGoal()
		}
		return b, b.remove()
	}
	if b.screen == board.OnWorkspaces {
		return b.inWorkspaces(key)
	}
	if b.screen != board.OnTheBoard {
		return b.onScreen(key)
	}
	if b.choosing {
		return b.choosingKind(key)
	}
	if b.filtering {
		return b.filteringBy(key), nil
	}
	if b.choosingView {
		return b.choosingAView(key), nil
	}
	if b.prompt != NoPrompt {
		return b.typing(key)
	}
	if key == switcherKey && b.roams() {
		return b, b.listing()
	}
	return b.onBoard(key)
}

func (b Board) aboutStore() bool {
	return b.aboutReason(migration.ErrStoreBehind) || b.aboutReason(migration.ErrNewerStore)
}

func (b Board) aboutReason(err error) bool {
	return b.agreeing != nil && b.agreeing.Reason == engine.RefusalOf(err).Reason
}

func (b Board) answeringStore(key string) (tea.Model, tea.Cmd) {
	if key == "esc" {
		return b, tea.Quit
	}
	if (key == "enter" || key == "return") && b.aboutReason(migration.ErrStoreBehind) {
		return b, b.carry(b.agreeing.Op, b.agreeing.In)
	}
	return b, nil
}

func (b Board) jumping(key string) (bool, tea.Model, tea.Cmd) {
	waiting := b.pending
	b.pending = ""
	if !b.jumpable() {
		return false, b, nil
	}
	if key == "shift+g" {
		return true, b.toEnd(), nil
	}
	if key == "g" {
		if waiting == "g" {
			return true, b.toStart(), nil
		}
		b.pending = "g"
		return true, b, nil
	}
	return false, b, nil
}

func (b Board) jumpable() bool {
	if b.screen != board.OnTheBoard || b.prompt != NoPrompt || b.choosing {
		return false
	}
	return b.tab == board.Objectives || b.task != 0
}

func (b Board) help() Board {
	if b.screen == board.OnHelp {
		b.screen = b.wasScreen
		return b
	}
	b.wasScreen = b.screen
	b.screen = board.OnHelp
	return b
}

func (b Board) onScreen(key string) (tea.Model, tea.Cmd) {
	if key == "esc" {
		b.screen, b.stepOn = b.wasScreen, 0
		if b.screen == board.OnHelp {
			b.screen = board.OnTheBoard
		}
		return b, nil
	}
	if way, ok := direction(key); ok {
		if b.screen == board.OnHelp {
			return b.readOn(way), nil
		}
		if b.screen == board.OnSaid {
			switch way {
			case goDown:
				b.saidAt++
			case goUp:
				b.saidAt = max(0, b.saidAt-1)
			}
			return b, nil
		}
		if b.screen == board.OnDetail {
			return b.walkingSteps(way), nil
		}
	}
	if b.screen == board.OnDetail && b.setting == "series" {
		b.setting = ""
		return b.seriesAct(key)
	}
	if b.screen == board.OnDetail {
		switch key {
		case "r":
			if b.heldRecurs() {
				b.setting = "series"
			}
			return b, nil
		case "e":
			return b.editForm()
		case "backspace", "delete":
			return b.askingToDelete(), nil
		case "f":
			return b, b.toggleFocus()
		case "ctrl+d":
			return b.scrollingSaying(true), nil
		case "ctrl+u":
			return b.scrollingSaying(false), nil
		}
	}
	return b, nil
}

func (b Board) onBoard(key string) (tea.Model, tea.Cmd) {
	if b.setting != "" {
		return b.settingByLetter(key)
	}
	if b.tab == board.Series {
		if model, cmd, took := b.inSeries(key); took {
			return model, cmd
		}
	}
	if b.tab == board.Time {
		if walked, ok := b.walkingWeeks(key); ok {
			return walked, walked.readingWeek()
		}
	}
	if b.tab == board.Tasks {
		if model, cmd, took := b.onTasks(key); took {
			return model, cmd
		}
	}
	if way, ok := direction(key); ok {
		return b.moveCursor(way), nil
	}
	switch key {
	case "esc":
		if b.query != "" {
			b.query = ""
			return b, nil
		}
		if b.inside != 0 {
			return b.leaveObjective(), nil
		}
	case "enter":
		if b.tab == board.Objectives {
			return b.openObjective(), nil
		}
	case "a":
		return b.addForm()
	case "shift+a":
		return b.showingArchived(), nil
	case "shift+f":
		return b.openingFilters(), nil
	case "v":
		b.choosingView = b.tab == board.Tasks && b.inside == 0
		return b, nil
	case "shift+x":
		if b.tab != board.Objectives {
			b.asking, b.sweeping = true, true
		}
		return b, nil
	case "/":
		b.prompt, b.typed = Searching, ""
		return b, nil
	case ":":
		b.prompt, b.typed = Commanding, ""
		return b, nil
	case "tab":
		b = b.otherTab()
		if b.tab == board.Time {
			return b, b.readingWeek()
		}
		return b, nil
	case "b":
		return b, b.telling(engine.BackupNew, engine.Input{})
	case "B", "shift+b":
		return b, b.telling(engine.BackupList, engine.Input{})
	case "e":
		if b.tab == board.Objectives {
			return b.editForm()
		}
	case "D", "shift+d":
		if b.tab == board.Objectives && b.goal != 0 {
			return b.deletingGoal(), nil
		}
	case "t":
		if b.tab == board.Objectives {
			return b.showingItsTasks(), nil
		}
	case "N", "shift+n":
		if b.tab == board.Objectives {
			return b.subGoalForm()
		}
	case "z":
		if b.tab == board.Objectives {
			return b.folding(), nil
		}
	case "J", "shift+j", "K", "shift+k":
		if b.tab == board.Objectives && b.goal != 0 {
			in, moves := board.Stepped(b.snap.Goals, b.goal, key == "K" || key == "shift+k")
			if !moves {
				return b, nil
			}
			return b, b.carry(engine.ObjectiveMove, in)
		}
	case "[":
		return b.walkingSegments(-1), nil
	case "]":
		return b.walkingSegments(1), nil
	}
	return b, nil
}

func (b Board) onTasks(key string) (tea.Model, tea.Cmd, bool) {
	switch key {
	case "ctrl+h":
		return b.toColumn(b.column - 1), nil, true
	case "ctrl+l":
		return b.toColumn(b.column + 1), nil, true
	case "ctrl+j":
		return b.intoBand(), nil, true
	case "ctrl+k":
		return b.outOfBand(), nil, true
	}
	if b.task == 0 {
		return b, nil, false
	}
	if way, ok := shifted(key); ok {
		model, cmd := b.push(way)
		return model, cmd, true
	}
	switch key {
	case "enter":
		b.screen = board.OnDetail
		return b, b.hearing(b.task), true
	case "e":
		model, cmd := b.editForm()
		return model, cmd, true
	case "f":
		return b, b.toggleFocus(), true
	case "o":
		model, cmd := b.linkedForm()
		return model, cmd, true
	case "D", "shift+d":
		b.asking = true
		return b, nil, true
	case "m":
		b.setting = "state"
		return b, nil, true
	case "p":
		b.setting = "priority"
		return b, nil, true
	case "r":
		if b.heldRecurs() {
			b.setting = "series"
		}
		return b, nil, true
	case "x":
		return b, b.puttingAway(), true
	}
	return b, nil, false
}

func (b Board) showingArchived() Board {
	b.archived = !b.archived
	return b
}

func (b Board) typing(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		b.prompt, b.typed = NoPrompt, ""
		if b.query != "" {
			b.query = ""
		}
		return b, nil
	case "enter":
		return b.accept()
	case "tab":
		return b.completed(), nil
	case "backspace":
		b.typed = board.Ending(b.typed).Back().Value
		return b.afterTyping(), nil
	}
	if letter, ok := typed(key); ok {
		b.typed = board.Ending(b.typed).Types(letter).Value
		return b.afterTyping(), nil
	}
	return b, nil
}

func (b Board) afterTyping() Board {
	if b.prompt == Searching {
		b.query = b.typed
	}
	return b
}

func typed(key string) (string, bool) {
	if rest, ok := strings.CutPrefix(key, "shift+"); ok && len([]rune(rest)) == 1 {
		return strings.ToUpper(rest), true
	}
	if key == "space" {
		return " ", true
	}
	if len([]rune(key)) == 1 {
		return key, true
	}
	return "", false
}

func (b Board) settingByLetter(key string) (tea.Model, tea.Cmd) {
	field := b.setting
	b.setting = ""
	if field == "series" {
		return b.seriesAct(key)
	}
	values := engine.States()
	if field == "priority" {
		values = board.Priorities()
	}
	for _, value := range values {
		if len(key) == 1 && strings.HasPrefix(value, key) {
			return b, b.telling(engine.TaskEdit, engine.Input{Given: map[string]string{
				"number": strconv.Itoa(b.task), field: value}})
		}
	}
	return b, nil
}
