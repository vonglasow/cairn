package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
)

type seriesState struct {
	seriesOn    int
	seriesFate  int
	fateOf      int
	seriesChord bool
}

// The offers of a series shown, and what the bar offers on its tab.
const (
	EditSeriesSays   = "Edit the series"
	SkipNextSays     = "Skip the next"
	StopSeriesSays   = "Stop"
	ResumeSeriesSays = "Resume"
	DeleteSeriesSays = "Delete…"
	NewSeriesSays    = "New series"
)

// SeriesListed is the series the tab lists for what is searched: the active ones, then the stopped ones.
func (w Window) SeriesListed() []engine.Series {
	return engine.ShownSeries(w.snap.Series, w.looking.Reads())
}

// SeriesShown is the series the right of the tab shows: the one chosen, or the first listed while none is.
func (w Window) SeriesShown() (engine.Series, bool) {
	all := w.SeriesListed()
	for _, s := range all {
		if s.Number == w.seriesOn {
			return s, true
		}
	}
	if len(all) == 0 {
		return engine.Series{}, false
	}
	return all[0], true
}

// SeriesRow is one line of the list of series: a head naming the active or the stopped ones and how many, or a series, and where it stands.
type SeriesRow struct {
	Head, Count string
	Series      engine.Series
	At          Box
}

var seriesHeads = map[string]string{"active": "Active", "stopped": "Stopped"}

// SeriesRows lays the list of series out, a head over the active ones and one over the stopped ones.
func (w Window) SeriesRows() []SeriesRow {
	list, _ := SplitOf(w.Layout())
	all := w.SeriesListed()
	var rows []SeriesRow
	y, state := list.Y+gap, ""
	for _, s := range all {
		if s.State != state {
			state = s.State
			count := 0
			for _, other := range all {
				if other.State == state {
					count++
				}
			}
			rows = append(rows, SeriesRow{Head: seriesHeads[state], Count: strconv.Itoa(count), At: NameIn(Box{X: list.X, Y: y - gap, W: list.W, H: headHigh})})
			y += headHigh - gap
		}
		rows = append(rows, SeriesRow{Series: s, At: Box{X: list.X + gap, Y: y, W: list.W - 2*gap, H: ItemHigh}})
		y += ItemHigh
	}
	return rows
}

// SeriesWhen is what a series' line says at its right: its next day, or the day it was stopped.
func (w Window) SeriesWhen(s engine.Series) string {
	if s.State == "stopped" {
		return "since " + engine.DaySaid(s.Stopped, w.now)
	}
	if len(s.Next) == 0 {
		return ""
	}
	return engine.DaySaid(s.Next[0], w.now)
}

// StandingOfSeries is the standing a series' state is drawn in: an active one as an active objective, a stopped one as one on hold.
func StandingOfSeries(s engine.Series) string {
	if s.State == "stopped" {
		return "on-hold"
	}
	return "active"
}

// OccurrenceLine is one occurrence the detail of a series lists, and where.
type OccurrenceLine struct {
	Occurrence engine.Occurrence
	At         Box
}

// SeriesDetail is the right of the series tab: the series shown and where each of its parts stands.
type SeriesDetail struct {
	Series                              engine.Series
	Box, Heading, Badge, Sentence, Next Box
	Template, SoFar, Now                Box
	TemplateFacts                       Facts
	Figures                             []Figure
	NowSays                             []string
	OccurrencesHead, Foot, FootSaid     Box
	OccurrencesSaid, OccurrencesAside   string
	NextSays, FootSays                  string
	Lines                               []string
	LinesAt                             Box
	Occurrences                         []OccurrenceLine
	Offers                              []string
	Buttons                             []Box
}

const (
	footLeast      = 80
	boxLeast       = 170
	seriesBoxHigh  = 6*lineHigh + 2*padding
	occurrenceHigh = SubHeight
)

// SeriesOffers is what the detail of a series offers: changing it, skipping its open occurrence, stopping or resuming it, deleting it.
func SeriesOffers(s engine.Series) []string {
	if s.State == "stopped" {
		return []string{EditSeriesSays, ResumeSeriesSays, DeleteSeriesSays}
	}
	if len(s.Open) == 0 {
		return []string{EditSeriesSays, StopSeriesSays, DeleteSeriesSays}
	}
	return []string{EditSeriesSays, SkipNextSays, StopSeriesSays, DeleteSeriesSays}
}

// SeriesDetail lays out the right of the series tab for the series it shows.
func (w Window) SeriesDetail() SeriesDetail {
	_, box := SplitOf(w.Layout())
	s, ok := w.SeriesShown()
	d := SeriesDetail{Box: box, Series: s}
	if !ok {
		return d
	}
	inside := Box{X: box.X + detailPad, Y: box.Y + detailPad, W: box.W - 2*detailPad, H: box.H - 2*detailPad}
	m := w.measured()
	badge := engine.StandingSaid(s)
	badgeWide := ends(m, badge, small) + 2*pillPad
	y := inside.Y
	d.Heading = Box{X: inside.X, Y: y, W: max(0, inside.W-badgeWide-gap), H: buttonHigh}
	d.Badge = Box{X: inside.X + inside.W - badgeWide, Y: y + (buttonHigh-pillHigh)/2, W: badgeWide, H: pillHigh}
	y += buttonHigh + between
	d.Sentence = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh
	d.NextSays = NextSays(s)
	d.Next = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh + padding

	d.Offers = SeriesOffers(s)
	foot := inside.Y + inside.H - buttonHigh
	end := inside.X + inside.W
	widths := w.widthsOf(d.Offers)
	if sumOf(widths)+(len(d.Offers)-1)*gap > inside.W {
		d.Offers = shortOffers(d.Offers)
		widths = w.tightWidths(d.Offers)
	}
	for len(d.Offers) > 1 && sumOf(widths)+(len(d.Offers)-1)*gap > inside.W {
		d.Offers, widths = d.Offers[1:], widths[1:]
	}
	for i := len(widths) - 1; i >= 0; i-- {
		end -= widths[i]
		d.Buttons = append([]Box{{X: end, Y: foot, W: widths[i], H: buttonHigh}}, d.Buttons...)
		end -= gap
	}
	d.Foot = Box{X: inside.X, Y: foot - between - 1, W: inside.W, H: 1}
	if room := end - inside.X; room >= footLeast {
		d.FootSays = FootSaid(s)
		d.FootSaid = Box{X: inside.X, Y: foot + (buttonHigh-lineHigh)/2, W: room, H: lineHigh}
	}

	third := (inside.W - 2*padding) / 3
	if s.State == "active" && third < boxLeast {
		d.Lines = []string{"Template: " + engine.TemplateSaid(s.Template), SoFarSays(s)}
		d.LinesAt = Box{X: inside.X, Y: y, W: inside.W, H: 2 * lineHigh}
		y += 2*lineHigh + padding
	}
	if s.State == "active" && third >= boxLeast && foot-y > seriesBoxHigh+lineHigh {
		d.Template = Box{X: inside.X, Y: y, W: third, H: seriesBoxHigh}
		d.SoFar = Box{X: inside.X + third + padding, Y: y, W: third, H: seriesBoxHigh}
		d.Now = Box{X: inside.X + 2*(third+padding), Y: y, W: inside.W - 2*(third+padding), H: seriesBoxHigh}
		d.TemplateFacts = templateFacts(s.Template)
		d.Figures = []Figure{{Figure: strconv.Itoa(s.Done), Word: "done"}, {Figure: strconv.Itoa(s.Skipped), Word: "skipped"},
			{Figure: strconv.Itoa(s.Missed), Word: "missed", Late: s.Missed > 0}}
		d.NowSays = w.nowSaid(s)
		y += seriesBoxHigh + padding
	}
	d.OccurrencesSaid = "Occurrences · " + strconv.Itoa(len(s.Occurrences))
	d.OccurrencesAside = "archived ones included · a click opens the task"
	d.OccurrencesHead = Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh + inset
	for _, o := range s.Occurrences {
		if y+occurrenceHigh > d.Foot.Y-between {
			break
		}
		d.Occurrences = append(d.Occurrences, OccurrenceLine{Occurrence: o, At: Box{X: inside.X, Y: y, W: inside.W, H: occurrenceHigh}})
		y += occurrenceHigh
	}
	return d
}

// NextSays is what the detail says under the rule: its next days, or the ones resuming it today would make.
func NextSays(s engine.Series) string {
	said := engine.NextSaid(s)
	if rest, ok := strings.CutPrefix(said, "if resumed today"); ok {
		return "If resumed today" + rest
	}
	return "Next: " + said
}

// SoFarSays is what a series did so far, in a line.
func SoFarSays(s engine.Series) string {
	said := "So far: " + strconv.Itoa(s.Done) + " done · " + strconv.Itoa(s.Skipped) + " skipped · " + strconv.Itoa(s.Missed) + " missed"
	if s.TimeSpent != "" {
		said += " · " + s.TimeSpent
	}
	return said
}

// FootSaid is what the foot of the detail says of a series beside its offers: when it was stopped, or how many occurrences it made.
func FootSaid(s engine.Series) string {
	if s.State == "stopped" {
		return "stopped " + s.Stopped
	}
	return strconv.Itoa(len(s.Occurrences)) + " occurrences"
}

func templateFacts(t engine.Template) Facts {
	steps := strconv.Itoa(len(t.Steps))
	if len(t.Steps) > 0 {
		steps += " — " + strings.Join(t.Steps, ", ")
	}
	return Facts{Lines: []Fact{
		{Name: "title", Value: t.Title},
		{Name: "priority", Value: t.Priority},
		{Name: "labels", Value: domainOrNothing(strings.Join(t.Labels, " "))},
		{Name: "steps", Value: steps},
		{Name: "estimate", Value: domainOrNothing(t.Estimate)},
	}}
}

func domainOrNothing(said string) string {
	if said == "" {
		return "—"
	}
	return said
}

func (w Window) nowSaid(s engine.Series) []string {
	if len(s.Open) == 0 {
		return []string{"Nothing open.", "The next is made on its day."}
	}
	c, found := w.snap.Card(s.Open[0])
	if !found {
		return []string{"#" + strconv.Itoa(s.Open[0])}
	}
	said := c.State
	if text, show, _, _ := engine.OccurrenceDay(c, w.now); show {
		said += " · " + text
	}
	if c.SubTotal > 0 {
		said += " · " + strconv.Itoa(c.SubDone) + " of " + strconv.Itoa(c.SubTotal) + " steps"
	}
	return []string{"#" + strconv.Itoa(c.Number) + " " + c.Title, said, "One open at a time; the next is made when it is done or skipped."}
}

// OccurrenceSays is what a line of the occurrences says: its day, its task, its title, where it stands and the time spent on it.
func (w Window) OccurrenceSays(o engine.Occurrence) (day, number, title, is, spent string) {
	day = engine.DaySaid(o.Day, w.now)
	number, title = "—", "no task made"
	if o.Task != 0 {
		number = "#" + strconv.Itoa(o.Task)
		title = w.occurrenceTitle(o.Task)
	}
	is = o.Is
	if o.Archived {
		is += " · archived"
	}
	spent = o.TimeSpent
	if spent == "" {
		spent = "—"
	}
	return day, number, title, is, spent
}

func (w Window) occurrenceTitle(n int) string {
	if c, found := w.snap.Card(n); found {
		return c.Title
	}
	if s, found := w.SeriesShown(); found {
		return s.Title
	}
	return ""
}

func (w Window) seriesKey(name string) (Window, bool) {
	switch name {
	case "up", "down":
		return w.walkingSeries(map[string]int{"up": -1, "down": 1}[name]), true
	case "enter":
		return w.openingOccurrence(), true
	}
	s, ok := w.SeriesShown()
	if !ok {
		return w, false
	}
	for _, offer := range SeriesOffers(s) {
		if Shortcut(offer) == name {
			return w.seriesActing(offer), true
		}
	}
	return w, false
}

func (w Window) walkingSeries(by int) Window {
	all := w.SeriesListed()
	if len(all) == 0 {
		return w
	}
	shown, _ := w.SeriesShown()
	at := 0
	for i, s := range all {
		if s.Number == shown.Number {
			at = i
		}
	}
	w.seriesOn = all[min(len(all)-1, max(0, at+by))].Number
	return w
}

func (w Window) openingOccurrence() Window {
	s, ok := w.SeriesShown()
	if !ok || len(s.Open) == 0 {
		return w
	}
	return w.openingTask(s.Open[0])
}

func (w Window) openingTask(n int) Window {
	if _, found := w.snap.Card(n); !found {
		return w
	}
	w.chosen, w.chosenGoal = n, false
	return w.Edit()
}

func (w Window) tightWidths(offers []string) []int {
	widths := make([]int, len(offers))
	for i, offer := range offers {
		key := KeyChipWide(w.OfferKey(offer))
		if key > 0 {
			key += hintSpace
		}
		widths[i] = buttonWordWide(w.measured(), offer) + key + padding + inset
	}
	return widths
}

func shortOffers(offers []string) []string {
	out := make([]string, len(offers))
	for i, offer := range offers {
		out[i] = offer
		if short, ok := seriesShort[offer]; ok {
			out[i] = short
		}
	}
	return out
}

var seriesShort = map[string]string{EditSeriesSays: "Edit", SkipNextSays: "Skip"}

func (w Window) seriesActing(offer string) Window {
	s, ok := w.SeriesShown()
	if !ok {
		return w
	}
	for long, short := range seriesShort {
		if offer == short {
			offer = long
		}
	}
	w.seriesOn = s.Number
	switch offer {
	case EditSeriesSays:
		w.form, w.screen, w.formMoved = asOpened(Form{Number: s.Number, ForSeries: true, fields: engine.SeriesFields(s)}), OnForm, Point{}
		return w.beside()
	case SkipNextSays:
		if len(s.Open) == 0 {
			return w
		}
		return w.asking(engine.TaskSkip, engine.Numbered(s.Open[0]))
	case StopSeriesSays:
		return w.asking(engine.SeriesStop, engine.Numbered(s.Number))
	case ResumeSeriesSays:
		return w.asking(engine.SeriesResume, engine.Numbered(s.Number))
	case DeleteSeriesSays:
		w.screen, w.fateOf, w.seriesFate = OnSeriesFates, s.Number, 0
		return w
	}
	return w
}

func (w Window) onSeries(n int) Window {
	w = w.onPage(Series)
	w.seriesOn = n
	return w
}

func (w Window) startingSeries() Window {
	fields := withRepeats(engine.AddingFields(engine.StateAt(1)))
	w.form, w.screen, w.formMoved = asOpened(Form{Adding: true, fields: fields}), OnForm, Point{}
	for i, f := range fields {
		if f.Arg == engine.RepeatsField().Arg {
			w.form = w.form.onto(i)
		}
	}
	return w.beside()
}

func (w Window) clickingSeries(p Point) (Window, bool) {
	if w.screen != OnTheBoard || w.tab != Series {
		return w, false
	}
	for _, row := range w.SeriesRows() {
		if row.Head == "" && row.At.In(p) {
			w.seriesOn = row.Series.Number
			return w, true
		}
	}
	d := w.SeriesDetail()
	for i, at := range d.Buttons {
		if at.In(p) {
			return w.seriesActing(d.Offers[i]), true
		}
	}
	for _, line := range d.Occurrences {
		if line.At.In(p) && line.Occurrence.Task != 0 {
			return w.openingTask(line.Occurrence.Task), true
		}
	}
	return w, true
}

func (w Window) seriesSlotAt(p Point) Box {
	for _, row := range w.SeriesRows() {
		if row.Head == "" && row.At.In(p) {
			return row.At
		}
	}
	d := w.SeriesDetail()
	for _, at := range d.Buttons {
		if at.In(p) {
			return at
		}
	}
	for _, line := range d.Occurrences {
		if line.At.In(p) {
			return line.At
		}
	}
	return Box{}
}

// The answers to the question saving an occurrence asks.
const (
	ThisOneOnlySays = "This one only"
	WholeSeriesSays = "The whole series"
)

func (w Window) asksWhole() bool {
	if w.form.Adding || w.form.ForObjective || w.form.ForSeries || w.form.answered || !w.pages.Recurs() {
		return false
	}
	c, found := w.snap.Card(w.form.Number)
	return found && engine.ChangesTheSeries(c, w.form.fields)
}

func (w Window) changedOfTheSeries() string {
	var names []string
	for _, f := range w.form.fields {
		if f.Value != f.Was && engine.ChangesTheSeries(engine.Card{Series: &engine.Occurs{}}, []engine.Field{f}) {
			names = append(names, f.Name)
		}
	}
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and its " + names[len(names)-1]
}

func (w Window) choosingOnSeries(name string) (Window, bool) {
	switch {
	case w.screen == OnSeriesFates && name == DeleteTheSeriesSays:
		return w.confirmingSeriesFate(), true
	case w.screen == OnSeriesFates:
		w.screen = OnTheBoard
		return w, true
	case w.screen == OnAsking && w.question == aboutDeletingSeriesTasks && name == "Delete":
		return w.deletingSeries(), true
	case w.screen == OnAsking && w.question == aboutDeletingSeriesTasks:
		w.screen, w.question = OnSeriesFates, aboutDeleting
		return w, true
	case w.screen == OnAsking && w.question == aboutSeriesSave && (name == ThisOneOnlySays || name == WholeSeriesSays):
		w.screen, w.question = OnForm, aboutDeleting
		w.form.whole, w.form.answered = name == WholeSeriesSays, true
		saved := w.Save()
		if saved.screen == OnForm {
			saved.form.whole, saved.form.answered = false, false
		}
		return saved, true
	case w.screen == OnAsking && w.question == aboutSeriesSave:
		w.screen, w.question = OnForm, aboutDeleting
		return w, true
	case name == NewSeriesSays:
		return w.startingSeries(), true
	}
	return w.actingOnOccurrence(name)
}

func (w Window) seriesChipAt(p Point) (int, bool) {
	if w.screen != OnTheBoard || w.tab != Tasks {
		return 0, false
	}
	for _, placed := range w.Placed() {
		if !placed.Within.In(p) || !placed.Slot.In(p) || placed.Card.Series == nil {
			continue
		}
		row := CardRowOf(placed.Card, PartsOf(placed.Slot).Number, w.measured(), w.now, w.pages)
		if row.SeriesAt.In(p) {
			return row.Series, true
		}
	}
	return 0, false
}

// The offers a chosen occurrence and the form of one add, and what they answer to.
const (
	SkipItSays      = "Skip it"
	ItsSeriesSays   = "Its series"
	SkipThisOneSays = "Skip this one"
	StopTheSeries   = "Stop the series"
	ResumeTheSeries = "Resume the series"
)

func (w Window) occurrenceChosen() (engine.Card, bool) {
	c, found := w.snap.Card(w.chosen)
	return c, found && w.chosen != 0 && !w.chosenGoal && c.Series != nil && w.pages.Recurs()
}

func (w Window) chording(name string) Window {
	c, ok := w.occurrenceChosen()
	if !ok {
		return w
	}
	switch name {
	case "s":
		return w.asking(engine.TaskSkip, engine.Numbered(c.Number))
	case "x":
		return w.asking(engine.SeriesStop, engine.Numbered(c.Series.Series))
	case "r":
		return w.asking(engine.SeriesResume, engine.Numbered(c.Series.Series))
	case "e":
		return w.onSeries(c.Series.Series)
	}
	return w
}

func (w Window) occurrenceOffers() []string {
	if w.form.Adding || w.form.ForObjective || w.form.ForSeries || !w.pages.Recurs() {
		return nil
	}
	c, found := w.snap.Card(w.form.Number)
	if !found || c.Series == nil || engine.Finished(c) {
		return nil
	}
	if c.Series.Stopped {
		return []string{SkipThisOneSays, ResumeTheSeries}
	}
	return []string{SkipThisOneSays, StopTheSeries}
}

func (w Window) actingOnOccurrence(offer string) (Window, bool) {
	c, _ := w.occurrenceChosen()
	if w.screen == OnForm {
		c, _ = w.snap.Card(w.form.Number)
	}
	if c.Series == nil {
		return w, false
	}
	leaving := func(w Window) Window {
		w.screen, w.form = OnTheBoard, Form{}
		return w
	}
	switch offer {
	case SkipThisOneSays, SkipItSays:
		return leaving(w).asking(engine.TaskSkip, engine.Numbered(c.Number)), true
	case StopTheSeries:
		return leaving(w).asking(engine.SeriesStop, engine.Numbered(c.Series.Series)), true
	case ResumeTheSeries:
		return leaving(w).asking(engine.SeriesResume, engine.Numbered(c.Series.Series)), true
	case ItsSeriesSays:
		return w.onSeries(c.Series.Series), true
	}
	return w, false
}

// OccurrenceBox is where the form of an occurrence says its series, under its details, and whether there is room for it.
func (w Window) OccurrenceBox() (Box, bool) {
	if w.screen != OnForm || w.Tracing() || len(w.occurrenceOffers()) == 0 {
		return Box{}, false
	}
	fields := w.form.Fields()
	bottom := FirstRow(w.formPanel())
	for i, f := range fields {
		if _, sectional := sectionOf(f); !sectional {
			continue
		}
		_, value, rows := FormPlace(w.formBox(), fields, i)
		frame := Framed(value, rows)
		bottom = max(bottom, frame.Y+frame.H)
	}
	right := FormBeside(w.formPanel())
	box := Box{X: right.X + padding, Y: bottom + padding, W: right.W - 2*padding, H: 2*padding + 2*lineHigh + between + w.occurrenceFacts().High()}
	if box.Y+box.H > FormRule(w.formPanel()).Y-padding {
		return Box{}, false
	}
	return box, true
}

func (w Window) occurrenceFacts() Facts {
	c, _ := w.snap.Card(w.form.Number)
	if c.Series == nil {
		return Facts{}
	}
	s, _ := w.snap.SeriesOf(c.Series.Series)
	next := engine.ShortDays(c.Series.Then[:min(1, len(c.Series.Then))])
	if c.Series.Stopped {
		next = "none, the series is stopped"
	}
	return Facts{Lines: []Fact{
		{Name: "this one", Value: engine.ShortDays([]string{c.Series.Day})},
		{Name: "next", Value: domainOrNothing(next), Aside: "made when this one is done or skipped"},
		{Name: "so far", Value: strconv.Itoa(s.Done) + " done · " + strconv.Itoa(s.Skipped) + " skipped · " + strconv.Itoa(s.Missed) + " missed"},
	}}
}

func drawOccurrenceBeside(k Kit, w Window) {
	at, ok := w.OccurrenceBox()
	if !ok {
		return
	}
	c, _ := w.snap.Card(w.form.Number)
	card := InfoBox{}.Draw(k, at)
	in := k.inside("info")
	inside := Box{X: card.X + padding, Y: card.Y + padding, W: card.W - 2*padding, H: card.H - 2*padding}
	wide := SeriesNumberWide(c.Series.Series, small)
	Head{Said: SeriesHead}.Draw(in, Box{X: inside.X, Y: inside.Y, W: inside.W - wide - gap, H: lineHigh})
	SeriesNumber{Number: c.Series.Series, Ink: in.skin.Colour.Accent.Base}.Draw(in, Box{X: inside.X + inside.W - wide, Y: inside.Y, W: wide, H: lineHigh})
	says := Box{X: inside.X, Y: inside.Y + lineHigh + between, W: inside.W, H: lineHigh}
	say(in.gtx, in.th, in.skin.Colour.Ink.OnTint, says, fits(in.gtx, in.th, c.Series.Says, says.W, reads), reads)
	facts := w.occurrenceFacts()
	facts.Draw(in, Box{X: inside.X, Y: says.Y + lineHigh + between, W: inside.W, H: facts.High()})
}

// SeriesHead is what the form of an occurrence heads the box saying its series with.
const SeriesHead = "Series"
