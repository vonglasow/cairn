package engine

import (
	"cairn/internal/engine/workspace"
	"cairn/internal/store/files"
)

// PagesOf is the pages an answer of config show says are shown, how many days a week shows, and whether tasks come back: what a board that read the configuration again draws from.
func PagesOf(c workspace.Config) Pages {
	var p Pages
	if c.Pages != nil {
		for _, page := range []struct {
			name string
			on   bool
		}{{PageTasks, c.Pages.Tasks}, {PageObjectives, c.Pages.Objectives}, {PageTime, c.Pages.Time == files.TimeOn}} {
			if !page.on {
				p.off = append(p.off, page.name)
			}
		}
	}
	p.noWeekend = c.Time != nil && c.Time.Weekends == files.WeekendsHidden
	p.noRecurrence = c.Recurrence == nil || !c.Recurrence.On
	return p
}
