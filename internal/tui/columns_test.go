package tui

import (
	"cairn/internal/board"
	"testing"
)

func TestUnderTheBand(t *testing.T) {
	if underTheBand(0) {
		t.Fatal("the band was said to lie under the backlog, and it does not")
	}
	for i := 1; i < board.Columns; i++ {
		if !underTheBand(i) {
			t.Fatalf("the band was said not to lie under column %d", i)
		}
	}
}
