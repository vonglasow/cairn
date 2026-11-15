package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/series"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// The offers of a series shown, and what the bar offers on its tab.
const (
	EditSeriesSays   = "Edit the series"
	SkipNextSays     = "Skip the next"
	StopSeriesSays   = "Stop"
	ResumeSeriesSays = "Resume"
	DeleteSeriesSays = "Delete…"
	NewSeriesSays    = "New series"
	OpenOccurrence   = "Open the occurrence"
	EditTheSeries    = "Edit the series…"
)

func (w Window) walkingSeriesState() Window {
	w.SeriesIn = (w.SeriesIn + 1) % len(board.SeriesStates)
	return w
}

// SeriesRow is one line of the list of series: a head naming the active or the stopped ones and how many, or a series, and where it stands.
type SeriesRow struct {
	Head, Count string
	Series      engine.Series
	At          geom.Box
}

var seriesHeads = map[string]string{"active": "Active", "stopped": "Stopped"}

// SeriesRows lays the list of series out, a head over the active ones and one over the stopped ones.
func (w Window) SeriesRows() []SeriesRow {
	list, _ := SplitOf(w.Layout())
	all := w.SeriesListed()
	var rows []SeriesRow
	y, state := list.Y+geom.Gap, ""
	for _, s := range all {
		if s.State != state {
			state = s.State
			count := 0
			for _, other := range all {
				if other.State == state {
					count++
				}
			}
			rows = append(rows, SeriesRow{Head: seriesHeads[state], Count: strconv.Itoa(count), At: NameIn(geom.Box{X: list.X, Y: y - geom.Gap, W: list.W, H: geom.HeadHigh})})
			y += geom.HeadHigh - geom.Gap
		}
		rows = append(rows, SeriesRow{Series: s, At: geom.Box{X: list.X + geom.Gap, Y: y, W: list.W - 2*geom.Gap, H: ItemHigh}})
		y += ItemHigh
	}
	return rows
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
	Occurrence series.Occurrence
	At         geom.Box
}

// SeriesDetail is the right of the series tab: the series shown and where each of its parts stands.
type SeriesDetail struct {
	Series                              engine.Series
	Box, Heading, Badge, Sentence, Next geom.Box
	Template, SoFar, Now                geom.Box
	TemplateFacts                       Facts
	Figures                             []Figure
	NowSays                             []string
	OccurrencesHead, Foot, FootSaid     geom.Box
	OccurrencesSaid, OccurrencesAside   string
	NextSays, FootSays                  string
	Lines                               []string
	LinesAt                             geom.Box
	Occurrences                         []OccurrenceLine
	Offers                              []string
	Buttons                             []geom.Box
}

const (
	footLeast      = 80
	boxLeast       = 170
	seriesBoxHigh  = 6*lineHigh + 2*padding
	occurrenceHigh = geom.SubHeight
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
	inside := geom.Box{X: box.X + detailPad, Y: box.Y + detailPad, W: box.W - 2*detailPad, H: box.H - 2*detailPad}
	m := w.measured()
	badge := engine.StandingSaid(s)
	badgeWide := ends(m, badge, small) + 2*pillPad
	y := inside.Y
	d.Heading = geom.Box{X: inside.X, Y: y, W: max(0, inside.W-badgeWide-geom.Gap), H: buttonHigh}
	d.Badge = geom.Box{X: inside.X + inside.W - badgeWide, Y: y + (buttonHigh-pillHigh)/2, W: badgeWide, H: pillHigh}
	y += buttonHigh + between
	d.Sentence = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh
	d.NextSays = NextSays(s)
	d.Next = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh + padding

	d.Offers = SeriesOffers(s)
	foot := inside.Y + inside.H - buttonHigh
	end := inside.X + inside.W
	widths := w.widthsOf(d.Offers)
	if sumOf(widths)+(len(d.Offers)-1)*geom.Gap > inside.W {
		d.Offers = shortOffers(d.Offers)
		widths = w.tightWidths(d.Offers)
	}
	for len(d.Offers) > 1 && sumOf(widths)+(len(d.Offers)-1)*geom.Gap > inside.W {
		d.Offers, widths = d.Offers[1:], widths[1:]
	}
	for i := len(widths) - 1; i >= 0; i-- {
		end -= widths[i]
		d.Buttons = append([]geom.Box{{X: end, Y: foot, W: widths[i], H: buttonHigh}}, d.Buttons...)
		end -= geom.Gap
	}
	d.Foot = geom.Box{X: inside.X, Y: foot - between - 1, W: inside.W, H: 1}
	if room := end - inside.X; room >= footLeast {
		d.FootSays = FootSaid(s)
		d.FootSaid = geom.Box{X: inside.X, Y: foot + (buttonHigh-lineHigh)/2, W: room, H: lineHigh}
	}

	third := (inside.W - 2*padding) / 3
	if s.State == "active" && third < boxLeast {
		d.Lines = []string{"Template: " + engine.TemplateSaid(s.Template), SoFarSays(s)}
		d.LinesAt = geom.Box{X: inside.X, Y: y, W: inside.W, H: 2 * lineHigh}
		y += 2*lineHigh + padding
	}
	if s.State == "active" && third >= boxLeast && foot-y > seriesBoxHigh+lineHigh {
		d.Template = geom.Box{X: inside.X, Y: y, W: third, H: seriesBoxHigh}
		d.SoFar = geom.Box{X: inside.X + third + padding, Y: y, W: third, H: seriesBoxHigh}
		d.Now = geom.Box{X: inside.X + 2*(third+padding), Y: y, W: inside.W - 2*(third+padding), H: seriesBoxHigh}
		d.TemplateFacts = templateFacts(s.Template)
		d.Figures = []Figure{{Figure: strconv.Itoa(s.Done), Word: "done"}, {Figure: strconv.Itoa(s.Skipped), Word: "skipped"},
			{Figure: strconv.Itoa(s.Missed), Word: "missed", Late: s.Missed > 0}}
		d.NowSays = w.NowSaid(s)
		y += seriesBoxHigh + padding
	}
	d.OccurrencesSaid = "Occurrences · " + strconv.Itoa(len(s.Occurrences))
	d.OccurrencesAside = OccurrencesAside(s)
	d.OccurrencesHead = geom.Box{X: inside.X, Y: y, W: inside.W, H: lineHigh}
	y += lineHigh + inset
	for _, o := range s.Occurrences {
		if y+occurrenceHigh > d.Foot.Y-between {
			break
		}
		d.Occurrences = append(d.Occurrences, OccurrenceLine{Occurrence: o, At: geom.Box{X: inside.X, Y: y, W: inside.W, H: occurrenceHigh}})
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

// FootSaid is what the foot of the detail says of a series beside its offers: who stopped it and when, or its number — written after the loop the foot draws — the day it was made and how many occurrences it made.
func FootSaid(s engine.Series) string {
	if s.State == "stopped" {
		said := "stopped"
		for _, a := range s.Acts {
			if a.Did == "stopped" && a.By != "" {
				said = "stopped by " + a.By
			}
		}
		if day, err := calendar.ParseDeadline(s.Stopped); err == nil && !day.IsZero() {
			said += " · " + day.Format("2 Jan")
		}
		return said
	}
	made := ""
	if day, err := calendar.ParseDeadline(s.From); err == nil && !day.IsZero() {
		made = " · made " + day.Format("2 Jan")
	}
	return strconv.Itoa(s.Number) + made + " · " + strconv.Itoa(len(s.Occurrences)) + " occurrences"
}

// OccurrencesAside is what the head of the occurrences says at its right: that every one is kept when all are archived, else that archived ones are there and how one is opened.
func OccurrencesAside(s engine.Series) string {
	archived := 0
	for _, o := range s.Occurrences {
		if o.Archived || o.Task == 0 {
			archived++
		}
	}
	if len(s.Occurrences) > 0 && archived == len(s.Occurrences) {
		return "all archived, all kept"
	}
	return "archived ones included · ⏎ opens the task"
}

func templateFacts(t series.Template) Facts {
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
	w.SeriesOn = all[min(len(all)-1, max(0, at+by))].Number
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
	if _, found := w.Snap.Card(n); !found {
		return w
	}
	w.Chosen, w.ChosenGoal = n, false
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
	w.SeriesOn = s.Number
	switch offer {
	case EditSeriesSays:
		w.Form, w.Screen, w.formMoved = board.AsOpened(board.Form{Number: s.Number, ForSeries: true, Fields: board.SeriesFields(s)}), board.OnForm, geom.Point{}
		return w.beside()
	case SkipNextSays:
		if len(s.Open) == 0 {
			return w
		}
		return w.asking(engine.TaskSkip, board.Numbered(s.Open[0]))
	case StopSeriesSays:
		return w.asking(engine.SeriesStop, board.Numbered(s.Number))
	case ResumeSeriesSays:
		return w.asking(engine.SeriesResume, board.Numbered(s.Number))
	case DeleteSeriesSays:
		w.Screen, w.FateOf, w.SeriesFate = board.OnSeriesFates, s.Number, 0
		return w
	}
	return w
}

func (w Window) onSeries(n int) Window {
	w = w.onPage(board.Series)
	w.SeriesOn = n
	return w
}

func (w Window) startingSeries() Window {
	fields := withRepeats(board.AddingFields(board.StateAt(1)))
	w.Form, w.Screen, w.formMoved = board.AsOpened(board.Form{Adding: true, Fields: fields}), board.OnForm, geom.Point{}
	for i, f := range fields {
		if f.Arg == board.RepeatsField().Arg {
			w.Form = w.Form.Onto(i)
		}
	}
	return w.beside()
}

func (w Window) clickingSeries(p geom.Point) (Window, bool) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Series {
		return w, false
	}
	list, detail := SplitOf(w.Layout())
	if page := (geom.Box{X: list.X, Y: list.Y, W: detail.X + detail.W - list.X, H: list.H}); !page.In(p) {
		return w, false
	}
	for _, row := range w.SeriesRows() {
		if row.Head == "" && row.At.In(p) {
			w.SeriesOn = row.Series.Number
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

func (w Window) seriesSlotAt(p geom.Point) geom.Box {
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
	return geom.Box{}
}

// The answers to the question saving an occurrence asks.
const (
	ThisOneOnlySays = "This one only"
	WholeSeriesSays = "The whole series"
)

func (w Window) choosingOnSeries(name string) (Window, bool) {
	switch {
	case w.Screen == board.OnSeriesFates && name == DeleteTheSeriesSays:
		return w.confirmingSeriesFate(), true
	case w.Screen == board.OnSeriesFates:
		w.Screen = board.OnTheBoard
		return w, true
	case w.Screen == board.OnAsking && w.Question == board.AboutDeletingSeriesTasks && name == "Delete":
		return w.deletingSeries(), true
	case w.Screen == board.OnAsking && w.Question == board.AboutDeletingSeriesTasks:
		w.Screen, w.Question = board.OnSeriesFates, board.AboutDeleting
		return w, true
	case w.Screen == board.OnAsking && w.Question == board.AboutSeriesSave && (name == ThisOneOnlySays || name == WholeSeriesSays):
		w.Screen, w.Question = board.OnForm, board.AboutDeleting
		w.Form.Whole, w.Form.Answered = name == WholeSeriesSays, true
		saved := w.Save()
		if saved.Screen == board.OnForm {
			saved.Form.Whole, saved.Form.Answered = false, false
		}
		return saved, true
	case w.Screen == board.OnAsking && w.Question == board.AboutSeriesSave:
		w.Screen, w.Question = board.OnForm, board.AboutDeleting
		return w, true
	case name == NewSeriesSays:
		return w.startingSeries(), true
	case w.Screen == board.OnTheBoard && w.Tab == board.Series && name == OpenOccurrence:
		return w.openingOccurrence(), true
	case w.Screen == board.OnTheBoard && w.Tab == board.Series && slices.Contains([]string{"Edit", "Skip", StopSeriesSays, ResumeSeriesSays}, name):
		return w.seriesActing(name), true
	case w.Screen == board.OnForm && name == EditTheSeries:
		if c, found := w.Snap.Card(w.Form.Number); found && c.Series != nil {
			w.Screen, w.Form = board.OnTheBoard, board.Form{}
			return w.onSeries(c.Series.Series), true
		}
	}
	return w.actingOnOccurrence(name)
}

func (w Window) seriesChipAt(p geom.Point) (int, bool) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Tasks {
		return 0, false
	}
	for _, placed := range w.Placed() {
		if !placed.Within.In(p) || !placed.Slot.In(p) || placed.Card.Series == nil {
			continue
		}
		row := CardRowOf(placed.Card, PartsOf(placed.Slot).Number, w.measured(), w.Now, w.Pages)
		if row.SeriesAt.In(p) {
			return row.Series, true
		}
	}
	return 0, false
}

// The offers a chosen occurrence and the form of one add, and what they answer to.
const (
	ItsSeriesSays   = "Open the series"
	SkipThisOneSays = "Skip this one"
	StopTheSeries   = "Stop the series"
	ResumeTheSeries = "Resume the series"
)

func (w Window) chording(name string) Window {
	c, ok := w.OccurrenceChosen()
	if !ok {
		return w
	}
	switch name {
	case "s":
		return w.asking(engine.TaskSkip, board.Numbered(c.Number))
	case "x":
		return w.asking(engine.SeriesStop, board.Numbered(c.Series.Series))
	case "r":
		return w.asking(engine.SeriesResume, board.Numbered(c.Series.Series))
	case "e":
		return w.onSeries(c.Series.Series)
	}
	return w
}

func (w Window) occurrenceOffers() []string {
	if w.Form.Adding || w.Form.ForObjective || w.Form.ForSeries || !w.Pages.Recurs() {
		return nil
	}
	c, found := w.Snap.Card(w.Form.Number)
	if !found || c.Series == nil || board.Finished(c) {
		return nil
	}
	if c.Series.Stopped {
		return []string{SkipThisOneSays, ResumeTheSeries}
	}
	return []string{SkipThisOneSays, StopTheSeries}
}

func (w Window) actingOnOccurrence(offer string) (Window, bool) {
	c, _ := w.OccurrenceChosen()
	if w.Screen == board.OnForm {
		c, _ = w.Snap.Card(w.Form.Number)
	}
	if c.Series == nil {
		return w, false
	}
	leaving := func(w Window) Window {
		w.Screen, w.Form = board.OnTheBoard, board.Form{}
		return w
	}
	switch offer {
	case SkipThisOneSays:
		return leaving(w).asking(engine.TaskSkip, board.Numbered(c.Number)), true
	case StopTheSeries:
		return leaving(w).asking(engine.SeriesStop, board.Numbered(c.Series.Series)), true
	case ResumeTheSeries:
		return leaving(w).asking(engine.SeriesResume, board.Numbered(c.Series.Series)), true
	case ItsSeriesSays:
		return w.onSeries(c.Series.Series), true
	}
	return w, false
}

// OccurrenceBox is where the form of an occurrence says its series, under its details, and whether there is room for it.
func (w Window) OccurrenceBox() (geom.Box, bool) {
	if w.Screen != board.OnForm || w.Tracing() || len(w.occurrenceOffers()) == 0 {
		return geom.Box{}, false
	}
	fields := w.Form.Fields
	bottom := FirstRow(w.formPanel())
	for i, f := range fields {
		if _, sectional := board.SectionOf(f); !sectional {
			continue
		}
		_, value, rows := FormPlace(w.formBox(), fields, i)
		frame := Framed(value, rows)
		bottom = max(bottom, frame.Y+frame.H)
	}
	right := FormBeside(w.formPanel())
	box := geom.Box{X: right.X + padding, Y: bottom + padding, W: right.W - 2*padding, H: 2*padding + occurrenceHeadHigh + 2*between + lineHigh + w.occurrenceFacts().High()}
	if box.Y+box.H > FormRule(w.formPanel()).Y-padding {
		return geom.Box{}, false
	}
	return box, true
}

const occurrenceHeadHigh = 24

// OccurrenceEdit is the button at the head of the box saying an occurrence's series, which opens the series in its tab.
func (w Window) OccurrenceEdit() (geom.Box, bool) {
	box, ok := w.OccurrenceBox()
	if !ok {
		return geom.Box{}, false
	}
	wide := buttonWordWide(w.measured(), EditTheSeries) + KeyChipWide("r e") + hintSpace + 2*padding
	return geom.Box{X: box.X + box.W - padding - wide, Y: box.Y + padding, W: wide, H: occurrenceHeadHigh}, true
}

func (w Window) occurrenceFacts() Facts {
	c, _ := w.Snap.Card(w.Form.Number)
	if c.Series == nil {
		return Facts{}
	}
	s, _ := w.Snap.SeriesOf(c.Series.Series)
	rule, _ := series.ParseRule(c.Series.Rule)
	anchored := "On the calendar"
	if rule.AfterDone() {
		anchored = "After the last is done"
	}
	next := series.ShortDays(c.Series.Then[:min(1, len(c.Series.Then))])
	if next != "" && rule.At() != "" {
		next += ", " + rule.At()
	}
	if c.Series.Stopped {
		next = "none, the series is stopped"
	}
	since := domainOrNothing(series.ShortDays([]string{s.From})) + " · ends " + endsSaid(rule.Parts())
	lines := []Fact{
		{Name: "anchored", Value: anchored},
		{Name: "next", Value: domainOrNothing(next)},
		{Name: "since", Value: since},
		{Name: "so far", Value: strconv.Itoa(s.Done) + " done · " + strconv.Itoa(s.Skipped) + " skipped"},
	}
	if missed := board.MissedDays(s); len(missed) > 0 {
		lines = append(lines, Fact{Name: "missed", Value: strconv.Itoa(len(missed)) + " · " + strings.Join(missed, ", "), Aside: "nobody was there; not made up", Late: true})
	}
	return Facts{Lines: lines}
}

func endsSaid(p series.RuleParts) string {
	switch {
	case p.Times == 1:
		return "after once"
	case p.Times > 1:
		return "after " + strconv.Itoa(p.Times) + " times"
	case !p.Until.IsZero():
		return "on " + series.ShortDay(p.Until)
	}
	return "never"
}

// OccurrenceHead is how the form of an occurrence is headed: its title behind the loop, its number, and the series it is one of with the day it is for.
type OccurrenceHead struct{ Title, Number, Under string }

// OccurrenceHeading is the heading of the form open on an occurrence, and whether the form is open on one.
func (w Window) OccurrenceHeading() (OccurrenceHead, bool) {
	if !formShown(w.Screen) || w.Form.Adding || w.Form.ForSeries || w.Form.ForObjective || w.Form.Number == 0 || !w.Pages.Recurs() {
		return OccurrenceHead{}, false
	}
	c, found := w.Snap.Card(w.Form.Number)
	if !found || c.Series == nil {
		return OccurrenceHead{}, false
	}
	return OccurrenceHead{Title: c.Title, Number: "#" + strconv.Itoa(c.Number),
		Under: "one of the series " + c.Series.Often + " · this one is " + series.ShortDays([]string{c.Series.Day})}, true
}

func drawOccurrenceBeside(k Kit, w Window) {
	at, ok := w.OccurrenceBox()
	if !ok {
		return
	}
	c, _ := w.Snap.Card(w.Form.Number)
	card := InfoBox{}.Draw(k, at)
	in := k.inside("info")
	inside := geom.Box{X: card.X + padding, Y: card.Y + padding, W: card.W - 2*padding, H: card.H - 2*padding}
	edit, _ := w.OccurrenceEdit()
	Head{Said: SeriesHead}.Draw(in, geom.Box{X: inside.X, Y: inside.Y + (occurrenceHeadHigh-lineHigh)/2, W: edit.X - inside.X - geom.Gap, H: lineHigh})
	Button{Said: EditTheSeries, Key: "r e", Keyed: true}.Draw(in, edit)
	says := geom.Box{X: inside.X, Y: inside.Y + occurrenceHeadHigh + between, W: inside.W, H: lineHigh}
	LoopMark{Ink: in.skin.Colour.Ink.OnTint}.Draw(in, geom.Box{X: says.X, Y: says.Y + (lineHigh-loopSide)/2, W: loopSide, H: loopSide})
	often := geom.Box{X: says.X + loopSide + inset, Y: says.Y, W: says.W - loopSide - inset, H: lineHigh}
	say(in.gtx, in.th, in.skin.Colour.Ink.OnTint, often, fits(in.gtx, in.th, c.Series.Says, often.W, reads), reads)
	facts := w.occurrenceFacts()
	facts.Draw(in, geom.Box{X: inside.X, Y: says.Y + lineHigh + between, W: inside.W, H: facts.High()})
}

// SeriesHead is what the form of an occurrence heads the box saying its series with.
const SeriesHead = "Series"
