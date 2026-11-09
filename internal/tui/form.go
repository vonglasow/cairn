package tui

import (
	"cairn/internal/engine"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// Form is what is on screen while the form is open.
type Form struct {
	Adding       bool
	Number       int
	ForObjective bool

	asked  *huh.Form
	held   map[string]*string
	keep   *bool
	fields []engine.Field
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

// LongTitle is a long field's name and how many lines it holds, which is the half of a bar's job a terminal box can do from its title.
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

func (f Form) typed() []engine.Field {
	out := make([]engine.Field, len(f.fields))
	for i, field := range f.fields {
		if held, there := f.held[field.Name]; there {
			field.Value = *held
		}
		out[i] = field
	}
	return out
}

func asking(fields []engine.Field, width, height int, skin Skin, about describer) Form {
	held := make(map[string]*string, len(fields))
	keep := new(bool)
	*keep = true
	half := (len(fields) + 1) / 2
	var left, right []huh.Field
	for i, field := range fields {
		value := field.Value
		held[field.Name] = &value
		one := oneField(field, held[field.Name], about)
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

func walkedByArrows() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	down := key.NewBinding(key.WithKeys("down", "tab", "enter"))
	up := key.NewBinding(key.WithKeys("up", "shift+tab"))
	keys.Input.Next, keys.Input.Prev = down, up
	keys.Text.Next = key.NewBinding(key.WithKeys("tab"))
	keys.Text.Prev = key.NewBinding(key.WithKeys("shift+tab"))
	keys.Text.NewLine = key.NewBinding(key.WithKeys("enter", "shift+enter", "alt+enter", "ctrl+j"))
	keys.Text.Submit = key.NewBinding(key.WithDisabled())
	keys.Select.Next, keys.Select.Prev = down, up
	keys.Confirm.Next, keys.Confirm.Prev = key.NewBinding(key.WithKeys("down", "tab")), up
	keys.Select.Up = key.NewBinding(key.WithKeys("shift+up"))
	keys.Select.Down = key.NewBinding(key.WithKeys("shift+down"))
	keys.Select.Left = key.NewBinding(key.WithKeys("left"))
	keys.Select.Right = key.NewBinding(key.WithKeys("right"))
	return keys
}

func oneField(field engine.Field, into *string, about describer) huh.Field {
	if len(field.Choices) > 0 {
		return huh.NewSelect[string]().
			Key(field.Name).
			Title(field.Name).
			Options(huh.NewOptions(field.Choices...)...).
			Inline(true).
			Value(into)
	}
	if field.Long {
		return huh.NewText().Key(field.Name).
			Title(LongTitle(field.Name, *into)).
			TitleFunc(func() string { return LongTitle(field.Name, *into) }, into).
			Lines(3).ShowLineNumbers(true).Value(into)
	}
	input := huh.NewInput().Key(field.Name).Title(field.Name).Value(into)
	if about != nil {
		if said := about(field, into); said != nil {
			input = input.Description(said()).DescriptionFunc(said, into)
		}
	}
	return input
}

func themeOf(skin Skin) *huh.Styles {
	theme := huh.ThemeBase(true)
	p := skin.palette
	for _, state := range []*huh.FieldStyles{&theme.Focused, &theme.Blurred} {
		state.FocusedButton = p.Accent.On(state.FocusedButton).Background(lipgloss.NoColor{}).Reverse(true).Bold(true)
		state.BlurredButton = p.Faint.On(state.BlurredButton).Background(lipgloss.NoColor{})
		state.TextInput.Placeholder = p.Faint.Style()
	}
	theme.Focused.Base = theme.Focused.Base.BorderForeground(p.Accent.Border())
	theme.Focused.Title = p.Accent.On(theme.Focused.Title).Bold(true)
	theme.Focused.Description = p.Faint.On(theme.Focused.Description)
	theme.Focused.SelectedOption = p.Accent.On(theme.Focused.SelectedOption)
	theme.Focused.SelectSelector = p.Accent.On(theme.Focused.SelectSelector)
	theme.Focused.TextInput.Prompt = p.Accent.On(theme.Focused.TextInput.Prompt)
	theme.Focused.TextInput.Cursor = p.Accent.On(theme.Focused.TextInput.Cursor)
	theme.Blurred.Base = theme.Blurred.Base.BorderForeground(p.Line.Border())
	theme.Blurred.Title = p.Faint.On(theme.Blurred.Title)
	return theme
}

func (b Board) formWidth() int  { return b.width - margin - margin - frame }
func (b Board) formHeight() int { return max(1, b.height-fixed) }

func (b Board) addForm() (Board, tea.Cmd) {
	if b.tab == Objectives {
		b.form = asking(engine.NewGoalFields(), b.formWidth(), b.formHeight(), b.skin, b.about(true, 0))
		b.form.Adding, b.form.ForObjective = true, true
		b.screen = OnForm
		return b.started()
	}
	state := engine.StateAt(b.column)
	if b.inBand {
		state = engine.Stalled[0]
	}
	b.form = asking(engine.NewTaskFields(state), b.formWidth(), b.formHeight(), b.skin, b.about(false, 0))
	b.form.Adding = true
	b.screen = OnForm
	return b.started()
}

func (b Board) linkedForm() (Board, tea.Cmd) {
	b.form = asking(engine.LinkedTaskFields(b.task), b.formWidth(), b.formHeight(), b.skin, b.about(false, 0))
	b.form.Adding = true
	b.screen = OnForm
	return b.started()
}

func (b Board) editForm() (Board, tea.Cmd) {
	if b.tab == Objectives {
		for _, g := range b.goals() {
			if g.Number != b.goal {
				continue
			}
			b.form = asking(engine.GoalFields(g), b.formWidth(), b.formHeight(), b.skin, b.about(true, g.Number))
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
	b.form = asking(engine.EditFields(held), b.formWidth(), b.formHeight(), b.skin, b.about(false, held.Number))
	b.form.Number = held.Number
	b.screen = OnForm
	return b.started()
}

func (b Board) started() (Board, tea.Cmd) {
	cmd := b.form.asked.Init()
	model, _ := b.form.asked.Update(cmd())
	if form, ok := model.(*huh.Form); ok {
		b.form.asked = form
	}
	for name, held := range b.form.held {
		b.form.opened[name] = *held
	}
	return b, cmd
}

func (b Board) inForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := b.form.asked.Update(msg)
	if form, ok := model.(*huh.Form); ok {
		b.form.asked = form
	}
	switch b.form.asked.State {
	case huh.StateAborted:
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
		return b.save()
	}
	return b, cmd
}

func (f Form) valueOf(name string) string {
	if held, ok := f.held[name]; ok {
		return *held
	}
	return ""
}

func (b Board) drawForm() string {
	drawn := b.form.asked.View()
	rows := min(b.formHeight()+1, len(strings.Split(drawn, "\n")))
	body := b.panel(b.formHead(),
		through(drawn, b.formWidth(), rows, 0),
		b.width-margin-margin)
	refused := ""
	if b.err != nil {
		refused = " " + b.skin.Wrong.Render(Clip(b.err.Error(), b.width-margin-margin))
	}
	return strings.Join([]string{
		b.tabs(), "",
		lipgloss.NewStyle().PaddingLeft(margin).Render(body),
		refused, b.bar(b.offer("⇥", "field"), b.offer("←→", "change"),
			b.offer("⏎", "next · a line"), b.offer("^S", "save"),
			b.offer("?", "help"), b.offer("esc", "cancel")),
	}, "\n")
}

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
	dressed := b.skin.palette.Stopped(standing).Style()
	if b.form.ForObjective {
		standing = b.form.valueOf("standing")
		dressed = b.skin.palette.Standing(standing).Style()
	}
	right := dressed.Render(standing) + "  " +
		b.skin.palette.Priority(priority).Style().Render(priority)
	return Spread(number+"   "+title, right, b.width-margin-margin-frame)
}

func (b Board) formStays() bool {
	return b.screen == OnForm || b.leaving || b.screen == OnHelp && b.wasScreen == OnForm
}

func (b Board) freshened() Board {
	if b.form.Adding {
		return b
	}
	fresh, ok := b.freshFields()
	if !ok {
		return b
	}
	moved := false
	for i, field := range fresh {
		now := b.form.valueOf(field.Name)
		if was, opened := b.form.opened[field.Name]; opened && now != was {
			fresh[i].Value = now
			continue
		}
		moved = moved || now != field.Value
	}
	if !moved {
		return b
	}
	focused := b.form.asked.GetFocusedField().GetKey()
	number, forObjective := b.form.Number, b.form.ForObjective
	b.form = asking(fresh, b.formWidth(), b.formHeight(), b.skin, b.about(forObjective, number))
	b.form.Number, b.form.ForObjective = number, forObjective
	b, _ = b.started()
	for _, field := range b.freshOpened() {
		b.form.opened[field.Name] = field.Value
	}
	for range 2 * len(fresh) {
		if b.form.asked.GetFocusedField().GetKey() == focused {
			break
		}
		before := b.form.asked.GetFocusedField().GetKey()
		b.form.asked.NextField()
		if b.form.asked.GetFocusedField().GetKey() == before {
			b.form.asked.NextGroup()
		}
	}
	return b
}

func (b Board) freshOpened() []engine.Field {
	fresh, _ := b.freshFields()
	return fresh
}

func (b Board) freshFields() ([]engine.Field, bool) {
	if b.form.ForObjective {
		for _, g := range b.goals() {
			if g.Number == b.form.Number {
				return engine.GoalFields(g), true
			}
		}
		return nil, false
	}
	for _, c := range b.cards() {
		if c.Number == b.form.Number {
			return engine.EditFields(c), true
		}
	}
	return nil, false
}
