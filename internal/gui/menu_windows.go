//go:build windows

package gui

// The window's menu bar on Windows: a Win32 menu set on the window Gio made,
// and its commands caught by a window procedure put in front of Gio's, which
// hands every other message on unchanged. See menubar.go.

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"gioui.org/app"
	"gioui.org/io/event"
)

const (
	mfString    = 0x0000
	mfPopup     = 0x0010
	gwlpWndProc = ^uintptr(3) // -4
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

// installMenu sets cairn's menu bar on the window once Gio has made it, and
// says whether it did. Anything the system refuses leaves the window as it
// was, with the button standing in for the menu bar.
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

// menuProc takes a command from the menu bar, and hands every message on to
// Gio's own procedure.
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

// utf16Of is a string as Win32 reads one. Turned into a number only inside
// the call it is handed to, so the collector cannot take it in between. It
// fails only on a NUL, which no line of the menu bar holds.
func utf16Of(s string) *uint16 {
	p, _ := windows.UTF16PtrFromString(s)
	return p
}
