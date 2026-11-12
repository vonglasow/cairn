package engine

import (
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestSinceWhen_readsTheDaySaidOrTakesToday(t *testing.T) {
	got, err := sinceWhen(Input{Given: map[string]string{"since": "2026-01-02"}})
	want, _ := domain.ParseDeadline("2026-01-02")
	if err != nil || !got.Equal(want) {
		t.Errorf("since 2026-01-02 reads %v, %v", got, err)
	}
	now := time.Now()
	got, err = sinceWhen(Input{})
	if err != nil || !got.Equal(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		t.Errorf("no day said reads %v, %v", got, err)
	}
}
