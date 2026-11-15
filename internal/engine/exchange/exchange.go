// Package exchange holds the export and the import of tasks, objectives, series, views and days.
package exchange

import (
	"cairn/internal/engine/backup"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
	"errors"
	"strconv"
)

// What an export says it is: its kind, the version written, and the oldest version read.
const (
	Kind        = "tasks"
	Version     = 7
	BeforeSteps = 1
)

// ErrNotAnExport is what reading a file that does not hold tasks written by task export is refused with.
var ErrNotAnExport = errors.New("the file holds no tasks written by cairn task export")

// ErrExportPointsOutside is what reading a file whose tasks point at a task or an objective it does not hold is refused with.
var ErrExportPointsOutside = errors.New("the file points outside itself")

// LabelFiled is a label as an export files it.
type LabelFiled struct {
	Label  string `json:"label"`
	Colour string `json:"colour,omitempty"`
	Scope  string `json:"scope,omitempty"`
}

// ObjectiveFiled is an objective as an export files it.
type ObjectiveFiled struct {
	objective.Shape
	Parent  int   `json:"parent,omitempty"`
	WaitsOn []int `json:"waits_on,omitempty"`
}

// Filed is a task as an export files it.
type Filed struct {
	task.Shape
	Parent  int               `json:"parent,omitempty"`
	Spent   map[string]string `json:"spent,omitempty"`
	Focused bool              `json:"focused,omitempty"`
}

// Exported says which file the tasks were written to, and what it holds.
type Exported struct {
	Exported string       `json:"exported"`
	Holds    *backup.Held `json:"holds"`
}

// ImportedFile says which file was read in, and how much of it was added.
type ImportedFile struct {
	Imported string      `json:"imported"`
	Added    backup.Held `json:"added"`
}

// Numbered is the numbers of an export, renumbered as the import made them.
func Numbered(old []int, now map[int]int) []string {
	out := make([]string, 0, len(old))
	for _, n := range old {
		out = append(out, strconv.Itoa(now[n]))
	}
	return out
}

// PlusAll is the numbers, each with a + before it.
func PlusAll(numbers []string) []string {
	out := make([]string, 0, len(numbers))
	for _, n := range numbers {
		out = append(out, "+"+n)
	}
	return out
}

// NumberOf is the number of an objective.
func NumberOf(o objective.Objective) int { return objective.OfObjective(o).Number }
