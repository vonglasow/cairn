//go:build darwin && !ios

package gui

/*
int cairn_less_motion(void);

// cairn_less_motion is written in appkit_darwin.go, beside the rest of what is said to AppKit.
*/
import "C"

func lessMotion() bool { return C.cairn_less_motion() != 0 }
