//go:build !(darwin && !ios)

package gui

import (
	"errors"
	"io"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/x/explorer"
)

type explorerFiles struct{ dialog *explorer.Explorer }

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

func refusedOrDeclined(err error) error {
	if errors.Is(err, explorer.ErrUserDecline) {
		return ErrDeclined
	}
	return err
}
