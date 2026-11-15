package engine

import (
	"cairn/internal/engine/label"
	"cairn/internal/engine/workspace"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func listedBy(t *testing.T, h Home) map[string]workspace.Workspace {
	t.Helper()
	answer, err := DoFrom(WorkspaceList, h, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]workspace.Workspace{}
	for _, one := range answer.Workspaces.Workspaces {
		by[one.Name] = one
	}
	return by
}

func TestWorkspaceList_aFolderGoneIsSaidNotFound(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	by := listedBy(t, h)
	if !by["home"].NotFound || by["work"].NotFound {
		t.Errorf("home not found %v, work %v; want home alone", by["home"].NotFound, by["work"].NotFound)
	}
	said := workspace.WorkspacesText([]workspace.Workspace{{Name: "home", Folder: "/h", NotFound: true, Colour: "teal"}})
	if said != "home  /h  not found  teal\n" {
		t.Errorf("it says %q", said)
	}
}

func TestLocate_aColourNoneOfTheLabelColoursIsRefusedNamingIt(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "[workspace]\ncolour = \"chartreuse\"\n")
	_, err := Locate(h)
	if !errors.Is(err, label.ErrNoSuchColour) || !strings.Contains(err.Error(), "workspace.colour") || !strings.Contains(err.Error(), "chartreuse") {
		t.Errorf("Locate = %v; want the colour refused, naming it and its key", err)
	}
}
