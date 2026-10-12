package tui

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/engine"
)

// What a key does, and where. Two rules hold everywhere and are worth stating
// before the table: hjkl and the arrows only ever navigate, and the question
// mark answers in every mode. Everything else belongs to one mode, and a key
// that belongs to no mode here does nothing rather than something surprising.

// keyName is what a key press is called, in one spelling. Terminals disagree
// about shifted letters: some send "shift+l" when they can, and the rest send
// "L". Both are the same gesture, so both arrive here as the same word.
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

// The directions, spelled once. A board reads a direction, never a letter, so
// hjkl and the arrows cannot drift apart.
const (
	goLeft  = "left"
	goDown  = "down"
	goUp    = "up"
	goRight = "right"
)

// direction turns a key into the way it points, and says whether it points
// anywhere at all.
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

// shifted turns a shifted key into the way it points.
func shifted(key string) (string, bool) {
	if !strings.HasPrefix(key, "shift+") {
		return "", false
	}
	return direction(strings.TrimPrefix(key, "shift+"))
}

// pressed is the whole of what a key does. It is written as a chain of
// answers rather than one switch: what is on screen decides before what mode
// the board is in, and a prompt taking what is typed decides before either.
func (b Board) pressed(key string) (tea.Model, tea.Cmd) {
	if key == "ctrl+c" {
		// The way out, from anywhere. `:q` was the only one, and the help did
		// not mention it — whoever had not read the documentation was shut
		// inside the alt-screen with no way to ask.
		return b, tea.Quit
	}
	if key == "?" {
		return b.help(), nil
	}
	if answered, model, cmd := b.jumping(key); answered {
		return model, cmd
	}
	// A g that led nowhere is forgotten here rather than inside jumping: what
	// it hands back when it did not answer is thrown away, and a pending key
	// nobody clears is a key that changes the meaning of every press after it.
	b.pending = ""
	// A question waiting to be agreed to takes the next key and nothing else
	// does: enter agrees, anything else backs out. It is asked before the mode
	// so that no key can slip past it.
	if b.asking {
		b.asking = false
		agreed := key == "enter" || key == "return"
		if b.leaving {
			b.leaving = false
			// Backing out of the question goes back to the form, which is the
			// whole point of having asked.
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

// jumping answers `G` and `gg`, and says whether it did.
//
// It sits in front of everything else that reads a key, because a `g` already
// typed changes what the next one means — and it clears itself on anything
// else, so a `g` pressed by mistake is forgotten rather than lying in wait.
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

// jumpable says whether there is a run of things to jump to the ends of: tasks
// held in insert, or the objectives. On a column nothing is held, and in a form
// or a prompt a g is a letter.
func (b Board) jumpable() bool {
	if b.screen != OnTheBoard || b.prompt != NoPrompt || b.choosing {
		return false
	}
	return b.tab == Objectives || b.mode == Insert
}

// help opens the help, or closes it when it is already open. It remembers the
// mode it was asked from so closing it puts the board back untouched.
func (b Board) help() Board {
	if b.screen == OnHelp {
		b.screen = b.wasScreen
		return b
	}
	b.wasScreen = b.screen
	b.screen = OnHelp
	return b
}

// onScreen is what a key does while a screen holds the window. The form is not
// here: it takes the key itself, in Update, because what it needs is the press
// and not the name this board gives it.
func (b Board) onScreen(key string) (tea.Model, tea.Cmd) {
	if key == "esc" {
		// Back to whatever the screen was asked over, form included — which is
		// what the help promises when it is closed by a second ?.
		b.screen = b.wasScreen
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
		// The detail's bar offers hjkl and the arrows, and offering a key
		// without serving it makes the bar a thing to distrust. They move
		// between tasks, and the detail follows the cursor.
		if b.screen == OnDetail {
			b = b.stepTo(way)
			return b, b.hearing(b.task)
		}
	}
	// The detail's own bar offers e and f, and offering a key without serving
	// it makes the bar a thing to distrust. They do here what they do on the
	// board, on the task the detail is showing — which is the task the cursor
	// is on, so there is nothing else to work out.
	if b.screen == OnDetail {
		switch key {
		case "e":
			return b.editForm()
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

// inNormal is what a key does while the cursor is on a column.
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
		// An accepted search outlives the prompt that typed it, and nothing on
		// the board says so. Escape widens it again before it does anything
		// else, so there is always a way back to the whole board.
		if b.query != "" {
			b.query = ""
			return b, nil
		}
		// Inside an objective there is somewhere to go back to. On the board
		// itself there is not, and escape does nothing rather than something
		// surprising.
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
		// The whole column off the board at once. It asks first: it moves
		// everything the column holds, and a gesture that moves twelve things
		// is not a gesture to find out about afterwards.
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
		// The objectives list has no insert mode — normal already moves between
		// objectives — so what belongs to insert on the board belongs to normal
		// here. On the tasks tab these two do nothing, and the bar says so.
		if b.tab == Objectives {
			return b.editForm()
		}
	case "d":
		// Nothing to delete, nothing to ask about: the question is only worth
		// putting when agreeing to it would take something away.
		if b.tab == Objectives && b.goal != 0 {
			b.asking = true
			return b, nil
		}
	}
	return b, nil
}

// inInsert is what a key does while the cursor is on a task.
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

// showingArchived puts back on the board what was taken off it, or takes it
// away again.
func (b Board) showingArchived() Board {
	b.archived = !b.archived
	return b
}

// typing is what a key does while a prompt at the foot is taking it.
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

// afterTyping narrows the board as the search is typed. A command is not run
// until it is accepted, so only a search answers here.
func (b Board) afterTyping() Board {
	if b.prompt == Searching {
		b.query = b.typed
	}
	return b
}

// typed turns a key into the character it puts into a field, and says whether
// it puts one at all. A capital arrives as a shifted letter, because that is
// how the board reads it everywhere else — and a field that could not take one
// would be a field nobody can write a title in.
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

// trimLast takes the last character off, counted as a character and not as a
// byte — a search that cannot delete an accent is a search nobody will use
// twice.
func trimLast(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
