package calendar

import (
	"testing"
	"time"
)

func TestParseEstimate_refusesANegativeDuration(t *testing.T) {
	if _, err := ParseEstimate("-30m"); err == nil {
		t.Error("ParseEstimate(-30m) should fail: time does not run backwards")
	}
}

func TestFormatDuration_writesSecondsWhenThereAreSome(t *testing.T) {
	if got := FormatDuration(90 * time.Second); got != "1m30s" {
		t.Errorf("FormatDuration(90s) = %q, want 1m30s", got)
	}
}

func TestParseMoment_readsBackWhatFormatMomentWrote(t *testing.T) {
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)
	back, err := ParseMoment(FormatMoment(at))
	if err != nil || !back.Equal(at) {
		t.Errorf("ParseMoment = %v, %v, want %v", back, err, at)
	}
	if _, err := ParseMoment("yesterday"); err == nil {
		t.Error("a word was read as a moment")
	}
}
