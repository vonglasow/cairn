package gui

import (
	"cairn/internal/board"
	"testing"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func masked(string) string { return "‹masked›" }

func TestWhatTheWindowShowsIsSaidByItsScreenTabAndCard(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnForm, Tab: board.Objectives, Chosen: 4}}
	if got := w.Seen(); got != "screen=form tab=objectives chosen=4" {
		t.Errorf("the window says %q", got)
	}
	if (Window{}).Seen() == w.Seen() {
		t.Error("two windows showing different things say the same")
	}
}

func TestAKeyboardEventIsWrittenDownAsItCame(t *testing.T) {
	for _, c := range []struct {
		event event.Event
		want  string
	}{
		{key.Event{Name: "A", State: key.Press, Modifiers: key.ModShift}, `key Press name="A" mods=Shift`},
		{key.EditEvent{Text: "é", Range: key.Range{Start: 2, End: 3}}, `edit text=‹masked› range=(2,3)`},
		{key.SnippetEvent{Start: 0, End: 4}, `snippet range=(0,4)`},
		{key.SelectionEvent{Start: 1, End: 1}, `selection range=(1,1)`},
		{key.FocusEvent{Focus: true}, `focus true`},
	} {
		if got, ok := KeyTrace(c.event, masked); !ok || got != c.want {
			t.Errorf("the trace reads %q, %v, want %q", got, ok, c.want)
		}
	}
	if _, ok := KeyTrace(pointer.Event{}, masked); ok {
		t.Error("a pointer event was written down as a keyboard one")
	}
}
