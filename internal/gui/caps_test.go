package gui

import (
	"strings"
	"testing"
)

func TestEveryKeyASentenceNamesIsDrawnOnItsCap(t *testing.T) {
	marks := []string{"⏎", "⌘", "⇧", "⌥", "⌫"}
	words := map[string]bool{"esc": true, "Esc": true, "enter": true, "Enter": true}
	w := aFixturesBoard(t)
	screens := append(everyScreen(w), settingsScreens([2]int{1200, 800})...)
	screens = append(screens, struct {
		name string
		w    Window
	}{"a toast naming a key", w.listingBackups().Key("enter")})
	var wrong []string
	for _, s := range screens {
		for _, one := range textDrawn(s.w) {
			if !strings.Contains(strings.TrimSpace(one.said), " ") {
				continue
			}
			for _, word := range strings.Fields(one.said) {
				named := words[strings.Trim(word, ",.:")]
				for _, mark := range marks {
					named = named || strings.Contains(word, mark)
				}
				if named {
					wrong = append(wrong, s.name+": "+one.said)
					break
				}
			}
		}
	}
	if len(wrong) > 0 {
		t.Errorf("%d sentence(s) name a key off its cap:\n%s", len(wrong), strings.Join(wrong, "\n"))
	}
}
