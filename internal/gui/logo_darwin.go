//go:build darwin && !ios

package gui

/*
void cairn_icon(const void *bytes, int length);
int cairn_icon_waits(void);

// cairn_icon and cairn_icon_waits are written in appkit_darwin.go, beside the rest of what is said to AppKit.
*/
import "C"

import (
	"unsafe"

	"gioui.org/io/event"
)

func showLogo() { C.cairn_icon(unsafe.Pointer(&logo[0]), C.int(len(logo))) }

func desktopHome() (data, starts string, ok bool) { return "", "", false }

func logoWaitsForLaunch() bool { return C.cairn_icon_waits() != 0 }

func dress(event.Event) {}
