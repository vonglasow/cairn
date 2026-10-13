package tui

import (
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

func initTui(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		newWorld(t).register(sc)
	}
}

func TestAcceptance_Board(t *testing.T) {
	t.Parallel()
	if test.Suite(t, "tui.feature", initTui(t), "").Run() != 0 {
		t.Fatal("the board's scenarios did not all pass")
	}
}
