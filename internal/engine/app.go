package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"cairn/internal/build"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
)

// What installing cairn as an application can be turned down for, whatever the system.
var (
	ErrNoApplicationHere = errors.New("an application is installed only on a Mac or a Linux desktop")
	ErrPassing           = errors.New("this cairn is a build go run makes and throws away; install cairn (make install), then run cairn app install from it")
)

// AppInstalled says where the application was written, which cairn it opens the window of, and where the person opens it from.
type AppInstalled struct {
	Installed  string `json:"installed" jsonschema:"where the application was written: the Mac application, or the Linux desktop entry"`
	Runs       string `json:"runs" jsonschema:"the cairn whose window it opens"`
	OpenedFrom string `json:"opened_from" jsonschema:"where the person opens it from"`
}

// AppInstall writes cairn as an application: a Mac application in the person's Applications folder, or an entry in a Linux desktop's applications menu.
var AppInstall = Operation{
	Thing:   "app",
	Verb:    "install",
	Short:   "Install cairn as an application, on a Mac or a Linux desktop",
	Does:    "On a Mac, write cairn.app in your Applications folder, so Spotlight and the Dock open the window. On a Linux desktop, write cairn.desktop in the applications menu ($XDG_DATA_HOME/applications, or ~/.local/share/applications) and the logo in the icon theme, then refresh the menu with xdg-desktop-menu when it is there. Either way it runs the cairn that installed it, so installing a newer cairn in the same place updates it too. Opened from the desktop it has no shell environment: it finds the store where the configuration file says, or in the default place. Installing again replaces it; one cairn did not write is left alone.",
	Example: "cairn app install",
	Level:   Write,
	Needs:   NoStore,
	Again:   ChangesNothingMore,
	Answers: AppInstalled{},
	Run: func(on On, _ Input) (Answer, error) {
		if on.Home.Dir == "" {
			return Answer{}, fmt.Errorf("%w: %w", ErrStore, on.Home.Missing)
		}
		running, err := build.Running()
		if err != nil {
			return Answer{}, err
		}
		if strings.Contains(running, string(filepath.Separator)+"go-build") {
			return Answer{}, fmt.Errorf("%s: %w", running, ErrPassing)
		}
		switch {
		case macapp.Here:
			app, err := macapp.Install(on.Home.Dir, running, build.Read().Version)
			if err != nil {
				return Answer{}, err
			}
			from := "Spotlight (⌘ space, cairn) or the Dock"
			return Answer{Installed: &app, Runs: &running, OpenedFrom: &from}, nil
		case linuxapp.Here:
			entry, starts, err := linuxapp.Install(linuxapp.DataHome(on.Home.Dir, on.Home.lookup()), running)
			if err != nil {
				return Answer{}, err
			}
			from := "your desktop's applications menu (cairn)"
			return Answer{Installed: &entry, Runs: &starts, OpenedFrom: &from}, nil
		}
		return Answer{}, ErrNoApplicationHere
	},
}

func appText(a AppInstalled) string {
	return "cairn is installed as an application in " + a.Installed + "\n" +
		"it opens the window of " + a.Runs + "\n\n" +
		"open it from " + a.OpenedFrom + "\n"
}
