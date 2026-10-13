package gui

import "testing"

// On Linux and FreeBSD the window keeps IBus out unless asked to let it in:
// Gio drops the letter IBus answers a key with until a composition has been
// noted, and notes it only once IBus has answered, so an é took two presses.
// An address nothing listens at makes Gio go without IBus, and the keyboard's
// own layout writes the letter. CAIRN_IBUS lets IBus back, for an input method
// a person cannot do without; an address already set elsewhere is not cairn's
// to replace. Nowhere else is IBus what reads the keyboard.
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
