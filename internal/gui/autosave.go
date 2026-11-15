package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"slices"
	"strconv"
	"time"
)

func (w Window) autosaved() Window {
	if w.Screen != board.OnForm || w.Form.WritingPart || w.Form.Adding && w.Form.ForObjective {
		w.paused = nil
		return w
	}
	now := w.Form.Values()
	paused := slices.Equal(now, w.paused)
	w.paused = now
	if w.Form.Adding {
		if !paused || slices.Equal(now, w.drafting) {
			return w
		}
		w.drafting = now
		return w.keepingDraft()
	}
	if !paused || !w.Form.Changed() || !w.Form.Titled() || slices.Equal(now, w.refused) {
		return w
	}
	saved := w.savedInPlace()
	if saved.Err != nil {
		saved.refused = now
		return saved
	}
	saved.savedAt = w.Now
	return saved
}

func (w Window) savedInPlace() Window {
	kept := w.Form
	w, number, err := w.stored()
	if err != nil {
		return w
	}
	w = w.Look()
	if kept.ForObjective {
		g, found := w.Snap.Goal(number)
		if !found {
			return w
		}
		w.Form = editingGoal(g)
	} else {
		c, found := w.Snap.Card(number)
		if !found {
			return w
		}
		w.Form = w.editing(c)
	}
	w.Chosen, w.ChosenGoal = number, kept.ForObjective
	w.Form = w.Form.Onto(kept.At)
	typed := kept.Fields[kept.At].Value
	w.Form = w.Form.Writes(board.Written{Value: typed, At: kept.Caret})
	w.Form.Opened[kept.At] = typed
	w.Form.PartAt, w.Form.PartsTop, w.Form.PartsTaken = kept.PartAt, kept.PartsTop, kept.PartsTaken
	w.paused = w.Form.Values()
	return w.readingActs()
}

// FormNote is what the foot of an open form says of its saving itself: how long ago it last did, and nothing before it has.
func (w Window) FormNote() string {
	if w.Screen != board.OnForm || w.savedAt.IsZero() {
		return ""
	}
	ago := w.Now.Sub(w.savedAt)
	if ago < time.Minute {
		return "saved " + strconv.Itoa(int(ago.Seconds())) + "s ago"
	}
	return "saved " + calendar.FormatDuration(ago.Truncate(time.Minute)) + " ago"
}
