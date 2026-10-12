package gui

// The faces the window writes in: Go's own, and the few symbols they lack —
// the keys of a Mac keyboard, a tick, a cross — cut from DejaVu Sans into a
// font of a dozen kilobytes (symbols/, its licence beside it). A character no
// face cairn carries has sends the shaper to the system's fonts: every
// process that drew one read them off the disk and kept them, and on a system
// without them it drew an empty box. What a person types still reaches the
// system's fonts; what the window writes never needs to. See docs/gui.md.

import (
	_ "embed"
	"sync"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
)

//go:embed symbols/symbols.ttf
var symbolsFont []byte

// Faces is every face the window writes in, read once.
var Faces = sync.OnceValue(func() []font.FontFace {
	symbols, err := opentype.ParseCollection(symbolsFont)
	if err != nil {
		// Embedded and read by a test: a font that does not parse is a
		// build nobody should have shipped.
		panic(err)
	}
	return append(append([]font.FontFace{}, gofont.Collection()...), symbols...)
})
