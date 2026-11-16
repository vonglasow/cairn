package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
	"cairn/internal/store/files"
)

// OpenTheFile is the offer handing the file the scope shown writes to the system, to open in its editor.
const OpenTheFile = "Open the file"

// WriteWhole is the offer writing the whole global configuration, every key explained, where none is written yet.
const WriteWhole = "Write the whole file"

// SettingsFoot is what the foot of the settings says: nothing waits for a button.
const SettingsFoot = "Saved as you change it"

// WritingFoot is what the foot says while a field is written: how it is saved, and how it is put back.
const WritingFoot = "Enter or a click elsewhere saves the field · Esc puts it back"

// The words a row says of where its value comes from.
const (
	SetHere    = "Only here"
	FromGlobal = "Shared"
	ByDefault  = "Default"
	InFile     = "Changed"
	WhenOpened = "Read when cairn opens"
)

// The offers a row makes beside where its value comes from.
const (
	ResetToGlobal  = "Use the shared value"
	ResetShort     = "Reset"
	ResetToDefault = "Back to default"
	SetHereOffer   = "Set only here"
)

// The scopes the settings write, as the panel names them.
const (
	ThisWorkspace = "This workspace · "
	AllWorkspaces = "All workspaces"
)

func (w Window) readingConfig() (workspace.Config, bool) {
	got, err := w.work.Ask(engine.ConfigShow, engine.Input{})
	if err != nil || got.Config == nil {
		return workspace.Config{}, false
	}
	return *got.Config, true
}

func (w Window) openingConfiguration() Window {
	got, ok := w.readingConfig()
	if !ok {
		return w.failing(errNoConfiguration)
	}
	if got.Problem == "" || w.Configured.File == "" {
		w.Configured = got
	}
	w = w.troubledBy(got)
	_, missing := os.Stat(w.Configured.File)
	w.ConfigHere = missing == nil
	w.Global = !w.HereOffered()
	w.Screen, w.RowOn, w.SettingsAt, w.ChooserOpen, w.KeyWriting, w.RefusedKey = board.OnConfiguration, -1, 0, false, "", ""
	w.IntrosOpen = false
	return w
}

type noConfiguration struct{}

func (noConfiguration) Error() string { return "the configuration cannot be read" }

var errNoConfiguration error = noConfiguration{}

func (w Window) writingConfiguration() Window {
	if _, err := w.work.Ask(engine.ConfigInit, engine.Input{}); err != nil {
		return w.failing(err)
	}
	return w.applyingOwn().openingConfiguration()
}

// HereOffered says whether the settings of the store shown are a scope of their own: the store is a workspace's, or settings are already beside it.
func (w Window) HereOffered() bool {
	return w.switcher.place.Configured() != "" || w.Configured.Settings != ""
}

// SettingsScope is the scope the settings write, as the panel names it: this workspace, or all of them.
func (w Window) SettingsScope() string {
	if w.Global {
		return AllWorkspaces
	}
	return ThisWorkspace + w.hereName()
}

func (w Window) hereName() string {
	if name := w.switcher.place.Name(); name != "" {
		return name
	}
	return "here"
}

// SettingsPath is what the panel says beside the scope: the name of the file written; the whole path is copied from the mark before it.
func (w Window) SettingsPath() string {
	return filepath.Base(w.SettingsWritten())
}

func tilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return "~/" + rest
	}
	return path
}

// SettingRow is one key of the settings as the panel lays it out: its key, its value in force in the scope shown, where that value comes from, what the other scope says, what it offers, and the boxes it is drawn in.
type SettingRow struct {
	Key                         files.Key
	Value, Comes, Aside, Theirs string
	Offers                      []string
	Off, Fixed                  bool
	Said                        string
	At, Name, Control, From     geom.Box
	OfferAt                     []geom.Box
	Parts                       []geom.Box
	Words                       []string
}

// SettingSection is a section of the settings: its name, how many rows it holds in the scope shown, and the row it starts at.
type SettingSection struct {
	Name, Count string
	First       int
	At          geom.Box
}

func (w Window) rowOfKey(key files.Key) SettingRow {
	c := w.Configured
	row := SettingRow{Key: key, Value: w.valueShown(key.Name), Words: key.Words}
	here, global := slices.Contains(c.WrittenHere, key.Name), slices.Contains(c.Written, key.Name)
	switch {
	case key.Name == "store" || key.Name == "notes":
		row.Comes, row.Fixed = WhenOpened, true
	case w.Global && global:
		row.Comes, row.Offers = InFile, []string{ResetToDefault}
	case w.Global:
		row.Comes = ByDefault
	case here:
		row.Comes, row.Offers = SetHere, []string{ResetToGlobal}
		if value, ok := c.Global[key.Name]; ok && key.Name != "workspace.colour" {
			row.Aside = "shared: " + value
		}
	case global:
		row.Comes = FromGlobal
	default:
		row.Comes = ByDefault
	}
	if w.Global && here && w.HereOffered() {
		row.Theirs = w.hereName() + " sets its own: " + workspace.ValueIn(c, key.Name)
	}
	if !row.Fixed && row.Comes != SetHere && row.Comes != InFile && w.hoverKey == key.Name && !w.Global {
		row.Offers = []string{SetHereOffer}
	}
	if w.narrow() && len(row.Offers) > 0 && row.Offers[0] == ResetToGlobal {
		row.Offers = []string{ResetShort}
	}
	if w.Troubled != nil && w.Troubled.File == w.SettingsWritten() {
		row.Off, row.Offers = true, nil
	}
	if w.RefusedKey == key.Name {
		row.Said = w.RefusedSay
	}
	return row
}

func (w Window) valueShown(key string) string {
	c := w.Configured
	switch key {
	case "store":
		return tilde(c.Store.At)
	case "notes":
		return tilde(c.Notes.At)
	}
	value := workspace.ValueIn(c, key)
	if global, ok := c.Global[key]; ok && w.Global {
		value = global
	}
	if value == workspace.NotSaid {
		return ""
	}
	return value
}

// SettingSections is every section of the settings in the scope shown, with how many rows each holds.
func (w Window) SettingSections() []SettingSection {
	shown := w.keysShown()
	var sections []SettingSection
	for _, name := range files.Sections() {
		first, count := -1, 0
		for i, key := range shown {
			if key.Section == name {
				count++
				if first < 0 {
					first = i
				}
			}
		}
		if count > 0 {
			sections = append(sections, SettingSection{Name: name, Count: itoa(count), First: first})
		}
	}
	if !w.Global {
		sections = append(sections, SettingSection{Name: files.SectionPlaces, Count: itoa(len(machineKeys())), First: -1})
	}
	nav := SettingsNav(w.SettingsBox())
	for i := range sections {
		sections[i].At = geom.Box{X: nav.X + inset*2, Y: nav.Y + padding + i*navRow, W: nav.W - inset*4, H: navRow}
	}
	return sections
}

const (
	settingsWide   = 960
	narrowBelow    = 860
	settingsHigh   = 700
	settingsMargin = 16
	navWide        = 176
	navRow         = 30
	fromWide       = 190
	controlColumn  = 260
	rowHigh        = 52
	stackedHigh    = 92
	sectionHigh    = 38
	controlHigh    = 28
	scopeHigh      = 28
	fieldLeast     = 90
	rowsPad        = 18
	sbarWide       = 4
	swatchRow      = 14
	stepWide       = 28
	stepField      = 56
)

// SettingsBox is the panel of the settings in a window of that size: as wide and as tall as it is drawn, or the room the window has less a margin.
func SettingsBox(width, height int) geom.Box {
	wide, high := min(settingsWide, width-2*settingsMargin), min(settingsHigh, height-2*settingsMargin)
	return geom.Box{X: (width - wide) / 2, Y: (height - high) / 2, W: wide, H: high}
}

func (w Window) settingsOpens() geom.Box {
	if w.IntrosOpen {
		return IntrosBox(w.width, w.height)
	}
	if w.ChooserOpen {
		return ChooserBox(w.width, w.height)
	}
	return SettingsBox(w.width, w.height)
}

// ChooserBox is the panel the theme chooser is drawn in, a little wider than the settings.
func ChooserBox(width, height int) geom.Box {
	wide, high := min(900, width-2*settingsMargin), min(settingsHigh, height-2*settingsMargin)
	return geom.Box{X: (width - wide) / 2, Y: (height - high) / 2, W: wide, H: high}
}

func (w Window) narrow() bool { return SettingsBox(w.width, w.height).W < narrowBelow }

// ScopeAt is the line under the heading of a chooser: what it previews, and how to keep it.
func ScopeAt(box geom.Box) geom.Box {
	head := HeadingAt(box)
	return geom.Box{X: box.X + padding, Y: head.Y + head.H + padding, W: box.W - 2*padding, H: scopeHigh}
}

// SettingsBody is what lies between the heading and the foot: the sections and the rows.
func SettingsBody(box geom.Box) geom.Box {
	head := HeadingAt(box)
	top := head.Y + head.H + padding
	foot := PanelOfferAt(box, 0).Y - between
	return geom.Box{X: box.X, Y: top, W: box.W, H: foot - top}
}

func chooserBody(box geom.Box) geom.Box {
	scope := ScopeAt(box)
	top := scope.Y + scope.H + padding
	foot := PanelOfferAt(box, 0).Y - between
	return geom.Box{X: box.X, Y: top, W: box.W, H: foot - top}
}

// SettingsNav is the column the sections are listed down at the left of the body, none at the narrowest.
func SettingsNav(box geom.Box) geom.Box {
	if box.W < narrowBelow {
		return geom.Box{}
	}
	body := SettingsBody(box)
	return geom.Box{X: body.X, Y: body.Y, W: navWide, H: body.H}
}

// SettingsRows is where the rows scroll: the body right of the sections.
func SettingsRows(box geom.Box) geom.Box {
	body, nav := SettingsBody(box), SettingsNav(box)
	x := body.X + nav.W + rowsPad
	return geom.Box{X: x, Y: body.Y + inset, W: body.X + body.W - rowsPad - x, H: body.H - inset}
}

// ScopeSegments is where the two scopes are laid, on the heading's line after it: this workspace, then all of them; one only when the store has no settings of its own.
func (w Window) ScopeSegments() []geom.Box {
	head := HeadingAt(w.SettingsBox())
	m := w.measured()
	var boxes []geom.Box
	x := head.X + ends(w.headMeasured(), Heads(w), design.HeadingSize) + padding + 2
	y := head.Y + (head.H-scopeHigh)/2
	for i, said := range w.scopeWords() {
		wide := ends(m, said, small) + 2*pickerPad
		if i == 0 && w.HereOffered() {
			wide += dotSide + between
		}
		boxes = append(boxes, geom.Box{X: x, Y: y + 2, W: wide, H: scopeHigh - 4})
		x += wide + 2
	}
	return boxes
}

func (w Window) scopeWords() []string {
	if !w.HereOffered() {
		return []string{AllWorkspaces}
	}
	return []string{ThisWorkspace + w.hereName(), AllWorkspaces}
}

// SettingRows is every row of the settings in the scope shown, laid out down the rows' column, scrolled.
func (w Window) SettingRows() []SettingRow {
	rows := SettingsRows(w.SettingsBox())
	keys := w.keysShown()
	out := make([]SettingRow, 0, len(keys))
	y, section := rows.Y-w.SettingsAt, ""
	for _, key := range keys {
		if key.Section != section {
			section = key.Section
			y += sectionHigh
		}
		row := w.rowOfKey(key)
		high := rowHigh
		if w.narrow() {
			high = stackedHigh
		}
		if row.Said != "" {
			high += lineHigh
		}
		row = w.laidRow(row, geom.Box{X: rows.X - padding + 2, Y: y, W: rows.W + 2*(padding-2), H: high})
		out = append(out, row)
		y += high
	}
	return out
}

// SectionHeads is where each section's head is written down the rows, scrolled with them.
func (w Window) SectionHeads() []SettingSection {
	rows := SettingsRows(w.SettingsBox())
	var heads []SettingSection
	section := ""
	for i, row := range w.SettingRows() {
		if row.Key.Section != section {
			section = row.Key.Section
			count := 0
			for _, other := range w.keysShown() {
				if other.Section == section {
					count++
				}
			}
			heads = append(heads, SettingSection{Name: section, Count: itoa(count), First: i, At: geom.Box{X: rows.X, Y: row.At.Y - sectionHigh + padding, W: rows.W, H: lineHigh}})
		}
	}
	return heads
}

func (w Window) settingsLength() int {
	rows := w.SettingRows()
	if len(rows) == 0 {
		return 0
	}
	last := rows[len(rows)-1]
	return last.At.Y + last.At.H + w.SettingsAt - SettingsRows(w.SettingsBox()).Y + padding
}

func (w Window) laidRow(row SettingRow, at geom.Box) SettingRow {
	row.At = at
	inner := geom.Box{X: at.X + padding - 2, Y: at.Y, W: at.W - 2*(padding-2), H: at.H}
	line := geom.Box{X: inner.X, Y: inner.Y + between, W: inner.W, H: rowHigh - 2*between}
	if w.narrow() {
		row.From = geom.Box{X: inner.X, Y: line.Y + line.H, W: inner.W, H: stackedHigh - rowHigh}
	} else {
		row.From = geom.Box{X: inner.X + inner.W - fromWide, Y: line.Y, W: fromWide, H: line.H}
	}
	right := inner.X + inner.W
	if !w.narrow() {
		right = row.From.X - padding
	}
	column := geom.Box{X: right - controlColumn, Y: line.Y + (line.H-controlHigh)/2, W: controlColumn, H: controlHigh}
	row.Control = geom.Box{X: column.X, Y: column.Y, W: w.controlWide(row), H: controlHigh}
	row.Name = geom.Box{X: inner.X, Y: line.Y, W: max(0, column.X-padding-inner.X), H: line.H}
	row.Parts = w.controlParts(row)
	row.OfferAt = w.offersOfRow(row)
	return row
}

func (w Window) controlWide(row SettingRow) int {
	m := w.measured()
	switch {
	case row.Key.Kind == files.KindBool || row.Key.Switch:
		return toggleWide
	case row.Key.Kind == files.KindWords:
		wide := 4
		for _, word := range row.Shown() {
			wide += ends(m, word, small) + 2*pickerPad
		}
		return wide
	case row.Key.Kind == files.KindTheme:
		return min(controlColumn, PickerWide(m, row.Value, false, true))
	case row.Key.Kind == files.KindIntro:
		return min(controlColumn, PickerWide(m, row.Value, false, false))
	case row.Key.Kind == files.KindColour:
		return len(label.LabelColours)*(swatchRow+inset) - inset
	case row.Key.Kind == files.KindNumber:
		return 2*stepWide + stepField + 2*inset + between + ends(m, row.Key.Unit, small)
	case row.Key.Kind == files.KindPercent:
		return controlColumn
	case row.Fixed:
		return min(monoWide(row.Value, small)+copyRoom, controlColumn)
	}
	return controlColumn
}

func (w Window) controlParts(row SettingRow) []geom.Box {
	c := row.Control
	switch {
	case row.Key.Kind == files.KindBool || row.Key.Switch:
		return []geom.Box{ToggleAt(c)}
	case row.Key.Kind == files.KindWords:
		m := w.measured()
		var parts []geom.Box
		x := c.X + 2
		for _, word := range row.Shown() {
			wide := ends(m, word, small) + 2*pickerPad
			parts = append(parts, geom.Box{X: x, Y: c.Y + 2, W: wide, H: c.H - 4})
			x += wide
		}
		return parts
	case row.Key.Kind == files.KindColour:
		var parts []geom.Box
		for i := range label.LabelColours {
			parts = append(parts, geom.Box{X: c.X + i*(swatchRow+inset), Y: c.Y + (c.H-swatchRow)/2, W: swatchRow, H: swatchRow})
		}
		return parts
	case row.Key.Kind == files.KindPercent:
		return []geom.Box{sliderTrack(c)}
	case row.Key.Kind == files.KindNumber:
		minus := geom.Box{X: c.X, Y: c.Y, W: stepWide, H: c.H}
		field := geom.Box{X: minus.X + minus.W + inset, Y: c.Y, W: stepField, H: c.H}
		plus := geom.Box{X: field.X + field.W + inset, Y: c.Y, W: stepWide, H: c.H}
		return []geom.Box{minus, field, plus}
	}
	return nil
}

func (w Window) offersOfRow(row SettingRow) []geom.Box {
	m := w.measured()
	x := row.From.X
	y := row.From.Y + row.From.H/2
	if w.narrow() {
		x += ends(m, strings.Join(row.OriginSays(), " · "), small) + padding
		y = row.From.Y + (row.From.H-lineHigh)/2
	}
	var boxes []geom.Box
	for _, offer := range row.Offers {
		wide := ends(m, offer, small)
		boxes = append(boxes, geom.Box{X: x, Y: y, W: wide, H: lineHigh})
		x += wide + padding
	}
	return boxes
}

const smallButton = 22

// SettingsFootSays is what the foot of the settings says at its left.
func (w Window) SettingsFootSays() string {
	switch {
	case w.IntrosOpen:
		return ""
	case w.ChooserOpen:
		return w.shortSaid()
	case w.KeyWriting != "":
		return WritingFoot
	case w.Troubled != nil && w.Troubled.File == w.SettingsWritten():
		if w.HereOffered() && !w.Global {
			return "All workspaces stay editable"
		}
		return "Fix the file and it is read again"
	}
	return SettingsFoot
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}

func (w Window) onConfiguration(name string) Window {
	if w.IntrosOpen {
		return w.inIntros(name)
	}
	if w.ChooserOpen {
		return w.inChooser(name)
	}
	if w.colourKey != "" {
		switch name {
		case "right", "l":
			w.ColourPick = min(len(label.LabelColours)-1, w.ColourPick+1)
		case "left", "h":
			w.ColourPick = max(0, w.ColourPick-1)
		case "enter", "space":
			return w.colouredSetting(w.ColourPick)
		case "esc":
			w.colourKey = ""
		}
		return w
	}
	if w.KeyWriting != "" {
		switch name {
		case "enter":
			return w.settingEdited()
		case "esc":
			w.KeyWriting = ""
		case "backspace":
			w.KeyWritten = w.KeyWritten.Back()
		case "left":
			w.KeyWritten = w.KeyWritten.Left()
		case "right":
			w.KeyWritten = w.KeyWritten.Right()
		}
		return w
	}
	rows := w.SettingRows()
	switch name {
	case "down", "j":
		w.RowOn = min(len(rows)-1, w.RowOn+1)
		return w.rowInSight()
	case "up", "k":
		w.RowOn = max(0, w.RowOn-1)
		return w.rowInSight()
	case "tab":
		return w.scoped(!w.Global)
	case "enter", "space", " ":
		if w.RowOn >= 0 && w.RowOn < len(rows) {
			return w.acting(rows[w.RowOn], -1)
		}
	case "backspace":
		if w.RowOn >= 0 && w.RowOn < len(rows) && len(rows[w.RowOn].Offers) > 0 && rows[w.RowOn].Offers[0] != SetHereOffer {
			return w.unsetting(rows[w.RowOn].Key.Name)
		}
	case "left", "right":
		if w.RowOn >= 0 && w.RowOn < len(rows) && rows[w.RowOn].Key.Kind == files.KindPercent {
			return w.steppingPercent(rows[w.RowOn], map[string]int{"left": -sliderStep, "right": sliderStep}[name])
		}
		if w.RowOn >= 0 && w.RowOn < len(rows) && rows[w.RowOn].Key.Kind == files.KindWords {
			row := rows[w.RowOn]
			at := slices.Index(row.Words, row.Value)
			if name == "left" {
				at = max(0, at-1)
			} else {
				at = min(len(row.Words)-1, at+1)
			}
			return w.acting(row, at)
		}
	}
	return w
}

func (w Window) rowInSight() Window {
	rows, region := w.SettingRows(), SettingsRows(w.SettingsBox())
	if w.RowOn < 0 || w.RowOn >= len(rows) {
		return w
	}
	row := rows[w.RowOn]
	switch {
	case row.At.Y < region.Y:
		w.SettingsAt -= region.Y - row.At.Y + sectionHigh
	case row.At.Y+row.At.H > region.Y+region.H:
		w.SettingsAt += row.At.Y + row.At.H - region.Y - region.H
	}
	w.SettingsAt = max(0, w.SettingsAt)
	return w
}

func (w Window) scoped(global bool) Window {
	if !global && !w.HereOffered() {
		return w
	}
	w.Global, w.SettingsAt, w.RowOn, w.KeyWriting, w.RefusedKey = global, 0, -1, "", ""
	return w
}

func (w Window) acting(row SettingRow, word int) Window {
	if row.Off || row.Fixed {
		return w
	}
	switch row.Key.Kind {
	case files.KindBool:
		value := "true"
		if row.Value == "true" {
			value = "false"
		}
		return w.setting(row.Key.Name, value)
	case files.KindWords:
		if word < 0 {
			word = (slices.Index(row.Words, row.Value) + 1) % len(row.Words)
		}
		if row.Words[word] == row.Value {
			return w
		}
		return w.setting(row.Key.Name, row.Words[word])
	case files.KindTheme:
		return w.openingChooser()
	case files.KindIntro:
		return w.openingIntros()
	case files.KindColour:
		w.colourKey, w.ColourPick = row.Key.Name, max(0, slices.Index(label.LabelColours, row.Value))
		return w
	case files.KindDay:
		return w.openingCalendarFor(row.Key.Name, row.Value)
	}
	w.KeyWriting, w.KeyWritten, w.RefusedKey = row.Key.Name, board.Written{}.Types(row.Value).All(), ""
	return w
}

func (w Window) filling(global bool) Window {
	in := engine.Input{Given: map[string]string{}}
	if global {
		in.Given["global"] = "true"
	}
	if _, err := w.work.Ask(engine.ConfigFill, in); err != nil {
		return w.failing(err)
	}
	return w
}

func (w Window) leavingSettings() Window {
	if w.KeyWriting != "" {
		w.Screen, w.Question = board.OnAsking, board.AboutLeavingSettings
		return w
	}
	return w.Choose("Close")
}

func (w Window) settingEdited() Window {
	key, value := w.KeyWriting, strings.TrimSpace(w.KeyWritten.Reads())
	w.KeyWriting = ""
	if key == "approach.label" || value != "" {
		return w.setting(key, value)
	}
	return w.unsetting(key)
}

func (w Window) setting(key, value string) Window {
	in := engine.Input{Given: map[string]string{"key": key, "value": value}}
	if w.Global {
		in.Given["global"] = "true"
	}
	if _, err := w.work.Ask(engine.ConfigSet, in); err != nil {
		w.RefusedKey, w.RefusedSay = key, err.Error()
		return w
	}
	w.RefusedKey = ""
	return w.applyingOwn()
}

func (w Window) unsetting(key string) Window {
	in := engine.Input{Given: map[string]string{"key": key}}
	if w.Global {
		in.Given["global"] = "true"
	}
	if _, err := w.work.Ask(engine.ConfigUnset, in); err != nil {
		w.RefusedKey, w.RefusedSay = key, err.Error()
		return w
	}
	w.RefusedKey = ""
	return w.applyingOwn()
}

func (w Window) applyingOwn() Window {
	got, ok := w.readingConfig()
	if !ok || got.Problem != "" {
		return w
	}
	w = w.takingConfig(got)
	w.Stamp = workspace.StampOf(got)
	return w
}

func (w Window) clickingSettings(p geom.Point) Window {
	if w.IntrosOpen {
		return w.clickingIntros(p)
	}
	if w.ChooserOpen {
		return w.clickingChooser(p)
	}
	w.colourKey = ""
	if w.KeyWriting != "" && !w.writingAt(p) {
		w = w.settingEdited()
	}
	for i, segment := range w.ScopeSegments() {
		if segment.In(p) {
			return w.scoped(w.scopeWords()[i] == AllWorkspaces)
		}
	}
	for _, section := range w.SettingSections() {
		if section.At.In(p) {
			return w.toSection(section)
		}
	}
	region := SettingsRows(w.SettingsBox())
	if !region.In(p) {
		return w
	}
	for i, row := range w.SettingRows() {
		if !row.At.In(p) {
			continue
		}
		w.RowOn = i
		for j, offer := range row.OfferAt {
			if offer.In(p) {
				return w.offeredOnRow(row, row.Offers[j])
			}
		}
		for j, part := range row.Parts {
			if part.In(p) {
				return w.actingOnPart(row, j)
			}
		}
		if row.Control.In(p) {
			return w.acting(row, -1)
		}
		return w
	}
	return w
}

func (w Window) actingOnPart(row SettingRow, part int) Window {
	switch {
	case row.Key.Kind == files.KindBool || row.Key.Switch:
		return w.acting(row, -1)
	case row.Key.Kind == files.KindColour:
		w.colourKey = row.Key.Name
		return w.colouredSetting(part)
	case row.Key.Kind == files.KindNumber && part != 1:
		n, _ := strconv.Atoi(row.Value)
		return w.setting(row.Key.Name, strconv.Itoa(max(1, n+part-1)))
	case row.Key.Kind == files.KindNumber:
		return w.acting(row, -1)
	}
	return w.acting(row, part)
}

func (w Window) writingAt(p geom.Point) bool {
	for _, row := range w.SettingRows() {
		if row.Key.Name == w.KeyWriting {
			return row.Control.In(p)
		}
	}
	return false
}

func (w Window) offeredOnRow(row SettingRow, offer string) Window {
	if row.Off {
		return w
	}
	if offer == SetHereOffer {
		value := row.Value
		if row.Key.Kind == files.KindText && value == "" {
			return w
		}
		return w.setting(row.Key.Name, value)
	}
	return w.unsetting(row.Key.Name)
}

func (w Window) toSection(section SettingSection) Window {
	if section.First < 0 {
		w = w.scoped(true)
		for _, other := range w.SettingSections() {
			if other.Name == section.Name {
				section = other
			}
		}
	}
	rows := w.SettingRows()
	if section.First >= 0 && section.First < len(rows) {
		w.SettingsAt = max(0, rows[section.First].At.Y-sectionHigh+w.SettingsAt-SettingsRows(w.SettingsBox()).Y)
	}
	return w
}

func (w Window) hoveringSettings(p geom.Point) Window {
	w.hoverKey = ""
	if w.ChooserOpen || w.IntrosOpen || !SettingsRows(w.SettingsBox()).In(p) {
		return w
	}
	for _, row := range w.SettingRows() {
		if row.At.In(p) {
			w.hoverKey = row.Key.Name
		}
	}
	return w
}

func (w Window) scrollingSettings(by int) Window {
	region := SettingsRows(w.SettingsBox())
	w.SettingsAt = min(max(0, w.settingsLength()-region.H), max(0, w.SettingsAt+by))
	return w
}

func (w Window) choosingOnSettings(name string) (Window, bool) {
	if w.Screen != board.OnConfiguration {
		return w, false
	}
	if w.IntrosOpen {
		return w.choosingOnIntros(name)
	}
	switch {
	case name == OpenTheFile:
		w.Opens = w.SettingsWritten()
		return w.filling(w.Global), true
	case name == WriteWhole:
		return w.writingConfiguration(), true
	case name == OpenThemes:
		w.Opens = themesFolder(w.Configured)
	case strings.HasPrefix(name, "Wear "):
		return w.wearing(), true
	case name == "Cancel" && w.ChooserOpen:
		return w.leavingChooser(), true
	default:
		return w, false
	}
	return w, true
}

func (w Window) settingsOffers() []string {
	if w.IntrosOpen {
		return w.introsOffers()
	}
	if w.ChooserOpen {
		return []string{"Wear " + w.Previewing, "Cancel", OpenThemes}
	}
	if w.Global && !w.ConfigHere {
		return []string{"Close", WriteWhole}
	}
	return []string{"Close", OpenTheFile}
}

func (w Window) settingsShortcut(offer string) string {
	switch {
	case w.IntrosOpen:
		return w.introsShortcut(offer)
	case w.KeyWriting != "":
		return ""
	case w.ChooserOpen && strings.HasPrefix(offer, "Wear "):
		return "enter"
	case offer == "Cancel" || offer == "Close":
		return "esc"
	}
	return ""
}

func (w Window) troubledBy(got workspace.Config) Window {
	if got.Problem == "" {
		w.Troubled = nil
		return w
	}
	trouble := workspace.TroubleIn(got, w.Configured)
	w.Troubled = &trouble
	return w
}

func (w Window) rereading() Window {
	if w.Configured.File == "" {
		return w
	}
	stamp, moved := w.Stamp.Moved()
	w.Stamp = stamp
	if w.Changed != nil {
		var gone bool
		if w.ChangedFor, gone = lingered(w.ChangedFor); gone {
			w.Changed, w.ChangedFor = nil, 0
		}
	}
	if !moved {
		return w
	}
	got, ok := w.readingConfig()
	if !ok {
		return w
	}
	if got.Problem != "" {
		return w.troubledBy(got)
	}
	changes := workspace.ConfigChanges(w.Configured, got)
	in := filepath.Base(got.File)
	for _, change := range changes {
		if slices.Contains(got.WrittenHere, change.Key) || slices.Contains(w.Configured.WrittenHere, change.Key) {
			in = files.SettingsFile
		}
	}
	w.Troubled = nil
	w = w.takingConfig(got)
	if len(changes) > 0 {
		w.Changed, w.ChangedIn, w.ChangedFor = changes, in, 0
	}
	return w
}

func (w Window) takingConfig(got workspace.Config) Window {
	was := w.Configured
	w.Configured, w.ReadAt = got, w.Now
	_, missing := os.Stat(got.File)
	w.ConfigHere = missing == nil
	if got.Window != nil && (was.Window == nil || was.Window.Theme != got.Window.Theme || w.skin == nil) {
		w = w.wornFrom(got)
	}
	w.switcher.placeRead = false
	w = w.readingPlace()
	if pages := engine.PagesOf(got); was.File == "" || !samePages(pages, engine.PagesOf(was)) {
		w = w.Paged(pages)
		if w.Version != 0 {
			w = w.Look()
		}
	}
	return w
}

func samePages(a, b engine.Pages) bool {
	for _, page := range []string{engine.PageTasks, engine.PageObjectives, engine.PageTime} {
		if a.Shows(page) != b.Shows(page) {
			return false
		}
	}
	return a.WeekDays() == b.WeekDays() && a.Recurs() == b.Recurs()
}

// ReadingConfiguration is the window having read the configuration it opens with, to know when it changes: what it reads again on its beat starts from here.
func (w Window) ReadingConfiguration() Window {
	got, ok := w.readingConfig()
	if !ok || got.Problem != "" {
		return w
	}
	w.Configured, w.ReadAt, w.Stamp = got, w.Now, workspace.StampOf(got)
	_, missing := os.Stat(got.File)
	w.ConfigHere = missing == nil
	return w
}

// OpenTaken says the loop has handed the file asked for to the system.
func (w Window) OpenTaken() Window {
	w.Opens = ""
	return w
}

// OpenCommand is the command that hands a file to the system, to open in the editor it keeps for that kind of file.
func OpenCommand(goos, path string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{"-t", path}
	case "windows":
		return "cmd", []string{"/c", "start", "", path}
	}
	return "xdg-open", []string{path}
}

// Opened takes what came of handing the file to the system: a command that could not start is said.
func (w Window) Opened(err error) Window {
	if err != nil {
		return w.failing(err)
	}
	return w
}

// HandOver starts the command that has the system open a file, and does not wait for it.
func HandOver(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)
	command := exec.Command(name, args...)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}

// SettingsBox is the panel of the settings in this window.
func (w Window) SettingsBox() geom.Box {
	return w.placedPanel(board.OnConfiguration, w.settingsOpens())
}

func machineKeys() []files.Key {
	var machine []files.Key
	for _, key := range files.Keys() {
		if key.Machine {
			machine = append(machine, key)
		}
	}
	return machine
}

func (w Window) keysShown() []files.Key {
	var shown []files.Key
	for _, key := range files.Keys() {
		if w.Global && key.Store || !w.Global && key.Machine {
			continue
		}
		shown = append(shown, key)
	}
	return shown
}

// GhostShows is how much of what is dragged shows: what the configuration says, or config.DefaultGhost.
func (w Window) GhostShows() float32 {
	if w.Configured.Window != nil && w.Configured.Window.Ghost > 0 {
		return float32(w.Configured.Window.Ghost) / 100
	}
	return float32(files.DefaultGhost) / 100
}
