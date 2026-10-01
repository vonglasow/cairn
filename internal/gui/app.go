package gui

// The loop, and the one thing in this package a test cannot reach: a window
// only a real display can open.
//
// What an event *means* is worked out by Pointed, Typed and Twice, which are
// ordinary functions over a Window and are covered like everything else. What
// is left here is plumbing, and plumbing that decides nothing cannot get
// anything wrong quietly.

import (
	"image"
	"io"
	"os"
	"strings"
	"time"
	"unicode"

	"gioui.org/io/clipboard"
	"gioui.org/io/transfer"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"cairn/internal/reading"
)

// Beat is how often the window asks the store whether anybody else has written
// to it. The same two seconds as the terminal board, and for the reason a
// second screen gives: a board nobody is looking at still has to be true.
// A function and not a constant, for the reason test/harness.go gives: the
// arithmetic of a package-level constant is never inside a covered block.
func Beat() time.Duration { return 2 * time.Second }

// Twin is how long after a press a second one at the same place is the same
// gesture. Worked out here because Gio reports presses and says nothing about
// double-clicks.
func Twin() time.Duration { return 400 * time.Millisecond }

// Dp turns where a pointer says it is into where the board thinks it is. A
// window speaks in pixels and this board works in dp; on a screen with two
// pixels to the dp, taking one for the other picks the card twice as far down
// the column.
func Dp(at f32.Point, perDp float32) Point {
	return Point{X: int(at.X / perDp), Y: int(at.Y / perDp)}
}

// Sized turns the size a window reports, in pixels, into the size this board
// works in.
func Sized(max image.Point, perDp float32) (width, height int) {
	return int(float32(max.X) / perDp), int(float32(max.Y) / perDp)
}

// Turned is the same for how far a wheel says it went.
func Turned(by, perDp float32) int { return int(by / perDp) }

// pasting hands the window what the clipboard held. A clipboard that will not
// read leaves the text alone and stops the waiting all the same: a paste that
// silently emptied a field would be worse than one that did nothing.
func pasting(w Window, said io.Reader) Window {
	read, err := io.ReadAll(said)
	if err != nil {
		return w.Pasted("")
	}
	return w.Pasted(string(read))
}

// Gesture is what a pointer has said so far: when the last press landed and
// where, which is the whole of what tells the press after it from the second
// half of a double-click.
type Gesture struct {
	At    time.Time
	Where Point
}

// Pointing is everything a pointer event does, the double-click included, and
// the gesture it leaves behind. One function and not a rule kept in the loop:
// what the loop decided itself, no scenario could reach — and the release
// after a double-click closed the task it had just opened, found by hand in a
// window rather than by any test.
func Pointing(w Window, g Gesture, kind pointer.Kind, at Point, by int, now time.Time) (Window, Gesture) {
	if kind == pointer.Press {
		if Twice(g.At, now, g.Where, at) {
			return w.DoubleClick(at), Gesture{}
		}
		g = Gesture{At: now, Where: at}
	}
	return Pointed(w, kind, at, by), g
}

// Pointed says what a pointer event does to a window.
func Pointed(w Window, kind pointer.Kind, at Point, by int) Window {
	switch kind {
	case pointer.Press:
		return w.PressAt(at)
	case pointer.Drag:
		return w.MoveTo(at)
	case pointer.Release:
		return w.LetGo(at)
	case pointer.Scroll:
		return w.Scroll(at, by)
	}
	return w
}

// Twice says whether a press is the second half of a double-click: soon enough
// after the last one, and close enough to it that the hand did not move.
func Twice(before, now time.Time, was, at Point) bool {
	if before.IsZero() || now.Sub(before) > Twin() {
		return false
	}
	return abs(at.X-was.X) <= Grip && abs(at.Y-was.Y) <= Grip
}

// Typed says what a key does. The few flat shortcuts, and nothing modal.
func Typed(w Window, name string, with key.Modifiers) Window {
	if with.Contain(key.ModCommand) || with.Contain(key.ModCtrl) {
		// Named first, like every other key. Lowercasing Gio's own spelling
		// left ⌘ with a special key arriving as "cmd+⏎" — a name nothing is
		// bound to, and nothing said so.
		return w.Key("cmd+" + named(name))
	}
	said := named(name)
	// Shift is a modifier Gio reports apart from the key, so a shifted letter
	// arrives as the plain letter and has to be put back together here. Only a
	// letter and tab: a shifted "/" is a different character, not a shifted
	// key. Nothing typed comes this way — text goes through EditEvent.
	// A named key or a letter: those are what this board spells with a shift+
	// prefix. Punctuation is left alone, because Gio has already applied shift
	// to it — the "/" key with shift is named "?" before it gets here.
	if _, isNamed := naming(name); with.Contain(key.ModShift) && (isNamed || aLetter(said)) {
		return w.Key("shift+" + said)
	}
	return w.Key(said)
}

// aLetter says whether a name is a single letter.
func aLetter(name string) bool {
	r := []rune(name)
	return len(r) == 1 && unicode.IsLetter(r[0])
}

// named turns what Gio calls a key into what this board calls it. Gio spells
// them for a keyboard; a board says what they do.
func named(name string) string {
	said, _ := naming(name)
	return said
}

// naming does the same and says whether the key was one this board has a word
// for. Shift is joined to those and to letters, and to nothing else: Gio has
// already applied it to punctuation.
func naming(name string) (string, bool) {
	switch key.Name(name) {
	case key.NameEscape:
		return "esc", true
	case key.NameTab:
		return "tab", true
	case key.NameReturn, key.NameEnter:
		return "enter", true
	case key.NameDeleteBackward:
		return "backspace", true
	case key.NameLeftArrow:
		return "left", true
	case key.NameRightArrow:
		return "right", true
	case key.NameUpArrow:
		return "up", true
	case key.NameDownArrow:
		return "down", true
	}
	return strings.ToLower(name), false
}

// reach is how far one turn of a wheel is allowed to say it went. Larger than
// any window, so nothing a hand can do falls outside it.
const reach = 1 << 20

// tag is what this window's events are addressed to. One for the whole window:
// what a point means is worked out from the layout, not from whichever widget
// happened to claim it.
var tag = new(int)

// Run opens a window and keeps it until it is closed.
//
// The shape is not a choice. Gio's Main has to be called from the program's
// main goroutine and blocks there for as long as the program lives, because
// some operating systems insist on owning that thread to run a window at all —
// macOS is one. So the window's own loop goes on a goroutine and Main stays
// here, on the thread the command was called on.
//
// Measured, by running it: without Main the command starts, waits, and shows
// nothing. Every test in this package draws with no window and so never needed
// it, which is exactly the kind of thing only the built binary says.
//
// It follows that this never returns on a desktop, and the process ends from
// inside — which is why leave exists rather than an exit code handed back.
func Run(work reading.Work, width, height int) error {
	go func() { leave(code(loop(work, width, height))) }()
	app.Main()
	return nil
}

// leave is how the process ends once the window is gone. A variable because a
// test that called the real one would take the suite with it.
var leave = os.Exit

// code is what closing the window exits with. Closing it is not a failure.
func code(err error) int {
	if err == nil {
		return 0
	}
	return 1
}

// loop is the window itself: the only code in this package no test runs.
func loop(work reading.Work, width, height int) error {
	window := new(app.Window)
	window.Option(app.Title("cairn"), app.Size(unit.Dp(width), unit.Dp(height)))

	board := New(work, width, height).Look()
	theme, skin := Theme(), Dark()

	var ops op.Ops
	var pressed time.Time
	var pressedAt Point
	var looked time.Time

	go func() {
		for range time.Tick(Beat()) {
			window.Invalidate()
		}
	}()

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			board = board.Resized(Sized(gtx.Constraints.Max, gtx.Metric.PxPerDp))
			if due(looked, gtx.Now) {
				board, looked = board.Wonder(), gtx.Now
			}
			board, pressed, pressedAt = heard(gtx, board, pressed, pressedAt)
			Draw(gtx, theme, skin, board)
			e.Frame(gtx.Ops)
		}
	}
}

// due says whether it is time to ask the store again. Its own function because
// it is the one decision inside Run, and Run is the only thing here no test can
// reach — a decision left in there is a decision nothing judges.
func due(looked, now time.Time) bool { return now.Sub(looked) >= Beat() }

// heard takes everything that has happened since the last frame and answers the
// window it leaves behind.
func heard(gtx layout.Context, w Window, pressed time.Time, pressedAt Point) (Window, time.Time, Point) {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	area.Pop()

	// Where the keyboard was before any of this frame's events were read. A
	// keystroke can arrive twice — once as a key with a name, once as the text
	// an input method made of it — and some systems send both where others
	// send one. Judging by where the window was when the frame began is what
	// keeps the key that opens a field from also writing in it.
	was := w.screen

	// The window asks for the keyboard and keeps it. There is one thing here to
	// type into at a time and it is always the form's own field, so a second
	// place for the focus to be would only be a second place to lose it.
	gtx.Execute(key.FocusCmd{Tag: tag})

	// What the window asked for, done here because a value cannot execute
	// anything. Copying goes out now; pasting comes back as an event above.
	if said, wants := w.Copying(); wants {
		gtx.Execute(clipboard.WriteCmd{Type: "application/text",
			Data: io.NopCloser(strings.NewReader(said))})
		w = w.Copied()
	}
	if w.Pasting() {
		gtx.Execute(clipboard.ReadCmd{Tag: tag})
	}

	for {
		raw, ok := gtx.Event(
			pointer.Filter{
				Target: tag,
				Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Scroll,
				// A range, and not the zero one: Gio delivers a scroll only
				// when it falls inside what was asked for, and the zero range
				// asks for scrolls of nothing. Found by a test that scrolled a
				// column and watched it stay where it was.
				ScrollY: pointer.ScrollRange{Min: -reach, Max: reach},
			},
			key.Filter{Focus: tag, Optional: key.ModCommand | key.ModCtrl | key.ModShift},
			// Named explicitly, because Gio hands a filter that names no key
			// everything the system has not already spoken for — and tab is
			// one of the ones it has.
			key.Filter{Focus: tag, Name: key.NameTab},
			key.Filter{Focus: tag, Name: key.NameTab, Required: key.ModShift},
			key.Filter{Focus: tag, Name: key.NameDeleteBackward},
			key.Filter{Focus: tag, Name: key.NameReturn},
			key.Filter{Focus: tag, Name: key.NameEscape},
			key.Filter{Focus: tag, Name: key.NameLeftArrow},
			key.Filter{Focus: tag, Name: key.NameRightArrow},
			key.Filter{Focus: tag, Name: key.NameUpArrow},
			key.Filter{Focus: tag, Name: key.NameDownArrow},
			// And what an input method sends, which is how text actually
			// arrives: a letter is not a key press with a name, it is an edit.
			key.FocusFilter{Target: tag},
			transfer.TargetFilter{Target: tag, Type: "application/text"},
		)
		if !ok {
			return w, pressed, pressedAt
		}
		switch e := raw.(type) {
		case pointer.Event:
			var g Gesture
			w, g = Pointing(w, Gesture{At: pressed, Where: pressedAt}, e.Kind,
				Dp(e.Position, gtx.Metric.PxPerDp),
				Turned(e.Scroll.Y, gtx.Metric.PxPerDp), gtx.Now)
			pressed, pressedAt = g.At, g.Where
		case key.Event:
			if e.State == key.Press {
				w = Typed(w, string(e.Name), e.Modifiers)
			}
		case transfer.DataEvent:
			// What the clipboard held, come back as an event: the read was
			// asked for a frame or more ago and answered when the system
			// could.
			w = pasting(w, e.Open())
		case key.EditEvent:
			// Text goes only where there was already somewhere to put it. The
			// `e` that opened this form arrived as a name first and as text
			// second, and taking the second would write the shortcut into the
			// field it just opened — measured: a title read "eWrite the spec",
			// and every search began with the slash that opened it.
			if takesTyping(was) {
				w = w.Types(e.Text)
			}
		}
	}
}

// takesTyping says whether a screen has somewhere for text to land. The board
// and the panels do not: what they do with a letter is a shortcut.
func takesTyping(s Screen) bool { return s == OnForm || s == OnSearch || s == OnChoosing }
