package engine_test

import (
	"os"
	"testing"

	"cairn/test"
)

func TestMain(m *testing.M) {
	os.Exit(test.Guarded(m, 1<<30))
}
