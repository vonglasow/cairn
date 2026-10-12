//go:build darwin && !ios

package gui

/*
#include <stdint.h>
#include <stdlib.h>

void cairn_save_panel(uintptr_t view, const char *name);
void cairn_open_panel(uintptr_t view, const char *extensions);
*/
import "C"

import (
	"io"
	"os"
	"strings"
	"unsafe"

	"gioui.org/app"
	"gioui.org/io/event"
)

// panelAnswers is where the system's file panel hands back what was chosen:
// one panel at a time, the window asks for no other until it is answered.
var panelAnswers = make(chan string, 1)

// appkitFiles is the system's file panels as Files, opened over the window's
// view once it has one.
type appkitFiles struct {
	window *app.Window
	view   *uintptr
}

// systemFiles is the system's file dialog for the window, and what it has to
// be told of every event.
func systemFiles(window *app.Window) (Files, func(event.Event)) {
	files := appkitFiles{window: window, view: new(uintptr)}
	return files, func(e event.Event) {
		if v, ok := e.(app.AppKitViewEvent); ok && made(v.View) {
			*files.view = v.View
		}
	}
}

func (f appkitFiles) ChooseFile(extensions ...string) (io.ReadCloser, error) {
	for i, ext := range extensions {
		extensions[i] = strings.TrimPrefix(ext, ".")
	}
	path, err := f.asked(func(view C.uintptr_t, said *C.char) { C.cairn_open_panel(view, said) }, strings.Join(extensions, ","))
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (f appkitFiles) CreateFile(name string) (io.WriteCloser, error) {
	path, err := f.asked(func(view C.uintptr_t, said *C.char) { C.cairn_save_panel(view, said) }, name)
	if err != nil {
		return nil, err
	}
	return os.Create(path)
}

// asked opens a panel on the main thread and waits for its answer: a path, or
// ErrDeclined when it was closed without one.
func (f appkitFiles) asked(open func(C.uintptr_t, *C.char), said string) (string, error) {
	view := *f.view
	f.window.Run(func() {
		c := C.CString(said)
		defer C.free(unsafe.Pointer(c))
		open(C.uintptr_t(view), c)
	})
	path := <-panelAnswers
	if path == "" {
		return "", ErrDeclined
	}
	return path, nil
}

//export cairnFileChosen
func cairnFileChosen(path *C.char) {
	panelAnswers <- C.GoString(path)
}
