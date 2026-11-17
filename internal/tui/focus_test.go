package tui

import (
	"cairn/internal/board"
	"strings"
	"testing"
)

func TestAnOpenTaskInFocusSaysSoAtItsHead(t *testing.T) {
	f := newFake()
	f.cards = []board.Card{
		{Number: 1, Title: "Write the spec", State: "current", Priority: "normal"},
		{Number: 2, Title: "Read the spec", State: "current", Priority: "normal"},
	}
	f.focus = []int{1}
	for n, says := range map[int]bool{1: true, 2: false} {
		b := standing(f)
		b.task = n
		b = hit(b, "e")
		if head := strip(b.formHead()); strings.Contains(head, "● in focus") != says {
			t.Errorf("task %d opened, its head reads %q", n, head)
		}
	}
}
