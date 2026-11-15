package board

import (
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/label"
	"cairn/internal/engine/workspace"
)

// Board is what a board holds at one moment, whatever draws it: what was read, the page and the screen in sight, what is chosen or gathered, the filters, the search, the form open.
type Board struct {
	Snap       Snapshot
	Version    int64
	Screen     Screen
	Tab        Tab
	Chosen     int
	ChosenGoal bool
	Gathered   []int
	Now        time.Time
	Form       Form
	Filters    Filters
	Looking    Written
	Pages      engine.Pages
	ViewsState
	SeriesTab
	LabelsState
	BackupsState
	CalendarState
	TimeState
	SettingsState
	Telling        string
	ToldFor        int
	ChoosingFor    string
	Filtering      Written
	ChoosingFrom   int
	LabelsOffered  []string
	LabelsCounted  map[string]int
	PickingLabels  bool
	DetailGoal     int
	Acts           []string
	FilterOn       int
	FilterOpen     bool
	FilterTrail    []int
	FilterPick     int
	FilterFromPill bool
	SearchFrom     Screen
	Archived       bool
	Question       Question
	Draft          Form
	TickedSet      map[int]bool
	ChoosingOn     int
	Fates          Deleting
	RefusedChoice  string
	RefusedThen    string
	ViewWalked     int
	FateRow        int
	Unfolded       []int
	Err            error
}

// ViewsState is the views as a board holds them: the one in force, the name being typed for one kept, the one the panel is on, the one renamed.
type ViewsState struct {
	Viewing      string
	KeepName     Written
	OfView       bool
	ViewOn       int
	RenamingView bool
	ViewRenamed  Written
}

// SeriesTab is the series tab as a board holds it: the series on, the state narrowed to, what deleting one asks.
type SeriesTab struct {
	SeriesOn    int
	SeriesFate  int
	FateOf      int
	SeriesChord bool
	SeriesIn    int
}

// LabelsState is the list of the labels as a board holds it: what each carries, the one on, renaming, making, colouring, searching.
type LabelsState struct {
	Carried    []label.Carried
	LabelOn    int
	LabelsFrom int
	Renaming   bool
	Making     bool
	Colouring  bool
	ColourPick int
	Renamed    Written
	Seeking    bool
	Sought     Written
}

// BackupsState is the list of the backups as a board holds it: the backups, the one on, the first in sight.
type BackupsState struct {
	Backups     []backup.Backup
	BackupOn    int
	BackupsFrom int
}

// Question is what the one question screen is asking about.
type Question int

// The questions a board asks.
const (
	AboutDeleting Question = iota
	AboutLeaving
	AboutLeavingSettings
	AboutSweeping
	AboutDeletingSeveral
	AboutArchivingSeveral
	AboutPuttingBack
	AboutDeletingLabel
	AboutFinishingParts
	AboutCairn
	AboutDraft
	AboutDeletingPart
	AboutHiding
	AboutDeletingParts
	AboutApplying
	AboutNewer
	AboutSeriesSave
	AboutDeletingSeriesTasks
)

// CalendarView is what a calendar shows: the days of a month, the months of a year, or the years.
type CalendarView int

// The views of a calendar.
const (
	ViewDays CalendarView = iota
	ViewMonths
	ViewYears
)

// CalendarState is the calendar a date is picked on: the day it is on and what it shows.
type CalendarState struct {
	Day          time.Time
	CalendarView CalendarView
}

// TimeState is the time tab as a board holds it: the week or month shown, its timesheet, the days laid out, the calendar it shows.
type TimeState struct {
	Week     time.Time
	Sheet    engine.Timesheet
	Monthly  bool
	Daily    bool
	DaysFrom [7]int
	Grid     CalendarView
}

// SettingsState is the settings as a board holds them: the configuration read and where, what changed in it, the row on, the key being written, the colour schemes and intros offered.
type SettingsState struct {
	Configured  workspace.Config
	ConfigHere  bool
	Opens       string
	Stamp       workspace.ConfigStamp
	ReadAt      time.Time
	Troubled    *workspace.Trouble
	Changed     []workspace.ConfigChange
	ChangedIn   string
	ChangedFor  int
	Global      bool
	SettingsAt  int
	RowOn       int
	KeyWriting  string
	KeyWritten  Written
	RefusedKey  string
	RefusedSay  string
	ChooserOpen bool
	IntrosOpen  bool
	IntroFrom   Tab
	IntroOn     int
	TileOn      int
	Previewing  string
}
