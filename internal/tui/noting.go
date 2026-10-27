package tui

import (
	"fmt"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/debug"
)

// Noting is the board writing down what happens to it in the file debug mode keeps.
func (b Board) Noting(noted *debug.Log) Board {
	b.noted = noted
	return b
}

var (
	screenNames = map[Screen]string{OnTheBoard: "board", OnHelp: "help", OnDetail: "detail", OnForm: "form", OnSaid: "said"}
	tabNames    = map[Tab]string{Tasks: "tasks", Objectives: "objectives", Time: "time"}
	modeNames   = map[Mode]string{Normal: "normal", Insert: "insert"}
	promptNames = map[Prompt]string{NoPrompt: "none", Searching: "search", Commanding: "command"}
)

// Seen is what the board shows, the way debug mode writes it down each time it changes: the screen, the tab, the mode, the prompt, whether a question is asked, and the task under the cursor.
func (b Board) Seen() string {
	return fmt.Sprintf("screen=%s tab=%s mode=%s prompt=%s asking=%v choosing=%v task=%d",
		screenNames[b.screen], tabNames[b.tab], modeNames[b.mode], promptNames[b.prompt], b.asking, b.choosing, b.task)
}

func (b Board) noting(msg tea.Msg) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.noted.Say("board", fmt.Sprintf("size %dx%d", m.Width, m.Height))
	case tea.KeyPressMsg:
		b.noted.Say("board", "key "+b.keyShown(m))
	}
}

func (b Board) keyShown(m tea.KeyPressMsg) string {
	typing := b.screen == OnForm || b.prompt != NoPrompt
	written := m.Text != "" && m.Mod&(tea.ModCtrl|tea.ModAlt) == 0
	if typing && written {
		for _, r := range m.Text {
			if !unicode.IsPrint(r) {
				return keyName(m)
			}
		}
		return b.noted.Typed(m.Text)
	}
	return keyName(m)
}
