package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/fang"

	"cairn/internal/engine"
)

func TestWhatToTypeStandsOnALineOfItsOwn(t *testing.T) {
	var out bytes.Buffer
	saidToAPerson(&out, fang.Styles{}, engine.Told(engine.ErrStoreBehind))
	alone := false
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.TrimSpace(line) == "cairn migration apply" {
			alone = true
		}
	}
	if !alone || strings.Contains(out.String(), "run `") {
		t.Errorf("the command is not alone on its line:\n%s", out.String())
	}
}
