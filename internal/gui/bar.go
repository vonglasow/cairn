package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"cairn/internal/build"
	"cairn/internal/engine"
)

// Pages is what the tabs are called, in the order they are drawn.
var Pages = []string{"TASKS", "OBJECTIVES", "TIME", "SERIES", "AGENTS"}

// Heads is what the panel that is open says it is, written across its top.
func Heads(w Window) string { return HeadsOn(w, w.Screen) }

// HeadsOn is what a given panel says it is, whether or not it is the one on top.
func HeadsOn(w Window, screen board.Screen) string {
	switch screen {
	case board.OnBackups:
		return "Backups"
	case board.OnFiltering:
		if w.FilterOpen {
			return Titled(w.Filters[w.FilterOn].Name)
		}
		return "Filters"
	case board.OnLabels:
		return "Labels"
	case board.OnViews:
		return ViewsSays
	case board.OnConfiguration:
		if w.ChooserOpen {
			return "Theme of " + w.hereName()
		}
		if w.IntrosOpen {
			return "Intro of " + w.hereName()
		}
		return "Settings"
	case board.OnHelp:
		return "Help"
	case board.OnSeriesFates:
		s, _ := w.Snap.SeriesOf(w.FateOf)
		return "Delete the series " + s.Title + "?"
	case board.OnAsking:
		if head, _, _, ok := w.agentQuestion(); ok {
			return head
		}
		if w.Question == board.AboutLeaving || w.Question == board.AboutLeavingSettings {
			return "Throw away what you typed?"
		}
		if w.AsksAboutSeveral() {
			return w.SeveralHead()
		}
		if w.Question == board.AboutPuttingBack {
			return "Restore backup " + strconv.Itoa(w.Backups[w.BackupOn].Number) + "?"
		}
		if w.Question == board.AboutDeletingLabel {
			return "Delete the label " + w.LabelChosen() + " everywhere?"
		}
		if w.Question == board.AboutSweeping {
			return "Archive every task in " + strings.ToUpper(w.Sweeps()) + "?"
		}
		if w.Question == board.AboutCairn {
			return "About cairn"
		}
		if w.Question == board.AboutDraft {
			return "Pick up the new task left unsaved?"
		}
		if w.Question == board.AboutHiding {
			return "Show cairn's logo on this desktop?"
		}
		if w.Question == board.AboutDeletingParts {
			return "Delete " + strconv.Itoa(len(w.Form.TakenParts())) + " steps?"
		}
		if w.Question == board.AboutDeletingPart {
			return "Delete the step “" + w.PartTitle() + "”?"
		}
		if w.Question == board.AboutFinishingParts {
			return "Put task " + w.agreeing.In.Given["number"] + " in done, ticking its steps?"
		}
		if w.Question == board.AboutApplying {
			return "Apply what this store is waiting for?"
		}
		if w.Question == board.AboutNewer {
			return "This store comes from a newer cairn"
		}
		if w.Question == board.AboutSeriesSave {
			return "Change " + w.SavedTitle() + "?"
		}
		if w.Question == board.AboutDeletingSeriesTasks {
			return w.DeletingAllSays()
		}
		return "Delete " + w.Kind() + " " + strconv.Itoa(w.Chosen) + "?"
	case board.OnForm:
		if w.Form.Adding {
			return "New " + w.FormKind()
		}
		head := w.FormKind() + " " + strconv.Itoa(w.Form.Number)
		if saved := w.SavedTitle(); saved != "" {
			head += " · " + saved
		}
		return head
	}
	return ""
}

// Shortcut is the key that does an offer without the pointer.
func Shortcut(offer string) string {
	switch strings.ToLower(offer) {
	case "new":
		return "cmd+n"
	case "save search":
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
	case "settings…":
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
	case "open":
		return "e"
	case strings.ToLower(KeepTheViewSays), strings.ToLower(SaveToViewSays):
		return "enter"
	case strings.ToLower(KeepAsNewSays):
		return "v"
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
	case strings.ToLower(AddSubSays):
		return "shift+n"
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
	case strings.ToLower(EditSeriesSays):
		return "e"
	case strings.ToLower(SkipNextSays), strings.ToLower(WholeSeriesSays), "skip":
		return "s"
	case strings.ToLower(StopSeriesSays):
		return "x"
	case strings.ToLower(ResumeSeriesSays):
		return "r"
	case strings.ToLower(DeleteSeriesSays):
		return "backspace"
	case strings.ToLower(NewSeriesSays):
		return "n"
	case strings.ToLower(OpenOccurrence):
		return "enter"
	case strings.ToLower(SkipThisOneSays):
		return "r s"
	case strings.ToLower(ItsSeriesSays):
		return "r e"
	case strings.ToLower(StopTheSeries):
		return "r x"
	case strings.ToLower(ResumeTheSeries):
		return "r r"
	case strings.ToLower(DeleteTheSeriesSays), strings.ToLower(ThisOneOnlySays):
		return "enter"
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
	case "backspace":
		return "⌫"
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
func (w Window) PanelClose() (geom.Box, bool) {
	switch w.Screen {
	case board.OnForm, board.OnHelp, board.OnBackups, board.OnLabels, board.OnFiltering, board.OnConfiguration, board.OnViews:
		return FormClose(w.panel()), true
	}
	return geom.Box{}, false
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
	case w.Screen == board.OnFates && strings.HasPrefix(offer, "Delete") && !w.Fates.Ready():
		return buttonFaint
	case w.Screen == board.OnAsking && offer == SendSays && !w.answeredAll():
		return buttonFaint
	case w.Screen == board.OnFates && strings.HasPrefix(offer, "Delete"):
		return buttonDanger
	case w.Screen == board.OnConfiguration && !strings.HasPrefix(offer, "Wear ") && !strings.HasPrefix(offer, KeepIntro):
		return buttonPlain
	case offer == "Delete" || offer == "Throw away" || offer == DeleteTheSeriesSays || offer == StopTheSeries || (i == 0 && w.Screen == board.OnAsking && w.Destroys()):
		return buttonDanger
	case i == 0 && !leaving[offer]:
		return buttonOn
	}
	return buttonPlain
}

var leaving = map[string]bool{"Cancel": true, "Close": true, "Back": true, "Keep editing": true, "No filter": true, "No day": true, "Keep": true, "Start afresh": true}

// Offers is what the bar says can be done here.
func (w Window) Offers() []string {
	switch w.Screen {
	case board.OnPalette:
		return nil
	case board.OnChoosing:
		if w.PickingLabels {
			return []string{"Done"}
		}
		if w.OneOnly() {
			return []string{"Choose", "Cancel"}
		}
		return []string{"Done", "Cancel"}
	case board.OnCalendar:
		return []string{"Take", "No day", "Cancel"}
	case board.OnRule:
		return []string{"Choose", "Cancel"}
	case board.OnBackups:
		return []string{"Restore", "Close"}
	case board.OnFiltering:
		if w.FilterOpen {
			return []string{"Choose", "Cancel"}
		}
		return []string{"Take off", "No filter", "Close"}
	case board.OnConfiguration:
		return w.settingsOffers()
	case board.OnLabels:
		if w.Renaming {
			return []string{"Save", "Cancel"}
		}
		if w.Colouring {
			return []string{"Choose", "Cancel"}
		}
		if _, ok := w.LabelCreating(); ok && w.Seeking {
			return []string{"Create", "Cancel"}
		}
		if w.Seeking {
			return []string{"Cancel"}
		}
		return []string{"New label", "Close"}
	case board.OnMenu, board.OnWorkspaces:
		return nil
	case board.OnKeeping:
		if !w.OfView {
			return []string{KeepTheViewSays, "Cancel"}
		}
		if w.ViewChanged() {
			return []string{SaveToViewSays, KeepAsNewSays, "Close"}
		}
		return []string{"Close"}
	case board.OnViews:
		if w.RenamingView {
			return []string{"Save", "Cancel"}
		}
		if len(w.Snap.Views) == 0 {
			return []string{"Close"}
		}
		return []string{"Open", "Rename", "Delete", "Close"}
	case board.OnForm:
		offers := []string{"Save", "Cancel"}
		if w.Recoverable() {
			offers = append(offers, "Recover")
		}
		return append(offers, w.occurrenceOffers()...)
	case board.OnAsking:
		if _, _, offers, ok := w.agentQuestion(); ok {
			return offers
		}
		if w.Question == board.AboutLeaving || w.Question == board.AboutLeavingSettings {
			return []string{"Throw away", "Keep editing"}
		}
		if w.Question == board.AboutCairn {
			return []string{"Close"}
		}
		if w.Question == board.AboutDraft {
			return []string{"Pick it up", "Start afresh", "Cancel"}
		}
		if w.Question == board.AboutApplying {
			return []string{"Apply", "Quit"}
		}
		if w.Question == board.AboutNewer {
			return []string{"Quit"}
		}
		if w.Question == board.AboutSeriesSave {
			return []string{ThisOneOnlySays, WholeSeriesSays, "Cancel"}
		}
		if w.AsksAboutSeveral() {
			return w.SeveralOffers()
		}
		return []string{w.AgreeingVerb(), "Cancel"}
	case board.OnHelp:
		return []string{"Close"}
	case board.OnFates:
		return []string{w.DeletesSays(), "Cancel"}
	case board.OnSeriesFates:
		return []string{DeleteTheSeriesSays, "Cancel"}
	case board.OnSearch:
		return []string{"Save search", "Close"}
	}
	return w.BoardOffers()
}

// BoardOffers is what the bar under the board says, whatever is open over it.
func (w Window) BoardOffers() []string {
	if w.Tab == board.Agents {
		return append(w.agentOffers(), "Help")
	}
	offers := []string{"New", "Search", "Filter", "Archived", "Help"}
	if !w.Looking.Empty() || len(w.Taken()) > 0 {
		offers = append(offers, "Clear")
	}
	if len(w.Taken()) > 0 {
		if w.ChosenGoal {
			return append(offers, "Delete")
		}
		return append(offers, "Archive", "Delete")
	}
	if w.Tab == board.Series {
		offers = []string{NewSeriesSays, "Search", "Help"}
		if shown, found := w.SeriesShown(); found {
			acts := []string{OpenOccurrence, "Edit", "Skip", StopSeriesSays}
			if shown.State == "stopped" {
				acts = []string{OpenOccurrence, "Edit", ResumeSeriesSays}
			} else if len(shown.Open) == 0 {
				acts = []string{"Edit", StopSeriesSays}
			}
			offers = append(acts, offers...)
		}
		if !w.Looking.Empty() {
			offers = append(offers, "Clear")
		}
		return offers
	}
	if w.Chosen == 0 {
		return append(offers, "Export", "Import")
	}
	if w.ChosenGoal {
		return append(offers, ShowItsTasks, "Edit", "Delete")
	}
	{
		card, _ := w.Snap.Card(w.Chosen)
		offers = append(offers, "Edit", board.Offered(board.AwayOp(card)), "Delete")
	}
	if c, ok := w.OccurrenceChosen(); ok {
		stop := StopTheSeries
		if c.Series.Stopped {
			stop = ResumeTheSeries
		}
		offers = append(offers, ItsSeriesSays, SkipThisOneSays, stop)
	}
	return offers
}

func (w Window) clickingList(p geom.Point) Window {
	box := w.ChoosingBox()
	if !box.In(p) {
		return w.clickedOff(p)
	}
	if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.ChoosingFrom, p); on {
		w.ChoosingOn = at
		return w.taking()
	}
	for _, taken := range w.LabelsTaken() {
		if taken.Cross.In(p) {
			return w.untaking(taken.Label)
		}
	}
	if looking := ChoosingLooking(box); !box.Unsearched && looking.In(p) {
		w.Filtering = w.Filtering.CaretAt(into(Inside(looking), p, w.Filtering.Reads(), w.measured()))
	}
	return w
}

func chosenOnACalendar(field board.Field) bool {
	return field.Arg == "deadline" || field.Arg == board.FromField().Arg
}

func chosenFromRules(field board.Field) bool { return field.Arg == board.RepeatsField().Arg }

func chosenFromAList(field board.Field) bool {
	return slices.Contains([]string{"objective", "label", "parent", "blocker", "link"}, field.Arg) || len(field.Choices) > 0
}

func (w Window) offerAt(p geom.Point) (string, bool) {
	panel := w.panel()
	alone := w.Screen == board.OnTheBoard || w.Screen == board.OnSearch
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

func (w Window) offersAlong(panel geom.Box) []geom.Box {
	widths := w.widthsOf(w.Offers())
	slices.Reverse(widths)
	right := 0
	for _, wide := range widths {
		right += wide
	}
	end := panel.X + panel.W - padding
	if w.formUnder() && panel == w.formPanel() || movable(w.Screen) && panel == w.panel() {
		end = FormGrip(panel).X + inset
	}
	laid := laidAlong(widths, func(i int) geom.Box { return PanelOfferAt(panel, i) }, end-right)
	slices.Reverse(laid)
	return laid
}

// BarOffers is where the bar under the board draws what it offers, and so where those are clicked: each as wide as what it says, as a panel's are.
func (w Window) BarOffers() []geom.Box { return w.barAlong(w.Offers()) }

func (w Window) barAlong(offers []string) []geom.Box {
	boxes, _ := w.barLaid(offers)
	return boxes
}

// barLineLeast is the room the bar's line keeps for what it says: the hints give their words up before it gives way, then the first of them.
const barLineLeast = 160

func (w Window) barWorded() bool {
	_, worded := w.barLaid(w.Offers())
	return worded
}

func (w Window) barLaid(offers []string) ([]geom.Box, bool) {
	l := w.Layout()
	full, keys := make([]int, len(offers)), make([]int, len(offers))
	for i, offer := range offers {
		full[i] = w.hintWide(offer)
		keys[i] = KeyChipWide(w.BarKey(offer)) + hintSpace + hintMargin
	}
	start := l.Bar.X + barInset + KeyChipWide(w.keyboard.Spelled(Pretty("cmd+k"))) + hintSpace
	widths, worded := barWidths(full, keys, l.Bar.X+l.Bar.W-barInset-start-barLineLeast)
	return laidAlong(widths, func(i int) geom.Box { return OfferAt(l, i) }, l.Bar.X+l.Bar.W-barInset-sumOf(widths)), worded
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
	if key := agentKeys[offer]; key != "" && w.Screen == board.OnTheBoard && w.Tab == board.Agents {
		return key
	}
	if w.Screen == board.OnConfiguration {
		return w.settingsShortcut(offer)
	}
	if w.Screen == board.OnAsking && offer == w.AgreeingVerb() {
		return "enter"
	}
	if w.Screen == board.OnLabels && (w.Renaming && offer == "Save" || offer == "Create") {
		return "enter"
	}
	if w.Screen == board.OnViews && w.RenamingView && offer == "Save" {
		return "enter"
	}
	if w.Screen == board.OnChoosing && offer == "Done" && w.ticksBySpace() {
		return "enter"
	}
	if w.Screen == board.OnFates && strings.HasPrefix(offer, "Delete") {
		return "enter"
	}
	return Shortcut(offer)
}

// KeyChipWide is how wide the chip a key is written on is: its figures in Go Mono, a key mark Go Mono has no glyph for a third wider than an em, the face standing in for it being wider than Go Mono and wider still on some systems, and a little room either side.
func KeyChipWide(key string) int {
	wide := 0
	for i, one := range capsOf(key) {
		wide += capWide(one) + min(i, 1)*capGap
	}
	return wide
}

func capWide(key string) int {
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

func laidAlong(widths []int, slot func(int) geom.Box, x int) []geom.Box {
	var boxes []geom.Box
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

// Asks is what the question says, a sentence to a line.
func (w Window) Asks() []Said {
	if _, said, _, ok := w.agentQuestion(); ok {
		return said
	}
	if w.AsksAboutSeveral() {
		return w.severalSays()
	}
	if w.Question == board.AboutLeavingSettings {
		return []Said{{Means: "The setting typed is not saved."}}
	}
	if w.Question == board.AboutLeaving {
		return []Said{
			{Means: "It is set aside: this " + w.FormKind() + "'s form brings it back."},
		}
	}
	if w.Question == board.AboutDraft {
		return []Said{
			{Means: "What was typed into a new task, and neither saved nor thrown away, was kept."},
			{Means: "Pick it up to go on with it, or start afresh to let it go."},
		}
	}
	if w.Question == board.AboutPuttingBack {
		return []Said{
			{Means: "A new backup of what the store holds now is made first."},
		}
	}
	if w.Question == board.AboutDeletingLabel {
		return []Said{
			{Means: "It is taken off every task and every objective carrying it."},
		}
	}
	if w.Question == board.AboutCairn {
		return AboutSays(build.Read())
	}
	if w.Question == board.AboutHiding {
		shows := "shows no icon"
		if w.hiding.Icon != "" {
			shows = "shows the icon “" + w.hiding.Icon + "” instead"
		}
		return []Said{
			{Means: "Its desktop entry, " + Known + ".desktop, was not written by cairn and " + shows + "."},
			{Means: "Replacing it puts cairn's own in its place; cancelled, it is kept until it changes."},
		}
	}
	if w.Question == board.AboutDeletingParts {
		return []Said{
			{Means: "They go at once, whether the form is saved or not."},
		}
	}
	if w.Question == board.AboutDeletingPart {
		return []Said{
			{Means: "It goes at once, whether the form is saved or not."},
		}
	}
	if w.Question == board.AboutFinishingParts {
		return []Said{
			{Means: "Its steps not yet done are ticked with it."},
		}
	}
	if w.Question == board.AboutApplying {
		return []Said{
			{Means: "Waiting: " + strings.Join(w.waiting, ", ") + "."},
			{Means: "A backup is set aside first."},
		}
	}
	if w.Question == board.AboutNewer {
		return []Said{
			{Means: "Close this window and open cairn again: it is the newer one that can read it."},
		}
	}
	if w.Question == board.AboutSeriesSave {
		return []Said{
			{Means: "You changed its " + w.ChangedOfTheSeries() + "."},
			{Means: "For this one only, or for this one and every one to come?"},
			{Means: "The ones already made keep what they had."},
		}
	}
	if w.Question == board.AboutDeletingSeriesTasks {
		return []Said{{Means: "Their notes and time go with them."}}
	}
	what := "task"
	if w.ChosenGoal {
		what = "objective"
	}
	said := []Said{{Means: "This " + what + " cannot be brought back."}}
	if c, ok := w.Snap.Card(w.Chosen); ok && !w.ChosenGoal {
		said = append(said, Said{Means: goingWith(c)})
	}
	return said
}

func goingWith(c board.Card) string {
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
func (w Window) QuestionBox() geom.Box {
	return AskingOver(w.width, w.height, len(w.Asks()))
}

func (w Window) sweeping() Window {
	state := w.Sweeps()
	w.Screen, w.Question = board.OnTheBoard, board.AboutDeleting
	if state == "" {
		return w
	}
	return w.asking(engine.TaskArchive, engine.Agreed(engine.Input{Given: map[string]string{"state": state}}))
}
