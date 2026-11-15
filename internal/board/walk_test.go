package board

import "testing"

func walkingCards() []Card {
	return []Card{
		{Number: 1, State: "backlog"}, {Number: 2, State: "backlog"},
		{Number: 3, State: "todo"},
		{Number: 4, State: "current"}, {Number: 5, State: "current"},
		{Number: 6, State: "blocked"}, {Number: 7, State: "paused"},
	}
}

func TestWalk_movesTheChoiceAcrossTheColumnsAndTheBandOrLeavesIt(t *testing.T) {
	for _, c := range []struct {
		name                 string
		chosen, across, down int
		want                 int
		spot                 Spot
	}{
		{"nothing chosen lands on the first card", 0, 0, 1, 2, Spot{Column: 0, At: 1}},
		{"down a column", 1, 0, 1, 2, Spot{Column: 0, At: 1}},
		{"up stops at the top", 1, 0, -1, 1, Spot{Column: 0}},
		{"across keeps the rank it can", 2, 1, 0, 3, Spot{Column: 1}},
		{"across skips an empty column", 2, 1, 0, 3, Spot{Column: 1}},
		{"down from the last card of a column goes into the lane under it", 3, 0, 1, 6, Spot{Column: 0, InBand: true}},
		{"up from the top of a lane goes to the last card above it", 6, 0, -1, 3, Spot{Column: 1}},
		{"across the band skips an empty lane", 6, 1, 0, 7, Spot{Column: 2, InBand: true}},
	} {
		got, spot, ok := Walk(walkingCards(), false, c.chosen, c.across, c.down)
		if !ok || got.Number != c.want || spot != c.spot {
			t.Errorf("%s: landed on %d at %+v (%v), want %d at %+v", c.name, got.Number, spot, ok, c.want, c.spot)
		}
	}
	for _, c := range []struct {
		name                 string
		chosen, across, down int
	}{
		{"across with no column beyond", 5, 1, 0},
		{"across the band with no lane beyond", 7, 1, 0},
		{"up from a lane with nothing above", 7, 0, -1},
	} {
		if got, _, ok := Walk(walkingCards(), false, c.chosen, c.across, c.down); ok {
			t.Errorf("%s: landed on %d, want the choice left where it is", c.name, got.Number)
		}
	}
	if _, _, ok := Walk(nil, false, 0, 0, 1); ok {
		t.Error("a walk on an empty board landed somewhere")
	}
}

func TestInSight_isTheLeastScrollThatShowsTheElement(t *testing.T) {
	for _, c := range []struct{ at, high, run, where, want int }{
		{0, 10, 30, 0, 0},
		{3, 10, 30, 0, 10},
		{1, 10, 30, 20, 10},
		{2, 10, 30, 5, 5},
	} {
		if got := InSight(c.at, c.high, c.run, c.where); got != c.want {
			t.Errorf("InSight(%d, %d, %d, %d) = %d, want %d", c.at, c.high, c.run, c.where, got, c.want)
		}
	}
}

func TestSpotOf_saysWhereACardSits(t *testing.T) {
	for number, want := range map[int]Spot{2: {Column: 0, At: 1}, 5: {Column: 2, At: 1}, 7: {Column: 2, InBand: true}} {
		if got, ok := SpotOf(walkingCards(), false, number); !ok || got != want {
			t.Errorf("card %d sits at %+v (%v), want %+v", number, got, ok, want)
		}
	}
	if _, ok := SpotOf(walkingCards(), false, 9); ok {
		t.Error("a card not on the board sits somewhere")
	}
}
