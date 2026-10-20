//go:build linux && !android

package gui

/*
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>

static void cairn_x11_icon(void *display, unsigned long window, const unsigned int *held, int n) {
	Display *d = display;
	long *wide = malloc(sizeof(long) * n);
	if (wide == NULL) {
		return;
	}
	for (int i = 0; i < n; i++) {
		wide[i] = held[i];
	}
	Atom icon = XInternAtom(d, "_NET_WM_ICON", False);
	XChangeProperty(d, (Window)window, icon, XA_CARDINAL, 32, PropModeReplace, (unsigned char *)wide, n);
	XFlush(d);
	free(wide);
}
*/
import "C"

import (
	"os"
	"os/exec"
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
)

var x11Dressed bool

func showLogo() {}

func desktopHome() (data, starts string, ok bool) {
	running, err := os.Executable()
	if err != nil {
		running = Known
	}
	return DataHome(), Starter(running, exec.LookPath), true
}

func dress(e event.Event) {
	view, ok := e.(app.X11ViewEvent)
	if !ok || !view.Valid() || x11Dressed {
		return
	}
	held := X11Icon()
	C.cairn_x11_icon(view.Display, C.ulong(view.Window), (*C.uint)(unsafe.Pointer(&held[0])), C.int(len(held)))
	x11Dressed = true
}
