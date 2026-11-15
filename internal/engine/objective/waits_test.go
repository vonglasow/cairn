package objective

import (
	"errors"
	"strings"
	"testing"
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
