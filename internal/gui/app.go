package gui

import (
	"errors"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"fmt"
	"image"
	"io"
	"os"
	"runtime"
	rdebug "runtime/debug"
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
	"gioui.org/io/system"
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
	return abs(at.X-was.X) <= Wobble && abs(at.Y-was.Y) <= Wobble
}

// Typed says what a key does.
func Typed(w Window, name string, with key.Modifiers) Window {
	w = w.interrupted()
	said, ok := keyOf(name, with)
	if !ok {
		return w
	}
	return w.Key(said)
}

var shiftedFigures = map[string]string{"!": "1", "@": "2", "#": "3", "$": "4", "%": "5", "^": "6", "&": "7", "*": "8", "(": "9"}

func keyOf(name string, with key.Modifiers) (string, bool) {
	said, isNamed := naming(name)
	if with.Contain(key.ModAlt) && !moving(said) && said != "backspace" {
		return "", false
	}
	if moving(said) {
		return modified(with, key.ModCommand|key.ModCtrl, "cmd+") +
			modified(with, key.ModAlt, "alt+") + modified(with, key.ModShift, "shift+") + said, true
	}
	if with.Contain(key.ModCommand) || with.Contain(key.ModCtrl) {
		if figure, ok := shiftedFigures[said]; ok && with.Contain(key.ModShift) {
			said = figure
		}
		if with.Contain(key.ModShift) && (aLetter(said) || aFigure(said)) {
			return "cmd+shift+" + said, true
		}
		return "cmd+" + said, true
	}
	if with.Contain(key.ModAlt) && said == "backspace" {
		return "alt+backspace", true
	}
	if with.Contain(key.ModShift) && (isNamed || aLetter(said)) && !deleting(said) {
		return "shift+" + said, true
	}
	return said, true
}

func aFigure(name string) bool { return len(name) == 1 && name[0] >= '0' && name[0] <= '9' }

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

// Here is where a window stands among the others on this machine: the store it shows, and how it reaches another workspace, folder or store to show in a window of its own.
type Here struct {
	Store string
	Reach Reach
}

// Run opens a window wearing the skin given and keeps it until it is closed; a window already showing the store is brought forward instead, and none opened.
func Run(work engine.Work, pages engine.Pages, look Skin, drafts Drafts, sizes SizesKept, noted *debug.Log, width, height int, here Here) {
	window := new(app.Window)
	windows := WindowsHere(os.LookupEnv)
	claim, err := windows.Claim(here.Store, func() { window.Perform(system.ActionRaise) })
	if errors.Is(err, ErrShownElsewhere) {
		_, _ = fmt.Fprintf(os.Stderr, "cairn: %v\n", err)
		return
	}
	if claim != nil {
		defer func() { _ = claim.Close() }()
	}
	leaveConsole()
	if address, kept := ibusKeptOut(runtime.GOOS, os.Getenv); kept {
		_ = os.Setenv("IBUS_ADDRESS", address)
	}
	knownAs()
	showLogo()
	go func() {
		leaving := func(code int) {
			if noted != nil {
				noted.Leave(os.Stderr)
			}
			os.Exit(code)
		}
		defer func() {
			if stopped := recover(); stopped != nil && noted != nil {
				noted.Panicked(stopped, rdebug.Stack())
				noted.Leave(os.Stderr)
				panic(stopped)
			}
		}()
		ended(loop(window, work, pages, look, drafts, sizes, noted, width, height, apartFrom(here, windows)), os.Stderr, leaving)
	}()
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

type elsewhereOf struct {
	reach   Reach
	windows Windows
}

func apartFrom(here Here, windows Windows) elsewhereOf {
	return elsewhereOf{reach: here.Reach, windows: windows}
}

func (a elsewhereOf) opening(to engine.Elsewhere) Went {
	if a.reach == nil {
		return Went{Err: errNoElsewhere}
	}
	return OpeningApart(to, a.reach, a.windows, StartCairn)
}

var errNoElsewhere = errors.New("this window cannot open another")

func loop(window *app.Window, work engine.Work, pages engine.Pages, skin Skin, drafts Drafts, sizes SizesKept, noted *debug.Log, width, height int, elsewhere elsewhereOf) error {
	seen := ""

	theme := Theme()
	board := New(work, width, height).Paged(pages).Drafting(drafts).Measuring(Measured(theme)).MeasuringHeadings(MeasuredHeadings(theme)).Seeing(elsewhere.windows.Shows)
	if sizes != nil {
		board = board.KeepingSizes(sizes)
	}
	board = board.At(time.Now()).Look().ReadingConfiguration().Wearing(skin)
	window.Option(app.Title(board.Title()), app.Size(unit.Dp(width), unit.Dp(height)), app.MinSize(unit.Dp(Narrowest), unit.Dp(Shortest)))
	if data, starts, ok := desktopHome(); ok {
		board = board.Settling(data, starts)
	}

	files, listen := systemFiles(window)
	filed := make(chan Filed, 1)
	went := make(chan Went, 1)
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
		if onTheWindowsThread(window.Run, func() bool { return installMenu(event, pages, picks, window.Invalidate) }) {
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
			case g := <-went:
				board = board.Went(g)
			default:
			}
			board = board.Resized(Sized(gtx.Constraints.Max, gtx.Metric.PxPerDp)).At(gtx.Now)
			if due(looked, gtx.Now) {
				board, looked = board.Wonder(), gtx.Now
			}
			board, pressed, pressedAt = heard(gtx, board, pressed, pressedAt, noted)
			if now := board.Seen(); noted != nil && now != seen {
				noted.Say("window", now)
				seen = now
			}
			if ask, asked := board.FileAsked(); asked {
				board = board.FileTaken()
				go func() {
					filed <- Filing(ask, files)
					window.Invalidate()
				}()
			}
			if to, asked := board.ApartAsked(); asked {
				board = board.ApartTaken()
				go func() {
					went <- elsewhere.opening(to)
					window.Invalidate()
				}()
			}
			if path, asked := board.OpenAsked(); asked {
				board = board.OpenTaken().Opened(HandOver(path))
			}
			board = board.Born(gtx.Now)
			board.tellingTheBars()
			Draw(NewKit(gtx, theme, board.Skin(), &Drawn{}), board)
			_, building := board.Intro()
			if _, rising := board.PaletteRise(); building || rising {
				gtx.Execute(op.InvalidateCmd{})
			}
			e.Frame(gtx.Ops)
			if board.Leaving() {
				window.Perform(system.ActionClose)
			}
		}
	}
}

// onTheWindowsThread runs f on the thread that runs the system's window, and says what f answered: called from this loop while the window waits for the loop to take an event, a call that waits on the window would wait for ever.
func onTheWindowsThread(run func(func()), f func() bool) bool {
	took := false
	run(func() { took = f() })
	return took
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

func heard(gtx layout.Context, w Window, pressed time.Time, pressedAt Point, noted *debug.Log) (Window, time.Time, Point) {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	event.Op(gtx.Ops, tag)
	area.Pop()

	cursors(gtx, w)

	was := w.screen
	named := ""

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
		if noted != nil {
			if said, ok := KeyTrace(raw, noted.Typed); ok {
				noted.Say("window", said)
			}
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
			if noted != nil && PointerNoted(e.Kind) {
				noted.Say("window", PointerTrace(e, gtx.Metric.PxPerDp, w))
			}
		case key.Event:
			if e.State == key.Press {
				named = string(e.Name)
				w = Typed(w, named, e.Modifiers)
			}
		case transfer.DataEvent:
			w = pasting(w, e.Open())
		case key.EditEvent:
			if takesTyping(was) {
				w = w.Edits(e.Range.Start, e.Range.End, e.Text)
			} else if symbol, ok := symbolTyped(e.Text, named); ok {
				w = w.Key(symbol)
			}
		}
	}
}

func cursors(gtx layout.Context, w Window) {
	if w.widening.On {
		area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
		cursorOf(w.PointerAt(Point{})).Add(gtx.Ops)
		area.Pop()
		return
	}
	scaled := inDp(gtx)
	defer scaled.Pop()
	for _, grab := range w.Grabs() {
		area := clip.Rect(rect(grab.Box)).Push(gtx.Ops)
		cursorOf(grab.Takes.Pointer()).Add(gtx.Ops)
		area.Pop()
	}
}

func cursorOf(p Pointer) pointer.Cursor {
	if p == RowResize {
		return pointer.CursorRowResize
	}
	return pointer.CursorColResize
}

func telling(gtx layout.Context, w Window) {
	said, _ := w.Typed()
	gtx.Execute(key.SnippetCmd{Tag: tag, Snippet: key.Snippet{
		Range: key.Range{End: len([]rune(said.Value))}, Text: said.Value}})
	gtx.Execute(key.SelectionCmd{Tag: tag, Range: key.Range{Start: said.At, End: said.At + said.Held}})
}

// symbolTyped is the key a symbol the system says was typed stands for, on a screen that takes no text: a lone mark that is neither a letter, a figure nor a space, and that the press before it did not name already — Windows names a key by its place on an American keyboard, so ? typed on another arrives as the press of its key and, apart, as the mark.
func symbolTyped(text, named string) (string, bool) {
	runes := []rune(text)
	if len(runes) != 1 || text == named || unicode.IsLetter(runes[0]) || unicode.IsDigit(runes[0]) || unicode.IsSpace(runes[0]) || !unicode.IsGraphic(runes[0]) {
		return "", false
	}
	return text, true
}

func takesTyping(s Screen) bool {
	return s == OnForm || s == OnSearch || s == OnChoosing || s == OnPalette || s == OnLabels || s == OnWorkspaces || s == OnKeeping || s == OnViews
}
