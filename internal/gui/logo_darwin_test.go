//go:build darwin && !ios

package gui

import "testing"

func TestTheLogoWaitsForTheApplicationToHaveLaunched(t *testing.T) {
	showLogo()
	if !logoWaitsForLaunch() {
		t.Error("the logo was handed to an application that has not finished launching, which takes it back")
	}
}
