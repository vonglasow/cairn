package config

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestTheCatalogueHoldsEveryKeyTheExampleWrites(t *testing.T) {
	said := regexp.MustCompile(`^#?\s*([a-z_]+)\s*=`)
	heading := regexp.MustCompile(`^\[([a-z]+)\]`)
	table := ""
	var written []string
	for _, line := range strings.Split(Example, "\n") {
		if m := heading.FindStringSubmatch(line); m != nil {
			table = m[1]
			continue
		}
		m := said.FindStringSubmatch(line)
		if m == nil || table == "workspaces" {
			continue
		}
		name := m[1]
		if table != "" {
			name = table + "." + name
		}
		written = append(written, name)
	}
	var catalogued []string
	for _, key := range Keys() {
		if !key.Store && !key.Answered {
			catalogued = append(catalogued, key.Name)
		}
	}
	slices.Sort(written)
	slices.Sort(catalogued)
	if !slices.Equal(written, catalogued) {
		t.Errorf("the example writes %v, the catalogue holds %v", written, catalogued)
	}
}

func TestTheCatalogueSaysThePlacesTheWorkspaceAnsweringGives(t *testing.T) {
	for name, want := range map[string]bool{"store": true, "notes": true, "plugin": false} {
		if key, ok := KeyNamed(name); !ok || key.Answered != want {
			t.Errorf("%s: %+v, want answered %v", name, key, want)
		}
	}
}

func TestTheCatalogueSaysWhatBelongsToAStoreAndToTheMachine(t *testing.T) {
	for name, want := range map[string][2]bool{"store": {true, false}, "plugin": {true, false}, "workspace.colour": {false, true}, "pages.time": {false, false}} {
		key, ok := KeyNamed(name)
		if !ok || key.Machine != want[0] || key.Store != want[1] {
			t.Errorf("%s is %+v, %v", name, key, ok)
		}
	}
	if _, ok := KeyNamed("workspaces"); ok {
		t.Error("the workspaces are a key the catalogue sets")
	}
}

func TestEveryKeyOfTheCatalogueHasASectionAndWords(t *testing.T) {
	for _, key := range Keys() {
		if key.Section == "" || key.Says == "" || key.Kind == "" {
			t.Errorf("%s is %+v", key.Name, key)
		}
		if key.Kind == KindWords && len(key.Words) < 2 {
			t.Errorf("%s takes words and names %v", key.Name, key.Words)
		}
	}
}

func TestAValueIsWrittenAsTheKeyTakesIt(t *testing.T) {
	for _, c := range []struct{ key, value, literal string }{
		{"pages.time", "hidden", `"hidden"`},
		{"pages.tasks", "false", "false"},
		{"series.on", "true", "true"},
		{"focus.session", "45m", `"45m"`},
		{"archive.start", "2026-10-21", "2026-10-21"},
		{"archive.days", "14", "14"},
		{"approach.label", "discuss", `"discuss"`},
		{"plugin", "~/plugins", `"~/plugins"`},
		{"window.theme", "nord", `"nord"`},
		{"window.intro", "flap", `"flap"`},
		{"workspace.colour", "green", `"green"`},
	} {
		key, _ := KeyNamed(c.key)
		got, err := key.Literal(c.value)
		if err != nil || got != c.literal {
			t.Errorf("%s = %q is written %q, %v; want %q", c.key, c.value, got, err, c.literal)
		}
	}
}

func TestAValueTheKeyDoesNotTakeIsRefusedNamingIt(t *testing.T) {
	for _, c := range []struct{ key, value, said string }{
		{"pages.time", "hiden", `"hiden" is not a word pages.time takes (on, hidden, off)`},
		{"series.on", "yes", `series.on: "yes" is true or false`},
		{"archive.days", "0", `archive.days: "0" is not a number of days`},
		{"archive.days", "x", `archive.days: "x" is not a number of days`},
		{"archive.start", "monday", `archive.start: "monday" is not a day`},
		{"plugin", "plugins", `"plugins"`},
		{"approach.label", "", "approach.label"},
		{"window.intro", "confetti", `"confetti" is not a word window.intro takes`},
	} {
		key, _ := KeyNamed(c.key)
		_, err := key.Literal(c.value)
		if err == nil || !strings.Contains(err.Error(), c.said) {
			t.Errorf("%s = %q is refused with %v, want %q", c.key, c.value, err, c.said)
		}
	}
}

func TestTheSettingsOfAStoreSayTheApproachLabelEmptyForNoLabelHere(t *testing.T) {
	label, _ := KeyNamed("approach.label")
	if got, err := label.LiteralHere(""); err != nil || got != `""` {
		t.Errorf("no label here is written %q, %v", got, err)
	}
	time, _ := KeyNamed("pages.time")
	if _, err := time.LiteralHere(""); err == nil {
		t.Error("the time is written empty in a store's settings")
	}
	if got, err := label.LiteralHere("discuss"); err != nil || got != `"discuss"` {
		t.Errorf("a label here is written %q, %v", got, err)
	}
}

func TestTheSettingsSayTabsAndSeries(t *testing.T) {
	series, _ := KeyNamed("series.on")
	if series.Says != "Series" || series.Section != "Tabs" {
		t.Errorf("series.on says %q under %q, want Series under Tabs", series.Says, series.Section)
	}
}
