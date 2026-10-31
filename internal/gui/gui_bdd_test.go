package gui

import (
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

func initGui(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		newWorld(t).register(sc)
	}
}

func TestAcceptance_Gui(t *testing.T) {
	t.Parallel()
	if test.Suite(t, "gui.feature", initGui(t), "~@later").Run() != 0 {
		t.Fatal("the window's scenarios did not all pass")
	}
}
