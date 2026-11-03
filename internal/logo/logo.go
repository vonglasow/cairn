// Package logo carries cairn's logo at every size a system asks for, drawn from its one SVG by scripts/logo.sh.
package logo

import (
	"embed"
	"strconv"
)

//go:embed cairn-*.png
var drawn embed.FS

// At is the logo as a PNG, size pixels square, its corners clear; an error for a size it was not drawn at.
func At(size int) ([]byte, error) { return drawn.ReadFile("cairn-" + strconv.Itoa(size) + ".png") }
