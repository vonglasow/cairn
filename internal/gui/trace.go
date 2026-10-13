package gui

import (
	"fmt"
	"io"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

// PointerTrace is one pointer event as the window read it: what the system sent in pixels, the ratio it was read at, the point it came to in dp, and the card in hand afterwards.
func PointerTrace(e pointer.Event, perDp float32, w Window) string {
	at := Dp(e.Position, perDp)
	return fmt.Sprintf("pointer %v id=%d buttons=%v px=(%g,%g) perDp=%g dp=(%d,%d) held=%d at=(%d,%d) moved=%v",
		e.Kind, e.PointerID, e.Buttons, e.Position.X, e.Position.Y, perDp, at.X, at.Y,
		w.held.Number, w.held.At.X, w.held.At.Y, w.held.Moved)
}

// KeyTrace is one event from the keyboard as it came, and whether it is one: a key with its name and state, text with the range it replaces, and what an input method says of the text and the caret it holds.
func KeyTrace(e event.Event) (string, bool) {
	switch e := e.(type) {
	case key.Event:
		return fmt.Sprintf("key %v name=%q mods=%v", e.State, e.Name, e.Modifiers), true
	case key.EditEvent:
		return fmt.Sprintf("edit text=%q range=(%d,%d)", e.Text, e.Range.Start, e.Range.End), true
	case key.SnippetEvent:
		return fmt.Sprintf("snippet range=(%d,%d)", e.Start, e.End), true
	case key.SelectionEvent:
		return fmt.Sprintf("selection range=(%d,%d)", e.Start, e.End), true
	case key.FocusEvent:
		return fmt.Sprintf("focus %v", e.Focus), true
	}
	return "", false
}

func tracing(getenv func(string) string, stderr io.Writer) io.Writer {
	if getenv("CAIRN_GUI_TRACE") == "" {
		return nil
	}
	return stderr
}
