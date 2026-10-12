//go:build !(darwin && !ios)

package gui

import (
	"errors"
	"io"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/x/explorer"
)

// explorerFiles is the system's file dialog as Files, where Gio's explorer
// opens it: closed without a file is ErrDeclined, whatever the system calls
// it. macOS has its own, in files_darwin.go.
type explorerFiles struct{ dialog *explorer.Explorer }

// systemFiles is the system's file dialog for the window, and what it has to
// be told of every event.
func systemFiles(window *app.Window) (Files, func(event.Event)) {
	dialog := explorer.NewExplorer(window)
	return explorerFiles{dialog}, dialog.ListenEvents
}

func (s explorerFiles) ChooseFile(extensions ...string) (io.ReadCloser, error) {
	file, err := s.dialog.ChooseFile(extensions...)
	return file, refusedOrDeclined(err)
}

func (s explorerFiles) CreateFile(name string) (io.WriteCloser, error) {
	file, err := s.dialog.CreateFile(name)
	return file, refusedOrDeclined(err)
}

// refusedOrDeclined says a dialog closed the way Files says it.
func refusedOrDeclined(err error) error {
	if errors.Is(err, explorer.ErrUserDecline) {
		return ErrDeclined
	}
	return err
}
