package reading

// Where every card goes. Four columns for the states work passes through, and
// one band underneath for the three that mean it is not passing through
// anything. The band lies under the last three columns and not under the
// backlog, which runs the whole height: the backlog is the heap everything
// lands in, and it holds more than the other three together.

// Column is one of the four, named as the board writes it.
type Column struct {
	State string
	Name  string
	Cards []Card
}

// columns are the four, left to right, in the order work goes through them.
var columns = []struct{ state, name string }{
	{"backlog", "BACKLOG"},
	{"todo", "TODO"},
	{"current", "CURRENT"},
	{"done", "DONE"},
}

// Columns is how many columns stand side by side: backlog, todo, current,
// done. The band underneath is not one of them — it lies across the last
// three.
//
// Read off the list rather than written again, so the two cannot disagree.
const Columns = 4

// Band is the name the three stopped states share. Not "waiting", because one
// of the three is called watching and two seven-letter words that contain one
// another get read for each other.
const Band = "STALLED"

// Deal puts every card where it belongs: in the column of its state, or in the
// band if its state is one of the three that are not moving.
//
// A sub-task is left off entirely. It is a part of the work its parent names,
// so the parent is what stands in a column and the part is read where the
// parent is read — in the box under its fields. They used to sit beside their
// parents as though they were work of their own, which made a board of 219
// tasks show 219 things when 170 of them were the work and the rest its
// pieces.
//
// The order inside a column is the order the cards arrived in, which is the
// order the engine keeps them in. A board that sorted them again would be a
// second place deciding where a task sits.
func Deal(cards []Card) ([]Column, []Card) {
	out := make([]Column, len(columns))
	for i, c := range columns {
		out[i] = Column{State: c.state, Name: c.name}
	}
	var band []Card
	for _, card := range OnTheBoard(cards) {
		if Stopped(card.State) {
			band = append(band, card)
			continue
		}
		for i := range out {
			if out[i].State == card.State {
				out[i].Cards = append(out[i].Cards, card)
				break
			}
		}
	}
	return out, band
}

// At answers the column of a given state, and whether there is one. The band's
// three states have no column of their own, which is what this says.
func At(all []Column, state string) (Column, bool) {
	for _, c := range all {
		if c.State == state {
			return c, true
		}
	}
	return Column{}, false
}

// IndexOf answers where a column stands, left to right.
func IndexOf(state string) int {
	for i, c := range columns {
		if c.state == state {
			return i
		}
	}
	return -1
}

// StateAt answers which column stands at a place, left to right.
func StateAt(i int) string {
	if i < 0 || i >= len(columns) {
		return ""
	}
	return columns[i].state
}

// OnTheBoard is what a board lays out: every task hanging under nothing.
//
// A sub-task is a part of the work its parent names, so the parent stands in a
// column and the part is read where the parent is read. Asked here rather than
// inside Deal, because what is counted and what the arrows walk have to be the
// same set as what is drawn — a board whose cursor could reach a card nobody
// drew is a board that loses the keyboard on it.
func OnTheBoard(cards []Card) []Card {
	var laid []Card
	for _, c := range cards {
		if c.Parent == 0 {
			laid = append(laid, c)
		}
	}
	return laid
}

// Found is what a board shows of a search: the work a query reaches, with a
// part that reached it answering as the work it is part of.
//
// A board draws only the work, so a sub-task has no line of its own to appear
// on. Dropping one that matched would make a search come back empty for a word
// plainly in the store — measured on a real store, 44 of 49 sub-tasks carry a
// word their parent does not, so that is most of them.
//
// Each parent once, however many of its parts matched, and in the order the
// board already holds them: a search answers what to work on, not how many of
// its pieces mentioned a word.
//
// The command line does not do this and should not. There a line carries a
// number somebody can act on, so naming the sub-task itself says more than
// naming what it hangs under.
func Found(cards []Card, query string) []Card {
	if query == "" {
		return OnTheBoard(cards)
	}
	reached := map[int]bool{}
	for _, c := range cards {
		if !ReachesCard(c, query) {
			continue
		}
		if c.Parent != 0 {
			reached[c.Parent] = true
			continue
		}
		reached[c.Number] = true
	}
	var found []Card
	for _, c := range OnTheBoard(cards) {
		if reached[c.Number] {
			found = append(found, c)
		}
	}
	return found
}
