package gui

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"cairn/internal/build"
	"cairn/internal/engine"
)

// Pages is what the two tabs are called, in the order they are drawn.
var Pages = []string{"TASKS", "OBJECTIVES", "TIME"}

// Heads is what the panel that is open says it is, written across its top.
func Heads(w Window) string { return HeadsOn(w, w.screen) }

// HeadsOn is what a given panel says it is, whether or not it is the one on top.
func HeadsOn(w Window, screen Screen) string {
	switch screen {
	case OnBackups:
		return "Backups"
	case OnFiltering:
		if w.filterOpen {
			return Titled(w.filters[w.filterOn].Name)
		}
		return "Filters"
	case OnLabels:
		return "Labels"
	case OnConfiguration:
		if name := w.switcher.place.Workspace; name != "" {
			return "Settings of " + name
		}
		return "Configuration"
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
			return "Put task " + w.agreeing.In.Given["number"] + " in done, ticking its steps?"
		}
		if w.question == aboutApplying {
			return "Apply what this store is waiting for?"
		}
		if w.question == aboutNewer {
			return "This store comes from a newer cairn"
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
	case "keep search":
		return "enter"
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
	case "configuration…":
		return "cmd+,"
	case "open a workspace…":
		return "cmd+shift+o"
	case "filter":
		return "shift+f"
	case "no filter":
		return "alt+backspace"
	case "take off":
		return "backspace"
	case "rename":
		return "r"
	case "new label":
		return "n"
	case "colour":
		return "c"
	case "archive", "unarchive":
		return "x"
	case "take":
		return "enter"
	case "keep":
		return "cmd+enter"
	case "edit":
		return "e"
	case "show its tasks":
		return "t"
	case "delete":
		return "cmd+backspace"
	case "quit":
		return "cmd+q"
	case "agree", "apply", "restore", "open the file", "write the file", "open settings.toml":
		return "enter"
	case "choose":
		return "enter"
	case "save", "done":
		return "cmd+enter"
	case "close", "cancel", "clear", "throw away":
		return "esc"
	case "back", "keep editing":
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
	case "cmd+backspace":
		return "⌘⌫"
	case "enter":
		return "⏎"
	}
	if letter, ok := strings.CutPrefix(key, "cmd+shift+"); ok {
		return "⌘⇧" + strings.ToUpper(letter)
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

// PanelClose is the cross at the top right of the panel open, which does what its esc does, and whether the panel open has one: a question is answered, not closed, and a list or a calendar has no heading to carry it.
func (w Window) PanelClose() (Box, bool) {
	switch w.screen {
	case OnForm, OnHelp, OnBackups, OnLabels, OnFiltering, OnConfiguration:
		return FormClose(w.panel()), true
	}
	return Box{}, false
}

func (w Window) leavingOffer() string {
	for _, offer := range w.Offers() {
		if w.shortcut(offer) == "esc" {
			return offer
		}
	}
	return "Close"
}

// OfferLook is how the panel draws its i-th offer: red when it destroys, filled with the accent when it is the verb the panel is for, plain when it leaves or does something beside.
func (w Window) OfferLook(i int, offer string) buttonLook {
	switch {
	case w.screen == OnFates && strings.HasPrefix(offer, "Delete") && !w.fates.Ready():
		return buttonFaint
	case w.screen == OnFates && strings.HasPrefix(offer, "Delete"):
		return buttonDanger
	case offer == "Delete" || offer == "Throw away" || (i == 0 && w.screen == OnAsking && w.destroys()):
		return buttonDanger
	case i == 0 && !leaving[offer]:
		return buttonOn
	}
	return buttonPlain
}

var leaving = map[string]bool{"Cancel": true, "Close": true, "Back": true, "Keep editing": true, "No filter": true, "No day": true, "Keep": true, "Start afresh": true}

func (w Window) destroys() bool {
	switch w.question {
	case aboutDeleting, aboutDeletingSeveral, aboutDeletingLabel, aboutDeletingPart, aboutDeletingParts, aboutLeaving:
		return true
	}
	return false
}

// Offers is what the bar says can be done here.
func (w Window) Offers() []string {
	switch w.screen {
	case OnPalette:
		return nil
	case OnChoosing:
		if w.pickingLabels {
			return []string{"Choose", "Close"}
		}
		if w.OneOnly() {
			return []string{"Choose", "Cancel"}
		}
		return []string{"Done", "Choose", "Cancel"}
	case OnCalendar:
		return []string{"Take", "No day", "Cancel"}
	case OnBackups:
		return []string{"Restore", "Close"}
	case OnFiltering:
		if w.filterOpen {
			return []string{"Choose", "Cancel"}
		}
		return []string{"Take off", "No filter", "Close"}
	case OnConfiguration:
		if w.configured.Settings != "" && w.switcher.place.Workspace != "" {
			return []string{OpenSettings, "Open the file", "Close"}
		}
		if w.configHere {
			return []string{"Open the file", "Close"}
		}
		return []string{"Write the file", "Close"}
	case OnLabels:
		if w.renaming {
			return []string{"Save", "Cancel"}
		}
		if w.colouring {
			return []string{"Choose", "Cancel"}
		}
		if _, ok := w.LabelCreating(); ok && w.seeking {
			return []string{"Create", "Cancel"}
		}
		if w.seeking {
			return []string{"Cancel"}
		}
		return []string{"New label", "Close"}
	case OnMenu, OnWorkspaces:
		return nil
	case OnForm:
		offers := []string{"Save", "Cancel"}
		if w.recoverable() {
			offers = append(offers, "Recover")
		}
		return offers
	case OnAsking:
		if w.question == aboutLeaving {
			return []string{"Throw away", "Keep editing"}
		}
		if w.question == aboutCairn {
			return []string{"Close"}
		}
		if w.question == aboutDraft {
			return []string{"Pick it up", "Start afresh", "Cancel"}
		}
		if w.question == aboutApplying {
			return []string{"Apply", "Quit"}
		}
		if w.question == aboutNewer {
			return []string{"Quit"}
		}
		if w.asksAboutSeveral() {
			return w.severalOffers()
		}
		return []string{w.agreeingVerb(), "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnFates:
		return []string{w.DeletesSays(), "Cancel"}
	case OnSearch:
		return []string{"Keep search", "Close"}
	}
	return w.BoardOffers()
}

// BoardOffers is what the bar under the board says, whatever is open over it.
func (w Window) BoardOffers() []string {
	offers := []string{"New", "Search", "Filter", "Archived", "Help"}
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
	if w.chosenGoal {
		return append(offers, ShowItsTasks, "Edit", "Delete")
	}
	{
		card, _ := w.snap.Card(w.chosen)
		offers = append(offers, "Edit", engine.AwayOp(card).Offered(), "Delete")
	}
	return offers
}

func (w Window) clickingList(p Point) Window {
	box := w.ChoosingBox()
	if !box.In(p) {
		return w.clickedOff(p)
	}
	if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
		w.choosing = at
		return w
	}
	for _, taken := range w.LabelsTaken() {
		if taken.Cross.In(p) {
			return w.untaking(taken.Label)
		}
	}
	if looking := ChoosingLooking(box); !box.Unsearched && looking.In(p) {
		w.filtering = w.filtering.CaretAt(into(Inside(looking), p, w.filtering.Reads(), w.measured()))
	}
	return w
}

func chosenOnACalendar(field engine.Field) bool { return field.Arg == "deadline" }

func chosenFromAList(field engine.Field) bool {
	return field.Arg == "objective" || field.Arg == "label" || field.Arg == "parent" || len(field.Choices) > 0
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
	widths := w.widthsOf(w.Offers())
	slices.Reverse(widths)
	right := 0
	for _, wide := range widths {
		right += wide
	}
	end := panel.X + panel.W - padding
	if w.formUnder() && panel == w.formPanel() {
		end = FormGrip(panel).X + inset
	}
	laid := laidAlong(widths, func(i int) Box { return PanelOfferAt(panel, i) }, end-right)
	slices.Reverse(laid)
	return laid
}

// BarOffers is where the bar under the board draws what it offers, and so where those are clicked: each as wide as what it says, as a panel's are.
func (w Window) BarOffers() []Box { return w.barAlong(w.Offers()) }

func (w Window) barAlong(offers []string) []Box {
	boxes, _ := w.barLaid(offers)
	return boxes
}

// barLineLeast is the room the bar's line keeps for what it says: the hints give their words up before it gives way, then the first of them.
const barLineLeast = 160

func (w Window) barWorded() bool {
	_, worded := w.barLaid(w.Offers())
	return worded
}

func (w Window) barLaid(offers []string) ([]Box, bool) {
	l := w.Layout()
	full, keys := make([]int, len(offers)), make([]int, len(offers))
	for i, offer := range offers {
		full[i] = w.hintWide(offer)
		keys[i] = KeyChipWide(w.BarKey(offer)) + hintSpace + hintMargin
	}
	start := l.Bar.X + barInset + KeyChipWide(w.keyboard.Spelled(Pretty("cmd+k"))) + hintSpace
	widths, worded := barWidths(full, keys, l.Bar.X+l.Bar.W-barInset-start-barLineLeast)
	return laidAlong(widths, func(i int) Box { return OfferAt(l, i) }, l.Bar.X+l.Bar.W-barInset-sumOf(widths)), worded
}

func barWidths(full, keys []int, room int) ([]int, bool) {
	if sumOf(full) <= room {
		return full, true
	}
	widths := slices.Clone(keys)
	for i := range widths {
		if sumOf(widths) <= room {
			break
		}
		widths[i] = 0
	}
	return widths, false
}

func sumOf(widths []int) int {
	total := 0
	for _, wide := range widths {
		total += wide
	}
	return total
}

const (
	barInset   = 8
	chipInset  = 4
	hintSpace  = 6
	hintMargin = 14
)

func (w Window) hintWide(offer string) int {
	return KeyChipWide(w.BarKey(offer)) + hintSpace + ends(w.measured(), offer, reads) + hintMargin
}

func (w Window) shortcut(offer string) string {
	if w.screen == OnAsking && offer == w.agreeingVerb() {
		return "enter"
	}
	if w.screen == OnLabels && (w.renaming && offer == "Save" || offer == "Create") {
		return "enter"
	}
	if w.screen == OnFates && strings.HasPrefix(offer, "Delete") {
		return "enter"
	}
	return Shortcut(offer)
}

func (w Window) agreeingVerb() string {
	switch w.question {
	case aboutDeleting, aboutDeletingLabel, aboutDeletingPart, aboutDeletingParts, aboutDeletingSeveral:
		return "Delete"
	case aboutSweeping, aboutArchivingSeveral:
		return "Archive"
	case aboutPuttingBack:
		return "Restore"
	case aboutFinishingParts:
		return "Finish"
	case aboutHiding:
		return "Replace"
	}
	return "Agree"
}

// KeyChipWide is how wide the chip a key is written on is: its figures in Go Mono, a key mark Go Mono has no glyph for a third wider than an em, the face standing in for it being wider than Go Mono and wider still on some systems, and a little room either side.
func KeyChipWide(key string) int {
	if key == "" {
		return 0
	}
	wide := 0
	for _, r := range key {
		if r < unicode.MaxLatin1 {
			wide += monoWide(string(r), small)
			continue
		}
		wide += small * 4 / 3
	}
	return wide + 2*chipInset
}

// CommandSays is what the command bar's line says at rest: the task chosen, by its number and title, or what can be typed into it.
func (w Window) CommandSays() string {
	if w.chosen != 0 && !w.chosenGoal {
		if c, ok := w.snap.Card(w.chosen); ok {
			return "#" + strconv.Itoa(c.Number) + " · " + c.Title
		}
	}
	return "Command, number or title…"
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
		key := KeyChipWide(w.OfferKey(offer))
		if key > 0 {
			key += hintSpace
		}
		widths = append(widths, max(offerWide, buttonWordWide(w.measured(), offer)+key+padding+inset))
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
		}
	}
	if w.question == aboutDeletingLabel {
		return []Said{
			{Means: "It is taken off every task and every objective carrying it."},
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
			{Means: "Replacing it puts cairn's own in its place; cancelled, it is kept until it changes."},
		}
	}
	if w.question == aboutDeletingParts {
		return []Said{
			{Means: "They go at once, whether the form is saved or not."},
		}
	}
	if w.question == aboutDeletingPart {
		return []Said{
			{Means: "It goes at once, whether the form is saved or not."},
		}
	}
	if w.question == aboutFinishingParts {
		return []Said{
			{Means: "Its steps not yet done are ticked with it."},
		}
	}
	if w.question == aboutApplying {
		return []Said{
			{Means: "Waiting: " + strings.Join(w.waiting, ", ") + "."},
			{Means: "A backup is set aside first."},
		}
	}
	if w.question == aboutNewer {
		return []Said{
			{Means: "Close this window and open cairn again: it is the newer one that can read it."},
		}
	}
	what := "task"
	if w.ChosenGoal() {
		what = "objective"
	}
	said := []Said{{Means: "This " + what + " cannot be brought back."}}
	if c, ok := w.snap.Card(w.chosen); ok && !w.ChosenGoal() {
		said = append(said, Said{Means: goingWith(c)})
	}
	return said
}

func goingWith(c engine.Card) string {
	var what []string
	switch c.SubTotal {
	case 0:
	case 1:
		what = append(what, "its 1 step")
	default:
		what = append(what, "its "+strconv.Itoa(c.SubTotal)+" steps")
	}
	what = append(what, "its labels")
	if c.Spent != "" {
		what = append(what, "its "+c.Spent+" spent")
	}
	listed := what[len(what)-1]
	if len(what) > 1 {
		listed = strings.Join(what[:len(what)-1], ", ") + " and " + listed
	}
	return strings.ToUpper(listed[:1]) + listed[1:] + " go with it."
}

// AboutSays is what About cairn says of a build: its version, the commit it was made from and when, what built it for which system, and where cairn is kept.
func AboutSays(b build.Info) []Said {
	return []Said{
		{Means: "cairn " + b.Version},
		{Means: "commit " + b.Commit + ", " + b.CommitDate},
		{Means: "built " + b.Built + " with " + b.Go + " for " + b.Platform},
		{Means: build.Home},
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
	aboutApplying
	aboutNewer
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
