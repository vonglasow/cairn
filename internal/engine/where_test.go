package engine

import (
	"path/filepath"
	"testing"
)

func TestPlaces_keepTheWindowsSizesBesideTheStore(t *testing.T) {
	at := Places{Store: filepath.Join("/somewhere", "cairn.db")}
	if got := at.Sizes(); got != filepath.Join("/somewhere", "cairn.db.sizes.json") {
		t.Errorf("the window's sizes are kept at %q", got)
	}
}

func TestDebugging_isTheEnvironmentsToSay(t *testing.T) {
	said := func(value string) Home {
		return Home{Lookup: func(k string) (string, bool) { return value, k == "CAIRN_DEBUG" && value != "" }}
	}
	if !Debugging(said("1")) {
		t.Error("CAIRN_DEBUG=1 does not turn debug mode on")
	}
	if Debugging(said("")) || Debugging(Home{}) {
		t.Error("debug mode is on with nothing said")
	}
}
