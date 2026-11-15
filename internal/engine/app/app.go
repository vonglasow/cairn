// Package app holds installing cairn as an application, its Claude Code plugin, and upgrading it.
package app

import (
	"errors"
)

// What installing cairn as an application can be turned down for, whatever the system.
var (
	ErrNoApplicationHere = errors.New("an application is installed only on a Mac or a Linux desktop")
	ErrPassing           = errors.New("this cairn is a build go run makes and throws away; install cairn (make install), then run cairn app install from it")
)

// Installed says where the application was written, which cairn it opens the window of, and where the person opens it from.
type Installed struct {
	Installed  string `json:"installed" jsonschema:"where the application was written: the Mac application, or the Linux desktop entry"`
	Runs       string `json:"runs" jsonschema:"the cairn whose window it opens"`
	OpenedFrom string `json:"opened_from" jsonschema:"where the person opens it from"`
}

// Text is app install for a person.
func Text(a Installed) string {
	return "cairn is installed as an application in " + a.Installed + "\n" +
		"it opens the window of " + a.Runs + "\n\n" +
		"open it from " + a.OpenedFrom + "\n"
}
