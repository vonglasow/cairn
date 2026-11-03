package engine

import (
	"fmt"

	"cairn/internal/build"
	"cairn/internal/macapp"
)

// AppInstalled says where the Mac application was written, and which cairn it opens the window of.
type AppInstalled struct {
	Installed string `json:"installed" jsonschema:"where the application was written"`
	Runs      string `json:"runs" jsonschema:"the cairn whose window it opens"`
}

// AppInstall writes cairn as a Mac application in the person's Applications folder.
var AppInstall = Operation{
	Thing:   "app",
	Verb:    "install",
	Short:   "Install cairn as a Mac application",
	Does:    "Write cairn.app in your Applications folder, so Spotlight and the Dock open the window: it runs the cairn that installed it, so installing a newer cairn in the same place updates it too. Opened from the Finder it has no shell environment: it finds the store where the configuration file says, or in the default place. Installing again replaces it. On a Mac only.",
	Example: "cairn app install",
	Level:   Write,
	Needs:   NoStore,
	Again:   ChangesNothingMore,
	Answers: AppInstalled{},
	Run: func(on On, _ Input) (Answer, error) {
		if on.Home.Dir == "" {
			return Answer{}, fmt.Errorf("%w: %w", ErrStore, on.Home.Missing)
		}
		running, err := macapp.Running()
		if err != nil {
			return Answer{}, err
		}
		app, err := macapp.Install(on.Home.Dir, running, build.Read().Version)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Installed: &app, Runs: &running}, nil
	},
}

func appText(a AppInstalled) string {
	return "cairn is installed as an application in " + a.Installed + "\n" +
		"it opens the window of " + a.Runs + "\n\n" +
		"open it from Spotlight (⌘ space, cairn) or the Dock\n"
}
