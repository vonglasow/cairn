package act

import "time"

// Store is where the acts are kept.
type Store interface {
	Record(Act) error
	ActsOn(int) ([]Act, error)
	ActsSince(time.Time) ([]Act, error)
}
