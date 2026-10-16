//go:build !(darwin && !ios) && !(linux && !android)

package gui

import "gioui.org/io/event"

func showLogo() {}

func dress(event.Event) {}
