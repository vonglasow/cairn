package gui

import (
	"gioui.org/app"

	cairnlogo "cairn/internal/logo"
)

// Known is the name the window is known to the system by: Wayland's app_id and X11's class, which a desktop entry of the same name gives the logo to.
const Known = "cairn"

var logo, _ = cairnlogo.At(1024)

// Logo is cairn's logo as the window hands it to a system that takes it from the program: a PNG, 1024 pixels square, its corners clear.
func Logo() []byte { return logo }

func knownAs() { app.ID = Known }
