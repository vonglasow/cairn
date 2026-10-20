//go:build !(darwin && !ios) && !(linux && !android)

package gui

import "gioui.org/io/event"

func showLogo() {}

func desktopHome() (data, starts string, ok bool) { return "", "", false }

func dress(event.Event) {}
