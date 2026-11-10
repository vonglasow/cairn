package gui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/config"
)

// The offers of the intro chooser beside keeping the one it is on.
const (
	ReplayIntro = "Replay"
	KeepIntro   = "Keep "
)

// IntroRow is an intro the chooser offers: its name in the configuration, what it is called, how long it lasts on the board shown, whether it is the one in force, and where it is drawn.
type IntroRow struct {
	Name, Says, Lasts string
	InForce           bool
	At                Box
}

const (
	introsWide = 420
	introHigh  = 46
)

// IntrosBox is the panel the intro chooser is drawn in: narrow, against the right of the window, so the board behind can be watched.
func IntrosBox(width, height int) Box {
	wide, high := min(introsWide, width-2*settingsMargin), height-2*settingsMargin
	return Box{X: width - settingsMargin - wide, Y: settingsMargin, W: wide, H: high}
}

func (w Window) openingIntros() Window {
	w.introsOpen, w.settingsAt = true, 0
	w.introOn = max(0, slices.Index(config.Intros(), w.IntroShown()))
	return w.replaying()
}

func (w Window) inIntros(name string) Window {
	step := map[string]int{"up": -1, "k": -1, "down": 1, "j": 1}[name]
	if step == 0 {
		return w
	}
	return w.onIntro(min(len(config.Intros())-1, max(0, w.introOn+step)))
}

func (w Window) onIntro(at int) Window {
	w.introOn = at
	rows, region := w.IntroRows(), w.introsRegion()
	if row := rows[at].At; row.Y < region.Y {
		w.settingsAt -= region.Y - row.Y
	} else if row.Y+row.H > region.Y+region.H {
		w.settingsAt += row.Y + row.H - region.Y - region.H
	}
	w.settingsAt = max(0, w.settingsAt)
	return w.replaying()
}

func (w Window) clickingIntros(p Point) Window {
	if !w.introsRegion().In(p) {
		return w
	}
	for i, row := range w.IntroRows() {
		if row.At.In(p) {
			if i == w.introOn {
				return w.keepingIntro()
			}
			return w.onIntro(i)
		}
	}
	return w
}

func (w Window) keepingIntro() Window {
	name := w.IntroShown()
	w = w.leavingIntros()
	return w.setting("window.intro", name)
}

func (w Window) leavingIntros() Window {
	w.introsOpen, w.settingsAt = false, 0
	return w.interrupted()
}

func (w Window) introsRegion() Box {
	box := w.SettingsBox()
	body := SettingsBody(box)
	return Box{X: body.X + padding, Y: body.Y + inset, W: body.W - 2*padding, H: body.H - inset}
}

// IntroRows is every intro the chooser offers, in the order the configuration names them, laid down its panel and scrolled.
func (w Window) IntroRows() []IntroRow {
	region, stage := w.introsRegion(), w.Stage()
	inForce := config.DefaultIntro
	if c := w.configured.Window; c != nil && c.Intro != "" {
		inForce = c.Intro
	}
	var rows []IntroRow
	for i, name := range config.Intros() {
		lasts := IntroLasts(name, stage)
		rows = append(rows, IntroRow{
			Name: name, Says: IntroSays(name), InForce: name == inForce,
			Lasts: strconv.FormatFloat(lasts.Seconds(), 'f', 1, 64) + " s",
			At:    Box{X: region.X, Y: region.Y - w.settingsAt + i*introHigh, W: region.W, H: introHigh},
		})
	}
	return rows
}

func (w Window) introsOffers() []string {
	return []string{KeepIntro + w.IntroShown(), ReplayIntro, "Cancel"}
}

func (w Window) introsShortcut(offer string) string {
	switch {
	case strings.HasPrefix(offer, KeepIntro):
		return "enter"
	case offer == ReplayIntro:
		return "r"
	case offer == "Cancel":
		return "esc"
	}
	return ""
}

func (w Window) choosingOnIntros(name string) (Window, bool) {
	switch {
	case strings.HasPrefix(name, KeepIntro):
		return w.keepingIntro(), true
	case name == ReplayIntro:
		return w.replaying(), true
	case name == "Cancel":
		return w.leavingIntros(), true
	}
	return w, false
}
