package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"gioui.org/font"
	"gioui.org/text"

	"cairn/internal/store/files"
	"cairn/internal/theme"
)

func drawConfiguration(k Kit, w Window) {
	if !w.IntrosOpen {
		Veil{}.Draw(k, geom.Box{W: w.width, H: w.height})
		k = k.inside("veil")
	}
	box := w.SettingsBox()
	plain := !w.IntrosOpen && !w.ChooserOpen
	head := Heads(w)
	if plain {
		head = ""
	}
	Panel{Head: head, Moves: true}.Draw(k, box)
	in := k.inside("panel")
	if plain {
		at := FormHeading(box)
		Heading{Said: Heads(w)}.Draw(in, geom.Box{X: at.X, Y: at.Y, W: ends(w.headMeasured(), Heads(w), design.HeadingSize), H: at.H})
	}
	if w.IntrosOpen {
		drawIntros(in, w, box)
	} else if w.ChooserOpen {
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
		Sentence{Said: says}.Draw(in, geom.Box{X: box.X + padding, Y: y + (geom.BarHigh-lineHigh)/2, W: max(0, end-box.X-padding), H: lineHigh})
	}
}

func drawScope(k Kit, w Window) {
	segments := w.ScopeSegments()
	on := 0
	if w.Global && len(segments) > 1 {
		on = 1
	}
	dot := ""
	if w.HereOffered() {
		dot = w.switcher.place.Colour
	}
	first, last := segments[0], segments[len(segments)-1]
	track := geom.Box{X: first.X - 2, Y: first.Y - 2, W: last.X + last.W + 2 - first.X + 2, H: first.H + 4}
	Segmented{Said: w.scopeWords(), At: segments, On: on, Small: true, Dot: dot}.Draw(k, track)
}

func drawSettings(k Kit, w Window, box geom.Box) {
	drawScope(k, w)
	body := SettingsBody(box)
	Rule{}.Draw(k, geom.Box{X: box.X, Y: body.Y - 1, W: box.W})
	if nav := SettingsNav(box); nav.W > 0 {
		Rule{Upright: true}.Draw(k, geom.Box{X: nav.X + nav.W, Y: nav.Y, H: nav.H})
		shown := w.sectionShown()
		for _, section := range w.SettingSections() {
			Row{Picked: section.Name == shown}.Draw(k, section.At)
			line := Inside(section.At)
			say(k.gtx, k.th, k.skin.Colour.Ink.Text, geom.Box{X: line.X + between, Y: line.Y, W: line.W, H: line.H}, section.Name, reads)
			sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, geom.Box{X: line.X, Y: line.Y, W: line.W - between, H: line.H}, section.Count, small, k.skin.Type.Mono, text.End)
		}
	}
	region := SettingsRows(box)
	rows, cut := k.within(region)
	for _, head := range w.SectionHeads() {
		Head{Said: strings.ToUpper(head.Name)}.Draw(rows, head.At)
	}
	all := w.SettingRows()
	for i, row := range all {
		drawSettingRow(rows, w, row, i == w.RowOn, i+1 < len(all) && all[i+1].Key.Section == row.Key.Section)
	}
	cut.Pop()
	Scrollbar{Lines: w.settingsLength(), Rows: region.H, At: w.SettingsAt}.Draw(k, geom.Box{X: region.X + region.W + rowsPad - 2*sbarWide, Y: region.Y, W: sbarWide, H: region.H})
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

func drawSettingRow(k Kit, w Window, row SettingRow, keyed, ruled bool) {
	Row{Hover: w.hoverKey == row.Key.Name && !row.Off}.Draw(k, row.At)
	r := k.inside("row")
	if row.Off {
		faded := r.faded(0.45)
		defer faded.Pop()
	}
	name := geom.Box{X: row.Name.X, Y: row.Name.Y + (row.Name.H-2*lineHigh)/2, W: row.Name.W, H: lineHigh}
	said, under := w.RowLines(row)
	sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, said, reads, font.Medium, 1, text.Start)
	look := font.Font{}
	if under == row.Key.Name {
		look = k.skin.Type.Mono
	}
	sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, geom.Box{X: name.X, Y: name.Y + lineHigh, W: name.W, H: lineHigh}, under, small, look, text.Start)
	if row.Said != "" {
		sayIn(k.gtx, k.th, k.skin.Colour.State.Wrong, row.SaidAt(), row.Said, small, font.Font{}, text.Start)
	}
	drawControl(r, w, row)
	drawOrigin(r, w, row)
	if left := row.Left; left != nil {
		Banner{Said: left.Said, Then: left.Then, Warn: true, Room: left.At.X + left.At.W - left.Leave.X}.Draw(r, left.At)
		Button{Said: NotesLeave, Small: true}.Draw(r.inside("banner"), left.Leave)
		Button{Said: NotesMoveThem, Small: true}.Draw(r.inside("banner"), left.Move)
	}
	if ruled {
		Rule{}.Draw(r, geom.Box{X: row.Name.X, Y: row.At.Y + row.At.H - 1, W: row.From.X + row.From.W - row.Name.X})
	}
	if keyed {
		Ring{Radius: k.skin.Radius.Row}.Draw(r, row.At)
	}
}

// RowLines is what a row writes at its left: its name, and under it what it does, or its key while the pointer rests on it.
func (w Window) RowLines(row SettingRow) (string, string) {
	if w.hoverKey == row.Key.Name {
		return row.Key.Says, row.Key.Name
	}
	return row.Key.Says, row.Key.Help
}

// Shown is the words a row of words shows, in the order it takes them.
func (row SettingRow) Shown() []string {
	if len(row.Key.Shown) > 0 {
		return row.Key.Shown
	}
	return row.Words
}

func drawControl(k Kit, w Window, row SettingRow) {
	switch {
	case row.Key.Kind == files.KindBool:
		Toggle{On: row.Value == "true"}.Draw(k, row.Parts[0])
	case row.Key.Switch:
		Toggle{On: row.Value == row.Words[0]}.Draw(k, row.Parts[0])
	case row.Key.Kind == files.KindWords:
		Segmented{Said: row.Shown(), At: row.Parts, On: slices.Index(row.Words, row.Value), Small: true}.Draw(k, row.Control)
	case row.Key.Kind == files.KindTheme:
		Picker{Said: row.Value, Strip: w.stripOf(row.Value)}.Draw(k, row.Control)
	case row.Key.Kind == files.KindIntro:
		Picker{Said: row.Value}.Draw(k, row.Control)
	case row.Key.Kind == files.KindColour:
		picked := row.Value
		if w.colourKey == row.Key.Name {
			picked = label.LabelColours[w.ColourPick]
		}
		for i, colour := range label.LabelColours {
			Swatch{Colour: colour, Picked: colour == picked}.Draw(k, row.Parts[i])
		}
	case row.Key.Kind == files.KindPercent:
		percent := w.PercentShown(row)
		Slider{Of: float32(percent-files.LeastGhost) / float32(files.MostGhost-files.LeastGhost)}.Draw(k, row.Parts[0])
		label := geom.Box{X: row.Parts[0].X + row.Parts[0].W + between, Y: row.Control.Y + (row.Control.H-lineHigh)/2, W: sliderLabel, H: lineHigh}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, label, strconv.Itoa(percent)+" %", small, k.skin.Type.Mono, text.Start)
	case row.Key.Kind == files.KindNumber:
		Button{Said: "−", Small: true}.Draw(k, row.Parts[0])
		drawField(k, w, row, row.Parts[1])
		Button{Said: "+", Small: true}.Draw(k, row.Parts[2])
		unit := geom.Box{X: row.Parts[2].X + row.Parts[2].W + between, Y: row.Control.Y + (row.Control.H-lineHigh)/2, W: ends(w.measured(), row.Key.Unit, small), H: lineHigh}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, unit, row.Key.Unit, small, font.Font{}, text.Start)
	case row.Key.Kind == files.KindPath && !row.Fixed:
		drawField(k, w, row, row.Parts[0])
		Button{Said: NewSpaceChoose, Small: true}.Draw(k, row.Parts[1])
	case row.Fixed:
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, geom.Box{X: row.Control.X + copyRoom, Y: row.Control.Y + (row.Control.H-lineHigh)/2, W: row.Control.W - copyRoom, H: lineHigh}, row.Value, small, k.skin.Type.Mono, text.Start)
	default:
		drawField(k, w, row, row.Control)
	}
}

func drawField(k Kit, w Window, row SettingRow, at geom.Box) {
	editing := w.KeyWriting == row.Key.Name
	field := Field{Lit: editing || row.Said != ""}.Draw(k, at)
	line := geom.Box{X: field.X + pickerPad, Y: field.Y + (field.H-lineHigh)/2, W: field.W - 2*pickerPad, H: lineHigh}
	mark := ""
	switch {
	case row.Key.Kind == files.KindDay:
		mark = DayMark
	case row.Key.Kind != files.KindNumber:
		mark = row.Key.Unit
	}
	if mark != "" {
		wide := ends(w.measured(), mark, small) + between
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Faint, geom.Box{X: line.X + line.W - wide, Y: line.Y, W: wide, H: line.H}, mark, small, font.Font{}, text.End)
		line.W -= wide + between
	}
	switch {
	case editing:
		writing(k.inside("field"), w.measured(), line, w.KeyWritten, true, small, 1, Follow)
	case row.Value == "" && row.Key.Name == "approach.label":
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, line, "no label here", small, font.Font{}, text.Start)
	default:
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Text, line, row.Value, small, k.skin.Type.Mono, text.Start)
	}
}

// DayMark is what a field holding a day shows at its right: the calendar it opens.
const DayMark = "▦"

// SaidAt is where a row says why its value was refused: under its control.
func (row SettingRow) SaidAt() geom.Box {
	return geom.Box{X: row.Control.X, Y: row.Control.Y + row.Control.H + 2, W: row.From.X + row.From.W - row.Control.X, H: lineHigh}
}

func (w Window) stripOf(slug string) []color.NRGBA {
	if slug == w.WornScheme() {
		return StripOf(w.Skin().Colour)
	}
	for _, seen := range w.Configured.Themes {
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
	said := strings.Join(row.OriginSays(), " · ")
	at := geom.Box{X: row.From.X, Y: row.From.Y + row.From.H/2 - lineHigh, W: row.From.W, H: lineHigh}
	if len(row.Offers) == 0 {
		at.Y = row.From.Y + (row.From.H-lineHigh)/2
	}
	if w.narrow() {
		at = geom.Box{X: row.From.X, Y: row.From.Y + (row.From.H-lineHigh)/2, W: ends(w.measured(), said, small), H: lineHigh}
	}
	ink := k.skin.Colour.Ink.Faint
	if row.Comes == SetHere || row.Comes == InFile {
		ink = k.skin.Colour.Standing["active"].Ink
	}
	sayIn(k.gtx, k.th, ink, at, said, small, font.Font{}, text.Start)
	for i, offer := range row.Offers {
		sayIn(k.gtx, k.th, k.skin.Colour.Accent.Base, row.OfferAt[i], offer, small, font.Font{}, text.Start)
	}
}

func drawChooser(k Kit, w Window, box geom.Box) {
	scope := ScopeAt(box)
	said := "Previewing " + w.Previewing + " · `⏎` wears it · `esc` goes back to " + w.InForce()
	Sentence{Said: said}.Draw(k, geom.Box{X: scope.X, Y: scope.Y + (scope.H-lineHigh)/2, W: scope.W, H: lineHigh})
	body := chooserBody(box)
	Rule{}.Draw(k, geom.Box{X: box.X, Y: body.Y - 1, W: box.W})
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
		length = last.Y + last.H + w.SettingsAt - region.Y + padding
	}
	Scrollbar{Lines: length, Rows: region.H, At: w.SettingsAt}.Draw(k, geom.Box{X: region.X + region.W + rowsPad - 2*sbarWide, Y: region.Y, W: sbarWide, H: region.H})
}

func drawIntros(k Kit, w Window, box geom.Box) {
	scope := ScopeAt(box)
	Sentence{Said: IntrosSays}.Draw(k, geom.Box{X: scope.X, Y: scope.Y + (scope.H-lineHigh)/2, W: scope.W, H: lineHigh})
	body := chooserBody(box)
	Rule{}.Draw(k, geom.Box{X: box.X, Y: body.Y - 1, W: box.W})
	region := w.introsRegion()
	rows, cut := k.within(region)
	all := w.IntroRows()
	for i, row := range all {
		Row{Picked: i == w.IntroOn}.Draw(rows, row.At)
		line := Inside(row.At)
		name := geom.Box{X: line.X + between, Y: row.At.Y + (row.At.H-2*lineHigh)/2, W: line.W - 2*between, H: lineHigh}
		sayStyled(k.gtx, k.th, k.skin.Colour.Ink.Text, name, row.Says, reads, font.Medium, 1, text.Start)
		said := row.Name + " · " + row.Lasts
		if row.InForce {
			said += " · in use"
		}
		sayIn(k.gtx, k.th, k.skin.Colour.Ink.Low, geom.Box{X: name.X, Y: name.Y + lineHigh, W: name.W, H: lineHigh}, said, small, k.skin.Type.Mono, text.Start)
	}
	cut.Pop()
	length := 0
	if len(all) > 0 {
		last := all[len(all)-1].At
		length = last.Y + last.H + w.SettingsAt - region.Y + padding
	}
	Scrollbar{Lines: length, Rows: region.H, At: w.SettingsAt}.Draw(k, geom.Box{X: region.X + region.W + padding/2, Y: region.Y, W: sbarWide, H: region.H})
}

// IntrosSays is what the intro chooser says over its list.
const IntrosSays = "Played on the board · `r` again · `⏎` keeps it"

// BoardBanner is what the board says over itself of its configuration: a file read again and what changed, or a file that cannot be read; what it offers, and where it is drawn.
type BoardBanner struct {
	Said, Then, Offer string
	Told              bool
	At, OfferAt       geom.Box
}

// BoardBanners is every banner the board shows of its configuration, the one about a file that cannot be read first, stacked above the bar.
func (w Window) BoardBanners() []BoardBanner {
	if w.Screen == board.OnConfiguration {
		return nil
	}
	var banners []BoardBanner
	if t, ok := w.Trouble(); ok {
		banners = append(banners, BoardBanner{
			Said: tilde(t.File) + " cannot be read", Then: t.String() + " · the window keeps what it read at " + w.ReadAt.Format("15:04") + " until the file reads again.",
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
	bottom := l.Bar.Y - geom.Gap
	m := w.measured()
	for i := len(banners) - 1; i >= 0; i-- {
		b := &banners[i]
		offer := buttonWordWide(m, b.Offer) + 2*padding
		wide := min(bannerWidest, w.width-2*settingsMargin*2, max(ends(m, b.Said, reads), ends(m, b.Then, small))+offer+3*bannerPad)
		inner := wide - offer - 3*bannerPad
		high := BannerHigh(m, b.Said, b.Then, inner+2*bannerPad)
		b.At = geom.Box{X: (w.width - wide) / 2, Y: bottom - high, W: wide, H: high}
		b.OfferAt = geom.Box{X: b.At.X + b.At.W - bannerPad - offer, Y: b.At.Y + bannerPad, W: offer, H: smallButton}
		bottom = b.At.Y - geom.Gap
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

func (w Window) clickingBanners(p geom.Point) (Window, bool) {
	for _, b := range w.BoardBanners() {
		if !b.OfferAt.In(p) {
			continue
		}
		if b.Told {
			w.Changed = nil
			return w.openingConfiguration(), true
		}
		w.Opens = w.Troubled.File
		return w, true
	}
	return w, false
}
