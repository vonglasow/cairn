package gui

import (
	"testing"
	"time"
)

func TestTheHelpIsClosedWhereItsCloseIsDrawn(t *testing.T) {
	middle := func(b Box) Point { return Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	open := New(newFake(), 1600, 2000).Key("?")
	if open.Screen() != OnHelp {
		t.Fatalf("the help did not open: %v", open.Screen())
	}
	box := open.HelpBox()
	if got := open.clicked(middle(PanelOfferAt(box, 0))); got.Screen() != OnTheBoard {
		t.Errorf("clicking Close where it is drawn left the window on %v", got.Screen())
	}
	under := Point{X: box.X + box.W/2, Y: box.Y + box.H + apart}
	if got := open.clicked(under); got.Screen() != OnTheBoard {
		t.Errorf("clicking under the help left the window on %v", got.Screen())
	}
}

func TestTheWheelWalksTheHelp(t *testing.T) {
	open := New(newFake(), 1024, 640).Key("?")
	inside := Point{X: open.HelpBox().X + 40, Y: open.HelpBox().Y + 80}
	turned := open.wheeling(3*lineHigh, inside, time.Time{})
	if got := turned.HelpFrom(); got != 3 {
		t.Errorf("three lines of wheel left the help %d lines down, want 3", got)
	}
	far := open.wheeling(1000*lineHigh, inside, time.Time{})
	if got, want := far.HelpFrom()+far.HelpRoom(), len(far.HelpRows()); got != want {
		t.Errorf("a long turn left the help ending at %d of %d", got, want)
	}
}

func TestTheHelpDrawsItsBarWhenItScrolls(t *testing.T) {
	open := New(newFake(), 1024, 640).Key("?")
	bar, ok := open.HelpBar()
	if !ok {
		t.Fatal("a help longer than its panel has no bar")
	}
	if got := ink(drawn(t, open), bar, Dark().Panel); got == 0 {
		t.Error("the bar beside the help is not drawn")
	}
}

func TestTheHelpIsLaidOutToThePixel(t *testing.T) {
	tall := New(newFake(), 1600, 2000).Key("?")
	if got, want := tall.HelpBox(), (Box{X: 200, Y: 166, W: 1200, H: 106 + len(tall.HelpRows())*apart}); got != want {
		t.Errorf("a tall window's help is %+v, want %+v — as tall as its %d lines", got, want, len(tall.HelpRows()))
	}
	name, value := HelpRowAt(tall.HelpBox(), 2)
	if want := (Box{X: 212, Y: 280, W: 158, H: 18}); name != want {
		t.Errorf("the third key is at %+v, want %+v", name, want)
	}
	if want := (Box{X: 370, Y: 280, W: 1002, H: 18}); value != want {
		t.Errorf("what the third key does is at %+v, want %+v", value, want)
	}

	short := New(newFake(), 1024, 660).Key("?")
	if got := short.HelpRoom(); got != 13 {
		t.Errorf("a short window's help shows %d lines, want 13", got)
	}
	bar, ok := short.HelpBar()
	if want := (Box{X: 874, Y: 105, W: 10, H: 402}); !ok || bar != want {
		t.Errorf("the bar beside a short help is %+v, want %+v", bar, want)
	}
}

func TestTheHelpWalksBackUp(t *testing.T) {
	down := New(newFake(), 1024, 640).Key("?").Key("down").Key("j")
	if got := down.HelpFrom(); got != 2 {
		t.Fatalf("two steps down left the help %d lines down", got)
	}
	if got := down.Key("k").HelpFrom(); got != 1 {
		t.Errorf("k left the help %d lines down, want 1", got)
	}
	if got := down.Key("up").HelpFrom(); got != 1 {
		t.Errorf("up left the help %d lines down, want 1", got)
	}
}
