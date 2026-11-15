package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"cairn/internal/gui/motion"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/store/files"
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
	At                geom.Box
}

const (
	introsWide = 420
	introHigh  = 46
)

// IntrosBox is the panel the intro chooser is drawn in: narrow, against the right of the window, so the board behind can be watched.
func IntrosBox(width, height int) geom.Box {
	wide, high := min(introsWide, width-2*settingsMargin), height-2*settingsMargin
	return geom.Box{X: width - settingsMargin - wide, Y: settingsMargin, W: wide, H: high}
}

func (w Window) openingIntros() Window {
	w.IntrosOpen, w.SettingsAt = true, 0
	w.IntroFrom, w.Tab = w.Tab, board.Tasks
	w.IntroOn = max(0, slices.Index(files.Intros(), w.IntroShown()))
	return w.replaying()
}

func (w Window) inIntros(name string) Window {
	step := map[string]int{"up": -1, "k": -1, "down": 1, "j": 1}[name]
	if step == 0 {
		return w
	}
	return w.onIntro(min(len(files.Intros())-1, max(0, w.IntroOn+step)))
}

func (w Window) onIntro(at int) Window {
	w.IntroOn = at
	rows, region := w.IntroRows(), w.introsRegion()
	if row := rows[at].At; row.Y < region.Y {
		w.SettingsAt -= region.Y - row.Y
	} else if row.Y+row.H > region.Y+region.H {
		w.SettingsAt += row.Y + row.H - region.Y - region.H
	}
	w.SettingsAt = max(0, w.SettingsAt)
	return w.replaying()
}

func (w Window) clickingIntros(p geom.Point) Window {
	if !w.introsRegion().In(p) {
		return w
	}
	for i, row := range w.IntroRows() {
		if row.At.In(p) {
			if i == w.IntroOn {
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
	w.IntrosOpen, w.SettingsAt = false, 0
	w.Tab = w.IntroFrom
	return w.interrupted()
}

func (w Window) introsRegion() geom.Box {
	box := w.SettingsBox()
	body := SettingsBody(box)
	return geom.Box{X: body.X + padding, Y: body.Y + inset, W: body.W - 2*padding, H: body.H - inset}
}

// IntroRows is every intro the chooser offers, in the order the configuration names them, laid down its panel and scrolled.
func (w Window) IntroRows() []IntroRow {
	region, stage := w.introsRegion(), w.Stage()
	inForce := files.DefaultIntro
	if c := w.Configured.Window; c != nil && c.Intro != "" {
		inForce = c.Intro
	}
	var rows []IntroRow
	for i, name := range files.Intros() {
		lasts := motion.IntroLasts(name, stage)
		rows = append(rows, IntroRow{
			Name: name, Says: motion.IntroSays(name), InForce: name == inForce,
			Lasts: strconv.FormatFloat(lasts.Seconds(), 'f', 1, 64) + " s",
			At:    geom.Box{X: region.X, Y: region.Y - w.SettingsAt + i*introHigh, W: region.W, H: introHigh},
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
