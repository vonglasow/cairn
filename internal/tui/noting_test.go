package tui

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/debug"
)

func TestTheBoardWritesDownItsKeysItsSizeAndWhatItShowsInDebugMode(t *testing.T) {
	noted, err := debug.Open(t.TempDir(), "tui", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	model, _ := standing(twoTasks()).Noting(noted).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	hit(model.(Board), "/", "F", "i", "esc", "?")
	_ = noted.Close()
	read, err := os.ReadFile(noted.Path())
	if err != nil {
		t.Fatal(err)
	}
	said := string(read)
	for _, want := range []string{"board size 120x40", "board key /", "board key ‹typed 1›", "board key esc", "screen=help"} {
		if !strings.Contains(said, want) {
			t.Errorf("the file does not say %q:\n%s", want, said)
		}
	}
	if strings.Contains(said, "key F") || strings.Contains(said, "key i") {
		t.Errorf("a letter typed at the prompt is written down as itself:\n%s", said)
	}
}

func TestABoardOutOfDebugModeWritesNothingDown(t *testing.T) {
	b := hit(standing(twoTasks()), "/", "F")
	if b.noted != nil {
		t.Error("a board nobody put in debug mode holds a file")
	}
}
