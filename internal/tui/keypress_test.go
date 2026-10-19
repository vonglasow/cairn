package tui

import tea "charm.land/bubbletea/v2"

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
	case "delete":
		return tea.KeyPressMsg{Code: tea.KeyDelete}
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

func cutCtrl(name string) (string, bool) {
	const ctrl = "ctrl+"
	if len(name) > len(ctrl) && name[:len(ctrl)] == ctrl {
		return name[len(ctrl):], true
	}
	return "", false
}

func cutShift(name string) (string, bool) {
	const shift = "shift+"
	if len(name) > len(shift) && name[:len(shift)] == shift {
		return name[len(shift):], true
	}
	return "", false
}
