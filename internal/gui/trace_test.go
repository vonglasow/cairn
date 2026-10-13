package gui

import (
	"bytes"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

func TestAPointerEventIsWrittenDownWithWhatTheWindowMadeOfIt(t *testing.T) {
	w := Window{held: Held{Number: 3, At: Point{X: 40, Y: 50}, Moved: true}}
	got := PointerTrace(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(120, 150), Buttons: pointer.ButtonPrimary, PointerID: 2}, 1.5, w)
	want := "pointer Drag id=2 buttons=ButtonPrimary px=(120,150) perDp=1.5 dp=(80,100) held=3 at=(40,50) moved=true"
	if got != want {
		t.Errorf("the trace reads\n%q\nwant\n%q", got, want)
	}
}

func TestPointerEventsAreWrittenDownOnlyWhenAsked(t *testing.T) {
	var stderr bytes.Buffer
	asked := func(name string) string {
		if name == "CAIRN_GUI_TRACE" {
			return "1"
		}
		return ""
	}
	if got := tracing(asked, &stderr); got != &stderr {
		t.Errorf("asked to, the window writes to %v", got)
	}
	if got := tracing(func(string) string { return "" }, &stderr); got != nil {
		t.Errorf("unasked, the window writes to %v", got)
	}
}

func TestAKeyboardEventIsWrittenDownAsItCame(t *testing.T) {
	for _, c := range []struct {
		event event.Event
		want  string
	}{
		{key.Event{Name: "A", State: key.Press, Modifiers: key.ModShift}, `key Press name="A" mods=Shift`},
		{key.EditEvent{Text: "é", Range: key.Range{Start: 2, End: 3}}, `edit text="é" range=(2,3)`},
		{key.SnippetEvent{Start: 0, End: 4}, `snippet range=(0,4)`},
		{key.SelectionEvent{Start: 1, End: 1}, `selection range=(1,1)`},
		{key.FocusEvent{Focus: true}, `focus true`},
	} {
		if got, ok := KeyTrace(c.event); !ok || got != c.want {
			t.Errorf("the trace reads %q, %v, want %q", got, ok, c.want)
		}
	}
	if _, ok := KeyTrace(pointer.Event{}); ok {
		t.Error("a pointer event was written down as a keyboard one")
	}
}
