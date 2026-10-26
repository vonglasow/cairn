package gui

import (
	"cairn/internal/engine"
	"fmt"
	"image"
	"io"
	"os"
	"runtime"
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
)

// Beat is how often the window asks the store whether anybody else has written to it.
func Beat() time.Duration { return 2 * time.Second }

// Still is how long the wheel has to be left alone for the next turn to be a turn of its own.
func Still() time.Duration { return 150 * time.Millisecond }

// Twin is how long after a press a second one at the same place is the same gesture.
func Twin() time.Duration { return 400 * time.Millisecond }

// Dp turns where a pointer says it is into where the board thinks it is.
func Dp(at f32.Point, perDp float32) Point {
	return Point{X: int(at.X / perDp), Y: int(at.Y / perDp)}
}

// Sized turns the size a window reports, in pixels, into the size this board works in.
func Sized(max image.Point, perDp float32) (width, height int) {
	return int(float32(max.X) / perDp), int(float32(max.Y) / perDp)
}

// Turned is the same for how far a wheel says it went.
func Turned(by, perDp float32) int { return int(by / perDp) }

func pasting(w Window, said io.Reader) Window {
	read, err := io.ReadAll(said)
	if err != nil {
		return w.Pasted("")
	}
	return w.Pasted(string(read))
}

// Acts says whether a pointer event carries a button this board answers as a click.
func Acts(b pointer.Buttons) bool { return b == 0 || b == pointer.ButtonPrimary }

// Gesture is what a pointer has said so far: when the last press landed and where, which is the whole of what tells the press after it from the second half of a double-click.
type Gesture struct {
	At     time.Time
	Where  Point
	Clicks int
}

// Pointing is everything a pointer event does, the double-click included, and the gesture it leaves behind.
func Pointing(w Window, g Gesture, kind pointer.Kind, at Point, by int, now time.Time) (Window, Gesture) {
	if kind == pointer.Press {
		w = w.interrupted()
	}
	if kind == pointer.Scroll {
		return w.wheeling(by, at, now), g
	}
	if kind == pointer.Press {
		if Twice(g.At, now, g.Where, at) {
			if g.Clicks >= 2 {
				return w.TripleClick(at), Gesture{}
			}
			after := w.DoubleClick(at)
			if after.screen != w.screen {
				return after, Gesture{}
			}
			return after, Gesture{At: now, Where: at, Clicks: 2}
		}
		g = Gesture{At: now, Where: at, Clicks: 1}
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
	case pointer.Move, pointer.Enter:
		return w.Hovering(at)
	case pointer.Leave:
		return w.Away()
	}
	return w
}

// Twice says whether a press is the second half of a double-click: soon enough after the last one, and close enough to it that the hand did not move.
func Twice(before, now time.Time, was, at Point) bool {
	if before.IsZero() || now.Sub(before) > Twin() {
		return false
	}
	return abs(at.X-was.X) <= Grip && abs(at.Y-was.Y) <= Grip
}

// Typed says what a key does.
func Typed(w Window, name string, with key.Modifiers) Window {
	w = w.interrupted()
	said, isNamed := naming(name)
	if with.Contain(key.ModAlt) && !moving(said) && said != "backspace" {
		return w
	}
	if moving(said) {
		return w.Key(modified(with, key.ModCommand|key.ModCtrl, "cmd+") +
			modified(with, key.ModAlt, "alt+") + modified(with, key.ModShift, "shift+") + said)
	}
	if with.Contain(key.ModCommand) || with.Contain(key.ModCtrl) {
		return w.Key("cmd+" + said)
	}
	if with.Contain(key.ModAlt) && said == "backspace" {
		return w.Key("alt+backspace")
	}
	if with.Contain(key.ModShift) && (isNamed || aLetter(said)) && !deleting(said) {
		return w.Key("shift+" + said)
	}
	return w.Key(said)
}

func deleting(said string) bool { return said == "backspace" || said == "delete-forward" }

func moving(said string) bool {
	switch said {
	case "left", "right", "up", "down", "home", "end":
		return true
	}
	return false
}

func modified(with, any key.Modifiers, word string) string {
	if with&any != 0 {
		return word
	}
	return ""
}

func aLetter(name string) bool {
	r := []rune(name)
	return len(r) == 1 && unicode.IsLetter(r[0])
}

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
	case key.NameHome:
		return "home", true
	case key.NameEnd:
		return "end", true
	case key.NameDeleteForward:
		return "delete-forward", true
	}
	return strings.ToLower(name), false
}

const reach = 1 << 20

var tag = new(int)

// Run opens a window and keeps it until it is closed.
func Run(work engine.Work, pages engine.Pages, drafts Drafts, sizes SizesKept, width, height int) {
	leaveConsole()
	if address, kept := ibusKeptOut(runtime.GOOS, os.Getenv); kept {
		_ = os.Setenv("IBUS_ADDRESS", address)
	}
	knownAs()
	showLogo()
	go func() { ended(loop(work, pages, drafts, sizes, width, height), os.Stderr, os.Exit) }()
	app.Main()
}

func ended(err error, said io.Writer, exit func(int)) {
	if err == nil {
		exit(0)
		return
	}
	_, _ = fmt.Fprintf(said, "cairn: the window failed: %v\n", err)
	exit(1)
}

func loop(work engine.Work, pages engine.Pages, drafts Drafts, sizes SizesKept, width, height int) error {
	window := new(app.Window)
	window.Option(app.Title("cairn"), app.Size(unit.Dp(width), unit.Dp(height)))
	traced := tracing(os.Getenv, os.Stderr)

	theme, skin := Theme(), Dark()
	board := New(work, width, height).Paged(pages).Drafting(drafts).Measuring(Measured(theme))
	if sizes != nil {
		board = board.KeepingSizes(sizes)
	}
	board = board.Look()
	if data, starts, ok := desktopHome(); ok {
		board = board.Settling(data, starts)
	}

	files, listen := systemFiles(window)
	filed := make(chan Filed, 1)
	picks := make(chan string, 8)

	var ops op.Ops
	var pressed time.Time
	var pressedAt Point
	var looked time.Time

	beating := time.NewTicker(Beat())
	defer beating.Stop()
	gone := make(chan struct{})
	defer close(gone)
	go func() {
		for {
			select {
			case <-beating.C:
				window.Invalidate()
			case <-gone:
				return
			}
		}
	}()

	for {
		event := window.Event()
		listen(event)
		dress(event)
		if installMenu(event, pages, picks, window.Invalidate) {
			board = board.Native(true)
		}
		switch e := event.(type) {
		case app.DestroyEvent:
			board.Closing()
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			select {
			case got := <-filed:
				board = board.Filed(got)
			case item := <-picks:
				board = board.Picked(item)
			default:
			}
			board = board.Resized(Sized(gtx.Constraints.Max, gtx.Metric.PxPerDp)).At(gtx.Now)
			if due(looked, gtx.Now) {
				board, looked = board.Wonder(), gtx.Now
			}
			board, pressed, pressedAt = heard(gtx, board, pressed, pressedAt, traced)
			if ask, asked := board.FileAsked(); asked {
				board = board.FileTaken()
				go func() {
					filed <- Filing(ask, files)
					window.Invalidate()
				}()
			}
			board = board.Born(gtx.Now)
			board.tellingTheBars()
			Draw(gtx, theme, skin, board)
			if _, building := board.Intro(); building {
				gtx.Execute(op.InvalidateCmd{})
			}
			e.Frame(gtx.Ops)
		}
	}
}

func due(looked, now time.Time) bool { return now.Sub(looked) >= Beat() }

func listening() []event.Filter {
	return []event.Filter{
		pointer.Filter{
			Target: tag,
			Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Scroll |
				pointer.Move | pointer.Enter | pointer.Leave,
			ScrollY: pointer.ScrollRange{Min: -reach, Max: reach},
		},
		key.Filter{Focus: tag, Optional: key.ModCommand | key.ModCtrl | key.ModShift | key.ModAlt},
		key.Filter{Focus: tag, Name: key.NameTab},
		key.Filter{Focus: tag, Name: key.NameTab, Required: key.ModShift},
		key.Filter{Focus: tag, Name: key.NameDeleteBackward},
		key.Filter{Focus: tag, Name: key.NameReturn},
		key.Filter{Focus: tag, Name: key.NameEscape},
		key.Filter{Focus: tag, Name: key.NameLeftArrow},
		key.Filter{Focus: tag, Name: key.NameRightArrow},
		key.Filter{Focus: tag, Name: key.NameUpArrow},
		key.Filter{Focus: tag, Name: key.NameDownArrow},
		key.FocusFilter{Target: tag},
		transfer.TargetFilter{Target: tag, Type: "application/text"},
	}
}

func heard(gtx layout.Context, w Window, pressed time.Time, pressedAt Point, traced io.Writer) (Window, time.Time, Point) {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	area.Pop()

	was := w.screen

	gtx.Execute(key.FocusCmd{Tag: tag})

	if said, wants := w.Copying(); wants {
		gtx.Execute(clipboard.WriteCmd{Type: "application/text",
			Data: io.NopCloser(strings.NewReader(said))})
		w = w.Copied()
	}
	if w.Pasting() {
		gtx.Execute(clipboard.ReadCmd{Tag: tag})
	}

	for {
		raw, ok := gtx.Event(listening()...)
		if !ok {
			telling(gtx, w)
			return w, pressed, pressedAt
		}
		if said, ok := KeyTrace(raw); ok && traced != nil {
			_, _ = fmt.Fprintln(traced, said)
		}
		switch e := raw.(type) {
		case pointer.Event:
			if e.Buttons == pointer.ButtonSecondary && e.Kind == pointer.Press {
				w = w.interrupted().ContextAt(Dp(e.Position, gtx.Metric.PxPerDp))
				continue
			}
			if !Acts(e.Buttons) {
				continue
			}
			var g Gesture
			w, g = Pointing(w.KeysDown(e.Modifiers), Gesture{At: pressed, Where: pressedAt}, e.Kind,
				Dp(e.Position, gtx.Metric.PxPerDp),
				Turned(e.Scroll.Y, gtx.Metric.PxPerDp), gtx.Now)
			pressed, pressedAt = g.At, g.Where
			if traced != nil {
				_, _ = fmt.Fprintln(traced, PointerTrace(e, gtx.Metric.PxPerDp, w))
			}
		case key.Event:
			if e.State == key.Press {
				w = Typed(w, string(e.Name), e.Modifiers)
			}
		case transfer.DataEvent:
			w = pasting(w, e.Open())
		case key.EditEvent:
			if takesTyping(was) {
				w = w.Edits(e.Range.Start, e.Range.End, e.Text)
			}
		}
	}
}

func telling(gtx layout.Context, w Window) {
	said, _ := w.Typed()
	gtx.Execute(key.SnippetCmd{Tag: tag, Snippet: key.Snippet{
		Range: key.Range{End: len([]rune(said.Value))}, Text: said.Value}})
	gtx.Execute(key.SelectionCmd{Tag: tag, Range: key.Range{Start: said.At, End: said.At + said.Held}})
}

func takesTyping(s Screen) bool { return s == OnForm || s == OnSearch || s == OnChoosing }
