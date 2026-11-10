//go:build !(darwin && !ios) && !linux && !windows

package gui

func lessMotion() bool { return false }
