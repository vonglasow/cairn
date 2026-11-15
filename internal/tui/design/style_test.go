package design

import (
	"cairn/internal/engine/label"
	"cairn/internal/theme"
	"image/color"
	"testing"
)

func TestPriorityColour(t *testing.T) {
	p := PaletteOf(theme.Default())
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
	p := PaletteOf(theme.Default())
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

func TestThePaletteIsTheSchemes(t *testing.T) {
	c := theme.Default()
	p := PaletteOf(c)
	for name, got := range map[string]struct {
		ink  Ink
		want color.NRGBA
	}{
		"low": {p.Low, c.Priority["low"]}, "normal": {p.Normal, c.Priority["normal"]}, "high": {p.High, c.Priority["high"]},
		"critical": {p.Critical, c.Priority["critical"]}, "blocked": {p.Blocked, c.Priority["critical"]},
		"watching": {p.Watching, c.Priority["normal"]}, "paused": {p.Paused, c.Ink.Faint}, "reached": {p.Reached, c.State.Done},
		"late": {p.Late, c.State.Late}, "faint": {p.Faint, c.Ink.Faint}, "bright": {p.Bright, c.Ink.Text},
		"line": {p.Line, c.Line.Edge}, "accent": {p.Accent, c.Accent.Base}, "stalled": {p.Stalled, c.Priority["high"]},
	} {
		if got.ink.Colour != got.want {
			t.Errorf("%s is drawn in %v, want the scheme's %v", name, got.ink.Colour, got.want)
		}
	}
	for _, name := range label.LabelColours {
		if p.Labels[name] != c.Labels[name] {
			t.Errorf("the label colour %s is %v, want the scheme's %v", name, p.Labels[name], c.Labels[name])
		}
	}
	for name, ink := range map[string]Ink{"accent": FadedOf(c).Accent, "late": FadedOf(c).Late} {
		if ink.Colour != c.Ink.Faint {
			t.Errorf("a reference draws %s in %v, want the scheme's faint %v", name, ink.Colour, c.Ink.Faint)
		}
	}
}
