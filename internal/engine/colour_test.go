package engine

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func listedBy(t *testing.T, h Home) map[string]Workspace {
	t.Helper()
	answer, err := DoFrom(WorkspaceList, h, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]Workspace{}
	for _, one := range answer.Workspaces.Workspaces {
		by[one.Name] = one
	}
	return by
}

func TestWorkspaceList_eachWearsTheColourItsSettingsSayOrOneReadOffItsName(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[workspace]\ncolour = \"Green\"\n")
	by := listedBy(t, h)
	if by["home"].Colour != "green" {
		t.Errorf("home wears %q; want the green its settings say", by["home"].Colour)
	}
	if by["work"].Colour != domain.ColourOf("work") {
		t.Errorf("work wears %q; want %q, read off its name", by["work"].Colour, domain.ColourOf("work"))
	}
}

func TestWorkspaceList_aColourTheSettingsCannotSayIsReadOffTheName(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[workspace]\ncolour = \"chartreuse\"\n")
	settingsIn(t, filepath.Join(h.Dir, "w"), "not toml at all [")
	by := listedBy(t, h)
	if by["home"].Colour != domain.ColourOf("home") || by["work"].Colour != domain.ColourOf("work") {
		t.Errorf("home wears %q, work %q; want each read off its name", by["home"].Colour, by["work"].Colour)
	}
}

func TestWorkspaceList_theColourOfTheStoreInUseIsNotLentToAnother(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "[workspace]\ncolour = \"green\"\n")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range answer.Config.Workspaces {
		if one.Name == "home" && one.Colour != domain.ColourOf("home") {
			t.Errorf("home wears %q; want its own, not work's", one.Colour)
		}
	}
}

func TestWorkspaceList_aFolderGoneIsSaidNotFound(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	by := listedBy(t, h)
	if !by["home"].NotFound || by["work"].NotFound {
		t.Errorf("home not found %v, work %v; want home alone", by["home"].NotFound, by["work"].NotFound)
	}
	said := workspacesText([]Workspace{{Name: "home", Folder: "/h", NotFound: true, Colour: "teal"}})
	if said != "home  /h  not found  teal\n" {
		t.Errorf("it says %q", said)
	}
}

func TestWorkspaceShow_saysTheColour(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[workspace]\ncolour = \"teal\"\n")
	h.Workspace = "home"
	answer, err := DoFrom(WorkspaceShow, h, Input{}, Acting{})
	if err != nil || answer.WorkspaceShown.Colour != "teal" {
		t.Fatalf("WorkspaceShow = %+v, %v; want teal", answer.WorkspaceShown, err)
	}
	if said := RenderText(WorkspaceShow, answer); !strings.Contains(said, "colour     teal\n") {
		t.Errorf("it says\n%s", said)
	}
	h.Workspace = "work"
	if answer, _ := DoFrom(WorkspaceShow, h, Input{}, Acting{}); answer.WorkspaceShown.Colour != domain.ColourOf("work") {
		t.Errorf("work wears %q; want it read off its name", answer.WorkspaceShown.Colour)
	}
}

func TestWorkspaceShow_aLooseStoreWearsNoColourUnlessItsSettingsSayOne(t *testing.T) {
	h := homeHolding(t, "", nil, "loose")
	h.Store = filepath.Join(h.Dir, "loose", "cairn.db")
	if answer, err := DoFrom(WorkspaceShow, h, Input{}, Acting{}); err != nil || answer.WorkspaceShown.Colour != "" {
		t.Errorf("WorkspaceShow = %+v, %v; want no colour", answer.WorkspaceShown, err)
	}
	settingsIn(t, filepath.Join(h.Dir, "loose"), "[workspace]\ncolour = \"peach\"\n")
	if answer, err := DoFrom(WorkspaceShow, h, Input{}, Acting{}); err != nil || answer.WorkspaceShown.Colour != "peach" {
		t.Errorf("WorkspaceShow = %+v, %v; want peach", answer.WorkspaceShown, err)
	}
}

func TestLocate_aColourNoneOfTheLabelColoursIsRefusedNamingIt(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "[workspace]\ncolour = \"chartreuse\"\n")
	_, err := Locate(h)
	if !errors.Is(err, domain.ErrNoSuchColour) || !strings.Contains(err.Error(), "workspace.colour") || !strings.Contains(err.Error(), "chartreuse") {
		t.Errorf("Locate = %v; want the colour refused, naming it and its key", err)
	}
}
