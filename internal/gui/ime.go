package gui

// IBus kept out of the window on Linux and FreeBSD. Gio hands every key to
// IBus on X11, and IBus answers an é with the letter before it answers the
// call; Gio takes that letter only once it has noted a composition, which it
// notes after the answer — so the first é was dropped and the second written.
// Without IBus, Gio writes the letter the keyboard's own layout makes, dead
// keys and compose included. See docs/gui.md.

// ibusAddress is an address nothing listens at: Gio fails to reach IBus there
// and goes without it.
const ibusAddress = "unix:path=/nonexistent/cairn-keeps-ibus-out"

// ibusKeptOut says whether to point Gio away from IBus on this system, and at
// what: on the systems Gio speaks to IBus on, unless CAIRN_IBUS asks for it,
// for an input method a person cannot do without, or an address is already
// set, which is somebody else's to change.
func ibusKeptOut(system string, getenv func(string) string) (string, bool) {
	if system != "linux" && system != "freebsd" {
		return "", false
	}
	if getenv("CAIRN_IBUS") != "" || getenv("IBUS_ADDRESS") != "" {
		return "", false
	}
	return ibusAddress, true
}
