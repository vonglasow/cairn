package gui

import (
	"image/color"
	"slices"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/config"
	"cairn/internal/theme"
)

func drawConfiguration(k Kit, w Window) {
	if !w.introsOpen {
		Veil{}.Draw(k, Box{W: w.width, H: w.height})
		k = k.inside("veil")
	}
	box := w.SettingsBox()
	Panel{Head: Heads(w), Moves: true}.Draw(k, box)
	in := k.inside("panel")
	if w.introsOpen {
		drawIntros(in, w, box)
	} else if w.chooserOpen {
		drawChooser(in, w, box)
	} else {
		drawSettings(in, w, box)
	}
	drawPanelOffers(in, box, w)
	if says := w.SettingsFootSays(); says != "" {
		offers := w.offersAlong(box)
		end := box.X + box.W - padding
		if len(offers) > 0 {
			end = offers[0].X - padding
		}
		row := offers
		y := PanelOfferAt(box, 0).Y
		if len(row) > 0 {
			y = row[0].Y
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: box.X + padding, Y: y + (barHigh-lineHigh)/2, W: max(0, end-box.X-padding), H: lineHigh}, says, small, font.Font{}, text.Start)
	}
}

func drawScope(k Kit, w Window) {
	segments := w.ScopeSegments()
	on := 0
	if w.global && len(segments) > 1 {
		on = 1
	}
	dot := ""
	if w.HereOffered() {
		dot = w.switcher.place.Colour
	}
	first, last := segments[0], segments[len(segments)-1]
	track := Box{X: first.X - 2, Y: first.Y - 2, W: last.X + last.W + 2 - first.X + 2, H: first.H + 4}
	Segmented{Said: w.scopeWords(), At: segments, On: on, Small: true, Dot: dot}.Draw(k, track)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, w.ScopePathAt(), w.SettingsPath(), small, k.skin.Type.Mono, text.Start)
}

func drawSettings(k Kit, w Window, box Box) {
	drawScope(k, w)
	body := SettingsBody(box)
	Rule{}.Draw(k, Box{X: box.X, Y: body.Y - 1, W: box.W})
	if nav := SettingsNav(box); nav.W > 0 {
		Rule{Upright: true}.Draw(k, Box{X: nav.X + nav.W, Y: nav.Y, H: nav.H})
		shown := w.sectionShown()
		for _, section := range w.SettingSections() {
			Row{Picked: section.Name == shown}.Draw(k, section.At)
			line := Inside(section.At)
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: line.X + between, Y: line.Y, W: line.W, H: line.H}, section.Name, reads)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: line.X, Y: line.Y, W: line.W - between, H: line.H}, section.Count, small, k.skin.Type.Mono, text.End)
		}
	}
	region := SettingsRows(box)
	rows, cut := k.within(region)
	for _, head := range w.SectionHeads() {
		Head{Said: head.Name, Count: head.Count}.Draw(rows, head.At)
	}
	for i, row := range w.SettingRows() {
		drawSettingRow(rows, w, row, i == w.rowOn)
	}
	cut.Pop()
	Scrollbar{Lines: w.settingsLength(), Rows: region.H, At: w.settingsAt}.Draw(k, Box{X: region.X + region.W + rowsPad - 2*sbarWide, Y: region.Y, W: sbarWide, H: region.H})
}

func (w Window) sectionShown() string {
	region := SettingsRows(w.SettingsBox())
	for _, row := range w.SettingRows() {
		if row.At.Y+row.At.H > region.Y {
			return row.Key.Section
		}
	}
	return ""
}

func drawSettingRow(k Kit, w Window, row SettingRow, keyed bool) {
	Row{Hover: w.hoverKey == row.Key.Name && !row.Off}.Draw(k, row.At)
	r := k.inside("row")
	if row.Off {
		faded := r.faded(0.45)
		defer faded.Pop()
	}
	name := Box{X: row.Name.X, Y: row.Name.Y + (row.Name.H-2*lineHigh)/2, W: row.Name.W, H: lineHigh}
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, w.rowSays(row), reads, font.Medium, 1, text.Start)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: name.X, Y: name.Y + lineHigh, W: name.W, H: lineHigh}, row.Key.Name, small, k.skin.Type.Mono, text.Start)
	if row.Said != "" {
		sayIn(k.gtx, k.th, k.skin.Colour.State.Wrong, Box{X: name.X, Y: row.From.Y + row.From.H - 2, W: row.At.W - 2*padding, H: lineHigh}, row.Said, small, font.Font{}, text.Start)
	}
	drawControl(r, w, row)
	drawOrigin(r, w, row)
	if keyed {
		Ring{Radius: k.skin.Radius.Row}.Draw(r, row.At)
	}
}

func (w Window) rowSays(row SettingRow) string {
	if row.Key.Name == "workspace.colour" {
		return row.Key.Says + " of " + w.hereName()
	}
	return row.Key.Says
}

func drawControl(k Kit, w Window, row SettingRow) {
	switch {
	case row.Key.Kind == config.KindBool:
		Toggle{On: row.Value == "true"}.Draw(k, row.Parts[0])
	case row.Key.Kind == config.KindWords:
		Segmented{Said: row.Words, At: row.Parts, On: slices.Index(row.Words, row.Value), Small: true}.Draw(k, row.Control)
	case row.Key.Kind == config.KindTheme:
		Picker{Said: row.Value, Strip: w.stripOf(row.Value)}.Draw(k, row.Control)
	case row.Key.Kind == config.KindIntro:
		Picker{Said: row.Value}.Draw(k, row.Control)
	case row.Key.Kind == config.KindColour:
		Picker{Said: row.Value, Dot: row.Value}.Draw(k, row.Control)
	case row.Fixed:
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, Box{X: row.Control.X + copyRoom, Y: row.Control.Y + (row.Control.H-lineHigh)/2, W: row.Control.W - copyRoom, H: lineHigh}, row.Value, small, k.skin.Type.Mono, text.Start)
	default:
		editing := w.keyWriting == row.Key.Name
		field := Field{Lit: editing}.Draw(k, row.Control)
		line := Box{X: field.X + pickerPad, Y: field.Y + (field.H-lineHigh)/2, W: field.W - 2*pickerPad, H: lineHigh}
		if row.Key.Unit != "" {
			unit := ends(w.measured(), row.Key.Unit, small)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: line.X + line.W - unit, Y: line.Y, W: unit, H: line.H}, row.Key.Unit, small, font.Font{}, text.End)
			line.W -= unit + between
		}
		switch {
		case editing:
			writing(k.inside("field"), w.measured(), line, w.keyWritten, true, small, 1, Follow)
		case row.Value == "" && row.Key.Kind == config.KindText:
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, line, "no label here", small, font.Font{}, text.Start)
		default:
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, row.Value, small, k.skin.Type.Mono, text.Start)
		}
	}
}

func (w Window) stripOf(slug string) []color.NRGBA {
	if slug == w.WornScheme() {
		return StripOf(w.Skin().Colour)
	}
	for _, seen := range w.configured.Themes {
		if seen.Slug == slug {
			if offered := offeredFrom(seen); offered.bad == "" {
				return StripOf(offered.colours)
			}
		}
	}
	return StripOf(theme.Default())
}

// OriginSays is what a row says of where its value comes from, part by part, in the order it is written.
func (row SettingRow) OriginSays() []string {
	var said []string
	for _, one := range []string{row.Comes, row.Aside, row.Theirs} {
		if one != "" {
			said = append(said, one)
		}
	}
	return said
}

func drawOrigin(k Kit, w Window, row SettingRow) {
	m := w.measured()
	said := row.OriginSays()
	wides := make([]int, len(said))
	total := 0
	for i, one := range said {
		wides[i] = ends(m, one, small)
		if one == SetHere {
			wides[i] += 2 * pillPad
		}
		total += wides[i] + gap
	}
	x := row.From.X
	if !w.narrow() {
		x = row.From.X + row.From.W - total
		if len(row.OfferAt) > 0 {
			x = row.OfferAt[0].X - total
		}
	}
	for i, one := range said {
		at := Box{X: x, Y: row.From.Y + (row.From.H-lineHigh)/2, W: wides[i], H: lineHigh}
		switch one {
		case SetHere:
			Pill{Said: one, Tone: ToneHere}.Draw(k, Box{X: x, Y: row.From.Y + (row.From.H-pillHigh)/2, W: wides[i], H: pillHigh})
		case ByDefault, WhenOpened:
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, at, one, small, font.Font{}, text.Start)
		default:
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, at, one, small, font.Font{}, text.Start)
		}
		x += wides[i] + gap
	}
	for i, offer := range row.Offers {
		look := buttonPlain
		if offer == SetHereOffer {
			look = buttonGhost
		}
		Button{Said: offer, Look: look, Small: true}.Draw(k, row.OfferAt[i])
	}
}

func drawChooser(k Kit, w Window, box Box) {
	scope := ScopeAt(box)
	said := "Previewing " + w.previewing + " · ⏎ wears it · esc goes back to " + w.inForce()
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: scope.X, Y: scope.Y + (scope.H-lineHigh)/2, W: scope.W, H: lineHigh}, said, small, font.Font{}, text.Start)
	body := SettingsBody(box)
	Rule{}.Draw(k, Box{X: box.X, Y: body.Y - 1, W: box.W})
	region := w.chooserRegion()
	tiles, cut := k.within(region)
	for _, head := range w.ChooserHeads() {
		Head{Said: head.Name, Count: head.Count}.Draw(tiles, head.At)
	}
	for _, tile := range w.ChooserTiles() {
		tile.Tile.Draw(tiles, tile.At)
	}
	cut.Pop()
	tilesAll := w.ChooserTiles()
	length := 0
	if len(tilesAll) > 0 {
		last := tilesAll[len(tilesAll)-1].At
		length = last.Y + last.H + w.settingsAt - region.Y + padding
	}
	Scrollbar{Lines: length, Rows: region.H, At: w.settingsAt}.Draw(k, Box{X: region.X + region.W + rowsPad - 2*sbarWide, Y: region.Y, W: sbarWide, H: region.H})
}

func drawIntros(k Kit, w Window, box Box) {
	scope := ScopeAt(box)
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, Box{X: scope.X, Y: scope.Y + (scope.H-lineHigh)/2, W: scope.W, H: lineHigh}, IntrosSays, small, font.Font{}, text.Start)
	body := SettingsBody(box)
	Rule{}.Draw(k, Box{X: box.X, Y: body.Y - 1, W: box.W})
	region := w.introsRegion()
	rows, cut := k.within(region)
	all := w.IntroRows()
	for i, row := range all {
		Row{Picked: i == w.introOn}.Draw(rows, row.At)
		line := Inside(row.At)
		name := Box{X: line.X + between, Y: row.At.Y + (row.At.H-2*lineHigh)/2, W: line.W - 2*between, H: lineHigh}
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, row.Says, reads, font.Medium, 1, text.Start)
		said := row.Name + " · " + row.Lasts
		if row.InForce {
			said += " · in use"
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, Box{X: name.X, Y: name.Y + lineHigh, W: name.W, H: lineHigh}, said, small, k.skin.Type.Mono, text.Start)
	}
	cut.Pop()
	length := 0
	if len(all) > 0 {
		last := all[len(all)-1].At
		length = last.Y + last.H + w.settingsAt - region.Y + padding
	}
	Scrollbar{Lines: length, Rows: region.H, At: w.settingsAt}.Draw(k, Box{X: region.X + region.W + padding/2, Y: region.Y, W: sbarWide, H: region.H})
}

// IntrosSays is what the intro chooser says over its list.
const IntrosSays = "Played on the board · r again · ⏎ keeps it"

func (w Window) inForce() string {
	for _, seen := range w.configured.Themes {
		if seen.InForce {
			return seen.Slug
		}
	}
	return theme.DefaultSlug
}

// BoardBanner is what the board says over itself of its configuration: a file read again and what changed, or a file that cannot be read; what it offers, and where it is drawn.
type BoardBanner struct {
	Said, Then, Offer string
	Told              bool
	At, OfferAt       Box
}

// BoardBanners is every banner the board shows of its configuration, the one about a file that cannot be read first, stacked above the bar.
func (w Window) BoardBanners() []BoardBanner {
	if w.screen == OnConfiguration {
		return nil
	}
	var banners []BoardBanner
	if t, ok := w.Trouble(); ok {
		banners = append(banners, BoardBanner{
			Said: tilde(t.File) + " cannot be read", Then: t.String() + " · the window keeps what it read at " + w.readAt.Format("15:04") + " until the file reads again.",
			Offer: "Open " + baseName(t.File),
		})
	}
	if in, changes := w.ReadAgain(); len(changes) > 0 {
		said := ""
		for i, change := range changes {
			if i > 0 {
				said += " · "
			}
			said += change.String()
		}
		banners = append(banners, BoardBanner{Said: in + " read again", Then: said, Offer: "Settings", Told: true})
	}
	l := w.Layout()
	bottom := l.Bar.Y - gap
	m := w.measured()
	for i := len(banners) - 1; i >= 0; i-- {
		b := &banners[i]
		offer := buttonWordWide(m, b.Offer) + 2*padding
		wide := min(bannerWidest, w.width-2*settingsMargin*2, max(ends(m, b.Said, reads), ends(m, b.Then, small))+offer+3*bannerPad)
		inner := wide - offer - 3*bannerPad
		high := BannerHigh(m, b.Said, b.Then, inner+2*bannerPad)
		b.At = Box{X: (w.width - wide) / 2, Y: bottom - high, W: wide, H: high}
		b.OfferAt = Box{X: b.At.X + b.At.W - bannerPad - offer, Y: b.At.Y + bannerPad, W: offer, H: smallButton}
		bottom = b.At.Y - gap
	}
	return banners
}

const bannerWidest = 760

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

func drawBoardBanners(k Kit, w Window) {
	for _, b := range w.BoardBanners() {
		Banner{Said: b.Said, Then: b.Then, Told: b.Told, Room: b.OfferAt.W + bannerPad}.Draw(k, b.At)
		Button{Said: b.Offer, Small: true}.Draw(k.inside("banner"), b.OfferAt)
	}
}

func (w Window) clickingBanners(p Point) (Window, bool) {
	for _, b := range w.BoardBanners() {
		if !b.OfferAt.In(p) {
			continue
		}
		if b.Told {
			w.changed = nil
			return w.openingConfiguration(), true
		}
		w.opens = w.trouble.File
		return w, true
	}
	return w, false
}
