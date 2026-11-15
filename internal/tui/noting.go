package tui

import (
	"fmt"
	"unicode"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/board"
	"cairn/internal/debug"
)

// Noting is the board writing down what happens to it in the file debug mode keeps.
func (b Board) Noting(noted *debug.Log) Board {
	b.noted = noted
	return b
}

var (
	screenNames = map[board.Screen]string{board.OnTheBoard: "board", board.OnHelp: "help", board.OnDetail: "detail", board.OnForm: "form", board.OnSaid: "said", board.OnWorkspaces: "workspaces"}
	tabNames    = map[board.Tab]string{board.Tasks: "tasks", board.Objectives: "objectives", board.Time: "time"}
	promptNames = map[Prompt]string{NoPrompt: "none", Searching: "search", Commanding: "command"}
)

// Seen is what the board shows, the way debug mode writes it down each time it changes: the screen, the tab, the mode, the prompt, whether a question is asked, and the task under the cursor.
func (b Board) Seen() string {
	return fmt.Sprintf("screen=%s tab=%s prompt=%s asking=%v choosing=%v task=%d",
		screenNames[b.screen], tabNames[b.tab], promptNames[b.prompt], b.asking, b.choosing, b.task)
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
	typing := b.screen == board.OnForm || b.prompt != NoPrompt
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
