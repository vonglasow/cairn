package sqlite

import (
	"testing"
)

func TestVersion_RefusesOnAClosedStore(t *testing.T) {
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("closing the store: %v", err)
	}
	if _, err := s.Version(); err == nil {
		t.Fatal("a closed store answered a version")
	}
}
