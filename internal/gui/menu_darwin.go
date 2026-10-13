//go:build darwin && !ios

package gui

/*
#include <stdlib.h>

void cairn_menu_group(const char *title);
void cairn_menu_item(const char *title, const char *key, int tag);
*/
import "C"

import (
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
)

var (
	menuPicks chan<- string
	menuWake  func()
	menuLines []string
	installed bool
)

func installMenu(e event.Event, picks chan<- string, wake func()) bool {
	view, ok := e.(app.AppKitViewEvent)
	if !ok || !made(view.View) || installed {
		return false
	}
	menuPicks, menuWake, installed = picks, wake, true
	for _, group := range MenuBar() {
		said(group.Name, func(title *C.char) { C.cairn_menu_group(title) })
		for _, item := range group.Items {
			tag, key := len(menuLines), macKey(item)
			menuLines = append(menuLines, item)
			said(item, func(title *C.char) {
				said(key, func(k *C.char) { C.cairn_menu_item(title, k, C.int(tag)) })
			})
		}
	}
	return true
}

func said(s string, to func(*C.char)) {
	c := C.CString(s)
	defer C.free(unsafe.Pointer(c))
	to(c)
}

//export cairnMenuPicked
func cairnMenuPicked(tag C.int) {
	line, ok := lineAt(menuLines, int(tag))
	if !ok {
		return
	}
	handedOver(line, menuPicks, menuWake)
}
