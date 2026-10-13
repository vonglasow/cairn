package gui

import "testing"

func TestIBusIsKeptOutUnlessAskedFor(t *testing.T) {
	none := func(string) string { return "" }
	for _, c := range []struct {
		system string
		env    map[string]string
		kept   bool
	}{
		{"linux", nil, true},
		{"freebsd", nil, true},
		{"linux", map[string]string{"CAIRN_IBUS": "1"}, false},
		{"linux", map[string]string{"IBUS_ADDRESS": "unix:path=/run/ibus"}, false},
		{"darwin", nil, false},
		{"windows", nil, false},
	} {
		getenv := none
		if c.env != nil {
			getenv = func(name string) string { return c.env[name] }
		}
		address, kept := ibusKeptOut(c.system, getenv)
		if kept != c.kept || kept && address == "" {
			t.Errorf("on %s with %v, IBus kept out: %v at %q", c.system, c.env, kept, address)
		}
	}
}
