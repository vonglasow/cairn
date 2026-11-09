package gui

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTheFixturesDrawNoPartOverItsNeighbour(t *testing.T) {
	work := onTheFixtures(t)
	for _, size := range pinnedSizes {
		var wrong []string
		for _, s := range everyScreen(New(work, size.X, size.Y).At(time.Now()).Look()) {
			drawn := Drawn{all: partsDrawn(s.w)}
			for _, crossed := range drawn.Overlaps() {
				wrong = append(wrong, fmt.Sprintf("%s: %s %+v over %s %+v", s.name, crossed[0].Part, crossed[0].At, crossed[1].Part, crossed[1].At))
			}
		}
		if len(wrong) > 0 {
			t.Errorf("at %dx%d, %d overlap(s):\n%s", size.X, size.Y, len(wrong), strings.Join(wrong, "\n"))
		}
	}
}
