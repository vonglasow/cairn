//go:build linux

package gui

import (
	"cairn/internal/gui/motion"
	"os/exec"
)

func lessMotion() bool {
	said, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "enable-animations").Output()
	return err == nil && motion.AnimationsOff(string(said))
}
