package tui

import (
	"cairn/internal/board"
	"strings"
	"testing"
	"time"
)

var today = time.Date(2026, 9, 1, 15, 30, 0, 0, time.UTC)

func on(day string) time.Time {
	parsed, err := time.Parse("2006-01-02", day)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestDetailBandOfADoneTaskKeepsTheDayWithoutCountingIt(t *testing.T) {
	b := Board{skin: plainSkin, glyphs: GlyphsFor(Unicode), clock: func() time.Time { return today }}
	done := b.detailBand(board.Card{Number: 1, Title: "Shipped", State: "done", Priority: "normal", Deadline: on("2026-08-29")})
	if !strings.Contains(done, "2026-08-29") || strings.Contains(done, "3d") {
		t.Fatalf("a done task's band reads %q, want its day and no count", done)
	}
	open := b.detailBand(board.Card{Number: 1, Title: "Late", State: "todo", Priority: "normal", Deadline: on("2026-08-29")})
	if !strings.Contains(open, "2026-08-29 · −3d") {
		t.Fatalf("an open task's band reads %q, want its day and its count", open)
	}
}
