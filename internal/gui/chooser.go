package gui

import (
	"path/filepath"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/theme"
)

// OpenThemes is the offer handing the folder a person's own schemes are kept in to the system.
const OpenThemes = "Open the themes folder"

type schemeOffered struct {
	seen    engine.ThemeSeen
	colours theme.Colours
	kind    string
	bad     string
}

// ChooserTile is a scheme where the chooser draws it, and what its tile says.
type ChooserTile struct {
	Slug string
	Tile SchemeTile
	At   Box
}

func labelColourNames() []string { return domain.LabelColours }

func themesFolder(c engine.Config) string { return filepath.Join(filepath.Dir(c.File), "themes") }

func offeredFrom(seen engine.ThemeSeen) schemeOffered {
	offered := schemeOffered{seen: seen, kind: "dark"}
	scheme, err := engine.SchemeOf(seen)
	switch {
	case seen.Problem != "":
		offered.bad = seen.Problem
	case err != nil:
		offered.bad = err.Error()
	default:
		offered.colours = scheme.Colours()
		if scheme.Light() {
			offered.kind = "light"
		}
	}
	if seen.From != theme.Carried && offered.bad == "" {
		offered.kind = "yours"
	}
	return offered
}

func (w Window) openingChooser() Window {
	w.tiles = nil
	for _, own := range []bool{false, true} {
		for _, seen := range w.configured.Themes {
			if (seen.From != theme.Carried) == own {
				w.tiles = append(w.tiles, offeredFrom(seen))
			}
		}
	}
	w.chooserOpen, w.tileOn, w.settingsAt = true, 0, 0
	worn := w.WornScheme()
	for i, one := range w.tiles {
		if one.seen.Slug == worn {
			w.tileOn = i
		}
	}
	w.previewing = worn
	return w
}

func (w Window) inChooser(name string) Window {
	cols := w.chooserColumns()
	step := map[string]int{"left": -1, "h": -1, "right": 1, "l": 1, "up": -cols, "k": -cols, "down": cols, "j": cols}[name]
	if step == 0 {
		return w
	}
	at := w.tileOn
	for next := at + step; next >= 0 && next < len(w.tiles); next += step {
		if w.tiles[next].bad == "" {
			at = next
			break
		}
	}
	return w.previewingTile(at)
}

func (w Window) previewingTile(at int) Window {
	w.tileOn, w.previewing = at, w.tiles[at].seen.Slug
	for _, tile := range w.ChooserTiles() {
		region := w.chooserRegion()
		if tile.Slug != w.previewing {
			continue
		}
		if tile.At.Y < region.Y {
			w.settingsAt -= region.Y - tile.At.Y + sectionHigh
		} else if tile.At.Y+tile.At.H > region.Y+region.H {
			w.settingsAt += tile.At.Y + tile.At.H - region.Y - region.H + padding
		}
		w.settingsAt = max(0, w.settingsAt)
	}
	return w
}

func (w Window) clickingChooser(p Point) Window {
	if !w.chooserRegion().In(p) {
		return w
	}
	for i, tile := range w.ChooserTiles() {
		if tile.At.In(p) && w.tiles[i].bad == "" {
			if w.tileOn == i {
				return w.wearing()
			}
			return w.previewingTile(i)
		}
	}
	return w
}

func (w Window) wearing() Window {
	slug := w.previewing
	w.chooserOpen, w.previewing, w.settingsAt = false, "", 0
	return w.setting("window.theme", slug)
}

func (w Window) leavingChooser() Window {
	w.chooserOpen, w.previewing, w.settingsAt = false, "", 0
	return w
}

func (w Window) chooserColumns() int {
	if w.narrow() {
		return 3
	}
	return 5
}

// ChooserHeads is the two heads of the chooser: the schemes cairn carries, and a person's own.
func (w Window) ChooserHeads() []SettingSection {
	region := w.chooserRegion()
	shipped, own := w.tilesCounted()
	heads := []SettingSection{{Name: "Shipped", Count: strconv.Itoa(shipped), At: Box{X: region.X, Y: region.Y - w.settingsAt + padding, W: region.W, H: lineHigh}}}
	rows := (shipped + w.chooserColumns() - 1) / w.chooserColumns()
	y := heads[0].At.Y + sectionHigh - padding + rows*(tileHigh+tileGap)
	return append(heads, SettingSection{Name: "Yours · " + tilde(themesFolder(w.configured)), Count: strconv.Itoa(own), At: Box{X: region.X, Y: y + padding, W: region.W, H: lineHigh}})
}

func (w Window) tilesCounted() (shipped, own int) {
	for _, one := range w.tiles {
		if one.seen.From == theme.Carried {
			shipped++
		} else {
			own++
		}
	}
	return shipped, own
}

const tileGap = 10

// ChooserTiles is every scheme the chooser offers, where its tile is drawn: the ones cairn carries under their head, then a person's own under theirs.
func (w Window) ChooserTiles() []ChooserTile {
	region := w.chooserRegion()
	cols := w.chooserColumns()
	wide := (region.W - (cols-1)*tileGap) / cols
	heads := w.ChooserHeads()
	var tiles []ChooserTile
	at := 0
	for i, one := range w.tiles {
		head := heads[0]
		if one.seen.From != theme.Carried {
			if at > 0 && w.tiles[i-1].seen.From == theme.Carried {
				at = 0
			}
			head = heads[1]
		}
		top := head.At.Y + head.At.H + between
		box := Box{X: region.X + (at%cols)*(wide+tileGap), Y: top + (at/cols)*(tileHigh+tileGap), W: wide, H: tileHigh}
		warn := ""
		if n := len(one.seen.Short); n > 0 {
			warn = "falls short ×" + strconv.Itoa(n)
		}
		kind := one.kind
		if one.bad != "" {
			kind = "not offered"
		}
		tiles = append(tiles, ChooserTile{Slug: one.seen.Slug, At: box, Tile: SchemeTile{
			Colours: one.colours, Name: one.seen.Slug, Kind: kind, Warn: warn, Bad: one.bad,
			Worn: one.seen.InForce, Keyed: i == w.tileOn,
		}})
		at++
	}
	return tiles
}

func (w Window) shortSaid() string {
	if w.tileOn < len(w.tiles) {
		if one := w.tiles[w.tileOn]; len(one.seen.Short) > 0 && !w.narrow() {
			return one.seen.Slug + ": " + one.seen.Short[0]
		}
	}
	return ""
}

// WornScheme is the scheme the window shows: the one previewed while the chooser is open, otherwise the one in force.
func (w Window) WornScheme() string {
	if w.chooserOpen && w.previewing != "" {
		return w.previewing
	}
	for _, seen := range w.configured.Themes {
		if seen.InForce {
			return seen.Slug
		}
	}
	return theme.DefaultSlug
}

// Skin is what the window wears now: the scheme previewed while the chooser is open, otherwise the one in force.
func (w Window) Skin() Skin {
	if w.chooserOpen && w.previewing != "" {
		for _, one := range w.tiles {
			if one.seen.Slug == w.previewing && one.bad == "" {
				return Skinned(one.colours)
			}
		}
	}
	if w.skin != nil {
		return *w.skin
	}
	return Dark()
}

// Wearing is the window wearing the skin given until its configuration says another scheme.
func (w Window) Wearing(s Skin) Window {
	w.skin = &s
	return w
}

func (w Window) wornFrom(got engine.Config) Window {
	for _, seen := range got.Themes {
		if !seen.InForce {
			continue
		}
		if offered := offeredFrom(seen); offered.bad == "" {
			s := Skinned(offered.colours)
			w.skin = &s
		}
	}
	return w
}

func (w Window) chooserRegion() Box {
	box := w.SettingsBox()
	body := SettingsBody(box)
	return Box{X: body.X + rowsPad, Y: body.Y + inset, W: body.W - 2*rowsPad, H: body.H - inset}
}
