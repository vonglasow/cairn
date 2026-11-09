package objective

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func waiting(number int, title string, state State, on ...int) Objective {
	return Made(Carried{Number: number, Title: title, State: state, WaitsOn: on})
}

func byNumber(all ...Objective) map[int]Objective {
	by := map[int]Objective{}
	for _, o := range all {
		by[o.number] = o
	}
	return by
}

func TestMade_keepsWhatAnObjectiveWaitsOn(t *testing.T) {
	if got := waiting(3, "Features", Active, 2, 1).Carried().WaitsOn; len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Errorf("WaitsOn = %v, want 2 and 1", got)
	}
}

func TestMayWaitOn_takesAnyOtherObjective(t *testing.T) {
	by := byNumber(waiting(1, "Enclosure", Active), waiting(2, "Firmware", Active, 1))
	if err := MayWaitOn(waiting(3, "Features", Active, 2, 1), by); err != nil {
		t.Errorf("MayWaitOn = %v, want it allowed", err)
	}
}

func TestMayWaitOn_refusesItself(t *testing.T) {
	err := MayWaitOn(waiting(1, "Firmware", Active, 1), byNumber(waiting(1, "Firmware", Active)))
	if !errors.Is(err, ErrWaitsOnItself) || !strings.Contains(err.Error(), "#1 Firmware") {
		t.Errorf("MayWaitOn = %v, want #1 refused waiting on itself", err)
	}
}

func TestMayWaitOn_refusesACircle(t *testing.T) {
	by := byNumber(waiting(1, "Enclosure", Active), waiting(2, "Firmware", Active, 1), waiting(3, "Features", Active, 2))
	err := MayWaitOn(waiting(1, "Enclosure", Active, 3), by)
	if !errors.Is(err, ErrWaitsOnItself) || !strings.Contains(err.Error(), "#3 Features already waits on #1 Enclosure, through #2 Firmware") {
		t.Errorf("MayWaitOn = %v, want the circle through #2 named", err)
	}
	err = MayWaitOn(waiting(1, "Enclosure", Active, 2), by)
	if err == nil || !strings.Contains(err.Error(), "#2 Firmware already waits on #1 Enclosure;") {
		t.Errorf("MayWaitOn = %v, want the circle of two named with nothing between", err)
	}
}

func TestUnmet_namesWhatIsNotReached(t *testing.T) {
	by := byNumber(waiting(1, "Enclosure", Reached), waiting(2, "Firmware", Active), waiting(3, "Features", Active, 2, 1))
	got := Unmet(by[3], by)
	if len(got) != 1 || got[0] != "#3 Features waits on #2 Firmware, not reached" {
		t.Errorf("Unmet = %q", got)
	}
	if got := Unmet(by[2], by); len(got) != 0 {
		t.Errorf("Unmet of one waiting on nothing = %q", got)
	}
}

func TestAwaited_namesWhatWaitsOnIt(t *testing.T) {
	if got := Awaited([]Objective{waiting(3, "Features", Active), waiting(4, "Scenes", Active)}); got != "\n#3 Features waits on it\n#4 Scenes waits on it\n" {
		t.Errorf("Awaited = %q", got)
	}
	if got := Awaited(nil); got != "" {
		t.Errorf("Awaited of nothing = %q", got)
	}
}

func TestText_saysWhatItWaitsOn(t *testing.T) {
	o := waiting(3, "Features", Active, 2)
	said := Text(o, domain.Progress{}, Review{}, Family{WaitsOn: []Objective{waiting(2, "Firmware", Active)}})
	if !strings.Contains(said, "waits on     #2 Firmware\n") {
		t.Errorf("Text =\n%s", said)
	}
}
