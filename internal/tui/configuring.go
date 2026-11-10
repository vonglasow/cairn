package tui

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"cairn/internal/config"
	"cairn/internal/engine"
)

const globalWord = "--global"

const completedAtMost = 5

type configHeld struct {
	config  engine.Config
	stamp   engine.ConfigStamp
	readAt  time.Time
	trouble *engine.Trouble
	notice  string
}

// Reread is the configuration as the beat found it: the files' stamp now, and the configuration read again when one of them said something else.
type Reread struct {
	Stamp  engine.ConfigStamp
	Config *engine.Config
	First  bool
}

// Configured is what a key set or taken out from the : line answered, and the configuration as it reads after the write.
type Configured struct {
	Text   string
	Config *engine.Config
	Stamp  engine.ConfigStamp
	Err    error
}

func (b Board) readingConfig() tea.Cmd {
	return func() tea.Msg {
		got, err := b.work.Ask(engine.ConfigShow, engine.Input{})
		if err != nil || got.Config == nil {
			return nil
		}
		return Reread{Stamp: engine.StampOf(*got.Config), Config: got.Config, First: true}
	}
}

func (b Board) rereading() tea.Cmd {
	held := b.configured
	if held.config.File == "" {
		return nil
	}
	return func() tea.Msg {
		stamp, moved := held.stamp.Moved()
		if !moved {
			return Reread{Stamp: stamp}
		}
		got, err := b.work.Ask(engine.ConfigShow, engine.Input{})
		if err != nil || got.Config == nil {
			return Reread{Stamp: stamp}
		}
		return Reread{Stamp: stamp, Config: got.Config}
	}
}

func (b Board) reread(m Reread) Board {
	b.configured.stamp = m.Stamp
	if m.Config == nil {
		return b
	}
	now, was := *m.Config, b.configured.config
	if now.Problem != "" {
		if m.First {
			b.configured.config = now
			was = now
		}
		trouble := engine.TroubleIn(now, was)
		b.configured.trouble = &trouble
		return b
	}
	b.configured.config, b.configured.trouble, b.configured.readAt = now, nil, b.clock()
	if m.First {
		return b
	}
	b = b.Paged(engine.PagesOf(now))
	if changes := engine.ConfigChanges(was, now); len(changes) > 0 {
		b.configured.notice = readAgain(was, now, changes)
	}
	return b
}

func readAgain(was, now engine.Config, changes []engine.ConfigChange) string {
	file := filepath.Base(now.File)
	said := make([]string, 0, len(changes))
	for _, change := range changes {
		if slices.Contains(now.WrittenHere, change.Key) || slices.Contains(was.WrittenHere, change.Key) {
			file = config.SettingsFile
		}
		said = append(said, change.String())
	}
	return file + " read again: " + strings.Join(said, ", ")
}

func (b Board) settingKey(line string, words []string) (tea.Model, tea.Cmd) {
	in := engine.Input{Given: map[string]string{}}
	var rest []string
	for _, word := range words[1:] {
		if word == globalWord {
			in.Given["global"] = "true"
			continue
		}
		rest = append(rest, word)
	}
	op := engine.ConfigSet
	if words[0] == "unset" {
		op = engine.ConfigUnset
	}
	if len(rest) == 0 || (op.Verb == "unset") != (len(rest) == 1) {
		b.err = unknown(line)
		return b, nil
	}
	in.Given["key"] = rest[0]
	if op.Verb == "set" {
		value := strings.Join(rest[1:], " ")
		if value == `""` {
			value = ""
		}
		in.Given["value"] = value
	}
	return b, b.configuring(op, in)
}

func (b Board) configuring(op engine.Operation, in engine.Input) tea.Cmd {
	return func() tea.Msg {
		answer, err := b.work.Ask(op, in)
		if err != nil {
			return Configured{Err: err}
		}
		said := Configured{Text: strings.TrimSpace(engine.RenderText(op, answer))}
		if got, err := b.work.Ask(engine.ConfigShow, engine.Input{}); err == nil && got.Config != nil {
			said.Config, said.Stamp = got.Config, engine.StampOf(*got.Config)
		}
		return said
	}
}

func (b Board) configuredBy(m Configured) Board {
	if m.Err != nil {
		b.err = m.Err
		return b
	}
	b.configured.notice = m.Text
	if m.Config == nil || m.Config.Problem != "" {
		return b
	}
	b.configured.config, b.configured.stamp, b.configured.trouble, b.configured.readAt = *m.Config, m.Stamp, nil, b.clock()
	return b.Paged(engine.PagesOf(*m.Config))
}

func (b Board) keyTyped() (string, bool) {
	if b.prompt != Commanding {
		return "", false
	}
	words := strings.Fields(b.typed)
	if len(words) == 0 || (words[0] != "set" && words[0] != "unset") {
		return "", false
	}
	var rest []string
	for _, word := range words[1:] {
		if word != globalWord {
			rest = append(rest, word)
		}
	}
	ended := strings.HasSuffix(b.typed, " ")
	switch {
	case len(rest) == 0:
		return "", ended
	case len(rest) == 1 && !ended:
		return rest[0], true
	}
	return "", false
}

func (b Board) completions() []config.Key {
	typed, ok := b.keyTyped()
	if !ok {
		return nil
	}
	var found []config.Key
	for _, key := range config.Keys() {
		if strings.HasPrefix(key.Name, typed) && len(found) < completedAtMost {
			found = append(found, key)
		}
	}
	return found
}

func (b Board) completed() Board {
	found := b.completions()
	if len(found) == 0 {
		return b
	}
	typed, _ := b.keyTyped()
	b.typed = strings.TrimSuffix(b.typed, typed) + found[0].Name + " "
	return b
}

func (b Board) completionLines() []string {
	found := b.completions()
	c := b.configured.config
	lines := make([]string, 0, len(found))
	for _, key := range found {
		value := engine.ValueIn(c, key.Name)
		from := "default"
		switch {
		case slices.Contains(c.WrittenHere, key.Name):
			from = "set here ← global: " + c.Global[key.Name]
		case slices.Contains(c.Written, key.Name):
			from = "from global"
		}
		said := fmt.Sprintf("   %-20s %-14s ", key.Name, value)
		lines = append(lines, b.skin.Title.Render(Clip(said, b.width-margin))+b.skin.Faint.Render(Clip(from, max(0, b.width-margin-len([]rune(said))))))
	}
	return lines
}

func (b Board) troubleLine() string {
	t := b.configured.trouble
	where := filepath.Base(t.File)
	if t.Line > 0 {
		where = fmt.Sprintf("%s:%d", where, t.Line)
	}
	return " " + b.skin.Wrong.Render(Clip(where+" "+t.Said, b.width-margin-margin))
}

func (b Board) keptFoot() string {
	return " " + b.skin.Faint.Render(Clip("the board keeps what it read at "+b.configured.readAt.Format("15:04")+" until the file reads again", b.width-margin-margin))
}
