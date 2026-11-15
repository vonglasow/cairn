package objective

import (
	"errors"
)

// ErrObjectiveNowhereToGo is returned when an objective is asked to go behind itself.
var ErrObjectiveNowhereToGo = errors.New("an objective goes first, last, behind another, or under a parent")

// ErrObjectiveParentAndAfter is returned when the parent named is not the one of the objective it is to go behind.
var ErrObjectiveParentAndAfter = errors.New("an objective behind another sits under that one's parent")
