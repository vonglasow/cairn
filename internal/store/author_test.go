package store

import (
	"testing"
	"time"

	"cairn/internal/trace"
)

func TestStore_keepsWhoDidAnAct(t *testing.T) {
	s := openStore(t)
	var a trace.AsTold
	a.On(1)
	a.Did(trace.Created)
	a.Through("gui")
	a.By("sam")
	a.At(time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC))
	if err := s.Record(a.Act()); err != nil {
		t.Fatal(err)
	}
	acts, err := s.ActsOn(1)
	if err != nil {
		t.Fatal(err)
	}
	if got := acts[0].Carried(); got.By != "sam" || got.Through != "gui" {
		t.Errorf("the act read back says %q through %q", got.By, got.Through)
	}
}
