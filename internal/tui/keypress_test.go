package tui

import tea "charm.land/bubbletea/v2"

// keyPress builds the message a terminal would have sent for a key, so a
// scenario presses what a person presses.
func keyPress(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "f1":
		// A key that types nothing, for the tests that need one.
		return tea.KeyPressMsg{Code: tea.KeyF1}
	}
	if rest, found := cutCtrl(name); found {
		r := []rune(rest)[0]
		return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl, Text: string(r)}
	}
	if rest, found := cutShift(name); found {
		r := []rune(rest)[0]
		return tea.KeyPressMsg{Code: r, Mod: tea.ModShift, Text: string(r)}
	}
	r := []rune(name)[0]
	return tea.KeyPressMsg{Code: r, Text: name}
}

// cutCtrl takes the "ctrl+" off a key's name, the way cutShift does its own.
func cutCtrl(name string) (string, bool) {
	const ctrl = "ctrl+"
	if len(name) > len(ctrl) && name[:len(ctrl)] == ctrl {
		return name[len(ctrl):], true
	}
	return "", false
}

// cutShift takes the "shift+" off a key's name.
func cutShift(name string) (string, bool) {
	const shift = "shift+"
	if len(name) > len(shift) && name[:len(shift)] == shift {
		return name[len(shift):], true
	}
	return "", false
}
