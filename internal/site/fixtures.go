package site

import (
	"image"
	"time"

	"cairn/internal/engine"
	"cairn/internal/gui"
	"cairn/internal/theme"
	"cairn/internal/tui"
)

const (
	terminalWide, terminalHigh = 132, 34
	windowWide, windowHigh     = 1440, 900
)

// Fixtures is the board every picture of the site is drawn from: the fixtures cairn carries, in a folder of their own.
type Fixtures struct {
	held  *engine.Engine
	pages engine.Pages
}

// OpenFixtures makes the fixtures afresh in the folder and holds them open.
func OpenFixtures(folder string) (Fixtures, error) {
	made, err := engine.MakeFixtures(folder, "", true)
	if err != nil {
		return Fixtures{}, err
	}
	h := engine.Home{Dir: folder, Lookup: func(name string) (string, bool) {
		switch name {
		case "CAIRN_STORE":
			return made.Store, true
		case "CAIRN_NOTES":
			return made.Notes, true
		}
		return "", false
	}}
	held, err := engine.Hold(h, engine.Through("site", h))
	if err != nil {
		return Fixtures{}, err
	}
	return Fixtures{held: held, pages: engine.PagesFrom(h)}, nil
}

// Close lets the fixtures go.
func (f Fixtures) Close() error { return f.held.Close() }

// Terminal is the terminal board as cairn draws it over the fixtures, escapes and all.
func (f Fixtures) Terminal() (string, error) {
	found, err := engine.Look(f.held)
	if err != nil {
		return "", err
	}
	board := tui.New(f.held, tui.GlyphsFor(tui.Unicode), time.Now, terminalWide, terminalHigh).Paged(f.pages)
	looked, _ := board.Update(tui.Looked{Snapshot: found})
	return looked.(tui.Board).Draw(), nil
}

// Window is the window as cairn draws it over the fixtures, in the scheme named.
func (f Fixtures) Window(slug string) (image.Image, error) {
	scheme, err := theme.Find("", slug)
	if err != nil {
		return nil, err
	}
	skin, th := gui.Skinned(scheme.Colours()), gui.Theme()
	w := gui.New(f.held, windowWide, windowHigh).Paged(f.pages).
		Measuring(gui.Measured(th)).MeasuringHeadings(gui.MeasuredHeadings(th)).
		At(time.Now()).Look().Wearing(skin).LessMotion(true)
	shot, err := gui.Shot(w, th, skin)
	if err != nil {
		return nil, err
	}
	return shot, nil
}
