//go:build !(darwin && !ios)

package gui

import (
	"errors"
	"io"
	"os/exec"
	"runtime"
	"strings"

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

func (s explorerFiles) ChooseFolder() (string, error) {
	name, args := FolderCommand(runtime.GOOS)
	picker, err := exec.LookPath(name)
	if err != nil {
		return "", ErrNoPicker
	}
	out, err := exec.Command(picker, args...).Output()
	var exited *exec.ExitError
	if errors.As(err, &exited) && exited.ExitCode() == 1 {
		return "", ErrDeclined
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func refusedOrDeclined(err error) error {
	if errors.Is(err, explorer.ErrUserDecline) {
		return ErrDeclined
	}
	return err
}
