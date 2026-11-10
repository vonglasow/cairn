//go:build linux

package gui

import "os/exec"

func lessMotion() bool {
	said, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "enable-animations").Output()
	return err == nil && AnimationsOff(string(said))
}
