package gui

import (
	"cairn/internal/gui/design"
	"reflect"
	"testing"
)

func TestTheWindowIsHandedTheSkinItWears(t *testing.T) {
	if opens := reflect.TypeOf(Run); opens.NumIn() < 3 || opens.In(2) != reflect.TypeFor[design.Skin]() {
		t.Errorf("the window is opened by %v, handed no skin to wear", opens)
	}
}
