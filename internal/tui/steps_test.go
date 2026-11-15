package tui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/tui/geom"
	"fmt"
	"strings"

	"cairn/internal/engine/task"
)

func (w *world) columnsRead(list string) error {
	screen := w.screen()
	for _, name := range strings.Split(list, ",") {
		name = strings.Trim(strings.TrimSpace(name), `"`)
		if !strings.Contains(screen, name) {
			return fmt.Errorf("no column called %q on the board:\n%s", name, screen)
		}
	}
	return nil
}

func (w *world) bandReads(list string) error {
	_, band := board.Deal(w.cards(), w.b.archived)
	var got, want []string
	for _, c := range band {
		got = append(got, c.Title)
	}
	for _, title := range strings.Split(list, ",") {
		want = append(want, strings.Trim(strings.TrimSpace(title), `"`))
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		return fmt.Errorf("the band reads %q, want %q", got, want)
	}
	return nil
}

func (w *world) columnCarries(column, title string) error {
	found, ok := board.At(w.columns(), lower(column))
	if !ok {
		return fmt.Errorf("there is no column called %q", column)
	}
	for _, c := range found.Cards {
		if c.Title == title {
			return nil
		}
	}
	return fmt.Errorf("%q does not carry %q", column, title)
}

func lower(name string) string { return strings.ToLower(name) }

func (w *world) cards() []board.Card {
	found, err := board.Look(w.held)
	if err != nil {
		w.t.Fatalf("looking: %v", err)
	}
	return found.Cards
}

func (w *world) columns() []board.Column {
	all, _ := board.Deal(w.cards(), w.b.archived)
	return all
}

func (w *world) saysItNeeds(columns int) error {
	if err := w.mustSay("bigger window"); err != nil {
		return err
	}
	return w.mustSay(fmt.Sprintf("%d by", columns))
}

func (w *world) saysTheWindowIs(width int) error {
	return w.mustSay(fmt.Sprintf("is %d by", width))
}

func (w *world) mustSay(text string) error {
	if !strings.Contains(w.screen(), text) {
		return fmt.Errorf("the board never says %q:\n%s", text, w.screen())
	}
	return nil
}

func (w *world) modeIs(name string) error {
	if got := w.b.modeName(); got != name {
		return fmt.Errorf("the board is in %q, not in %q", got, name)
	}
	return nil
}

func (w *world) tabIs(name string) error { return w.mustSay(name) }

func (w *world) cursorOnTask(title string) error {
	for _, c := range w.cards() {
		if c.Number == w.b.Where().Task {
			if c.Title == title {
				return nil
			}
			return fmt.Errorf("the cursor is on %q, not on %q", c.Title, title)
		}
	}
	return fmt.Errorf("the cursor is on no task at all, expected %q", title)
}

func (w *world) taskInColumn(number int, state string) error {
	found, err := w.givens().Task(number)
	if err != nil {
		return err
	}
	told := task.OfTask(found)
	if told.State != state {
		return fmt.Errorf("task %d is in %q, not in %q", number, told.State, state)
	}
	return nil
}

func (w *world) taskTitled(number int, title string) error {
	found, err := w.givens().Task(number)
	if err != nil {
		return err
	}
	if told := task.OfTask(found); told.Title != title {
		return fmt.Errorf("task %d is called %q, not %q", number, told.Title, title)
	}
	return nil
}

func (w *world) barOffers(text string) error { return w.mustSay(text) }

func (w *world) barDoesNotOffer(text string) error {
	if strings.Contains(w.screen(), text) {
		return fmt.Errorf("the bar offers %q where it should not:\n%s", text, w.screen())
	}
	return nil
}

func (w *world) helpIsShown() error {
	if got := w.b.modeName(); got != "HELP" {
		return fmt.Errorf("the board is showing %q, not the help", got)
	}
	return nil
}

func (w *world) boardSaysNothingOf(title string) error {
	if strings.Contains(w.screen(), title) {
		return fmt.Errorf("the board draws %q:\n%s", title, w.screen())
	}
	return nil
}

func (w *world) boardCarries(want int) error {
	got := w.b.Where().Showing
	if got != want {
		return fmt.Errorf("the board carries %d tasks, not %d", got, want)
	}
	return nil
}

func (w *world) cardSays(number int, text string) error {
	return w.rowsOf(number, func(rows []string) error {
		for _, row := range rows {
			if strings.Contains(row, text) {
				return nil
			}
		}
		return fmt.Errorf("the card for task %d never says %q: %q", number, text, rows)
	})
}

func (w *world) rowsOf(number int, check func([]string) error) error {
	for _, c := range w.cards() {
		if c.Number == number {
			return check(Rows(c, GlyphsFor(Unicode), PlainSkin(), stood, engine.Pages{}, geom.LayoutFor(w.width, w.high).Interior))
		}
	}
	return fmt.Errorf("there is no task %d", number)
}

func (w *world) cardInFocus(number int) error {
	for _, c := range w.cards() {
		if c.Number == number {
			if c.InFocus {
				return w.mustSay("╔")
			}
			return fmt.Errorf("task %d is not the one in focus", number)
		}
	}
	return fmt.Errorf("there is no task %d", number)
}

func (w *world) boardSays(text string) error { return w.mustSay(text) }

func (w *world) boardSaysBefore(first, second string) error {
	screen := w.screen()
	a, b := strings.Index(screen, first), strings.Index(screen, second)
	if a < 0 || b < 0 || a > b {
		return fmt.Errorf("the board does not say %q before %q:\n%s", first, second, screen)
	}
	return nil
}

func (w *world) boardDoesNotSay(text string) error {
	if strings.Contains(w.screen(), text) {
		return fmt.Errorf("the board says %q where it should not:\n%s", text, w.screen())
	}
	return nil
}

func (w *world) columnReads(column, list string) error {
	found, ok := board.At(w.columns(), lower(column))
	if !ok {
		return fmt.Errorf("there is no column called %q", column)
	}
	var want []string
	for _, title := range strings.Split(list, ",") {
		want = append(want, strings.Trim(strings.TrimSpace(title), `"`))
	}
	if len(found.Cards) != len(want) {
		return fmt.Errorf("%q holds %d tasks, not %d", column, len(found.Cards), len(want))
	}
	for i, card := range found.Cards {
		if card.Title != want[i] {
			return fmt.Errorf("%q reads %q at place %d, want %q", column, card.Title, i, want[i])
		}
	}
	return nil
}
