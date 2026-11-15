package gui

import "cairn/internal/board"

// SizesKept keeps how wide and how tall the window was left, past it closing, for the next one to open with.
type SizesKept interface {
	Keep(Sizes) error
	Kept() (Sizes, error)
}

// SizesOf keeps the sizes through the engine holding the work, when the configuration says to.
func SizesOf(work board.Work) SizesKept { return askedSizes{windowKept{work, "sizes"}} }

type askedSizes struct{ windowKept }

func (a askedSizes) Keep(s Sizes) error { return a.keep(s) }

func (a askedSizes) Kept() (Sizes, error) {
	var s Sizes
	return s, a.read(&s)
}

type noSizes struct{}

func (noSizes) Keep(Sizes) error     { return nil }
func (noSizes) Kept() (Sizes, error) { return Sizes{}, nil }

// KeepingSizes is the window opening with the sizes kept where it is told, and keeping there every size it is given.
func (w Window) KeepingSizes(k SizesKept) Window {
	w.kept = k
	sizes, err := k.Kept()
	if err != nil {
		return w.failing(err)
	}
	w.sizes = sizes
	return w
}

func (w Window) resized(s Sizes) Window {
	w.sizes = s
	if err := w.kept.Keep(s); err != nil {
		return w.failing(err)
	}
	return w
}

// Sizes is every size a person gave the window by hand.
func (w Window) Sizes() Sizes { return w.sizes }
