package gui

import "strings"

// AnimationsOff says whether what a GNOME desktop answers of its animations turns them off, which is how it asks for less motion.
func AnimationsOff(said string) bool { return strings.TrimSpace(said) == "false" }
