package gui

import (
	"testing"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func masked(string) string { return "‹masked›" }

func TestAPointerEventIsWrittenDownWithWhatTheWindowMadeOfIt(t *testing.T) {
	w := Window{held: Held{Number: 3, At: Point{X: 40, Y: 50}, Moved: true}}
	got := PointerTrace(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(120, 150), Buttons: pointer.ButtonPrimary, PointerID: 2}, 1.5, w)
	want := "pointer Drag id=2 buttons=ButtonPrimary px=(120,150) perDp=1.5 dp=(80,100) held=3 at=(40,50) moved=true"
	if got != want {
		t.Errorf("the trace reads\n%q\nwant\n%q", got, want)
	}
}

func TestAHoverIsNotWrittenDown(t *testing.T) {
	for kind, want := range map[pointer.Kind]bool{
		pointer.Move: false, pointer.Enter: false, pointer.Leave: false,
		pointer.Press: true, pointer.Drag: true, pointer.Release: true, pointer.Scroll: true,
	} {
		if got := PointerNoted(kind); got != want {
			t.Errorf("a %v is written down: %v, want %v", kind, got, want)
		}
	}
}

func TestWhatTheWindowShowsIsSaidByItsScreenTabAndCard(t *testing.T) {
	w := Window{screen: OnForm, tab: Objectives, chosen: 4}
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
