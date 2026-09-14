package tui

import "testing"

func TestPriorityColour(t *testing.T) {
	p := Default
	for named, want := range map[string]any{
		"low":      p.Low,
		"normal":   p.Normal,
		"high":     p.High,
		"critical": p.Critical,
		"":         p.Normal,
		"unheard":  p.Normal,
	} {
		if got := p.Priority(named); got != want {
			t.Fatalf("%q is drawn in %v, want %v", named, got, want)
		}
	}
}

func TestStoppedColour(t *testing.T) {
	p := Default
	for named, want := range map[string]any{
		"blocked":  p.Blocked,
		"watching": p.Watching,
		"paused":   p.Paused,
		"anything": p.Paused,
	} {
		if got := p.Stopped(named); got != want {
			t.Fatalf("%q is drawn in %v, want %v", named, got, want)
		}
	}
}
