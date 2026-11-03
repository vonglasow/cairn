package engine

import (
	"strings"
	"testing"
)

func TestConfigText_saysTheApproachLabel(t *testing.T) {
	said := configText(Config{Approach: &ApproachConfig{Label: "discuss"}})
	if !strings.Contains(said, "\n[approach]\nlabel = \"discuss\"\n") {
		t.Errorf("config show does not say the approach label:\n%s", said)
	}
	none := configText(Config{Approach: &ApproachConfig{}})
	if !strings.Contains(none, "\n[approach]\n# label is not said: no label asks for the approach to be agreed first\n") {
		t.Errorf("config show does not say no label is said:\n%s", none)
	}
}

func TestConfigText_saysWhatItIgnored(t *testing.T) {
	said := configText(Config{Ignored: []string{`"vault" is not a key this cairn knows — cairn accepts store; it is ignored`}})
	if !strings.Contains(said, "\n# ignored: \"vault\" is not a key this cairn knows") {
		t.Errorf("config show does not say what it ignored:\n%s", said)
	}
}
