package tui

import (
	"fmt"
	"strings"

	"cairn/internal/engine"
	"cairn/internal/task"
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

func (w *world) bandNamed(name string) error {
	if !strings.Contains(w.screen(), name) {
		return fmt.Errorf("the band is not called %q:\n%s", name, w.screen())
	}
	return nil
}

func (w *world) bandCarries(want int) error {
	_, band := engine.Deal(w.cards(), w.b.archived)
	if len(band) != want {
		return fmt.Errorf("the band carries %d tasks, not %d", len(band), want)
	}
	return nil
}

func (w *world) bandReads(list string) error {
	_, band := engine.Deal(w.cards(), w.b.archived)
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
	found, ok := engine.At(w.columns(), lower(column))
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

func (w *world) cards() []engine.Card {
	found, err := engine.Look(w.held)
	if err != nil {
		w.t.Fatalf("looking: %v", err)
	}
	return found.Cards
}

func (w *world) columns() []engine.Column {
	all, _ := engine.Deal(w.cards(), w.b.archived)
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

func (w *world) cursorOnColumn(name string) error {
	if got := engine.StateAt(w.b.Where().Column); got != lower(name) {
		return fmt.Errorf("the cursor is on %q, not on %q", got, lower(name))
	}
	return nil
}

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

func (w *world) asksWhichOf(first, second, third string) error {
	for _, kind := range []string{first, second, third} {
		if err := w.mustSay(kind); err != nil {
			return err
		}
	}
	return w.mustSay("which kind of stopped")
}

func (w *world) asksNothing() error {
	if strings.Contains(w.screen(), "which kind of stopped") {
		return fmt.Errorf("the board asked something it should not have:\n%s", w.screen())
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

func (w *world) cardDescribed(number int) error {
	return w.rowsOf(number, func(rows []string) error {
		if strings.Contains(rows[0], GlyphsFor(Unicode).Described) {
			return nil
		}
		return fmt.Errorf("the card for task %d is not marked as described: %q", number, rows[0])
	})
}

func (w *world) cardArchived(number int) error {
	return w.rowsOf(number, func(rows []string) error {
		if strings.Contains(rows[0], GlyphsFor(Unicode).Archived) {
			return nil
		}
		return fmt.Errorf("the card for task %d is not marked as archived: %q", number, rows[0])
	})
}

func (w *world) rowsOf(number int, check func([]string) error) error {
	for _, c := range w.cards() {
		if c.Number == number {
			return check(Rows(c, GlyphsFor(Unicode), PlainSkin(), stood, LayoutFor(w.width, w.high).Interior))
		}
	}
	return fmt.Errorf("there is no task %d", number)
}

func (w *world) goalCardSays(number int, text string) error {
	found, err := engine.Look(w.held)
	if err != nil {
		return err
	}
	for _, g := range found.Goals {
		if g.Number != number {
			continue
		}
		for _, row := range w.b.with(PlainSkin()).goalRows(g, w.width-6) {
			if strings.Contains(row, text) {
				return nil
			}
		}
		return fmt.Errorf("the card for objective %d never says %q", number, text)
	}
	return fmt.Errorf("there is no objective %d", number)
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

func (w *world) boardSaysOnce(text string) error {
	if got := strings.Count(w.screen(), text); got != 1 {
		return fmt.Errorf("the board says %q %d times, not once:\n%s", text, got, w.screen())
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
	found, ok := engine.At(w.columns(), lower(column))
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

func (w *world) ruleBetweenColumns() error {
	l := LayoutFor(w.width, w.high)
	lines := strings.Split(w.screen(), "\n")
	head := -1
	for i, line := range lines {
		if strings.Contains(line, "BACKLOG") {
			head = i
			break
		}
	}
	if head < 0 {
		return fmt.Errorf("no column header on the board:\n%s", w.screen())
	}
	for gap := 0; gap < engine.Columns-1; gap++ {
		at := margin + l.Column + gap*(l.Column+1)
		for _, line := range lines[head : head+2+l.Top] {
			runes := []rune(line)
			if len(runes) <= at || runes[at] != '│' {
				return fmt.Errorf("no rule at column %d of %q:\n%s", at, line, w.screen())
			}
		}
	}
	return nil
}
