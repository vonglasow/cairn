//go:build windows

package gui

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"gioui.org/app"
	"gioui.org/io/event"
)

const (
	mfString    = 0x0000
	mfPopup     = 0x0010
	gwlpWndProc = ^uintptr(3)
)

var (
	user32          = windows.NewLazySystemDLL("user32.dll")
	createMenu      = user32.NewProc("CreateMenu")
	createPopupMenu = user32.NewProc("CreatePopupMenu")
	appendMenu      = user32.NewProc("AppendMenuW")
	setMenu         = user32.NewProc("SetMenu")
	setWindowLong   = user32.NewProc("SetWindowLongPtrW")
	callWindowProc  = user32.NewProc("CallWindowProcW")

	menuPicks chan<- string
	menuWake  func()
	menuLines []string
	gioProc   uintptr
	installed bool
)

func installMenu(e event.Event, picks chan<- string, wake func()) bool {
	view, ok := e.(app.Win32ViewEvent)
	if !ok || !made(view.HWND) || installed {
		return false
	}
	bar, _, _ := createMenu.Call()
	if !made(bar) {
		return false
	}
	for _, group := range MenuBar() {
		popup, _, _ := createPopupMenu.Call()
		if !made(popup) {
			return false
		}
		for _, item := range group.Items {
			named := utf16Of(winTitle(item))
			appendMenu.Call(popup, mfString, commandOf(len(menuLines)), uintptr(unsafe.Pointer(named)))
			menuLines = append(menuLines, item)
		}
		named := utf16Of(winMenu(group.Name))
		appendMenu.Call(bar, mfPopup, popup, uintptr(unsafe.Pointer(named)))
	}
	menuPicks, menuWake, installed = picks, wake, true
	gioProc, _, _ = setWindowLong.Call(view.HWND, gwlpWndProc, windows.NewCallback(menuProc))
	if !made(gioProc) {
		return false
	}
	done, _, _ := setMenu.Call(view.HWND, bar)
	return made(done)
}

func menuProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if at, ok := commandLine(msg, wParam); ok {
		if line, ok := lineAt(menuLines, at); ok {
			handedOver(line, menuPicks, menuWake)
			return 0
		}
	}
	handed, _, _ := callWindowProc.Call(gioProc, hwnd, msg, wParam, lParam)
	return handed
}

func utf16Of(s string) *uint16 {
	p, _ := windows.UTF16PtrFromString(s)
	return p
}
