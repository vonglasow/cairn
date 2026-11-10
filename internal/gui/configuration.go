package gui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"cairn/internal/config"
	"cairn/internal/engine"
)

// OpenSettings is the offer handing the settings of the workspace shown to the system, to open in its editor.
const OpenSettings = "Open settings.toml"

// OpenGlobal is the offer handing the global configuration to the system, to open in its editor.
const OpenGlobal = "Open config.toml"

// WriteWhole is the offer writing the whole global configuration, every key explained, where none is written yet.
const WriteWhole = "Write the whole file"

// SettingsFoot is what the foot of the settings says: nothing waits for a button, and the file is read again when it changes.
const SettingsFoot = "Saved as it is changed · the file is read again when it changes"

// The words a row says of where its value comes from.
const (
	SetHere    = "set here"
	FromGlobal = "from global"
	ByDefault  = "default"
	InFile     = "in config.toml"
	WhenOpened = "read when cairn opens"
)

// The offers a row makes beside where its value comes from.
const (
	ResetToGlobal  = "Reset to global"
	ResetShort     = "Reset"
	ResetToDefault = "Reset to default"
	SetHereOffer   = "Set here"
)

type configurationState struct {
	configured  engine.Config
	configHere  bool
	opens       string
	stamp       engine.ConfigStamp
	readAt      time.Time
	trouble     *engine.Trouble
	changed     []engine.ConfigChange
	changedIn   string
	changedFor  int
	global      bool
	settingsAt  int
	rowOn       int
	hoverKey    string
	keyWriting  string
	keyWritten  Written
	refusedKey  string
	refusedSay  string
	chooserOpen bool
	tileOn      int
	previewing  string
	tiles       []schemeOffered
	skin        *Skin
}

func (w Window) readingConfig() (engine.Config, bool) {
	got, err := w.work.Ask(engine.ConfigShow, engine.Input{})
	if err != nil || got.Config == nil {
		return engine.Config{}, false
	}
	return *got.Config, true
}

func (w Window) openingConfiguration() Window {
	got, ok := w.readingConfig()
	if !ok {
		return w.failing(errNoConfiguration)
	}
	if got.Problem == "" || w.configured.File == "" {
		w.configured = got
	}
	w = w.troubledBy(got)
	_, missing := os.Stat(w.configured.File)
	w.configHere = missing == nil
	w.global = !w.HereOffered()
	w.screen, w.rowOn, w.settingsAt, w.chooserOpen, w.keyWriting, w.refusedKey = OnConfiguration, -1, 0, false, "", ""
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
	return w.switcher.place.Workspace != "" || w.configured.Settings != ""
}

// SettingsScope is the scope the settings write: the workspace's name, the folder of the store, or Global.
func (w Window) SettingsScope() string {
	if w.global {
		return "Global"
	}
	return w.hereName()
}

func (w Window) hereName() string {
	if name := w.switcher.place.Name(); name != "" {
		return name
	}
	return "here"
}

// SettingsPath is what the panel says beside the scope: the file written, and the one it lies over or under.
func (w Window) SettingsPath() string {
	if w.global {
		if w.HereOffered() {
			return tilde(w.configured.File) + " · under every workspace's own settings"
		}
		return tilde(w.configured.File)
	}
	if w.narrow() {
		return tilde(engine.SettingsOf(w.configured))
	}
	return tilde(engine.SettingsOf(w.configured)) + " over " + tilde(w.configured.File)
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
	Key                         config.Key
	Value, Comes, Aside, Theirs string
	Offers                      []string
	Off, Fixed                  bool
	Said                        string
	At, Name, Control, From     Box
	OfferAt                     []Box
	Parts                       []Box
	Words                       []string
}

// SettingSection is a section of the settings: its name, how many rows it holds in the scope shown, and the row it starts at.
type SettingSection struct {
	Name, Count string
	First       int
	At          Box
}

func (w Window) keysShown() []config.Key {
	var shown []config.Key
	for _, key := range config.Keys() {
		if w.global && key.Store || !w.global && key.Machine {
			continue
		}
		shown = append(shown, key)
	}
	return shown
}

func (w Window) rowOfKey(key config.Key) SettingRow {
	c := w.configured
	row := SettingRow{Key: key, Value: w.valueShown(key.Name), Words: key.Words}
	here, global := slices.Contains(c.WrittenHere, key.Name), slices.Contains(c.Written, key.Name)
	switch {
	case key.Name == "store" || key.Name == "notes":
		row.Comes, row.Fixed = WhenOpened, true
	case w.global && global:
		row.Comes, row.Offers = InFile, []string{ResetToDefault}
	case w.global:
		row.Comes = ByDefault
	case here:
		row.Comes, row.Offers = SetHere, []string{ResetToGlobal}
		if value, ok := c.Global[key.Name]; ok && key.Name != "workspace.colour" {
			row.Aside = "← global: " + value
		}
	case global:
		row.Comes = FromGlobal
	default:
		row.Comes = ByDefault
	}
	if w.global && here && w.HereOffered() {
		row.Theirs = "· " + w.hereName() + " sets its own: " + engine.ValueIn(c, key.Name)
	}
	if !row.Fixed && row.Comes != SetHere && row.Comes != InFile && w.hoverKey == key.Name && !w.global {
		row.Offers = []string{SetHereOffer}
	}
	if w.narrow() && len(row.Offers) > 0 && row.Offers[0] == ResetToGlobal {
		row.Offers = []string{ResetShort}
	}
	if w.trouble != nil && w.trouble.File == w.scopeFile() {
		row.Off, row.Offers = true, nil
	}
	if w.refusedKey == key.Name {
		row.Said = w.refusedSay
	}
	return row
}

func (w Window) scopeFile() string {
	if w.global {
		return w.configured.File
	}
	return engine.SettingsOf(w.configured)
}

func (w Window) valueShown(key string) string {
	c := w.configured
	switch key {
	case "store":
		return tilde(c.Store.At)
	case "notes":
		return tilde(c.Notes.At)
	}
	value := engine.ValueIn(c, key)
	if global, ok := c.Global[key]; ok && w.global {
		value = global
	}
	if value == engine.NotSaid {
		return ""
	}
	return value
}

// SettingSections is every section of the settings in the scope shown, with how many rows each holds.
func (w Window) SettingSections() []SettingSection {
	shown := w.keysShown()
	var sections []SettingSection
	for _, name := range config.Sections() {
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
	if !w.global {
		sections = append(sections, SettingSection{Name: config.SectionPlaces, Count: "global", First: -1})
	}
	nav := SettingsNav(w.SettingsBox())
	for i := range sections {
		sections[i].At = Box{X: nav.X + inset*2, Y: nav.Y + padding + i*navRow, W: nav.W - inset*4, H: navRow}
	}
	return sections
}

const (
	settingsWide   = 860
	settingsHigh   = 700
	settingsMargin = 16
	navWide        = 176
	navRow         = 30
	fromWide       = 318
	rowHigh        = 46
	stackedHigh    = 72
	sectionHigh    = 38
	controlHigh    = 28
	scopeHigh      = 28
	fieldLeast     = 90
	rowsPad        = 18
	sbarWide       = 4
)

// SettingsBox is the panel of the settings in a window of that size: as wide and as tall as it is drawn, or the room the window has less a margin.
func SettingsBox(width, height int) Box {
	wide, high := min(settingsWide, width-2*settingsMargin), min(settingsHigh, height-2*settingsMargin)
	return Box{X: (width - wide) / 2, Y: (height - high) / 2, W: wide, H: high}
}

// SettingsBox is the panel of the settings in this window.
func (w Window) SettingsBox() Box {
	if w.chooserOpen {
		return ChooserBox(w.width, w.height)
	}
	return SettingsBox(w.width, w.height)
}

// ChooserBox is the panel the theme chooser is drawn in, a little wider than the settings.
func ChooserBox(width, height int) Box {
	wide, high := min(900, width-2*settingsMargin), min(settingsHigh, height-2*settingsMargin)
	return Box{X: (width - wide) / 2, Y: (height - high) / 2, W: wide, H: high}
}

func (w Window) narrow() bool { return SettingsBox(w.width, w.height).W < settingsWide }

// ScopeAt is the line under the heading: the scope chosen, and the file it writes.
func ScopeAt(box Box) Box {
	head := HeadingAt(box)
	return Box{X: box.X + padding, Y: head.Y + head.H + padding, W: box.W - 2*padding, H: scopeHigh}
}

// SettingsBody is what lies between the scope and the foot: the sections and the rows.
func SettingsBody(box Box) Box {
	scope := ScopeAt(box)
	top := scope.Y + scope.H + padding
	foot := PanelOfferAt(box, 0).Y - between
	return Box{X: box.X, Y: top, W: box.W, H: foot - top}
}

// SettingsNav is the column the sections are listed down at the left of the body, none at the narrowest.
func SettingsNav(box Box) Box {
	if box.W < settingsWide {
		return Box{}
	}
	body := SettingsBody(box)
	return Box{X: body.X, Y: body.Y, W: navWide, H: body.H}
}

// SettingsRows is where the rows scroll: the body right of the sections.
func SettingsRows(box Box) Box {
	body, nav := SettingsBody(box), SettingsNav(box)
	x := body.X + nav.W + rowsPad
	return Box{X: x, Y: body.Y + inset, W: body.X + body.W - rowsPad - x, H: body.H - inset}
}

// ScopeSegments is where the two scopes are laid on the scope's line: the workspace's, then Global; one only when the store has no settings of its own.
func (w Window) ScopeSegments() []Box {
	scope := ScopeAt(w.SettingsBox())
	m := w.measured()
	var boxes []Box
	x := scope.X + 2
	for i, said := range w.scopeWords() {
		wide := ends(m, said, small) + 2*pickerPad
		if i == 0 && w.HereOffered() {
			wide += dotSide + between
		}
		boxes = append(boxes, Box{X: x, Y: scope.Y + 2, W: wide, H: scope.H - 4})
		x += wide + 2
	}
	return boxes
}

func (w Window) scopeWords() []string {
	if !w.HereOffered() {
		return []string{"Global"}
	}
	return []string{w.hereName(), "Global"}
}

// ScopePathAt is where the path the scope writes is said, after its segments.
func (w Window) ScopePathAt() Box {
	scope, segments := ScopeAt(w.SettingsBox()), w.ScopeSegments()
	last := segments[len(segments)-1]
	x := last.X + last.W + 2 + padding
	return Box{X: x, Y: scope.Y + (scope.H-lineHigh)/2, W: max(0, scope.X+scope.W-x), H: lineHigh}
}

// SettingRows is every row of the settings in the scope shown, laid out down the rows' column, scrolled.
func (w Window) SettingRows() []SettingRow {
	rows := SettingsRows(w.SettingsBox())
	keys := w.keysShown()
	out := make([]SettingRow, 0, len(keys))
	y, section := rows.Y-w.settingsAt, ""
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
		row = w.laidRow(row, Box{X: rows.X - padding + 2, Y: y, W: rows.W + 2*(padding-2), H: high})
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
			heads = append(heads, SettingSection{Name: section, Count: itoa(count), First: i, At: Box{X: rows.X, Y: row.At.Y - sectionHigh + padding, W: rows.W, H: lineHigh}})
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
	return last.At.Y + last.At.H + w.settingsAt - SettingsRows(w.SettingsBox()).Y + padding
}

func (w Window) laidRow(row SettingRow, at Box) SettingRow {
	row.At = at
	inner := Box{X: at.X + padding - 2, Y: at.Y, W: at.W - 2*(padding-2), H: at.H}
	line := Box{X: inner.X, Y: inner.Y + between, W: inner.W, H: rowHigh - 2*between}
	if w.narrow() {
		row.From = Box{X: inner.X, Y: line.Y + line.H, W: inner.W, H: stackedHigh - rowHigh}
	} else {
		row.From = Box{X: inner.X + inner.W - fromWide, Y: line.Y, W: fromWide, H: line.H}
	}
	right := inner.X + inner.W
	if !w.narrow() {
		right = row.From.X - padding
	}
	wide := w.controlWide(row)
	row.Control = Box{X: right - wide, Y: line.Y + (line.H-controlHigh)/2, W: wide, H: controlHigh}
	row.Name = Box{X: inner.X, Y: line.Y, W: max(0, row.Control.X-padding-inner.X), H: line.H}
	row.Parts = w.controlParts(row)
	row.OfferAt = w.offersOfRow(row)
	return row
}

func (w Window) controlWide(row SettingRow) int {
	m := w.measured()
	switch row.Key.Kind {
	case config.KindBool:
		return toggleWide
	case config.KindWords:
		wide := 4
		for _, word := range row.Words {
			wide += ends(m, word, small) + 2*pickerPad
		}
		return wide
	case config.KindTheme:
		return PickerWide(m, row.Value, false, true)
	case config.KindColour:
		return PickerWide(m, row.Value, true, false)
	}
	if row.Fixed {
		return min(monoWide(row.Value, small)+copyRoom, 260)
	}
	least := fieldLeast
	switch {
	case row.Key.Kind == config.KindPath:
		least = 200
	case row.Key.Kind == config.KindText:
		least = 130
	case row.Key.Unit != "":
		least = 80 + ends(m, row.Key.Unit, small)
	}
	said := row.Value
	if w.keyWriting == row.Key.Name {
		said = w.keyWritten.Reads()
	}
	return min(260, max(least, monoWide(said, small)+2*pickerPad+ends(m, row.Key.Unit, small)))
}

func (w Window) controlParts(row SettingRow) []Box {
	switch row.Key.Kind {
	case config.KindBool:
		return []Box{ToggleAt(row.Control)}
	case config.KindWords:
		m := w.measured()
		var parts []Box
		x := row.Control.X + 2
		for _, word := range row.Words {
			wide := ends(m, word, small) + 2*pickerPad
			parts = append(parts, Box{X: x, Y: row.Control.Y + 2, W: wide, H: row.Control.H - 4})
			x += wide
		}
		return parts
	}
	return nil
}

func (w Window) offersOfRow(row SettingRow) []Box {
	m := w.measured()
	x := row.From.X + row.From.W
	var boxes []Box
	for i := len(row.Offers) - 1; i >= 0; i-- {
		wide := ends(m, row.Offers[i], small) + 2*between + 4
		x -= wide
		boxes = append([]Box{{X: x, Y: row.From.Y + (row.From.H-smallButton)/2, W: wide, H: smallButton}}, boxes...)
		x -= gap
	}
	return boxes
}

const smallButton = 22

// SettingsFootSays is what the foot of the settings says at its left.
func (w Window) SettingsFootSays() string {
	switch {
	case w.chooserOpen:
		return w.shortSaid()
	case w.trouble != nil && w.trouble.File == w.scopeFile():
		if w.HereOffered() && !w.global {
			return "Global stays editable"
		}
		return "Fix the file and it is read again"
	case w.narrow():
		return "Saved as it is changed"
	}
	return SettingsFoot
}

// Configured is the configuration the window holds in force: the last one read that could be read.
func (w Window) Configured() engine.Config { return w.configured }

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
	if w.chooserOpen {
		return w.inChooser(name)
	}
	if w.keyWriting != "" {
		switch name {
		case "enter":
			return w.settingEdited()
		case "esc":
			w.keyWriting = ""
		case "backspace":
			w.keyWritten = w.keyWritten.Back()
		case "left":
			w.keyWritten = w.keyWritten.Left()
		case "right":
			w.keyWritten = w.keyWritten.Right()
		}
		return w
	}
	rows := w.SettingRows()
	switch name {
	case "down", "j":
		w.rowOn = min(len(rows)-1, w.rowOn+1)
		return w.rowInSight()
	case "up", "k":
		w.rowOn = max(0, w.rowOn-1)
		return w.rowInSight()
	case "tab":
		return w.scoped(!w.global)
	case "enter", "space", " ":
		if w.rowOn >= 0 && w.rowOn < len(rows) {
			return w.acting(rows[w.rowOn], -1)
		}
	case "backspace":
		if w.rowOn >= 0 && w.rowOn < len(rows) && len(rows[w.rowOn].Offers) > 0 && rows[w.rowOn].Offers[0] != SetHereOffer {
			return w.unsetting(rows[w.rowOn].Key.Name)
		}
	case "left", "right":
		if w.rowOn >= 0 && w.rowOn < len(rows) && rows[w.rowOn].Key.Kind == config.KindWords {
			row := rows[w.rowOn]
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
	if w.rowOn < 0 || w.rowOn >= len(rows) {
		return w
	}
	row := rows[w.rowOn]
	switch {
	case row.At.Y < region.Y:
		w.settingsAt -= region.Y - row.At.Y + sectionHigh
	case row.At.Y+row.At.H > region.Y+region.H:
		w.settingsAt += row.At.Y + row.At.H - region.Y - region.H
	}
	w.settingsAt = max(0, w.settingsAt)
	return w
}

func (w Window) scoped(global bool) Window {
	if !global && !w.HereOffered() {
		return w
	}
	w.global, w.settingsAt, w.rowOn, w.keyWriting, w.refusedKey = global, 0, -1, "", ""
	return w
}

func (w Window) acting(row SettingRow, word int) Window {
	if row.Off || row.Fixed {
		return w
	}
	switch row.Key.Kind {
	case config.KindBool:
		value := "true"
		if row.Value == "true" {
			value = "false"
		}
		return w.setting(row.Key.Name, value)
	case config.KindWords:
		if word < 0 {
			word = (slices.Index(row.Words, row.Value) + 1) % len(row.Words)
		}
		if row.Words[word] == row.Value {
			return w
		}
		return w.setting(row.Key.Name, row.Words[word])
	case config.KindTheme:
		return w.openingChooser()
	case config.KindColour:
		return w.setting(row.Key.Name, nextColour(row.Value))
	}
	w.keyWriting, w.keyWritten, w.refusedKey = row.Key.Name, Written{}.Types(row.Value).All(), ""
	return w
}

func (w Window) settingEdited() Window {
	key, value := w.keyWriting, strings.TrimSpace(w.keyWritten.Reads())
	w.keyWriting = ""
	if key == "approach.label" || value != "" {
		return w.setting(key, value)
	}
	return w.unsetting(key)
}

func (w Window) setting(key, value string) Window {
	in := engine.Input{Given: map[string]string{"key": key, "value": value}}
	if w.global {
		in.Given["global"] = "true"
	}
	if _, err := w.work.Ask(engine.ConfigSet, in); err != nil {
		w.refusedKey, w.refusedSay = key, err.Error()
		return w
	}
	w.refusedKey = ""
	return w.applyingOwn()
}

func (w Window) unsetting(key string) Window {
	in := engine.Input{Given: map[string]string{"key": key}}
	if w.global {
		in.Given["global"] = "true"
	}
	if _, err := w.work.Ask(engine.ConfigUnset, in); err != nil {
		w.refusedKey, w.refusedSay = key, err.Error()
		return w
	}
	w.refusedKey = ""
	return w.applyingOwn()
}

func (w Window) applyingOwn() Window {
	got, ok := w.readingConfig()
	if !ok || got.Problem != "" {
		return w
	}
	w = w.takingConfig(got)
	w.stamp = engine.StampOf(got)
	return w
}

func nextColour(now string) string {
	colours := labelColourNames()
	return colours[(slices.Index(colours, now)+1)%len(colours)]
}

func (w Window) clickingSettings(p Point) Window {
	if w.chooserOpen {
		return w.clickingChooser(p)
	}
	for i, segment := range w.ScopeSegments() {
		if segment.In(p) {
			return w.scoped(w.scopeWords()[i] == "Global")
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
		w.rowOn = i
		for j, offer := range row.OfferAt {
			if offer.In(p) {
				return w.offeredOnRow(row, row.Offers[j])
			}
		}
		for j, part := range row.Parts {
			if part.In(p) {
				if row.Key.Kind == config.KindBool {
					return w.acting(row, -1)
				}
				return w.acting(row, j)
			}
		}
		if row.Control.In(p) {
			return w.acting(row, -1)
		}
		return w
	}
	return w
}

func (w Window) offeredOnRow(row SettingRow, offer string) Window {
	if row.Off {
		return w
	}
	if offer == SetHereOffer {
		value := row.Value
		if row.Key.Kind == config.KindText && value == "" {
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
		w.settingsAt = max(0, rows[section.First].At.Y-sectionHigh+w.settingsAt-SettingsRows(w.SettingsBox()).Y)
	}
	return w
}

func (w Window) hoveringSettings(p Point) Window {
	w.hoverKey = ""
	if w.chooserOpen || !SettingsRows(w.SettingsBox()).In(p) {
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
	w.settingsAt = min(max(0, w.settingsLength()-region.H), max(0, w.settingsAt+by))
	return w
}

func (w Window) choosingOnSettings(name string) (Window, bool) {
	if w.screen != OnConfiguration {
		return w, false
	}
	switch {
	case name == OpenSettings:
		w.opens = engine.SettingsOf(w.configured)
	case name == OpenGlobal:
		w.opens = w.configured.File
	case name == WriteWhole:
		return w.writingConfiguration(), true
	case name == OpenThemes:
		w.opens = themesFolder(w.configured)
	case strings.HasPrefix(name, "Wear "):
		return w.wearing(), true
	case name == "Cancel" && w.chooserOpen:
		return w.leavingChooser(), true
	default:
		return w, false
	}
	return w, true
}

func (w Window) settingsOffers() []string {
	if w.chooserOpen {
		return []string{"Wear " + w.previewing, "Cancel", OpenThemes}
	}
	offers := []string{"Close"}
	switch {
	case w.global && !w.configHere:
		return append(offers, WriteWhole)
	case w.global:
		return append(offers, OpenGlobal)
	}
	return append(offers, OpenSettings)
}

func (w Window) settingsShortcut(offer string) string {
	switch {
	case w.keyWriting != "":
		return ""
	case w.chooserOpen && strings.HasPrefix(offer, "Wear "):
		return "enter"
	case offer == "Cancel" || offer == "Close":
		return "esc"
	}
	return ""
}

func (w Window) troubledBy(got engine.Config) Window {
	if got.Problem == "" {
		w.trouble = nil
		return w
	}
	trouble := engine.TroubleIn(got, w.configured)
	w.trouble = &trouble
	return w
}

// Trouble is the file the window cannot read while it is open, and why: nothing when every file reads.
func (w Window) Trouble() (engine.Trouble, bool) {
	if w.trouble == nil {
		return engine.Trouble{}, false
	}
	return *w.trouble, true
}

// ReadAgain is what the window says of a configuration it read again: the file, and each key that changed with what it was.
func (w Window) ReadAgain() (string, []engine.ConfigChange) { return w.changedIn, w.changed }

func (w Window) rereading() Window {
	if w.configured.File == "" {
		return w
	}
	stamp, moved := w.stamp.Moved()
	w.stamp = stamp
	if w.changed != nil {
		var gone bool
		if w.changedFor, gone = lingered(w.changedFor); gone {
			w.changed, w.changedFor = nil, 0
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
	changes := engine.ConfigChanges(w.configured, got)
	in := filepath.Base(got.File)
	for _, change := range changes {
		if slices.Contains(got.WrittenHere, change.Key) || slices.Contains(w.configured.WrittenHere, change.Key) {
			in = config.SettingsFile
		}
	}
	w.trouble = nil
	w = w.takingConfig(got)
	if len(changes) > 0 {
		w.changed, w.changedIn, w.changedFor = changes, in, 0
	}
	return w
}

func (w Window) takingConfig(got engine.Config) Window {
	was := w.configured
	w.configured, w.readAt = got, w.now
	_, missing := os.Stat(got.File)
	w.configHere = missing == nil
	if got.Window != nil && (was.Window == nil || was.Window.Theme != got.Window.Theme || w.skin == nil) {
		w = w.wornFrom(got)
	}
	w.switcher.placeRead = false
	w = w.readingPlace()
	if pages := engine.PagesOf(got); was.File == "" || !samePages(pages, engine.PagesOf(was)) {
		w = w.Paged(pages)
		if w.version != 0 {
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
	w.configured, w.readAt, w.stamp = got, w.now, engine.StampOf(got)
	_, missing := os.Stat(got.File)
	w.configHere = missing == nil
	return w
}

// OpenAsked is the file the window wants the system to open, if any.
func (w Window) OpenAsked() (string, bool) { return w.opens, w.opens != "" }

// OpenTaken says the loop has handed the file asked for to the system.
func (w Window) OpenTaken() Window {
	w.opens = ""
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
