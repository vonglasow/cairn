package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/engine/label"
	"cairn/internal/gui/geom"
	"errors"
	"path"
	"strings"

	"cairn/internal/store/files"
)

// The words of the dialog a workspace is made in.
const (
	NewSpaceHead      = "New workspace"
	NewSpaceSays      = "Its own board, its own notes and its own settings, in a folder of its own."
	NewSpaceName      = "Name"
	NewSpaceFolder    = "Folder"
	NewSpaceNotes     = "Notes"
	NewSpaceNotesHint = "A folder of a vault lets Obsidian show them."
	NewSpaceColour    = "Colour"
	NewSpaceSwitch    = "Switch to it now"
	NewSpaceChoose    = "Choose…"
	NewSpaceHint      = "A folder that holds a board already is opened as it is."
	NewSpaceCancel    = "Cancel"
	NewSpaceCreate    = "Create"
)

type making struct {
	open, apart, folderTyped, notesTyped bool
	name, folder, notes                  board.Written
	on, colour, choosing                 int
	said                                 string
}

// NewSpaceParts is the dialog a workspace is made in, laid out: where each of its parts is drawn.
type NewSpaceParts struct {
	Box, Head, Says, NameSaid, Name, FolderSaid, Folder, Choose, Hint, NotesSaid, Notes, ChooseNotes, NotesHint, ColourSaid, Switch, Said, Cancel, Create geom.Box
	Swatches                                                                                                                                              []ColourSwatch
}

const (
	newSpaceWide = 440
	newSpaceRow  = 32
	chooseWide   = 96
)

// MakingSpace says whether the dialog a workspace is made in is open.
func (w Window) MakingSpace() bool { return w.Screen == board.OnWorkspaces && w.switcher.making.open }

func (w Window) openingNewSpace() Window {
	w.switcher.making = making{open: true, apart: true, colour: -1}
	return w
}

// NewSpaceTyped is what the dialog's field holds: its name, its folder, or its notes.
func (w Window) NewSpaceTyped(field int) board.Written {
	return *w.switcher.making.field(field)
}

func (m *making) field(at int) *board.Written {
	return []*board.Written{&m.name, &m.folder, &m.notes}[at]
}

// NewSpaceOn is the field the keyboard writes in: 0 for the name, 1 for the folder, 2 for the notes.
func (w Window) NewSpaceOn() int { return w.switcher.making.on }

// NewSpaceAway says whether the workspace made opens in a window of its own.
func (w Window) NewSpaceAway() bool { return w.switcher.making.apart }

// NewSpaceRefused is what the engine said of the last workspace refused, empty while none is.
func (w Window) NewSpaceRefused() string { return w.switcher.making.said }

func (w Window) newSpaceName() string { return strings.TrimSpace(w.switcher.making.name.Reads()) }

// NewSpaceFolderShown is the folder the workspace is made in: the one typed or chosen, or one named after it in cairn's folder.
func (w Window) NewSpaceFolderShown() string {
	if w.switcher.making.folderTyped {
		return strings.TrimSpace(w.switcher.making.folder.Reads())
	}
	return path.Join("~", files.DefaultFolder, w.newSpaceName())
}

// NewSpaceNotesShown is the folder the workspace's notes are written in: the one typed or chosen, or the notes folder inside its own.
func (w Window) NewSpaceNotesShown() string {
	if notes := strings.TrimSpace(w.switcher.making.notes.Reads()); w.switcher.making.notesTyped && notes != "" {
		return notes
	}
	return path.Join(w.NewSpaceFolderShown(), "notes")
}

// NewSpaceColour is the colour picked, or the one read off the name until one is.
func (w Window) NewSpaceColour() string {
	if c := w.switcher.making.colour; c >= 0 {
		return label.LabelColours[c]
	}
	if name := w.newSpaceName(); name != "" {
		return label.ColourOf(name)
	}
	return ""
}

// NewSpaceLaid lays the dialog out in the middle of the window.
func (w Window) NewSpaceLaid() NewSpaceParts {
	wide := min(newSpaceWide, w.width-2*settingsMargin)
	line := func(y int) geom.Box { return geom.Box{X: 0, Y: y, W: wide - 2*padding, H: lineHigh} }
	var p NewSpaceParts
	y := padding
	p.Head = geom.Box{Y: y, W: wide - 2*padding, H: titleHi}
	y += titleHi + between
	p.Says = line(y)
	y += lineHigh + padding
	p.NameSaid = line(y)
	y += lineHigh + inset
	p.Name = geom.Box{Y: y, W: wide - 2*padding, H: newSpaceRow}
	y += newSpaceRow + padding
	p.FolderSaid = line(y)
	y += lineHigh + inset
	p.Folder = geom.Box{Y: y, W: wide - 2*padding - chooseWide - between, H: newSpaceRow}
	p.Choose = geom.Box{X: p.Folder.W + between, Y: y, W: chooseWide, H: newSpaceRow}
	y += newSpaceRow + inset
	p.Hint = line(y)
	y += lineHigh + padding
	p.NotesSaid = line(y)
	y += lineHigh + inset
	p.Notes = geom.Box{Y: y, W: p.Folder.W, H: newSpaceRow}
	p.ChooseNotes = geom.Box{X: p.Choose.X, Y: y, W: chooseWide, H: newSpaceRow}
	y += newSpaceRow + inset
	p.NotesHint = line(y)
	y += lineHigh + padding
	p.ColourSaid = line(y)
	y += lineHigh + inset
	for i, colour := range label.LabelColours {
		p.Swatches = append(p.Swatches, ColourSwatch{Colour: colour, At: geom.Box{X: i * (swatchSide + swatchGap), Y: y, W: swatchSide, H: swatchSide}})
	}
	y += swatchSide + padding
	p.Switch = line(y)
	y += lineHigh + between
	if w.switcher.making.said != "" {
		p.Said = geom.Box{Y: y, W: wide - 2*padding, H: 2 * lineHigh}
		y += 2*lineHigh + between
	}
	y += between
	create := buttonWordWide(w.measured(), NewSpaceCreate) + 2*padding
	cancel := buttonWordWide(w.measured(), NewSpaceCancel) + 2*padding
	p.Create = geom.Box{X: wide - 2*padding - create, Y: y, W: create, H: geom.BarHigh}
	p.Cancel = geom.Box{X: p.Create.X - between - cancel, Y: y, W: cancel, H: geom.BarHigh}
	y += geom.BarHigh + padding
	box := geom.Box{X: (w.width - wide) / 2, Y: max(settingsMargin, (w.height-y)/2), W: wide, H: y}
	p.Box = box
	shift := func(b geom.Box) geom.Box {
		if b == (geom.Box{}) {
			return b
		}
		return geom.Box{X: b.X + box.X + padding, Y: b.Y + box.Y, W: b.W, H: b.H}
	}
	for _, b := range []*geom.Box{&p.Head, &p.Says, &p.NameSaid, &p.Name, &p.FolderSaid, &p.Folder, &p.Choose, &p.Hint, &p.NotesSaid, &p.Notes, &p.ChooseNotes, &p.NotesHint, &p.ColourSaid, &p.Switch, &p.Said, &p.Cancel, &p.Create} {
		*b = shift(*b)
	}
	for i := range p.Swatches {
		p.Swatches[i].At = shift(p.Swatches[i].At)
	}
	return p
}

func (w Window) onNewSpace(name string) Window {
	m := &w.switcher.making
	switch name {
	case "esc":
		return w.leavingNewSpace()
	case "enter":
		return w.makingNewSpace()
	case "tab":
		m.on = (m.on + 1) % 3
	case "shift+tab":
		m.on = (m.on + 2) % 3
	case "backspace":
		return w.writingNewSpace(board.Written.Back)
	case "left":
		return w.writingNewSpace(board.Written.Left)
	case "right":
		return w.writingNewSpace(board.Written.Right)
	}
	return w
}

func (w Window) writingNewSpace(change func(board.Written) board.Written) Window {
	m := &w.switcher.making
	switch {
	case m.on == 1 && !m.folderTyped:
		m.folder, m.folderTyped = board.Written{}, true
	case m.on == 2:
		m.notesTyped = true
	}
	*m.field(m.on) = change(*m.field(m.on))
	return w
}

func (w Window) leavingNewSpace() Window {
	w.switcher.making = making{}
	w.Screen = board.OnTheBoard
	return w
}

func (w Window) clickingNewSpace(p geom.Point) Window {
	laid := w.NewSpaceLaid()
	m := &w.switcher.making
	switch {
	case laid.Name.In(p):
		m.on = 0
	case laid.Folder.In(p):
		m.on = 1
	case laid.Notes.In(p):
		m.on = 2
	case laid.Choose.In(p), laid.ChooseNotes.In(p):
		m.choosing = 1
		if laid.ChooseNotes.In(p) {
			m.choosing = 2
		}
		w.fileAsk = &FileAsk{Folder: true}
	case laid.Switch.In(p):
		m.apart = !m.apart
	case laid.Cancel.In(p):
		return w.leavingNewSpace()
	case laid.Create.In(p):
		return w.makingNewSpace()
	}
	for i, swatch := range laid.Swatches {
		if swatch.At.In(p) {
			m.colour = i
		}
	}
	return w
}

func (w Window) folderForNewSpace(got Filed) Window {
	m := &w.switcher.making
	switch {
	case errors.Is(got.Err, ErrNoPicker):
		m.on = max(1, m.choosing)
	case got.Declined:
	case got.Err != nil:
		return w.failing(got.Err)
	case m.choosing == 2:
		m.notes, m.notesTyped, m.on = board.Written{}.Types(tilde(got.Chosen)).All(), true, 2
	default:
		m.folder, m.folderTyped, m.on = board.Written{}.Types(tilde(got.Chosen)).All(), true, 1
	}
	return w
}

func (w Window) makingNewSpace() Window {
	name := w.newSpaceName()
	given := map[string]string{"name": name, "folder": w.NewSpaceFolderShown()}
	if c := w.switcher.making.colour; c >= 0 {
		given["colour"] = label.LabelColours[c]
	}
	if notes := strings.TrimSpace(w.switcher.making.notes.Reads()); w.switcher.making.notesTyped && notes != "" {
		given["notes"] = notes
	}
	if _, err := w.work.Ask(engine.WorkspaceAdd, engine.Input{Given: given}); err != nil {
		w.switcher.making.said = err.Error()
		return w
	}
	apart := w.switcher.making.apart
	w.switcher.making = making{}
	if apart {
		return w.askingApart(engine.Elsewhere{Workspace: name}, name)
	}
	w.Screen = board.OnTheBoard
	return w.telling(name + " is made")
}

// NewSpaceAsks is what the dialog asks, in the order it asks it.
func NewSpaceAsks() []string {
	return []string{NewSpaceName, NewSpaceFolder, NewSpaceNotes, NewSpaceColour, NewSpaceSwitch}
}
