package gui

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
		panic(err)
	}
	return append(append([]font.FontFace{}, gofont.Collection()...), symbols...)
})
