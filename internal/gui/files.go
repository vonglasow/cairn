package gui

// The store carried out of the window and brought into it, as a file chosen
// in the system's own dialog. The engine writes the store out and puts one in
// place; what is here is the plumbing between it and the dialog, which only
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

// FileAsk is what the window wants of the dialog: a file to save the store
// already written at From to, under Name, or a store to bring in.
type FileAsk struct {
	Save bool
	Name string
	From string
}

// Filed is what came of it: the file the dialog named, the window's own copy
// of a store chosen, or the dialog closed, or what went wrong.
type Filed struct {
	Ask      FileAsk
	Called   string
	Chosen   string
	Declined bool
	Err      error
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

// exportingStore writes the store out through the engine, into a file of the
// window's own, and asks for the dialog that says where it goes.
func (w Window) exportingStore() Window {
	dir, err := os.MkdirTemp("", "cairn-export-")
	if err != nil {
		return w.failing(err)
	}
	from := filepath.Join(dir, "cairn.db")
	if _, err := w.work.Ask(engine.BackupExport, engine.Input{Given: map[string]string{"to": from}}); err != nil {
		_ = os.RemoveAll(dir)
		return w.failing(err)
	}
	w.fileAsk = &FileAsk{Save: true, Name: "cairn-" + domain.FormatDay(w.now) + ".db", From: from}
	return w
}

// importingStore asks for the dialog a store is chosen in.
func (w Window) importingStore() Window {
	w.fileAsk = &FileAsk{}
	return w
}

// Filed takes what the dialog answered: a store written out is said, a store
// chosen is asked about before it replaces anything, and a dialog closed
// changes nothing.
func (w Window) Filed(got Filed) Window {
	if got.Declined {
		return w
	}
	if got.Err != nil {
		return w.failing(got.Err)
	}
	if got.Ask.Save {
		return w.telling("the store was written to " + got.Called)
	}
	w.bringing, w.bringingName = got.Chosen, filepath.Base(got.Called)
	w.screen, w.question = OnAsking, aboutBringingIn
	return w
}

// bringingIn puts the store chosen in place of this one, once agreed to. The
// window's copy of it goes whatever the engine answered.
func (w Window) bringingIn() Window {
	from, name := w.bringing, w.bringingName
	w.screen, w.question, w.bringing = OnTheBoard, aboutDeleting, ""
	defer func() { _ = os.RemoveAll(filepath.Dir(from)) }()
	if _, err := w.work.Ask(engine.BackupImport, engine.Input{Given: map[string]string{"from": from}}); err != nil {
		return w.failing(err)
	}
	return w.Look().telling("the store now holds what " + name + " held")
}

// notBringingIn lets go of the store chosen without bringing it in.
func (w Window) notBringingIn() Window {
	_ = os.RemoveAll(filepath.Dir(w.bringing))
	w.screen, w.question, w.bringing = OnTheBoard, aboutDeleting, ""
	return w
}

// Filing is what the loop does on the side while the dialog is open: it asks
// the system, and copies the store one way or the other. It blocks until the
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
	from, err := files.ChooseFile(".db")
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
	got.Chosen = filepath.Join(dir, "cairn.db")
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
