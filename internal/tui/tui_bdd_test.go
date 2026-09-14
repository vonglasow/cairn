package tui

// Executable acceptance scenarios for bdd/tui.feature. The Gherkin is the
// contract; the world that drives the board lives in world_test.go and what
// the scenarios look at afterwards lives in steps_test.go.

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
