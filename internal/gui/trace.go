package gui

// What the window makes of each pointer event, written down when asked. A
// drag on Windows drew its card at the top left of the window rather than
// under the pointer, on a machine this code is not written on; everything
// between the system and the drawing is pure, and this is the line that says
// which step of it went wrong there. See docs/gui.md.

import (
	"fmt"
	"io"

	"gioui.org/io/pointer"
)

// PointerTrace is one pointer event as the window read it: what the system
// sent in pixels, the ratio it was read at, the point it came to in dp, and
// the card in hand afterwards.
func PointerTrace(e pointer.Event, perDp float32, w Window) string {
	at := Dp(e.Position, perDp)
	return fmt.Sprintf("pointer %v id=%d buttons=%v px=(%g,%g) perDp=%g dp=(%d,%d) held=%d at=(%d,%d) moved=%v",
		e.Kind, e.PointerID, e.Buttons, e.Position.X, e.Position.Y, perDp, at.X, at.Y,
		w.held.Number, w.held.At.X, w.held.At.Y, w.held.Moved)
}

// tracing is where pointer events are written down: standard error when
// CAIRN_GUI_TRACE says anything, for a system the window misbehaves on that
// nobody here can sit at, and nowhere otherwise.
func tracing(getenv func(string) string, stderr io.Writer) io.Writer {
	if getenv("CAIRN_GUI_TRACE") == "" {
		return nil
	}
	return stderr
}
