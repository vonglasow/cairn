package tui

func (f Form) typedInto(name, value string) Form {
	if held, there := f.held[name]; there {
		*held = value
	}
	return f
}

// Showing is how many whole cards a run of rows has room for.
func Showing(rows int) int { return rows / CardHeight }

// PlainSkin is a skin with no colour in it.
func PlainSkin() Skin { return Skin{palette: Default} }
