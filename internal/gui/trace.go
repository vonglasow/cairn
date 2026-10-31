package gui

import (
	"fmt"
	"strings"

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

// PointerNoted says whether debug mode writes down a pointer event of this kind: a press, a drag, a release, a scroll — never the pointer merely passing over, which would bury the rest.
func PointerNoted(kind pointer.Kind) bool {
	return kind != pointer.Move && kind != pointer.Enter && kind != pointer.Leave
}

// KeyTrace is one event from the keyboard as it came, and whether it is one: a key with its name and state, text — masked, since it is what the person wrote — with the range it replaces, and what an input method says of the text and the caret it holds.
func KeyTrace(e event.Event, mask func(string) string) (string, bool) {
	switch e := e.(type) {
	case key.Event:
		return fmt.Sprintf("key %v name=%q mods=%v", e.State, e.Name, e.Modifiers), true
	case key.EditEvent:
		return fmt.Sprintf("edit text=%s range=(%d,%d)", mask(e.Text), e.Range.Start, e.Range.End), true
	case key.SnippetEvent:
		return fmt.Sprintf("snippet range=(%d,%d)", e.Start, e.End), true
	case key.SelectionEvent:
		return fmt.Sprintf("selection range=(%d,%d)", e.Start, e.End), true
	case key.FocusEvent:
		return fmt.Sprintf("focus %v", e.Focus), true
	}
	return "", false
}

var screenNames = map[Screen]string{
	OnTheBoard: "board", OnForm: "form", OnSearch: "search", OnAsking: "asking", OnHelp: "help",
	OnChoosing: "choosing", OnCalendar: "calendar", OnMenu: "menu", OnBackups: "backups",
	OnLabels: "labels", OnFiltering: "filtering", OnConfiguration: "configuration",
}

// Seen is what the window shows, the way debug mode writes it down each time it changes: the screen, the page, and the card or objective chosen.
func (w Window) Seen() string {
	return fmt.Sprintf("screen=%s tab=%s chosen=%d", screenNames[w.screen], strings.ToLower(Pages[w.tab]), w.chosen)
}
