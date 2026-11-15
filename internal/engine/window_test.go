package engine

import "testing"

func TestWindowKeep_keepsWhatTheWindowLeavesUntilItIsDropped(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(WindowKeep, given("name", "draft", "value", `{"title":"Ship"}`)); err != nil {
		t.Fatal(err)
	}
	kept, err := e.Ask(WindowKept, given("name", "draft"))
	if err != nil || kept.Kept == nil || *kept.Kept != `{"title":"Ship"}` {
		t.Fatalf("kept %v, %v: want the draft as it was left", kept.Kept, err)
	}
	if _, err := e.Ask(WindowDrop, given("name", "draft")); err != nil {
		t.Fatal(err)
	}
	if gone, err := e.Ask(WindowKept, given("name", "draft")); err != nil || gone.Kept == nil || *gone.Kept != "" {
		t.Errorf("kept %v, %v: want nothing once dropped", gone.Kept, err)
	}
}

func TestWindowKeep_keepsNoSizesUnlessTheConfigurationSaysTo(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(WindowKeep, given("name", "sizes", "value", `{"band":[1,2]}`)); err != nil {
		t.Fatal(err)
	}
	if kept, err := e.Ask(WindowKept, given("name", "sizes")); err != nil || kept.Kept == nil || *kept.Kept != "" {
		t.Errorf("kept %v, %v: want no sizes, the configuration keeps none", kept.Kept, err)
	}
}
