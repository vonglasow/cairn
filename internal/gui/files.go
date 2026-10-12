package gui

// Tasks carried out of the window and into it, as a file chosen in the
// system's own dialog. The engine writes the tasks out and adds the tasks of
// a file; what is here is the plumbing between it and the dialog, which only
// the loop can open: a dialog blocks until it is answered, and the window is
// a value that never waits. The window says what it wants (FileAsked), the
// loop asks the system on the side (Filing), and hands the answer back
// (Filed). See docs/gui.md.

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// ErrDeclined is what a dialog answers when it was closed without a file.
var ErrDeclined = errors.New("no file was chosen")

// Files is the system's file dialog: a file chosen to be read, or one to be
// written under a name it suggests.
type Files interface {
	ChooseFile(extensions ...string) (io.ReadCloser, error)
	CreateFile(name string) (io.WriteCloser, error)
}

// FileAsk is what the window wants of the dialog: a file to save the tasks
// already written at From to, under Name, or a file of tasks to add.
type FileAsk struct {
	Save bool
	Name string
	From string
}

// Filed is what came of it: the file the dialog named, the window's own copy
// of a file chosen, or the dialog closed, or what went wrong.
type Filed struct {
	Ask      FileAsk
	Called   string
	Chosen   string
	Declined bool
	Err      error
}

// filesState is what the window keeps of the system's file dialog: the
// dialog it wants opened, until the loop takes it.
type filesState struct {
	fileAsk *FileAsk
}

// FileAsked is the dialog the window wants opened, if any.
func (w Window) FileAsked() (FileAsk, bool) {
	if w.fileAsk == nil {
		return FileAsk{}, false
	}
	return *w.fileAsk, true
}

// FileTaken says the loop has opened the dialog asked for.
func (w Window) FileTaken() Window {
	w.fileAsk = nil
	return w
}

// exportingTasks writes the tasks out through the engine, into a file of the
// window's own, and asks for the dialog that says where it goes.
func (w Window) exportingTasks() Window {
	dir, err := os.MkdirTemp("", "cairn-export-")
	if err != nil {
		return w.failing(err)
	}
	from := filepath.Join(dir, "tasks.json")
	if _, err := w.work.Ask(engine.TaskExport, engine.Input{Given: map[string]string{"to": from}}); err != nil {
		_ = os.RemoveAll(dir)
		return w.failing(err)
	}
	w.fileAsk = &FileAsk{Save: true, Name: "cairn-tasks-" + domain.FormatDay(w.now) + ".json", From: from}
	return w
}

// importingTasks asks for the dialog a file of tasks is chosen in.
func (w Window) importingTasks() Window {
	w.fileAsk = &FileAsk{}
	return w
}

// Filed takes what the dialog answered: tasks written out are said, the tasks
// of a file chosen are added at once — nothing is replaced, so nothing is
// asked — and a dialog closed changes nothing.
func (w Window) Filed(got Filed) Window {
	if got.Declined {
		return w
	}
	if got.Err != nil {
		return w.failing(got.Err)
	}
	if got.Ask.Save {
		return w.telling("the tasks were written to " + got.Called)
	}
	return w.addingFrom(got.Chosen, filepath.Base(got.Called))
}

// addingFrom adds the tasks of the window's copy of a file chosen, and lets
// the copy go whatever the engine answered.
func (w Window) addingFrom(from, name string) Window {
	defer func() { _ = os.RemoveAll(filepath.Dir(from)) }()
	answer, err := w.work.Ask(engine.TaskImport, engine.Input{Given: map[string]string{"from": from}})
	if err != nil {
		return w.failing(err)
	}
	var added domain.Held
	if answer.Holds != nil {
		added = *answer.Holds
	}
	return w.Look().telling(fmt.Sprintf("%d task(s) and %d objective(s) were added from %s", added.Tasks, added.Objectives, name))
}

// Filing is what the loop does on the side while the dialog is open: it asks
// the system, and copies the file one way or the other. It blocks until the
// dialog is answered, which is why it is not the window's to call.
func Filing(ask FileAsk, files Files) Filed {
	got := Filed{Ask: ask}
	if ask.Save {
		defer func() { _ = os.RemoveAll(filepath.Dir(ask.From)) }()
		to, err := files.CreateFile(ask.Name)
		if err != nil {
			got.Declined, got.Err = errors.Is(err, ErrDeclined), declined(err)
			return got
		}
		got.Called, got.Err = calledOf(to, ask.Name), copied(ask.From, to)
		return got
	}
	from, err := files.ChooseFile(".json")
	if err != nil {
		got.Declined, got.Err = errors.Is(err, ErrDeclined), declined(err)
		return got
	}
	got.Called = calledOf(from, "the file chosen")
	held, err := io.ReadAll(from)
	_ = from.Close()
	if err != nil {
		got.Err = fmt.Errorf("reading %s: %w", got.Called, err)
		return got
	}
	dir, err := os.MkdirTemp("", "cairn-import-")
	if err != nil {
		got.Err = err
		return got
	}
	got.Chosen = filepath.Join(dir, "tasks.json")
	got.Err = os.WriteFile(got.Chosen, held, 0o600)
	return got
}

// declined is what went wrong, unless it was only the dialog being closed.
func declined(err error) error {
	if errors.Is(err, ErrDeclined) {
		return nil
	}
	return err
}

// copied writes a file into what the dialog opened, and closes it: some
// systems write the file only once it is closed.
func copied(from string, to io.WriteCloser) error {
	source, err := os.Open(from)
	if err != nil {
		_ = to.Close()
		return err
	}
	defer func() { _ = source.Close() }()
	if _, err := io.Copy(to, source); err != nil {
		_ = to.Close()
		return err
	}
	return to.Close()
}

// calledOf is the name of the file a dialog opened, when it says one.
func calledOf(file any, otherwise string) string {
	if named, ok := file.(interface{ Name() string }); ok {
		return named.Name()
	}
	return otherwise
}
