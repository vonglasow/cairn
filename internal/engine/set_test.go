package engine

import (
	"errors"
	"testing"
)

func TestChangedNumbers_refusesNoughtAndWhatDoesNotRead(t *testing.T) {
	notANumber, bothWays, atBirth := errors.New("nan"), errors.New("both"), errors.New("birth")
	for _, said := range []string{"0", "99999999999999999999"} {
		if _, err := changedNumbers([]string{said}, nil, false, notANumber, bothWays, atBirth); !errors.Is(err, notANumber) {
			t.Errorf("%q gave %v", said, err)
		}
	}
}
