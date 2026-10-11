package gui

import (
	"bytes"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
)

// A pointer event written down says everything the window made of it: what
// the system sent, in pixels, the ratio it was read at, where that lands in dp,
// and where a card in hand now stands. It is how a drag that draws its card in
// the wrong place on one system is told apart from a drag that is told the
// wrong place by it.
func TestAPointerEventIsWrittenDownWithWhatTheWindowMadeOfIt(t *testing.T) {
	w := Window{held: Held{Number: 3, At: Point{X: 40, Y: 50}, Moved: true}}
	got := PointerTrace(pointer.Event{Kind: pointer.Drag, Position: f32.Pt(120, 150), Buttons: pointer.ButtonPrimary, PointerID: 2}, 1.5, w)
	want := "pointer Drag id=2 buttons=ButtonPrimary px=(120,150) perDp=1.5 dp=(80,100) held=3 at=(40,50) moved=true"
	if got != want {
		t.Errorf("the trace reads\n%q\nwant\n%q", got, want)
	}
}

// Pointer events are written to standard error only when CAIRN_GUI_TRACE says
// something.
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
