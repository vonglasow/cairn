package exchange

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/engine/series"
)

func TestActFiled_refusesWhatNoSeriesDoes(t *testing.T) {
	if _, err := ActFiled(series.Act{Did: "flew", Day: "2026-11-15"}); !errors.Is(err, ErrNotAnExport) {
		t.Errorf("err = %v, want ErrNotAnExport", err)
	}
	if _, err := ActFiled(series.Act{Did: "made", Day: "2026-11-15", At: "yesterday"}); !errors.Is(err, ErrNotAnExport) {
		t.Errorf("err = %v, want a moment not written RFC 3339 refused", err)
	}
	if kept, err := ActFiled(series.Act{Did: "made", Task: 3, Day: "2026-11-15"}); err != nil || kept.Task != 3 || kept.At.IsZero() {
		t.Errorf("kept %+v, %v: want the act, at now", kept, err)
	}
}

func TestNumbered_saysTheNumbersTheImportMade(t *testing.T) {
	if got := PlusAll(Numbered([]int{1, 2}, map[int]int{1: 7, 2: 9})); !slices.Equal(got, []string{"+7", "+9"}) {
		t.Errorf("got %v", got)
	}
}
