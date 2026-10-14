//go:build darwin && !ios

package gui

/*
void cairn_icon(const void *bytes, int length);

// cairn_icon is written in appkit_darwin.go, beside the rest of what is said to AppKit.
*/
import "C"

import "unsafe"

func showLogo() { C.cairn_icon(unsafe.Pointer(&logo[0]), C.int(len(logo))) }
