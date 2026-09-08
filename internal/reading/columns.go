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
// The order inside a column is the order the cards arrived in, which is the
// order the engine keeps them in. A board that sorted them again would be a
// second place deciding where a task sits.
func Deal(cards []Card) ([]Column, []Card) {
	out := make([]Column, len(columns))
	for i, c := range columns {
		out[i] = Column{State: c.state, Name: c.name}
	}
	var band []Card
	for _, card := range cards {
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
