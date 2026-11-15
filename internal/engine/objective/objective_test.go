package objective

import (
	"testing"
)

func TestParseObjectiveState_readsNothingAsActive(t *testing.T) {
	got, err := ParseObjectiveState("")
	if err != nil {
		t.Fatalf("ParseObjectiveState(%q): %v", "", err)
	}
	if got != Active {
		t.Errorf("ParseObjectiveState(%q) = %q, want %q", "", got, Active)
	}
}
