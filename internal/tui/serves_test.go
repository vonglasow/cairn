package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestRows_cutWhatACardServesFromItsStart(t *testing.T) {
	c := board.Card{Number: 22, Title: "Flash the board", State: "todo", Serves: []string{"Connected home automation › Firmware update"}}
	rows := Rows(c, GlyphsFor(Unicode), PlainSkin(), stood, engine.Pages{}, 24)
	last := rows[len(rows)-1]
	if !strings.HasPrefix(last, "◇ …") || !strings.HasSuffix(last, "› Firmware update") || lipgloss.Width(last) > 24 {
		t.Errorf("facts = %q, want what it serves cut from its start within 24", last)
	}
}

func TestACardInsideAnObjectiveLeavesThatObjectiveUnsaid(t *testing.T) {
	b := Board{snap: board.Snapshot{Goals: treeFake().goals}, inside: 2}
	c := board.Card{Number: 1, Objectives: []int{2, 4}, Serves: []string{"Home automation › Enclosure", "Features"}}
	if got := b.servingNamed(c).Serves; len(got) != 1 || got[0] != "Features" {
		t.Errorf("inside objective 2 the card serves %q, want Features alone", got)
	}
}
