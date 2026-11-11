package gui

import (
	"testing"

	"cairn/internal/engine"
)

func TestEveryPanelMovesByItsHeadingAndResizesByItsGrip(t *testing.T) {
	opening := map[Screen]func(Window) Window{
		OnConfiguration: Window.openingConfiguration,
		OnHelp:          func(w Window) Window { return w.Key("?") },
	}
	for screen, open := range opening {
		w := open(holding(configuredWith(engine.Config{}), engine.Card{Number: 1, Title: "Spec", State: "todo"}))
		if w.Screen() != screen {
			t.Fatalf("opened %v, want %v", w.Screen(), screen)
		}
		opens := w.PanelBox()
		head := Point{X: opens.X + opens.W/2, Y: opens.Y + 8}
		moved := w.PressAt(head).MoveTo(Point{X: head.X - 40, Y: head.Y + 30})
		moved = moved.LetGo(Point{X: head.X - 40, Y: head.Y + 30})
		if got := moved.PanelBox(); got.X != opens.X-40 || got.Y != opens.Y+30 || got.W != opens.W {
			t.Errorf("%v dragged by its heading stands at %+v, from %+v", screen, got, opens)
		}
		grip := FormGrip(moved.PanelBox())
		corner := Point{X: grip.X + grip.W/2, Y: grip.Y + grip.H/2}
		grown := moved.PressAt(corner).MoveTo(Point{X: corner.X - 5000, Y: corner.Y - 5000})
		grown = grown.LetGo(Point{X: corner.X - 5000, Y: corner.Y - 5000})
		if got := grown.PanelBox(); got.W < PanelLeast.X || got.H < PanelLeast.Y || got.W >= opens.W {
			t.Errorf("%v shrunk by its grip is %+v, want at least %+v and less than %d wide", screen, got, PanelLeast, opens.W)
		}
		again := open(grown.Key("esc"))
		if got := again.PanelBox(); got != grown.PanelBox() {
			t.Errorf("%v opened again stands at %+v, want where it was left, %+v", screen, got, grown.PanelBox())
		}
	}
}
