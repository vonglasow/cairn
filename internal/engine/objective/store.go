package objective

import (
	"cairn/internal/engine/filter"
	"errors"
)

// Store is where the objectives are kept.
type Store interface {
	KeepObjective(Objective) (Objective, error)
	FindObjective(int) (Objective, error)
	AllObjectives() ([]Objective, error)
	NarrowedObjectives(*filter.Narrowing) ([]Objective, error)
	UpdateObjective(Objective) error
	DeleteObjective(int) error
	PlaceObjective(number, parent, after int) error
}

// ErrNoSuchObjective is returned when an objective is asked for by a number nothing carries.
var ErrNoSuchObjective = errors.New("there is no such objective")
