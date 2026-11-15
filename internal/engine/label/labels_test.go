package label

import (
	"errors"
	"testing"
)

func TestLabelChange_refusesAPlainLabelNoLabelMayBe(t *testing.T) {
	if _, err := ParseLabelChange([]string{"#urgent"}); !errors.Is(err, ErrLabelHolds) {
		t.Errorf("err = %v, want ErrLabelHolds", err)
	}
}

func TestLabelChange_refusesToReplaceAndChangeAtOnce(t *testing.T) {
	if _, err := ParseLabelChange([]string{"+urgent", "=draft"}); !errors.Is(err, ErrLabelsBothWays) {
		t.Errorf("err = %v, want ErrLabelsBothWays", err)
	}
}

func TestParseLabels_ignoresEmptyPlacesInAList(t *testing.T) {
	got, err := ParseLabels([]string{"spec,,team,"})
	if err != nil {
		t.Fatalf("ParseLabels: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %v, want two labels", got)
	}
}
