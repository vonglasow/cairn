package gui

import (
	"strconv"
	"strings"

	"cairn/internal/build"
	"cairn/internal/engine"
)

// Pages is what the two tabs are called, in the order they are drawn.
var Pages = []string{"TASKS", "OBJECTIVES", "TIME"}

func (w Window) formNamed() string {
	if w.form.Adding {
		return "the new " + w.formKind()
	}
	return w.formKind() + " " + strconv.Itoa(w.form.Number)
}

// Heads is what the panel that is open says it is, written across its top.
func Heads(w Window) string { return HeadsOn(w, w.screen) }

// HeadsOn is what a given panel says it is, whether or not it is the one on top.
func HeadsOn(w Window, screen Screen) string {
	switch screen {
	case OnChoosing:
		what := "objectives for"
		if w.choosingFor == "label" {
			what = "labels for"
		}
		if w.OneOnly() {
			what = "a " + w.form.Fields()[w.form.On()].Name + " for"
		}
		return what + " " + w.formNamed()
	case OnCalendar:
		return "a deadline for " + w.formNamed() + " — " + w.day.Format("January 2006")
	case OnBackups:
		return "Backups — enter restores one"
	case OnLabels:
		if w.renaming {
			return "Rename " + w.labelChosen() + " to " + w.renamed.Reads()
		}
		return "Labels — r renames one, d takes it off everything"
	case OnHelp:
		return "What everything does"
	case OnAsking:
		if w.question == aboutLeaving {
			return "Throw away what you typed?"
		}
		if w.asksAboutSeveral() {
			return w.severalHead()
		}
		if w.question == aboutPuttingBack {
			return "Restore backup " + strconv.Itoa(w.backups[w.backupOn].Number) + "?"
		}
		if w.question == aboutDeletingLabel {
			return "Delete the label " + w.labelChosen() + " everywhere?"
		}
		if w.question == aboutSweeping {
			return "Archive every task in " + strings.ToUpper(w.sweeps()) + "?"
		}
		if w.question == aboutCairn {
			return "About cairn"
		}
		if w.question == aboutDraft {
			return "Pick up the new task left unsaved?"
		}
		if w.question == aboutHiding {
			return "Show cairn's logo on this desktop?"
		}
		if w.question == aboutDeletingParts {
			return "Delete " + strconv.Itoa(len(w.form.takenParts())) + " steps?"
		}
		if w.question == aboutDeletingPart {
			return "Delete the step “" + w.partTitle() + "”?"
		}
		if w.question == aboutFinishingParts {
			return "Put task " + w.finishingIn.Given["number"] + " in done, ticking its steps?"
		}
		return "Delete " + w.kind() + " " + strconv.Itoa(w.chosen) + "?"
	case OnForm:
		if w.form.Adding {
			return "New " + w.formKind()
		}
		head := w.formKind() + " " + strconv.Itoa(w.form.Number)
		if saved := w.savedTitle(); saved != "" {
			head += " · " + saved
		}
		return head
	}
	return ""
}

func (w Window) kind() string {
	if w.chosenGoal {
		return "objective"
	}
	return "task"
}

func (w Window) formKind() string {
	if w.form.ForObjective {
		return "objective"
	}
	return "task"
}

// Shortcut is the key that does an offer without the pointer.
func Shortcut(offer string) string {
	switch strings.ToLower(offer) {
	case "new":
		return "cmd+n"
	case "search":
		return "/"
	case "help":
		return "?"
	case "archived":
		return "shift+a"
	case "export":
		return "cmd+e"
	case "import":
		return "cmd+o"
	case "new backup":
		return "cmd+b"
	case "backups":
		return "shift+b"
	case "labels":
		return "shift+l"
	case "rename":
		return "r"
	case "archive", "unarchive":
		return "x"
	case "take":
		return "enter"
	case "keep":
		return "cmd+enter"
	case "edit":
		return "e"
	case "delete":
		return "d"
	case "agree":
		return "enter"
	case "choose":
		return "enter"
	case "save":
		return "cmd+enter"
	case "close", "cancel", "clear", "throw away":
		return "esc"
	case "back", "all possible":
		return "enter"
	case "skip flagged":
		return "s"
	case "recover":
		return "cmd+z"
	case "no day":
		return "backspace"
	}
	return ""
}

// Pretty is how the bar writes a key, which is not how a key arrives.
func Pretty(key string) string {
	switch key {
	case "cmd+enter":
		return "⌘⏎"
	case "enter":
		return "⏎"
	}
	if letter, ok := strings.CutPrefix(key, "cmd+"); ok {
		return "⌘" + strings.ToUpper(letter)
	}
	if letter, ok := strings.CutPrefix(key, "shift+"); ok {
		return "⇧" + strings.ToUpper(letter)
	}
	return key
}

// Said is one line of a panel that only says something: a key, or nothing, and what it means.
type Said struct {
	Key, Means string
}

// Helping is what everything does, as a person reads it.
func (w Window) Helping() []Said {
	lines := helping()
	for i, line := range lines {
		lines[i] = Said{Key: w.keyboard.Spelled(line.Key), Means: w.keyboard.Spelled(line.Means)}
	}
	return lines
}

func helping() []Said {
	return []Said{
		{Key: "click", Means: "choose a card"},
		{Key: "double-click", Means: "open it, and see everything it carries"},
		{Key: "drag", Means: "move it — the column it lands in is its new state"},
		{Key: "drag into a lane", Means: "blocked, watching or paused"},
		{Key: "drag a gutter", Means: "widen the column or lane on one side of it, narrow the other — until the window closes"},
		{Key: "right click", Means: "what can be done to what is under it — the board's offers where there is nothing"},
		{Key: "wheel", Means: "scroll what is under the pointer — a column, a lane, the objectives, a trace"},
		{Key: "wheel on a trace", Means: "scroll it — the bar beside it can be pulled too"},
		{Key: "on an open task", Means: "every text can be held: drag across it, double-click a word, triple-click a paragraph — ⌘A holds it all, ⌘C copies"},
		{Key: "drag a form", Means: "move it, from anywhere in it that holds no text — it keeps that place until it closes"},
		{Key: "^D ^U on an open task", Means: "walk its description half a box at a time — the wheel and the bar do too"},
		{Key: "e", Means: "edit what is chosen, on the board or on an open task"},
		{Key: "f", Means: "focus or unfocus what is chosen — several in focus at once is fine"},
		{Key: "o", Means: "open a new task linked to what is chosen — in todo, the two seen from either side"},
		{Key: "d", Means: "delete what is chosen — it asks first"},
		{Key: "x", Means: "archive or unarchive what is chosen"},
		{Key: "⇧X", Means: "archive every task in that card's column — it asks first"},
		{Key: "←→↑↓", Means: "walk the board, the band included"},
		{Key: "⇧←→↑↓", Means: "move what is chosen — sideways changes its column"},
		{Key: "⏎ on the board", Means: "open what is chosen"},
		{Key: "⏎ on a question", Means: "agree to it"},
		{Key: "⏎ in a form", Means: "a new line in the description, the next field anywhere else — it never saves"},
		{Key: "↑↓ in the description", Means: "walk its lines — tab leaves it, and the wheel and the bar scroll it"},
		{Key: "⌘⏎ or ⌘S in a form", Means: "save now and close — a form being changed also saves itself once the typing pauses, a new one only then"},
		{Key: "esc esc in a form", Means: "throw it away — set aside: ⌘Z on the same form brings it back"},
		{Key: "⌘A ⌘C ⌘X ⌘V", Means: "hold all of a field, copy, cut, paste"},
		{Key: "⇧←→ in a field", Means: "hold a run of text — dragging across it does the same"},
		{Key: "⌘←→ or home end", Means: "the start or the end of the line — ⌥←→ walks a word at a time, and ⇧ holds as either goes"},
		{Key: "⌦ ⌘⌫ ⌥⌫", Means: "take out what is after the caret, back to the start of the line, back a word"},
		{Key: "⇧↑↓ in the description", Means: "hold whole lines of it"},
		{Key: "⏎ in a list", Means: "take the row · ⌘⏎ keeps the list · j and k walk it"},
		{Key: "⏎ on a deadline", Means: "a calendar — the arrows walk days and weeks, ⇧←→ months, t today, ⌫ no day"},
		{Key: "⇧A", Means: "what was archived, on the board or off it — it reads faintly"},
		{Key: "⌘N", Means: "new"},
		{Key: "⌘B", Means: "a new backup of the store — it says where it lies"},
		{Key: "⇧B", Means: "the backups — enter restores one, after asking, a new backup made first"},
		{Key: "⇧L", Means: "every label, with what carries it — r renames one everywhere, d takes it off everything, after asking"},
		{Key: "⌘E", Means: "export every task to a file chosen in the system's dialog"},
		{Key: "⌘O", Means: "import the tasks of a file chosen there, as new tasks"},
		{Key: "/ or ⌘F", Means: "search — it reaches the description and the reason"},
		{Key: "⏎ in a search", Means: "keep what it narrowed and go back to the board"},
		{Key: "tab", Means: "the next page — shift+tab the one before"},
		{Key: "←→ on TIME", Means: "the week before or after — t this week"},
		{Key: "tab, ↑↓ in a form", Means: "walk its fields — inside the description ↑↓ walk its lines"},
		{Key: "←→", Means: "change a field that has a closed set"},
		{Key: "esc", Means: "back out of whatever is open"},
		{Key: "?", Means: "this help"},
	}
}

// Offers is what the bar says can be done here.
func (w Window) Offers() []string {
	switch w.screen {
	case OnChoosing:
		return []string{"Take", "Keep", "Cancel"}
	case OnCalendar:
		return []string{"Take", "No day", "Cancel"}
	case OnBackups:
		return []string{"Cancel"}
	case OnLabels:
		if w.renaming {
			return []string{"Take", "Cancel"}
		}
		return []string{"Rename", "Delete", "Cancel"}
	case OnMenu:
		return nil
	case OnForm:
		offers := []string{"Save", "Cancel"}
		if on := w.form.Fields()[w.form.On()]; chosenFromAList(on) || chosenOnACalendar(on) {
			offers = []string{"Choose", "Save", "Cancel"}
		}
		if w.recoverable() {
			offers = append(offers, "Recover")
		}
		return offers
	case OnAsking:
		if w.question == aboutLeaving {
			return []string{"Throw away", "Back"}
		}
		if w.question == aboutCairn {
			return []string{"Close"}
		}
		if w.question == aboutDraft {
			return []string{"Pick it up", "Start afresh", "Cancel"}
		}
		if w.asksAboutSeveral() {
			return w.severalOffers()
		}
		return []string{"Agree", "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnSearch:
		return []string{"Close"}
	}
	return w.BoardOffers()
}

// BoardOffers is what the bar under the board says, whatever is open over it.
func (w Window) BoardOffers() []string {
	offers := []string{"New", "Search", "Archived", "Help"}
	if !w.looking.Empty() || len(w.Taken()) > 0 {
		offers = append(offers, "Clear")
	}
	if len(w.Taken()) > 0 {
		if w.chosenGoal {
			return append(offers, "Delete")
		}
		return append(offers, "Archive", "Delete")
	}
	if w.chosen == 0 {
		return append(offers, "Export", "Import")
	}
	{
		card, _ := w.snap.Card(w.chosen)
		offers = append(offers, "Edit", engine.AwayOp(card).Offered(), "Delete")
	}
	return offers
}

func (w Window) clickingList(p Point) Window {
	box := w.ChoosingBox()
	if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
		w.choosing = at
		return w
	}
	if looking := ChoosingLooking(box); looking.In(p) {
		w.filtering = w.filtering.CaretAt(into(Inside(looking), p, w.filtering.Reads(), w.measured()))
	}
	return w
}

func chosenOnACalendar(field engine.Field) bool { return field.Arg == "deadline" }

func chosenFromAList(field engine.Field) bool {
	return field.Arg == "objective" || field.Arg == "label" || len(field.Choices) > 0
}

func (w Window) offerAt(p Point) (string, bool) {
	panel := w.panel()
	alone := w.screen == OnTheBoard || w.screen == OnSearch
	along, bar := w.offersAlong(panel), w.BarOffers()
	for i, offer := range w.Offers() {
		if alone && bar[i].In(p) {
			return offer, true
		}
		if !alone && along[i].In(p) {
			return offer, true
		}
	}
	return "", false
}

func (w Window) offersAlong(panel Box) []Box {
	return laidAlong(w.widthsOf(w.Offers()), func(i int) Box { return PanelOfferAt(panel, i) }, panel.X+padding)
}

// BarOffers is where the bar under the board draws what it offers, and so where those are clicked: each as wide as what it says, as a panel's are.
func (w Window) BarOffers() []Box { return w.barAlong(w.Offers()) }

func (w Window) barAlong(offers []string) []Box {
	l := w.Layout()
	return laidAlong(w.widthsOf(offers), func(i int) Box { return OfferAt(l, i) }, l.Bar.X)
}

func laidAlong(widths []int, slot func(int) Box, x int) []Box {
	var boxes []Box
	for i, wide := range widths {
		box := slot(i)
		box.X, box.W = x, wide
		boxes = append(boxes, box)
		x += wide
	}
	return boxes
}

func (w Window) offerWidths() []int { return w.widthsOf(w.Offers()) }

func (w Window) widthsOf(offers []string) []int {
	var widths []int
	for _, offer := range offers {
		widths = append(widths, max(offerWide, ends(w.measured(), w.Said(offer), reads)+padding))
	}
	return widths
}

func (w Window) savedTitle() string {
	if w.form.ForObjective {
		g, _ := w.snap.Goal(w.form.Number)
		return g.Title
	}
	c, _ := w.snap.Card(w.form.Number)
	return c.Title
}

// Asks is what the question says, a sentence to a line.
func (w Window) Asks() []Said {
	if w.asksAboutSeveral() {
		return w.severalSays()
	}
	if w.question == aboutLeaving {
		return []Said{
			{Means: "It is set aside: this " + w.formKind() + "'s form brings it back."},
		}
	}
	if w.question == aboutDraft {
		return []Said{
			{Means: "What was typed into a new task, and neither saved nor thrown away, was kept."},
			{Means: "Pick it up to go on with it, or start afresh to let it go."},
		}
	}
	if w.question == aboutPuttingBack {
		return []Said{
			{Means: "A new backup of what the store holds now is made first."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	if w.question == aboutDeletingLabel {
		return []Said{
			{Means: "It is taken off every task and every objective carrying it."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	if w.question == aboutCairn {
		return AboutSays(build.Read())
	}
	if w.question == aboutHiding {
		shows := "shows no icon"
		if w.hiding.Icon != "" {
			shows = "shows the icon “" + w.hiding.Icon + "” instead"
		}
		return []Said{
			{Means: "Its desktop entry, " + Known + ".desktop, was not written by cairn and " + shows + "."},
			{Means: "Enter replaces it with cairn's own, or Cancel keeps it until it changes."},
		}
	}
	if w.question == aboutDeletingParts {
		return []Said{
			{Means: "They go at once, whether the form is saved or not."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	if w.question == aboutDeletingPart {
		return []Said{
			{Means: "It goes at once, whether the form is saved or not."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	if w.question == aboutFinishingParts {
		return []Said{
			{Means: "Its steps not yet done are ticked with it."},
			{Means: "Enter to agree, or Cancel."},
		}
	}
	what := "task"
	if w.ChosenGoal() {
		what = "objective"
	}
	return []Said{
		{Means: "This " + what + " cannot be brought back."},
		{Means: "Delete again to agree, or Cancel."},
	}
}

// AboutSays is what About cairn says of a build: its version, the commit it was made from and when, and what built it for which system.
func AboutSays(b build.Info) []Said {
	return []Said{
		{Means: "cairn " + b.Version},
		{Means: "commit " + b.Commit + ", " + b.CommitDate},
		{Means: "built " + b.Built + " with " + b.Go + " for " + b.Platform},
	}
}

// QuestionBox is the box a question is asked in: as wide as its sentences want and as tall as they are, not the panel a form is drawn in.
func (w Window) QuestionBox() Box {
	return AskingOver(w.width, w.height, len(w.Asks()))
}

// Question is what the one question screen is asking about.
type Question int

const (
	aboutDeleting Question = iota
	aboutLeaving
	aboutSweeping
	aboutDeletingSeveral
	aboutArchivingSeveral
	aboutPuttingBack
	aboutDeletingLabel
	aboutFinishingParts
	aboutCairn
	aboutDraft
	aboutDeletingPart
	aboutHiding
	aboutDeletingParts
)

func (w Window) sweeps() string {
	if w.chosen == 0 || w.chosenGoal {
		return ""
	}
	state := w.stateOf(w.chosen)
	if engine.Stopped(state) {
		return ""
	}
	return state
}

func (w Window) sweeping() Window {
	state := w.sweeps()
	w.screen, w.question = OnTheBoard, aboutDeleting
	if state == "" {
		return w
	}
	return w.asking(engine.TaskArchive, engine.Agreed(engine.Input{Given: map[string]string{"state": state}}))
}
