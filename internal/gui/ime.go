package gui

const ibusAddress = "unix:path=/nonexistent/cairn-keeps-ibus-out"

func ibusKeptOut(system string, getenv func(string) string) (string, bool) {
	if system != "linux" && system != "freebsd" {
		return "", false
	}
	if getenv("CAIRN_IBUS") != "" || getenv("IBUS_ADDRESS") != "" {
		return "", false
	}
	return ibusAddress, true
}
