package motion

import "testing"

func TestADesktopTurningItsAnimationsOffAsksForLessMotion(t *testing.T) {
	for said, off := range map[string]bool{"false\n": true, "true\n": false, "": false} {
		if AnimationsOff(said) != off {
			t.Errorf("%q is read as animations off: %v", said, !off)
		}
	}
}
