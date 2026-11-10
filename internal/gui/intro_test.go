package gui

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"cairn/internal/config"
	"cairn/internal/engine"
)

func introBoard() Window {
	cards := []engine.Card{
		{Number: 18, Title: "Write the packaging spec", State: "backlog", Priority: "high"},
		{Number: 39, Title: "Back up before the migration", State: "backlog", Priority: "critical"},
		{Number: 20, Title: "Settle the multi-chart question", State: "backlog", Priority: "normal"},
		{Number: 38, Title: "Fix the boot loop", State: "todo", Priority: "high"},
		{Number: 21, Title: "Write the Flows section", State: "todo", Priority: "normal"},
		{Number: 24, Title: "Test a rollback on an empty store", State: "current", Priority: "high", InFocus: true},
		{Number: 30, Title: "Update the README", State: "current", Priority: "normal"},
		{Number: 19, Title: "Read ADR 0014", State: "done", Priority: "normal"},
		{Number: 31, Title: "Send the invoice", State: "done", Priority: "low"},
	}
	return holding(newFake(), cards...).At(time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC))
}

func playing(w Window, name string, at time.Duration) Window {
	w = w.Born(w.now).Introducing(name)
	return w.At(w.now.Add(at))
}

func TestEveryIntroEndsOnTheStillBoard(t *testing.T) {
	w := introBoard()
	stage := w.Stage()
	for _, name := range config.Intros() {
		lasts := IntroLasts(name, stage)
		if lasts <= 0 || lasts > 2200*time.Millisecond {
			t.Errorf("%s lasts %v", name, lasts)
		}
		end := Staging(name, lasts, stage)
		if !end.Camera.Resting() || !end.Top.Resting() || !end.Bar.Resting() || end.Wave != nil || end.Sweep != nil || end.Grains != nil || end.Beat != nil || end.Stones || len(end.Order) > 0 {
			t.Errorf("%s does not end on the still board: %+v", name, end)
		}
		for i, pose := range append(end.Cards, end.Grounds...) {
			if !pose.Resting() {
				t.Errorf("%s ends with its %d-th thing posed %+v", name, i, pose)
			}
		}
		moving := false
		for _, share := range []float64{0.25, 0.5} {
			mid := Staging(name, time.Duration(float64(lasts)*share), stage)
			for _, pose := range append(mid.Cards, mid.Grounds...) {
				moving = moving || !pose.Resting()
			}
		}
		if !moving {
			t.Errorf("%s holds the board still all the way through", name)
		}
	}
}

func TestAnIntroIsAPureFunctionOfTime(t *testing.T) {
	stage := introBoard().Stage()
	for _, name := range config.Intros() {
		at := 400 * time.Millisecond
		if a, b := Staging(name, at, stage), Staging(name, at, stage); !reflect.DeepEqual(a, b) {
			t.Errorf("%s drew two boards at the same instant", name)
		}
	}
}

func TestABusyBoardArrivesNoSlower(t *testing.T) {
	var cards []engine.Card
	for n := range 80 {
		cards = append(cards, engine.Card{Number: n + 1, Title: "Task " + strconv.Itoa(n+1), State: []string{"backlog", "todo", "current", "done"}[n%4]})
	}
	stage := holding(newFake(), cards...).Stage()
	for _, name := range config.Intros() {
		if lasts := IntroLasts(name, stage); lasts > 2200*time.Millisecond {
			t.Errorf("%s lasts %v on a busy board", name, lasts)
		}
	}
}

func TestTheIntroIsTheOneTheConfigurationNamesAndNoneWithLessMotion(t *testing.T) {
	w := introBoard()
	if got := w.IntroShown(); got != config.IntroRipple {
		t.Errorf("with nothing said, the intro is %q", got)
	}
	w.configured.Window = &engine.WindowConfig{Intro: config.IntroFlap}
	if got := w.IntroShown(); got != config.IntroFlap {
		t.Errorf("the configuration names flap, the intro is %q", got)
	}
	if _, on := w.Born(w.now).Intro(); !on {
		t.Error("no intro plays")
	}
	if _, on := w.LessMotion(true).Born(w.now).Intro(); on {
		t.Error("an intro plays though the system asks for less motion")
	}
}

func TestADesktopTurningItsAnimationsOffAsksForLessMotion(t *testing.T) {
	for said, off := range map[string]bool{"false\n": true, "true\n": false, "": false} {
		if AnimationsOff(said) != off {
			t.Errorf("%q is read as animations off: %v", said, !off)
		}
	}
}

func TestTheIntroChooserWalksPlaysAndKeeps(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"}).At(time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC))
	w.screen = OnConfiguration
	w = w.openingIntros()
	if got := w.IntroShown(); got != config.IntroRipple {
		t.Errorf("the chooser opens on %q", got)
	}
	w = Typed(w, "↓", 0)
	if _, on := w.Intro(); !on || w.IntroShown() != config.Intros()[1] {
		t.Errorf("walking down plays %q: %v", w.IntroShown(), on)
	}
	if rows := w.IntroRows(); len(rows) != len(config.Intros()) || !rows[0].InForce || rows[1].Says != IntroSays(config.Intros()[1]) {
		t.Errorf("the chooser offers %+v", rows)
	}
	if got := Heads(w); got != "Intro of here" {
		t.Errorf("the chooser is headed %q", got)
	}
	for _, size := range [][2]int{{720, 560}, {1200, 800}} {
		sized := w.Resized(size[0], size[1])
		box := sized.SettingsBox()
		for _, offer := range sized.offersAlong(box) {
			if offer.X < box.X || offer.X+offer.W > box.X+box.W {
				t.Errorf("at %v an offer %+v runs out of the chooser %+v", size, offer, box)
			}
		}
	}
	w = w.clickingIntros(Point{X: w.IntroRows()[3].At.X + 4, Y: w.IntroRows()[3].At.Y + 4})
	if w.IntroShown() != config.Intros()[3] {
		t.Errorf("a click plays %q", w.IntroShown())
	}
	w = w.Key("esc")
	if w.introsOpen || w.screen != OnConfiguration {
		t.Error("esc does not go back to the settings")
	}
}

func TestEachIntroIsDrawnAtAnyInstantAndEndsOnTheStillBoard(t *testing.T) {
	still := introBoard()
	shots := os.Getenv("INTRO_SHOTS")
	want := drawn(t, still)
	for _, name := range config.Intros() {
		lasts := IntroLasts(name, still.Stage())
		for _, share := range []float64{0.1, 0.3, 0.6} {
			at := time.Duration(float64(lasts) * share)
			shot := drawn(t, playing(still, name, at))
			if bytes.Equal(shot.Pix, want.Pix) {
				t.Errorf("%s at %v draws the still board", name, at)
			}
			if shots != "" {
				keep(t, filepath.Join(shots, name+"-"+strconv.Itoa(int(at/time.Millisecond))+".png"), shot)
			}
		}
		if end := drawn(t, playing(still, name, lasts)); !bytes.Equal(end.Pix, want.Pix) {
			t.Errorf("%s does not end on the still board", name)
		}
	}
}

func keep(t *testing.T, path string, shot *image.RGBA) {
	t.Helper()
	var said bytes.Buffer
	if err := png.Encode(&said, shot); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, said.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}
