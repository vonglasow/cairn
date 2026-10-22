package tui

import (
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
				b.screen = OnForm
				return b, nil
			}
			b.screen = OnTheBoard
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
		if b.finishing != nil {
			held := *b.finishing
			b.finishing = nil
			if !agreed {
				return b, b.look()
			}
			return b, b.carry(held.Op, engine.Agreed(held.In))
		}
		if !agreed {
			return b, nil
		}
		if b.tab == Objectives {
			return b, b.removeGoal()
		}
		return b, b.remove()
	}
	if b.screen != OnTheBoard {
		return b.onScreen(key)
	}
	if b.choosing {
		return b.choosingKind(key)
	}
	if b.prompt != NoPrompt {
		return b.typing(key)
	}
	if b.mode == Insert {
		return b.inInsert(key)
	}
	return b.inNormal(key)
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
	if b.screen != OnTheBoard || b.prompt != NoPrompt || b.choosing {
		return false
	}
	return b.tab == Objectives || b.mode == Insert
}

func (b Board) help() Board {
	if b.screen == OnHelp {
		b.screen = b.wasScreen
		return b
	}
	b.wasScreen = b.screen
	b.screen = OnHelp
	return b
}

func (b Board) onScreen(key string) (tea.Model, tea.Cmd) {
	if key == "esc" {
		b.screen, b.stepOn = b.wasScreen, 0
		if b.screen == OnHelp {
			b.screen = OnTheBoard
		}
		return b, nil
	}
	if way, ok := direction(key); ok {
		if b.screen == OnHelp {
			return b.readOn(way), nil
		}
		if b.screen == OnSaid {
			switch way {
			case goDown:
				b.saidAt++
			case goUp:
				b.saidAt = max(0, b.saidAt-1)
			}
			return b, nil
		}
		if b.screen == OnDetail {
			return b.walkingSteps(way), nil
		}
	}
	if b.screen == OnDetail {
		switch key {
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

func (b Board) inNormal(key string) (tea.Model, tea.Cmd) {
	if b.tab == Time {
		if walked, ok := b.walkingWeeks(key); ok {
			return walked, walked.readingWeek()
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
	case "i", "enter":
		return b.takeHold(), nil
	case "a":
		return b.addForm()
	case "shift+a":
		return b.showingArchived(), nil
	case "shift+x":
		if b.tab != Objectives {
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
		if b.tab == Time {
			return b, b.readingWeek()
		}
		return b, nil
	case "b":
		return b, b.telling(engine.BackupNew, engine.Input{})
	case "B", "shift+b":
		return b, b.telling(engine.BackupList, engine.Input{})
	case "e":
		if b.tab == Objectives {
			return b.editForm()
		}
	case "d":
		if b.tab == Objectives && b.goal != 0 {
			b.asking = true
			return b, nil
		}
	}
	return b, nil
}

func (b Board) inInsert(key string) (tea.Model, tea.Cmd) {
	if way, ok := shifted(key); ok {
		return b.push(way)
	}
	if way, ok := direction(key); ok {
		return b.stepTo(way), nil
	}
	switch key {
	case "esc":
		b.mode = Normal
		return b, nil
	case "enter":
		b.screen = OnDetail
		return b, b.hearing(b.task)
	case "e":
		return b.editForm()
	case "f":
		return b, b.toggleFocus()
	case "o":
		return b.linkedForm()
	case "d":
		if b.task != 0 {
			b.asking = true
		}
		return b, nil
	case "x":
		return b, b.puttingAway()
	case "shift+a":
		return b.showingArchived(), nil
	}
	return b, nil
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
	case "backspace":
		b.typed = trimLast(b.typed)
		return b.afterTyping(), nil
	}
	if letter, ok := typed(key); ok {
		b.typed += letter
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

func trimLast(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
