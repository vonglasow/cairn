package gui

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// Drafts keeps what was typed into the form of a task being added, past the window closing, until it is saved or thrown away.
type Drafts interface {
	Keep(values map[string]string) error
	Held() (map[string]string, error)
	Drop() error
}

// DraftsAt keeps the draft in a file of its own.
func DraftsAt(path string) Drafts { return fileDrafts(path) }

type fileDrafts string

func (f fileDrafts) Keep(values map[string]string) error {
	written, err := json.Marshal(values)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(string(f)), 0o700); err != nil {
		return err
	}
	return os.WriteFile(string(f), written, 0o600)
}

func (f fileDrafts) Held() (map[string]string, error) {
	read, err := os.ReadFile(string(f))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var values map[string]string
	if err := json.Unmarshal(read, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func (f fileDrafts) Drop() error {
	if err := os.Remove(string(f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

type noDrafts struct{}

func (noDrafts) Keep(map[string]string) error     { return nil }
func (noDrafts) Held() (map[string]string, error) { return nil, nil }
func (noDrafts) Drop() error                      { return nil }

// Drafting is the window keeping the draft of a new task where it is told.
func (w Window) Drafting(d Drafts) Window {
	w.drafts = d
	return w
}

func (f Form) drafted() map[string]string {
	values := map[string]string{}
	for _, field := range f.fields {
		if field.Value != field.Was {
			values[field.Name] = field.Value
		}
	}
	return values
}

func (w Window) keepingDraft() Window {
	if !w.form.Changed() {
		return w.droppingDraft()
	}
	if err := w.drafts.Keep(w.form.drafted()); err != nil {
		return w.failing(err)
	}
	return w
}

func (w Window) droppingDraft() Window {
	if err := w.drafts.Drop(); err != nil {
		return w.failing(err)
	}
	return w
}

func (w Window) startingNew() Window {
	w.form = adding(w.tab, w.addingOn())
	if w.form.ForObjective {
		w.screen = OnForm
		return w
	}
	held, err := w.drafts.Held()
	if err != nil {
		return w.failing(err)
	}
	if len(held) == 0 {
		w.screen = OnForm
		return w
	}
	w.screen, w.question = OnAsking, aboutDraft
	return w
}

func (w Window) pickingUp() Window {
	held, err := w.drafts.Held()
	if err != nil {
		return w.failing(err)
	}
	for name, value := range held {
		w.form = w.form.Set(name, value)
	}
	w.screen, w.question = OnForm, aboutDeleting
	return w
}

func (w Window) startingAfresh() Window {
	w.screen, w.question = OnForm, aboutDeleting
	return w.droppingDraft()
}

func (w Window) leavingUnsaved() Window {
	w = w.keepingDraft()
	if w.err != nil {
		return w
	}
	w.screen, w.form, w.question = OnTheBoard, Form{}, aboutDeleting
	return w
}

// Closing is the window as it closes: a task being added and not saved is kept as a draft with everything typed into it, the last keys too.
func (w Window) Closing() Window {
	if !formShown(w.screen) || !w.form.Adding || w.form.ForObjective {
		return w
	}
	return w.keepingDraft()
}
