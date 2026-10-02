package tui

import (
	"cairn/internal/reading"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// The form. One form adds and one form edits, and it is the same one: two forms
// would be two places where the fields of a task are listed, and the day a
// thirteenth arrives one of them would be missed.
//
// Nothing it holds reaches the store until it is saved. Writing each field as it
// was left was weighed and dropped: cancelling would have meant nothing, and a
// form that wrote as it went could not have added a task, since there would have
// been nothing to write to.
//
// One thing the board's own rule gives up here, and it is worth saying rather
// than discovering: inside a field being typed into, a letter is a letter.
// `hjkl` navigate everywhere else; in a title they spell.

// Form is what is on screen while the form is open.
type Form struct {
	Adding       bool
	Number       int
	ForObjective bool

	asked  *huh.Form
	held   map[string]*string
	keep   *bool
	fields []reading.Field
	// opened is what the fields held when the form was opened, so leaving it
	// can tell a form somebody typed into from one they only looked at. Ten
	// minutes of typing used to go on an esc with no question, while a
	// deletion of two words was guarded.
	opened map[string]string
}

// Changed says whether anything in the form differs from what it opened with.
func (f Form) Changed() bool {
	for name, was := range f.opened {
		if now, there := f.held[name]; there && *now != was {
			return true
		}
	}
	return false
}

// LongTitle is a long field's name and how many lines it holds, which is the
// half of a bar's job a terminal box can do from its title.
func LongTitle(name, value string) string {
	if value == "" {
		return name
	}
	count := strings.Count(value, "\n") + 1
	if count == 1 {
		return name + " · 1 line"
	}
	return fmt.Sprintf("%s · %d lines", name, count)
}

// typed is the form's fields carrying what was typed into them, which is what
// a save hands over. The fields themselves keep what the form opened with —
// that is what Changed compares against — and handing those over saved the
// title and nothing else, since the board first had a form.
func (f Form) typed() []reading.Field {
	out := make([]reading.Field, len(f.fields))
	for i, field := range f.fields {
		if held, there := f.held[field.Name]; there {
			field.Value = *held
		}
		out[i] = field
	}
	return out
}

// typedInto puts a value into a field, the way the form itself would. Written
// here rather than in a test, because what a form holds is held behind pointers
// and a test reaching for them would be a test reaching past the form.
func (f Form) typedInto(name, value string) Form {
	if held, there := f.held[name]; there {
		*held = value
	}
	return f
}

// asking builds the form over a set of fields. Every value is held behind a
// pointer the form writes into, and read back off those pointers when it is
// saved — so what is typed goes nowhere until then.
//
// The fields are dealt into two columns rather than stacked into one long
// list. A task is a dozen small facts, and a dozen small facts in a single
// column is a scroll rather than a shape.
func asking(fields []reading.Field, width, height int, skin Skin) Form {
	held := make(map[string]*string, len(fields))
	// Save is what the last question is answered with unless somebody says
	// otherwise: a form is opened to change something, so keeping the change is
	// the ordinary way out. Escape is the other one, and it writes nothing.
	keep := new(bool)
	*keep = true
	half := (len(fields) + 1) / 2
	var left, right []huh.Field
	for i, field := range fields {
		value := field.Value
		held[field.Name] = &value
		one := oneField(field, held[field.Name])
		if i < half {
			left = append(left, one)
			continue
		}
		right = append(right, one)
	}
	right = append(right, huh.NewConfirm().
		Key("keep").
		Title("").
		Affirmative("save").
		Negative("cancel").
		Value(keep))
	form := huh.NewForm(huh.NewGroup(left...), huh.NewGroup(right...)).
		WithLayout(huh.LayoutColumns(2)).
		// Told its width rather than left to work one out. Measured: until it
		// is told, a field in place shows four characters of its value — the
		// first drawing of a column read "← to" where the column was todo.
		WithWidth(max(20, width)).
		WithTheme(huh.ThemeFunc(func(bool) *huh.Styles { return themeOf(skin) })).
		WithKeyMap(walkedByArrows()).
		WithShowHelp(false).
		WithHeight(max(6, height))
	opened := make(map[string]string, len(fields))
	for _, field := range fields {
		opened[field.Name] = field.Value
	}
	return Form{asked: form, held: held, keep: keep, fields: fields, opened: opened}
}

// walkedByArrows is the form's clavier: the arrows walk the fields, and a field
// with a closed set of values is changed in place with left and right.
//
// It is arrows and not hjkl on purpose. Everywhere else on the board hjkl
// navigate; in a field being typed into they spell, so the one place that has
// to be different is the one place a letter means itself.
func walkedByArrows() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	down := key.NewBinding(key.WithKeys("down", "tab", "enter"))
	up := key.NewBinding(key.WithKeys("up", "shift+tab"))
	keys.Input.Next, keys.Input.Prev = down, up
	// In the description the arrows walk its lines and enter writes one, the
	// same as in the window; tab is what leaves it. Down used to leave for the
	// next field and enter to submit, so the only way through a long text was
	// a character at a time and the key a hand reaches for to break a line
	// finished the form instead.
	keys.Text.Next = key.NewBinding(key.WithKeys("tab"))
	keys.Text.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	keys.Text.NewLine = key.NewBinding(key.WithKeys("enter", "shift+enter", "alt+enter", "ctrl+j"))
	keys.Text.Submit = key.NewBinding(key.WithDisabled())
	keys.Select.Next, keys.Select.Prev = down, up
	keys.Confirm.Next, keys.Confirm.Prev = key.NewBinding(key.WithKeys("down", "tab")), up
	// A list picked in place is walked with left and right; up and down are
	// spent on the fields, so nothing here may take them.
	keys.Select.Up = key.NewBinding(key.WithKeys("shift+up"))
	keys.Select.Down = key.NewBinding(key.WithKeys("shift+down"))
	keys.Select.Left = key.NewBinding(key.WithKeys("left"))
	keys.Select.Right = key.NewBinding(key.WithKeys("right"))
	return keys
}

// oneField builds the one kind of field this one is: a list for a closed set, a
// box for prose, a line for everything else.
func oneField(field reading.Field, into *string) huh.Field {
	if len(field.Choices) > 0 {
		// Picked in place rather than out of a list that unrolls: seven columns
		// and four priorities are a value to cycle, not a page to read.
		return huh.NewSelect[string]().
			Key(field.Name).
			Title(field.Name).
			Options(huh.NewOptions(field.Choices...)...).
			Inline(true).
			Value(into)
	}
	if field.Long {
		// What a bar says, in the two halves a terminal box can carry: how much
		// there is, beside the name, and where in it the cursor stands, down
		// the side. huh draws the box and offers no bar, and a description the
		// arrows could not walk through read as one cut off.
		// Title first, for what it says the moment the form opens: huh works
		// a TitleFunc out on the first update and draws nothing until then.
		return huh.NewText().Key(field.Name).
			Title(LongTitle(field.Name, *into)).
			TitleFunc(func() string { return LongTitle(field.Name, *into) }, into).
			Lines(3).ShowLineNumbers(true).Value(into)
	}
	return huh.NewInput().Key(field.Name).Title(field.Name).Value(into)
}

// themeOf dresses the form in the board's own colours, so a form does not look
// like something that wandered in from another application.
func themeOf(skin Skin) *huh.Styles {
	theme := huh.ThemeBase(true)
	p := skin.palette
	theme.Focused.Base = theme.Focused.Base.BorderForeground(p.Accent)
	theme.Focused.Title = theme.Focused.Title.Foreground(p.Accent).Bold(true)
	theme.Focused.Description = theme.Focused.Description.Foreground(p.Faint)
	theme.Focused.SelectedOption = theme.Focused.SelectedOption.Foreground(p.Accent)
	theme.Focused.SelectSelector = theme.Focused.SelectSelector.Foreground(p.Accent)
	theme.Focused.FocusedButton = theme.Focused.FocusedButton.Foreground(p.Ink).Background(p.Accent).Bold(true)
	theme.Focused.BlurredButton = theme.Focused.BlurredButton.Foreground(p.Faint)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(p.Accent)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(p.Accent)
	theme.Blurred.Base = theme.Blurred.Base.BorderForeground(p.Line)
	theme.Blurred.Title = theme.Blurred.Title.Foreground(p.Faint)
	return theme
}

// formWidth is what the form has to draw in: the window less the margins and
// the panel it sits in. formHeight is the same for the rows.
//
// Both are one expression used everywhere — where the form is told its size,
// and where the board holds it to that size. Two expressions would let the
// second be wider than the first, and a window wider than what it holds cuts
// nothing: the mistake would be invisible.
func (b Board) formWidth() int  { return b.width - margin - margin - frame }
func (b Board) formHeight() int { return max(1, b.height-fixed) }

// addForm opens the form empty, with the column already filled in from where the
// cursor stands — a task added on a column belongs to that column, and asking
// again would be asking something the board already knows.
func (b Board) addForm() (Board, tea.Cmd) {
	if b.tab == Objectives {
		b.form = asking(reading.GoalFields(reading.Goal{State: "active", Priority: "normal"}), b.formWidth(), b.formHeight(), b.skin)
		b.form.Adding, b.form.ForObjective = true, true
		b.screen = OnForm
		return b.started()
	}
	start := reading.Card{State: reading.StateAt(b.column), Priority: "normal"}
	if b.inBand {
		start.State = kinds[0]
	}
	b.form = asking(reading.TaskFields(start), b.formWidth(), b.formHeight(), b.skin)
	b.form.Adding = true
	b.screen = OnForm
	return b.started()
}

// editForm opens the form over what the cursor is holding.
func (b Board) editForm() (Board, tea.Cmd) {
	if b.tab == Objectives {
		for _, g := range b.goals() {
			if g.Number != b.goal {
				continue
			}
			b.form = asking(reading.GoalFields(g), b.formWidth(), b.formHeight(), b.skin)
			b.form.Number, b.form.ForObjective = g.Number, true
			b.screen = OnForm
			return b.started()
		}
		return b, nil
	}
	held, ok := b.held()
	if !ok {
		return b, nil
	}
	// The parts come with it. They are reachable here and nowhere else: a part
	// is off the board, so a task that carried one could neither be changed
	// nor deleted once it did.
	fields := append(reading.TaskFields(held), reading.Field{
		Name: "parts", Value: reading.PartsSaid(reading.Hanging(b.snap.Cards, held.Number)), Long: true,
	})
	b.form = asking(fields, b.formWidth(), b.formHeight(), b.skin)
	b.form.Number = held.Number
	b.screen = OnForm
	return b.started()
}

// started lets the form set itself up.
//
// Init is called for what it does as much as for what it returns: it is what
// makes the first group active, and a form whose Init was never called draws
// its fields and types into the first of them and then never moves — focused
// without being running. Measured.
//
// What it returns is a sequence, and asking a sequence for its message is
// instant: what would be slow is running the commands inside it, and that is
// bubbletea's to do. Handing the sequence straight back to the form is what
// settles its fields into place; what the form says about it afterwards is not
// needed, because the runtime will ask again.
func (b Board) started() (Board, tea.Cmd) {
	// Init never answers nothing — it always asks for a sequence — so there is
	// no branch here for a case that cannot happen.
	cmd := b.form.asked.Init()
	model, _ := b.form.asked.Update(cmd())
	if form, ok := model.(*huh.Form); ok {
		b.form.asked = form
	}
	// What the form holds once it has settled, and not what the fields were
	// built from: a closed set opened on nothing comes back holding its first
	// choice, and comparing with the value before that made every form read as
	// one somebody had typed into.
	for name, held := range b.form.held {
		b.form.opened[name] = *held
	}
	return b, cmd
}

// inForm hands a key to the form, and answers what the form did about it.
// Escape gives the window back without writing; reaching the end of it saves or
// cancels, depending on which of the two was chosen.
func (b Board) inForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := b.form.asked.Update(msg)
	if form, ok := model.(*huh.Form); ok {
		b.form.asked = form
	}
	switch b.form.asked.State {
	case huh.StateAborted:
		// A form somebody typed into is asked about before it goes.
		if b.form.Changed() {
			b.asking, b.leaving, b.screen = true, true, OnTheBoard
			return b, nil
		}
		b.screen = OnTheBoard
		return b, nil
	case huh.StateCompleted:
		if !*b.form.keep {
			b.screen = OnTheBoard
			return b, nil
		}
		// Left open until the write comes back without a refusal. Closed here,
		// a title left empty or a date mistyped put the refusal in the footer
		// and everything that had been typed in the bin.
		return b.save()
	}
	return b, cmd
}

// valueOf reads a field back off the form by name.
func (f Form) valueOf(name string) string {
	if held, ok := f.held[name]; ok {
		return *held
	}
	return ""
}

// drawForm is the form, whichever of the two things it is filling in. It is
// headed by the thing itself rather than by the word "form": what is being
// changed is a task, and it should look like one while it is being changed.
func (b Board) drawForm() string {
	drawn := b.form.asked.View()
	// Clipped as well as given a height: Huh is told how tall it may be and
	// draws taller than that when its columns are narrow enough to wrap, so the
	// board holds it to what the window actually has. Measured — at sixty
	// columns the form ran four rows past the bottom of the screen. And no
	// taller than what it drew, so a short form is not padded out with a screen
	// of nothing.
	rows := min(b.formHeight()+1, len(strings.Split(drawn, "\n")))
	body := b.panel(b.formHead(),
		through(drawn, b.formWidth(), rows, 0),
		b.width-margin-margin)
	return strings.Join([]string{
		b.tabs(), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(body),
		"", b.bar(b.offer("⇥", "field"), b.offer("←→", "change"),
			b.offer("⏎", "next · a line"), b.offer("^S", "save"),
			b.offer("?", "help"), b.offer("esc", "cancel")),
	}, "\n")
}

// formHead is the line at the top of the form: what is being changed, drawn the
// way a card draws it — the number, the title, and where it stands, in the
// colours it stands in. It reads off the form as it is now, so it follows what
// is being typed.
func (b Board) formHead() string {
	number := b.skin.Faint.Render("new")
	if b.form.Number != 0 {
		number = b.skin.Faint.Render(strconv.Itoa(b.form.Number))
	}
	title := b.form.valueOf("title")
	if title == "" {
		title = b.skin.Faint.Render("untitled")
	} else {
		title = b.skin.Title.Render(title)
	}
	standing, priority := b.form.valueOf("column"), b.form.valueOf("priority")
	dressed := lipgloss.NewStyle().Foreground(b.skin.palette.Stopped(standing))
	if b.form.ForObjective {
		standing = b.form.valueOf("standing")
		dressed = lipgloss.NewStyle().Foreground(b.skin.palette.Standing(standing))
	}
	right := dressed.Render(standing) + "  " +
		lipgloss.NewStyle().Foreground(b.skin.palette.Priority(priority)).Render(priority)
	return Spread(number+"   "+title, right, b.width-margin-margin-frame)
}
