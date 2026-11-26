package board

// Screen is what has the window.
type Screen int

const (
	// OnTheBoard is the board itself, which is where a window opens.
	OnTheBoard Screen = iota
	// OnForm is the modal over the board, adding or editing.
	OnForm
	// OnSearch is the board narrowed to what a search reaches, while it is still being typed.
	OnSearch
	// OnAsking is the one question this window asks: deleting loses work, and losing work without being asked is how work is lost.
	OnAsking
	// OnHelp is what everything does, which is the one thing here that has to be read rather than pointed at.
	OnHelp
	// OnChoosing is the list of objectives a task may serve, open over the form.
	OnChoosing
	// OnCalendar is the calendar a deadline is chosen on, over the form.
	OnCalendar
	// OnMenu is what a right click offers, beside the pointer.
	OnMenu
	// OnBackups is the list of backups, one of which may be put back.
	OnBackups
	// OnLabels is the list of every label, one of which may be renamed or taken off everything.
	OnLabels
	// OnFiltering is the board's filters, each walked through its values, the board narrowing as they move.
	OnFiltering
	// OnConfiguration is the configuration in use, key by key, read from the window, with the file to be opened or written.
	OnConfiguration
	// OnPalette is the command bar opened: a command, a number or a title typed, and what answers it.
	OnPalette
	// OnWorkspaces is the switcher dropped under the pill: the workspaces and the recent folders, one of them opened in a window of its own.
	OnWorkspaces
	// OnFates is what deleting an objective asks, row by row, before anything goes.
	OnFates
	// OnKeeping is the popover a view is kept from, its name typed, or the one saying what the view in force holds and shows.
	OnKeeping
	// OnViews is the panel of the views: renamed, moved, deleted, or opened on the board.
	OnViews
	// OnSeriesFates is what deleting a series asks before anything goes: what becomes of the tasks it made.
	OnSeriesFates
	// OnRule is the picker a rule is built in by its parts, over the form, under its repeats field.
	OnRule
	// OnDetail is one task open, whole: its facts, its steps, what was said on it.
	OnDetail
	// OnSaid is what a command typed at the foot answered, until esc.
	OnSaid
	// TooSmallToDraw is a window too small to say anything useful in.
	TooSmallToDraw
)

// Tab is which of the two is being looked at.
type Tab int

const (
	// Tasks is the work itself, and is where a window opens.
	Tasks Tab = iota
	// Objectives is what the work is for.
	Objectives
	// Time is how long the work took, a week at a time: see week.go.
	Time
	// Series is the tasks that come back, each series beside the one chosen: see series.go.
	Series
	// Agents is the Claude sessions running in the workspace: see agents.go.
	Agents
)
