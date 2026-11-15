package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"slices"
	"strconv"
	"strings"
)

// Note is what the window last said that is not something gone wrong: where a backup lies, which one was put back.
func (b Board) Note() string { return b.Telling }

// Destroys says whether what the question agrees to loses work.
func (b Board) Destroys() bool {
	switch b.Question {
	case AboutDeleting, AboutDeletingSeveral, AboutDeletingLabel, AboutDeletingPart, AboutDeletingParts, AboutLeaving, AboutDeletingSeriesTasks:
		return true
	}
	return false
}

// AgreeingVerb is the word the question's yes says.
func (b Board) AgreeingVerb() string {
	switch b.Question {
	case AboutDeleting, AboutDeletingLabel, AboutDeletingPart, AboutDeletingParts, AboutDeletingSeveral, AboutDeletingSeriesTasks:
		return "Delete"
	case AboutSweeping, AboutArchivingSeveral:
		return "Archive"
	case AboutPuttingBack:
		return "Restore"
	case AboutFinishingParts:
		return "Finish"
	case AboutHiding:
		return "Replace"
	}
	return "Agree"
}

// Showing is the work narrowed to what the filters keep and a search reaches, which is what everything draws from.
func (b Board) Showing() Snapshot {
	if b.Looking.Empty() && !b.Filters.Narrows() && b.Viewing == "" {
		return b.Snap
	}
	cards := b.Snap.Viewed(b.Viewing, b.Snap.Cards)
	if kept, err := b.Filters.Kept(cards); err == nil {
		cards = kept
	}
	found := Snapshot{Focus: b.Snap.Focus, Version: b.Snap.Version}
	found.Cards = Found(cards, b.Looking.Reads(), b.Archived)
	found.Goals = FoundGoals(b.Filters.KeptGoals(b.Snap.Goals), b.Looking.Reads())
	return found
}

// LaneSays is what a lane of the band writes above itself: what it is, and how many it holds.
func (b Board) LaneSays(kind string) string {
	lane := slices.Index(Stalled, kind)
	if lane < 0 {
		return kind + "  0"
	}
	return kind + "  " + strconv.Itoa(len(b.InLane(lane)))
}

// ChoosingFromRow is the first row of the open list in sight, and ChoosingRoom how many of them are in sight at once.
func (b Board) ChoosingFromRow() int { return b.ChoosingFrom }

// ValueColumns is the values of the open filter laid out in columns, an objective's sub-objectives in the column after it.
func (b Board) ValueColumns() [][]int {
	f := b.Filters[b.FilterOn]
	if f.Arg != "objective" {
		all := make([]int, 0, len(f.Values))
		for i := 1; i < len(f.Values); i++ {
			all = append(all, i)
		}
		return [][]int{all}
	}
	under := func(parent int) []int {
		var out []int
		for i, value := range f.Values {
			if g, found := b.GoalOfValue(value); found && (g.Parent == parent || parent == 0 && !slices.Contains(f.Values, strconv.Itoa(g.Parent))) {
				out = append(out, i)
			}
		}
		return out
	}
	columns := [][]int{under(0)}
	for _, n := range b.FilterTrail {
		own := []int{slices.Index(f.Values, strconv.Itoa(n))}
		if itself := slices.Index(f.Values, strconv.Itoa(n)+" itself"); itself >= 0 {
			own = append(own, itself)
		}
		columns = append(columns, append(own, under(n)...))
	}
	return columns
}

// HoldsValue says whether a value of the objective filter holds another among the objectives it reaches.
func (b Board) HoldsValue(value, want int) bool {
	f := b.Filters[b.FilterOn]
	g, found := b.GoalOfValue(f.Values[value])
	if !found || len(g.Children) == 0 {
		return false
	}
	reached := b.Snap.Reach(g.Number)
	n, _ := strings.CutSuffix(f.Values[want], " itself")
	number, err := strconv.Atoi(n)
	return err == nil && number != g.Number && slices.Contains(reached, number)
}

// OpensAColumn says whether a value of the objective filter opens a column of its sub-objectives to the right.
func (b Board) OpensAColumn(value, column int) bool {
	f := b.Filters[b.FilterOn]
	g, found := b.GoalOfValue(f.Values[value])
	if !found || len(g.Children) == 0 {
		return false
	}
	return column == 0 || b.FilterTrail[column-1] != g.Number
}

// GoalOpened says whether a value of the objective filter is the one whose sub-objectives a column shows.
func (b Board) GoalOpened(value, column int) bool {
	g, found := b.GoalOfValue(b.Filters[b.FilterOn].Values[value])
	return found && b.FilterTrail[column] == g.Number
}

// FilterCounts is how many tasks each value of the open filter would keep under the other filters, in the order its values are listed.
func (b Board) FilterCounts() []int {
	if !b.FilterOpen || b.Screen != OnFiltering {
		return nil
	}
	counts, err := b.Filters.Counts(b.Reading().Cards, b.FilterOn)
	if err != nil {
		return nil
	}
	return counts
}

// FilterTaken says whether a value of the open filter is one it is on.
func (b Board) FilterTaken(value int) bool {
	f := b.Filters[b.FilterOn]
	return slices.Contains(f.Taken(), f.Values[value])
}

// FilterHint says how the open filter's values are taken: several ticked together, or one at a time.
func (b Board) FilterHint() string {
	if b.Filters[b.FilterOn].Many {
		return "ticked ones together"
	}
	return "one at a time"
}

// Tree is the objectives drawn, in the order the tab lists them: each of level one followed by its sub-objectives when it is unfolded; a search draws what it reaches, under the parent holding it.
func (b Board) Tree() []Branch {
	if !b.Looking.Empty() {
		return FoundBranches(b.Filters.KeptGoals(b.Snap.Goals), b.Looking.Reads())
	}
	var out []Branch
	for _, br := range Branches(b.Showing().Goals) {
		if br.Under && b.Folded(br.Goal.Parent) {
			continue
		}
		out = append(out, br)
	}
	return out
}

// Folded says whether a parent's sub-objectives are put away: every parent starts so, until it is unfolded in this window or holds what is chosen; nothing is folded while a search is typed.
func (b Board) Folded(n int) bool {
	g, ok := b.Snap.Goal(n)
	if !ok || len(g.Children) == 0 || !b.Looking.Empty() || slices.Contains(b.Unfolded, n) {
		return false
	}
	if c, chosen := b.Snap.Goal(b.Chosen); b.ChosenGoal && chosen && c.Parent == n {
		return false
	}
	return true
}

// GoalsCount is what the head of the list counts: the objectives shown, or, while a search is typed, those it reaches of all there are.
func (b Board) GoalsCount() string {
	if b.Looking.Empty() {
		return strconv.Itoa(len(b.Showing().Goals))
	}
	reached := 0
	for _, br := range b.Tree() {
		if !br.Holds {
			reached++
		}
	}
	return strconv.Itoa(reached) + " of " + strconv.Itoa(len(b.Snap.Goals))
}

// Recoverable says whether a draft kept is of the form open, unchanged since it opened, and so may be taken back.
func (b Board) Recoverable() bool {
	d := b.Draft
	return d.Fields != nil && !b.Form.Changed() &&
		d.Adding == b.Form.Adding && d.ForObjective == b.Form.ForObjective && d.Number == b.Form.Number
}

// ChoosingAt is where in that list the keyboard sits.
func (b Board) ChoosingAt() int { return b.ChoosingOn }

// Ticked says whether an objective is in, as the open list has it.
func (b Board) Ticked(number int) bool { return b.TickedSet[number] }

// ChoosingRefused is what the open list refused last, said at its foot: nothing when the last tick was taken.
func (b Board) ChoosingRefused() string { return b.RefusedChoice }

// OneOnly says whether the open list takes one answer rather than several.
func (b Board) OneOnly() bool { return len(b.ClosedSet()) > 0 || b.ChoosingFor == "parent" }

// ChoosingSearched says whether the open list has a line to type into: only the labels', a label being written as often as picked.
func (b Board) ChoosingSearched() bool { return b.ChoosingFor == "label" }

// ChoosingThen is what a refusal at the foot of the open list says under it: what to untick to put the task where it was ticked, and that nothing was changed.
func (b Board) ChoosingThen() string {
	if b.RefusedChoice == "" {
		return ""
	}
	return b.RefusedThen
}

// LabelsCarried is every label, as the engine listed them, with what carries each.
func (b Board) LabelsCarried() []label.Carried { return b.Carried }

// LabelChosen is the label the list is on.
func (b Board) LabelChosen() string { return b.LabelsShown()[b.LabelOn].Label }

// LabelsShown is the labels the list holds: every one, or those whose name holds what was typed after a slash.
func (b Board) LabelsShown() []label.Carried {
	sought := strings.ToLower(strings.TrimSpace(b.Sought.Reads()))
	if sought == "" {
		return b.Carried
	}
	var shown []label.Carried
	for _, one := range b.Carried {
		if strings.Contains(one.Label, sought) {
			shown = append(shown, one)
		}
	}
	return shown
}

// LabelCreating is the label the list offers to make, what was sought when no label holds it, and whether it offers one.
func (b Board) LabelCreating() (string, bool) {
	sought := strings.TrimSpace(b.Sought.Reads())
	return sought, sought != "" && len(b.LabelsShown()) == 0
}

// Picking says the labels are picked from the board: the list hangs under the chosen card, no form behind it, and each label taken is written at once.
func (b Board) Picking() bool { return b.PickingLabels }

// NarrowingTaken is the one objective the board narrows to, by a filter or by the view in force, and whether it narrows to one.
func (b Board) NarrowingTaken() (string, bool) {
	if taken, ok := b.ObjectiveTaken(); ok {
		return taken, true
	}
	if v, in := b.InView(); in && len(v.Holds.Objective) == 1 && !v.Holds.Itself {
		return strconv.Itoa(v.Holds.Objective[0]), true
	}
	return "", false
}

// Detailed is the objective the right of the tab shows: the one chosen, or the first of the tree while none is, a parent drawn only as what holds a match passed over.
func (b Board) Detailed() (Goal, bool) {
	tree := b.Tree()
	shown := b.DetailGoal
	if b.ChosenGoal {
		shown = b.Chosen
	}
	for _, br := range tree {
		if shown != 0 && br.Goal.Number == shown && !br.Holds {
			return br.Goal, true
		}
	}
	for _, br := range tree {
		if !br.Holds {
			return br.Goal, true
		}
	}
	return Goal{}, false
}

// SeriesListed is the series the tab lists for what is searched and the state it is narrowed to: the active ones, then the stopped ones.
func (b Board) SeriesListed() []engine.Series {
	return InState(ShownSeries(b.Snap.Series, b.Looking.Reads()), b.SeriesState())
}

// SeriesState is the state the series' tab is narrowed to: all, active or stopped.
func (b Board) SeriesState() string { return SeriesStates[b.SeriesIn] }

// SeriesShown is the series the right of the tab shows: the one chosen, or the first listed while none is.
func (b Board) SeriesShown() (engine.Series, bool) {
	all := b.SeriesListed()
	for _, s := range all {
		if s.Number == b.SeriesOn {
			return s, true
		}
	}
	if len(all) == 0 {
		return engine.Series{}, false
	}
	return all[0], true
}

// OccurrenceSays is what a line of the occurrences says: its day, its task, its title, where it stands and the time spent on it.
func (b Board) OccurrenceSays(o series.Occurrence) (day, number, title, is, spent string) {
	day = DaySaid(o.Day, b.Now)
	number, title = "—", "no task made"
	if o.Task != 0 {
		number = "#" + strconv.Itoa(o.Task)
		title = b.OccurrenceTitle(o.Task)
	}
	is = o.Is
	if o.Archived {
		is += " · archived"
	}
	if !o.Archived && o.Task != 0 && slices.Contains(engine.States(), o.Is) && o.Is != "done" {
		is = "on the board"
	}
	spent = o.TimeSpent
	if spent == "" {
		spent = "—"
	}
	return day, number, title, is, spent
}

// OccurrenceTitle is the title of an occurrence, or of its series when the task is no longer read.
func (b Board) OccurrenceTitle(n int) string {
	if c, found := b.Snap.Card(n); found {
		return c.Title
	}
	if s, found := b.SeriesShown(); found {
		return s.Title
	}
	return ""
}

// DeletingAllSays is what deleting a series asks about the tasks it made.
func (b Board) DeletingAllSays() string {
	s, _ := b.Snap.SeriesOf(b.FateOf)
	return "Delete all " + strconv.Itoa(len(MadeStill(s))) + " tasks too?"
}

// Runs is the numbers a stretch gathers along: the columns and the lanes, or the objectives in order.
func (b Board) Runs() [][]int {
	numbers := func(cards []Card) []int {
		var out []int
		for _, c := range cards {
			out = append(out, c.Number)
		}
		return out
	}
	if b.Tab == Objectives {
		var goals []int
		for _, g := range Sorted(b.Showing().Goals) {
			goals = append(goals, g.Number)
		}
		return [][]int{goals}
	}
	var runs [][]int
	for i := range Columns {
		runs = append(runs, numbers(b.InColumn(i)))
	}
	for i := range Stalled {
		runs = append(runs, numbers(b.InLane(i)))
	}
	return runs
}

// AsksAboutSeveral says whether the question is about several gathered.
func (b Board) AsksAboutSeveral() bool {
	return b.Question == AboutDeletingSeveral || b.Question == AboutArchivingSeveral
}

// Flags is what the board flags among those gathered before they are deleted or archived.
func (b Board) Flags() []Flag {
	if b.Question == AboutArchivingSeveral {
		return ArchivingTasks(b.Snap, b.Taken())
	}
	if b.ChosenGoal {
		return DeletingGoals(b.Snap, b.Taken())
	}
	return DeletingTasks(b.Snap, b.Taken())
}

// SeveralOffers is the answers the question about several offers.
func (b Board) SeveralOffers() []string {
	if slices.ContainsFunc(b.Flags(), func(f Flag) bool { return !f.Refused }) {
		return []string{b.AgreeingVerb(), "Skip flagged", "Cancel"}
	}
	return []string{b.AgreeingVerb(), "Cancel"}
}

// SeveralHead is the head of the question about several.
func (b Board) SeveralHead() string {
	verb := "Delete "
	if b.Question == AboutArchivingSeveral {
		verb = "Archive "
	}
	return verb + strconv.Itoa(len(b.Taken())) + " " + b.Kind() + "s?"
}

// Naming is the name being typed for a view about to be kept.
func (b Board) Naming() Written { return b.KeepName }

// InView is the view in force, and whether there is one.
func (b Board) InView() (Viewing, bool) {
	if b.Viewing == "" {
		return Viewing{}, false
	}
	return b.Snap.View(b.Viewing)
}

// ViewChanged says whether a filter or a search stands on top of the view in force.
func (b Board) ViewChanged() bool {
	_, in := b.InView()
	return in && (b.Filters.Narrows() || !b.Looking.Empty())
}

// KeepInput is what keeping the filters and the search as a view is asked with.
func (b Board) KeepInput() engine.Input {
	v, _ := b.InView()
	return b.Filters.Holding(v.View, b.Looking.Reads())
}

// KeepingFoot is the keys the foot of the popover says, each with what it does.
func (b Board) KeepingFoot() [][2]string {
	if !b.OfView {
		return [][2]string{{"⏎", "keep"}, {"esc", "close"}}
	}
	if b.ViewChanged() {
		return [][2]string{{"⏎", "save"}, {"v", "keep as new"}, {"esc", "close"}}
	}
	return [][2]string{{"esc", "close"}}
}

// ChosenView is the view the panel of the views is on, and whether there is one.
func (b Board) ChosenView() (Viewing, bool) {
	if b.ViewOn >= len(b.Snap.Views) {
		return Viewing{}, false
	}
	return b.Snap.Views[b.ViewOn], true
}

// Wrong is what went wrong, if anything did.
func (b Board) Wrong() error { return b.Err }

// InColumn is the cards a column holds.
func (b Board) InColumn(column int) []Card {
	dealt, _ := Deal(b.Showing().Cards, b.Archived)
	return dealt[column].Cards
}

// InLane is the cards a lane of the band holds.
func (b Board) InLane(lane int) []Card {
	_, stalled := Deal(b.Showing().Cards, b.Archived)
	return InLanes(stalled)[lane]
}
