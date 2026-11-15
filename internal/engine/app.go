package engine

import (
	"cairn/internal/engine/app"
	"fmt"
	"path/filepath"
	"strings"

	"cairn/internal/build"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
)

// appInstall writes cairn as an application: a Mac application in the person's Applications folder, or an entry in a Linux desktop's applications menu.
var appInstall = Operation{
	Thing:   "app",
	Verb:    "install",
	Short:   "Install cairn as an application, on a Mac or a Linux desktop",
	Does:    "On a Mac, write cairn.app in your Applications folder, so Spotlight and the Dock open the window. On a Linux desktop, write cairn.desktop in the applications menu ($XDG_DATA_HOME/applications, or ~/.local/share/applications) and the logo in the icon theme, then refresh the menu with xdg-desktop-menu when it is there. Either way it runs the cairn that installed it, so installing a newer cairn in the same place updates it too. Opened from the desktop it has no shell environment: it finds the store where the configuration file says, or in the default place. Installing again replaces it; one cairn did not write is left alone.",
	Example: "cairn app install",
	Level:   write,
	Needs:   noStore,
	Again:   ChangesNothingMore,
	Answers: app.Installed{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		if on.Home.Dir == "" {
			return Answer{}, fmt.Errorf("%w: %w", errStore, on.Home.Missing)
		}
		running, err := build.Running()
		if err != nil {
			return Answer{}, err
		}
		if strings.Contains(running, string(filepath.Separator)+"go-build") {
			return Answer{}, fmt.Errorf("%s: %w", running, app.ErrPassing)
		}
		switch {
		case macapp.Here:
			installed, err := macapp.Install(on.Home.Dir, running, build.Read().Version)
			if err != nil {
				return Answer{}, err
			}
			from := "Spotlight (⌘ space, cairn) or the Dock"
			return Answer{Installed: &installed, Runs: &running, OpenedFrom: &from}, nil
		case linuxapp.Here:
			entry, starts, err := linuxapp.Install(linuxapp.DataHome(on.Home.Dir, on.Home.lookup()), running)
			if err != nil {
				return Answer{}, err
			}
			from := "your desktop's applications menu (cairn)"
			return Answer{Installed: &entry, Runs: &starts, OpenedFrom: &from}, nil
		}
		return Answer{}, app.ErrNoApplicationHere
	},
}

// Upgrade puts the newest release of cairn in the place of the one that was run; a person asks it, never a machine.
var Upgrade = Operation{
	Verb:  "upgrade",
	Short: "Put the newest release of cairn in the place of this one",
	Does: "Take the newest release of cairn from GitHub, check its archive against the release's SHA256SUMS, and put its binary in the place of the cairn that was run. " +
		"Whatever runs cairn from there runs the new one: the terminal, the Mac application, and a cairn mcp serve already running, at its next call. " +
		"Its place is tried before anything is fetched, and one that cannot be written is refused. " +
		"A cairn already the newest is left as it is; one built from its source, or installed by Homebrew, is left to whoever keeps it. " +
		"With --json it answers from, to and binary; the newest already answers to equal to from.",
	Example: "cairn upgrade\ncairn upgrade --json",
	Level:   write,
	Needs:   noStore,
	Again:   ChangesNothingMore,
	Answers: app.Upgraded{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		running, err := build.Running()
		if err != nil {
			return Answer{}, err
		}
		upgraded, err := app.Upgrade(on.Home.version(), running, build.Read().Platform)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Upgraded: &upgraded}, nil
	},
}
