package gui

import (
	"embed"
	"sync"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
)

//go:embed symbols/symbols.ttf
var symbolsFont []byte

//go:embed fonts/*.ttf
var textFonts embed.FS

// TextFace is the typeface the window writes its words in.
const TextFace = "Figtree"

// Faces is every face the window writes in, read once: Figtree first, then Go's own for figures and what Figtree lacks, then the symbols.
var Faces = sync.OnceValue(func() []font.FontFace {
	var faces []font.FontFace
	for _, weight := range []string{"Regular", "Medium", "SemiBold"} {
		raw, err := textFonts.ReadFile("fonts/Figtree-" + weight + ".ttf")
		if err != nil {
			panic(err)
		}
		face, err := opentype.ParseCollection(raw)
		if err != nil {
			panic(err)
		}
		faces = append(faces, face...)
	}
	symbols, err := opentype.ParseCollection(symbolsFont)
	if err != nil {
		panic(err)
	}
	return append(append(faces, gofont.Collection()...), symbols...)
})
